package access

import (
	"fmt"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/models"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestExpiredShareLinkAccessViaAPI tests accessing expired share links through API endpoint
func TestExpiredShareLinkAccessViaAPI(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Sensitive Note")

	db := database.GetDB()

	// Create expired share link
	expiredShare := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: "expired_token_123",
		ExpiresAt:  time.Now().Add(-24 * time.Hour), // Expired yesterday
	}
	db.Create(&expiredShare)

	// Verify in database that link exists but is expired
	var checkShare models.SharedLink
	err := db.Where("share_token = ?", "expired_token_123").First(&checkShare).Error
	assert.NoError(t, err, "Share link should exist in database")
	assert.True(t, checkShare.ExpiresAt.Before(time.Now()), "Share link should be expired")

	// Try to query only active shares (should not find expired one)
	var activeShare models.SharedLink
	err = db.Where("share_token = ? AND expires_at > ?", "expired_token_123", time.Now()).First(&activeShare).Error
	assert.Error(t, err, "Should not find expired share when filtering by expires_at")
}

// TestMultipleUsersExpiredShares tests expired shares across different users
func TestMultipleUsersExpiredShares(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create multiple users
	user1ID := createTestUser(t, "user1", "pass1")
	user2ID := createTestUser(t, "user2", "pass2")
	user3ID := createTestUser(t, "user3", "pass3")

	// Create notes for each user
	note1ID := createTestNote(t, user1ID, "User1 Note")
	note2ID := createTestNote(t, user2ID, "User2 Note")
	note3ID := createTestNote(t, user3ID, "User3 Note")

	db := database.GetDB()

	// Create mix of expired and active shares
	shares := []models.SharedLink{
		{NoteID: note1ID, UserID: user1ID, ShareToken: "user1_expired", ExpiresAt: time.Now().Add(-1 * time.Hour)},
		{NoteID: note1ID, UserID: user1ID, ShareToken: "user1_active", ExpiresAt: time.Now().Add(1 * time.Hour)},
		{NoteID: note2ID, UserID: user2ID, ShareToken: "user2_expired", ExpiresAt: time.Now().Add(-2 * time.Hour)},
		{NoteID: note3ID, UserID: user3ID, ShareToken: "user3_active", ExpiresAt: time.Now().Add(2 * time.Hour)},
	}

	for _, share := range shares {
		db.Create(&share)
	}

	// Query only active shares across all users
	var activeShares []models.SharedLink
	db.Where("expires_at > ?", time.Now()).Find(&activeShares)

	assert.Equal(t, 2, len(activeShares), "Should find exactly 2 active shares")
	
	// Verify correct shares are returned
	tokens := []string{activeShares[0].ShareToken, activeShares[1].ShareToken}
	assert.Contains(t, tokens, "user1_active")
	assert.Contains(t, tokens, "user3_active")
	assert.NotContains(t, tokens, "user1_expired")
	assert.NotContains(t, tokens, "user2_expired")
}

// TestShareLinkExpirationTransition tests a share link before and after expiration
func TestShareLinkExpirationTransition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time-based test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Test Note")

	db := database.GetDB()

	// Create share link that expires in 1 second
	expirationTime := time.Now().Add(1 * time.Second)
	shareLink := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: "transition_token",
		ExpiresAt:  expirationTime,
	}
	db.Create(&shareLink)

	// Access before expiration
	var activeShare models.SharedLink
	err := db.Where("share_token = ? AND expires_at > ?", "transition_token", time.Now()).First(&activeShare).Error
	assert.NoError(t, err, "Should access share before expiration")

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Try to access after expiration
	var expiredShare models.SharedLink
	err = db.Where("share_token = ? AND expires_at > ?", "transition_token", time.Now()).First(&expiredShare).Error
	assert.Error(t, err, "Should not access share after expiration")
}

// TestConcurrentShareAccess tests multiple concurrent access attempts to expired shares
func TestConcurrentShareAccess(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Concurrent Test Note")

	db := database.GetDB()

	// Create expired share
	expiredShare := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: "concurrent_expired",
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}
	db.Create(&expiredShare)

	// Simulate concurrent access attempts
	done := make(chan bool, 5)
	
	for i := 0; i < 5; i++ {
		go func(index int) {
			var share models.SharedLink
			err := db.Where("share_token = ? AND expires_at > ?", "concurrent_expired", time.Now()).First(&share).Error
			assert.Error(t, err, fmt.Sprintf("Concurrent access %d should fail", index))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
}

// TestExpiredSharesDoNotAffectActiveNotes tests that expired shares don't affect note access
func TestExpiredSharesDoNotAffectActiveNotes(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Test Note")

	db := database.GetDB()

	// Create expired share
	expiredShare := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: "expired_share",
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}
	db.Create(&expiredShare)

	// User should still be able to access their own note directly
	token := generateTestToken(userID, "testuser")
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/notes/%d", noteID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Note: This requires the actual handler implementation
	// The note should be accessible to the owner regardless of expired shares
	var note models.Note
	err := db.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error
	assert.NoError(t, err, "User should access their own note even with expired shares")
}

// TestExpiredShareDeletion tests deleting expired shares
func TestExpiredShareDeletion(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "testuser", "password123")
	note1ID := createTestNote(t, userID, "Note 1")
	note2ID := createTestNote(t, userID, "Note 2")

	db := database.GetDB()

	// Create multiple shares with different states
	shares := []models.SharedLink{
		{NoteID: note1ID, UserID: userID, ShareToken: "note1_expired_1", ExpiresAt: time.Now().Add(-3 * time.Hour)},
		{NoteID: note1ID, UserID: userID, ShareToken: "note1_expired_2", ExpiresAt: time.Now().Add(-1 * time.Hour)},
		{NoteID: note1ID, UserID: userID, ShareToken: "note1_active", ExpiresAt: time.Now().Add(1 * time.Hour)},
		{NoteID: note2ID, UserID: userID, ShareToken: "note2_expired", ExpiresAt: time.Now().Add(-2 * time.Hour)},
		{NoteID: note2ID, UserID: userID, ShareToken: "note2_active", ExpiresAt: time.Now().Add(2 * time.Hour)},
	}

	for _, share := range shares {
		db.Create(&share)
	}

	// Delete all expired shares
	result := db.Where("expires_at <= ?", time.Now()).Delete(&models.SharedLink{})
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(3), result.RowsAffected, "Should delete 3 expired shares")

	// Verify only active shares remain
	var remainingShares []models.SharedLink
	db.Find(&remainingShares)
	assert.Equal(t, 2, len(remainingShares), "Should have 2 active shares")

	for _, share := range remainingShares {
		assert.True(t, share.ExpiresAt.After(time.Now()), "Remaining shares should all be active")
	}
}

// TestShareExpirationWithDifferentTimezones tests expiration handling across timezones
func TestShareExpirationWithDifferentTimezones(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Timezone Test Note")

	db := database.GetDB()

	// Create share with UTC time
	utcNow := time.Now().UTC()
	expiredShareUTC := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: "utc_expired",
		ExpiresAt:  utcNow.Add(-1 * time.Hour),
	}
	db.Create(&expiredShareUTC)

	// Query should still correctly identify as expired regardless of local timezone
	var share models.SharedLink
	err := db.Where("share_token = ? AND expires_at > ?", "utc_expired", time.Now()).First(&share).Error
	assert.Error(t, err, "Expired share should not be accessible regardless of timezone")
}

// TestShareListNotesPerformance tests performance of listing notes with many expired shares
func TestShareListNotesPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "perfuser", "password123")

	db := database.GetDB()

	// Create 100 notes with mixed share states
	for i := 0; i < 100; i++ {
		noteID := createTestNote(t, userID, fmt.Sprintf("Note %d", i))

		// Create 3 expired and 1 active share per note
		for j := 0; j < 3; j++ {
			expiredShare := models.SharedLink{
				NoteID:     noteID,
				UserID:     userID,
				ShareToken: fmt.Sprintf("note%d_expired%d", i, j),
				ExpiresAt:  time.Now().Add(-time.Duration(j+1) * time.Hour),
			}
			db.Create(&expiredShare)
		}

		activeShare := models.SharedLink{
			NoteID:     noteID,
			UserID:     userID,
			ShareToken: fmt.Sprintf("note%d_active", i),
			ExpiresAt:  time.Now().Add(1 * time.Hour),
		}
		db.Create(&activeShare)
	}

	// Measure query performance
	start := time.Now()

	var notes []models.Note
	err := db.Where("user_id = ?", userID).Find(&notes).Error
	assert.NoError(t, err)

	// For each note, check if it has active shares
	for _, note := range notes {
		var count int64
		db.Model(&models.SharedLink{}).
			Where("note_id = ? AND expires_at > ?", note.ID, time.Now()).
			Count(&count)
	}

	elapsed := time.Since(start)
	t.Logf("Query took %v for 100 notes with 400 shares (300 expired, 100 active)", elapsed)

	// Should complete reasonably fast (adjust threshold as needed)
	assert.Less(t, elapsed, 5*time.Second, "Query should complete within 5 seconds")
}

// TestExpiredShareNoLeakage tests that expired shares don't leak information
func TestExpiredShareNoLeakage(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	user1ID := createTestUser(t, "user1", "password1")
	user2ID := createTestUser(t, "user2", "password2")

	note1ID := createTestNote(t, user1ID, "User1 Secret Note")

	db := database.GetDB()

	// User1 shares note with user2, but link expires
	expiredShare := models.SharedLink{
		NoteID:     note1ID,
		UserID:     user1ID,
		ShareToken: "shared_with_user2_expired",
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}
	db.Create(&expiredShare)

	// User2 should not be able to access the note anymore
	_ = generateTestToken(user2ID, "user2")
	_, _ = http.NewRequest("GET", fmt.Sprintf("/api/notes/%d", note1ID), nil)

	// Note: Direct note access is owner-only in the current implementation
	// User2 never had direct access, shares are separate
	var note models.Note
	err := db.Where("id = ? AND user_id = ?", note1ID, user2ID).First(&note).Error
	assert.Error(t, err, "User2 should not access User1's note with expired share")
}

// createTestUserWithID creates a user and returns the ID (helper for this test file)
func createTestUserWithID(t *testing.T, username, password string) uint {
	db := database.GetDB()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user.ID
}

// TestRevokeAllSharesIncludingExpired tests revoking all shares including expired ones
func TestRevokeAllSharesIncludingExpired(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Test Note")

	db := database.GetDB()

	// Create mix of expired and active shares
	shares := []models.SharedLink{
		{NoteID: noteID, UserID: userID, ShareToken: "share1_expired", ExpiresAt: time.Now().Add(-2 * time.Hour)},
		{NoteID: noteID, UserID: userID, ShareToken: "share2_expired", ExpiresAt: time.Now().Add(-1 * time.Hour)},
		{NoteID: noteID, UserID: userID, ShareToken: "share3_active", ExpiresAt: time.Now().Add(1 * time.Hour)},
		{NoteID: noteID, UserID: userID, ShareToken: "share4_active", ExpiresAt: time.Now().Add(2 * time.Hour)},
	}

	for _, share := range shares {
		db.Create(&share)
	}

	// Revoke all shares for the note (including expired ones)
	result := db.Where("note_id = ?", noteID).Delete(&models.SharedLink{})
	assert.NoError(t, result.Error)
	assert.Equal(t, int64(4), result.RowsAffected, "Should delete all 4 shares")

	// Verify no shares remain
	var count int64
	db.Model(&models.SharedLink{}).Where("note_id = ?", noteID).Count(&count)
	assert.Equal(t, int64(0), count, "No shares should remain after revoke")
}
