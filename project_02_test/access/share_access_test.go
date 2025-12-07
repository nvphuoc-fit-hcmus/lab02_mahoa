package access

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lab02_mahoa/server/auth"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/handlers"
	"lab02_mahoa/server/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupTestDB initializes a test database
func setupTestDB(t *testing.T) {
	// Initialize in-memory test database
	err := database.InitTestDB(&models.User{}, &models.Note{}, &models.SharedLink{})
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
}

// teardownTestDB cleans up the test database
func teardownTestDB(t *testing.T) {
	db := database.GetDB()
	if db != nil {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// createTestUser creates a user for testing
func createTestUser(t *testing.T, username, password string) uint {
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

// createTestNote creates a note for testing
func createTestNote(t *testing.T, userID uint, title string) uint {
	db := database.GetDB()

	note := models.Note{
		UserID:           userID,
		Title:            title,
		EncryptedContent: "encrypted_content_test",
		EncryptedKey:     "encrypted_key_test",
		IV:               "test_iv",
	}

	if err := db.Create(&note).Error; err != nil {
		t.Fatalf("Failed to create test note: %v", err)
	}

	return note.ID
}

// generateTestToken generates a JWT token for testing
func generateTestToken(userID uint, username string) string {
	token, err := auth.GenerateJWT(userID, username)
	if err != nil {
		return ""
	}
	return token
}

// TestAccessActiveShareLink tests accessing an active (non-expired) share link
func TestAccessActiveShareLink(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Test Note")

	// Create an active share link (expires in 1 hour)
	db := database.GetDB()
	shareLink := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: fmt.Sprintf("share_%d_%d", noteID, time.Now().Unix()),
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create share link successfully")

	// Verify the share link exists and is not expired
	var retrievedLink models.SharedLink
	err = db.Where("share_token = ? AND expires_at > ?", shareLink.ShareToken, time.Now()).First(&retrievedLink).Error
	assert.NoError(t, err, "Active share link should be accessible")
	assert.Equal(t, shareLink.ShareToken, retrievedLink.ShareToken)
}

// TestAccessExpiredShareLink tests that expired share links cannot be accessed
func TestAccessExpiredShareLink(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Test Note")

	// Create an expired share link (expired 1 hour ago)
	db := database.GetDB()
	shareLink := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: fmt.Sprintf("share_%d_%d", noteID, time.Now().Unix()),
		ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create expired share link")

	// Try to access the expired share link
	var retrievedLink models.SharedLink
	err = db.Where("share_token = ? AND expires_at > ?", shareLink.ShareToken, time.Now()).First(&retrievedLink).Error
	assert.Error(t, err, "Expired share link should not be accessible")
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Should return record not found error")
}

// TestMultipleExpiredShareLinks tests handling multiple expired links
func TestMultipleExpiredShareLinks(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Test Note")

	db := database.GetDB()

	// Create multiple share links with different expiration times
	shareLinks := []models.SharedLink{
		{
			NoteID:     noteID,
			UserID:     userID,
			ShareToken: "share_expired_1",
			ExpiresAt:  time.Now().Add(-2 * time.Hour), // Expired 2 hours ago
		},
		{
			NoteID:     noteID,
			UserID:     userID,
			ShareToken: "share_expired_2",
			ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		},
		{
			NoteID:     noteID,
			UserID:     userID,
			ShareToken: "share_active",
			ExpiresAt:  time.Now().Add(1 * time.Hour), // Active (expires in 1 hour)
		},
	}

	for _, link := range shareLinks {
		err := db.Create(&link).Error
		assert.NoError(t, err, "Should create share link")
	}

	// Query only active links
	var activeLinks []models.SharedLink
	err := db.Where("note_id = ? AND expires_at > ?", noteID, time.Now()).Find(&activeLinks).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, len(activeLinks), "Should only find 1 active link")
	assert.Equal(t, "share_active", activeLinks[0].ShareToken)
}

// TestListNotesWithExpiredShares tests that notes with only expired shares show as not shared
func TestListNotesWithExpiredShares(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and notes
	userID := createTestUser(t, "testuser", "password123")
	noteID1 := createTestNote(t, userID, "Note with expired share")
	noteID2 := createTestNote(t, userID, "Note with active share")

	db := database.GetDB()

	// Create expired share for note 1
	expiredShare := models.SharedLink{
		NoteID:     noteID1,
		UserID:     userID,
		ShareToken: "share_expired",
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}
	db.Create(&expiredShare)

	// Create active share for note 2
	activeShare := models.SharedLink{
		NoteID:     noteID2,
		UserID:     userID,
		ShareToken: "share_active",
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}
	db.Create(&activeShare)

	// Test ListNotesHandler to verify IsShared status
	token := generateTestToken(userID, "testuser")
	req, _ := http.NewRequest("GET", "/api/notes", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.ListNotesHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Should return 200 OK")

	var response struct {
		Notes []models.NoteResponse `json:"notes"`
		Count int                   `json:"count"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "Should parse response")
	assert.Equal(t, 2, response.Count, "Should have 2 notes")

	// Find the notes in the response
	for _, note := range response.Notes {
		if note.ID == noteID1 {
			assert.False(t, note.IsShared, "Note with expired share should show as not shared")
		} else if note.ID == noteID2 {
			assert.True(t, note.IsShared, "Note with active share should show as shared")
		}
	}
}

// TestRevokeExpiredShare tests revoking already expired shares
func TestRevokeExpiredShare(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Test Note")

	db := database.GetDB()

	// Create expired share link
	expiredShare := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: "share_expired",
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}
	db.Create(&expiredShare)

	// Revoke all shares for the note
	token := generateTestToken(userID, "testuser")
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/notes/%d/revoke", noteID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.RevokeShareHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Should revoke successfully")

	// Verify the share link is deleted
	var count int64
	db.Model(&models.SharedLink{}).Where("note_id = ?", noteID).Count(&count)
	assert.Equal(t, int64(0), count, "All shares should be deleted")
}

// TestCreateShareWithCustomExpiration tests creating shares with different expiration times
func TestCreateShareWithCustomExpiration(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Test Note")

	token := generateTestToken(userID, "testuser")

	// Test creating share with 1 hour expiration
	reqBody := models.CreateShareRequest{
		DurationHours: 1,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/notes/%d/share", noteID), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.CreateShareHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Should create share successfully")

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)
	
	assert.True(t, response["success"].(bool))
	assert.NotEmpty(t, response["share_token"])

	// Verify expiration time is approximately 1 hour from now
	expiresAt, _ := time.Parse(time.RFC3339, response["expires_at"].(string))
	expectedExpiration := time.Now().Add(1 * time.Hour)
	timeDiff := expiresAt.Sub(expectedExpiration).Abs()
	assert.Less(t, timeDiff, 1*time.Minute, "Expiration should be approximately 1 hour from now")
}

// TestShareLinkExpirationBoundary tests share link at exact expiration moment
func TestShareLinkExpirationBoundary(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Test Note")

	db := database.GetDB()

	// Create share link that expires exactly now
	shareLink := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: "share_boundary",
		ExpiresAt:  time.Now(),
	}
	db.Create(&shareLink)

	// Wait a tiny bit to ensure time has passed
	time.Sleep(10 * time.Millisecond)

	// Try to access the share link
	var retrievedLink models.SharedLink
	err := db.Where("share_token = ? AND expires_at > ?", shareLink.ShareToken, time.Now()).First(&retrievedLink).Error
	assert.Error(t, err, "Share link at exact expiration should not be accessible")
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

// TestCleanupExpiredShares tests bulk cleanup of expired shares
func TestCleanupExpiredShares(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and notes
	userID := createTestUser(t, "testuser", "password123")
	noteID1 := createTestNote(t, userID, "Note 1")
	noteID2 := createTestNote(t, userID, "Note 2")
	noteID3 := createTestNote(t, userID, "Note 3")

	db := database.GetDB()

	// Create multiple expired and active shares
	shares := []models.SharedLink{
		{NoteID: noteID1, UserID: userID, ShareToken: "share_1_expired", ExpiresAt: time.Now().Add(-2 * time.Hour)},
		{NoteID: noteID1, UserID: userID, ShareToken: "share_1_active", ExpiresAt: time.Now().Add(2 * time.Hour)},
		{NoteID: noteID2, UserID: userID, ShareToken: "share_2_expired", ExpiresAt: time.Now().Add(-1 * time.Hour)},
		{NoteID: noteID3, UserID: userID, ShareToken: "share_3_active", ExpiresAt: time.Now().Add(3 * time.Hour)},
	}

	for _, share := range shares {
		db.Create(&share)
	}

	// Cleanup expired shares
	result := db.Where("expires_at <= ?", time.Now()).Delete(&models.SharedLink{})
	assert.NoError(t, result.Error, "Should cleanup expired shares")
	assert.Equal(t, int64(2), result.RowsAffected, "Should delete 2 expired shares")

	// Verify only active shares remain
	var remainingShares []models.SharedLink
	db.Find(&remainingShares)
	assert.Equal(t, 2, len(remainingShares), "Should have 2 active shares remaining")
}

// TestUnauthorizedAccessToExpiredShare tests that unauthorized users cannot access expired shares
func TestUnauthorizedAccessToExpiredShare(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create two users
	userID1 := createTestUser(t, "user1", "password123")
	userID2 := createTestUser(t, "user2", "password456")
	
	// User1 creates a note
	noteID := createTestNote(t, userID1, "User1's Note")

	db := database.GetDB()

	// Create expired share for user1's note
	expiredShare := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID1,
		ShareToken: "share_expired",
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}
	db.Create(&expiredShare)

	// User2 tries to access the note directly
	token := generateTestToken(userID2, "user2")
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/notes/%d", noteID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.GetNoteHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "User2 should not access User1's note")
}

// TestShareLinkTokenUniqueness tests that share tokens are unique
func TestShareLinkTokenUniqueness(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Test Note")

	db := database.GetDB()

	// Create first share link
	share1 := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: "unique_token_123",
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}
	err := db.Create(&share1).Error
	assert.NoError(t, err, "Should create first share link")

	// Try to create duplicate token
	share2 := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: "unique_token_123", // Same token
		ExpiresAt:  time.Now().Add(2 * time.Hour),
	}
	err = db.Create(&share2).Error
	assert.Error(t, err, "Should fail to create duplicate share token")
}
