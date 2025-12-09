package access

import (
	"fmt"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMaxAccessCountEnforcement tests that shares with max_access_count work correctly
func TestMaxAccessCountEnforcement(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "maxaccess_user", "password123")
	noteID := createTestNote(t, userID, "Max Access Test")

	db := database.GetDB()

	// Create share with max_access_count = 2
	shareLink := models.SharedLink{
		NoteID:          noteID,
		UserID:          userID,
		ShareToken:      fmt.Sprintf("share_max_%d", time.Now().Unix()),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		MaxAccessCount:  2,
		AccessCount:     0,
		RequirePassword: false,
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create share link with max_access_count")

	t.Logf("✅ Created share with max_access_count=2")

	// Access 1: increment counter
	shareLink.AccessCount++
	db.Save(&shareLink)
	assert.Equal(t, 1, shareLink.AccessCount, "Access count should be 1")
	assert.Less(t, shareLink.AccessCount, shareLink.MaxAccessCount, "Should still be under limit")
	t.Logf("✅ Access 1: count=%d, still active", shareLink.AccessCount)

	// Access 2: increment counter (should reach limit)
	shareLink.AccessCount++
	db.Save(&shareLink)
	assert.Equal(t, 2, shareLink.AccessCount, "Access count should be 2")
	assert.GreaterOrEqual(t, shareLink.AccessCount, shareLink.MaxAccessCount, "Should reach limit")
	t.Logf("✅ Access 2: count=%d, limit reached", shareLink.AccessCount)

	// Check if share should be deleted (access_count >= max_access_count)
	var exhaustedShares []models.SharedLink
	db.Where("max_access_count > 0 AND access_count >= max_access_count").Find(&exhaustedShares)
	assert.Equal(t, 1, len(exhaustedShares), "Should find 1 exhausted share")
	assert.Equal(t, shareLink.ID, exhaustedShares[0].ID, "Should be our share")
	t.Logf("✅ Share exhausted and ready for cleanup")
}

// TestMaxAccessCountZeroUnlimited tests that max_access_count=0 means unlimited
func TestMaxAccessCountZeroUnlimited(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "unlimited_user", "password123")
	noteID := createTestNote(t, userID, "Unlimited Test")
	db := database.GetDB()

	// Create share with max_access_count = 0 (unlimited)
	shareLink := models.SharedLink{
		NoteID:          noteID,
		UserID:          userID,
		ShareToken:      fmt.Sprintf("share_unlimited_%d", time.Now().Unix()),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		MaxAccessCount:  0,
		AccessCount:     0,
		RequirePassword: false,
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create unlimited share")

	// Access multiple times
	for i := 1; i <= 5; i++ {
		shareLink.AccessCount = i
		db.Save(&shareLink)
	}

	assert.Equal(t, 5, shareLink.AccessCount, "Access count should be 5")

	// Check that it's NOT in exhausted shares
	var exhaustedShares []models.SharedLink
	db.Where("max_access_count > 0 AND access_count >= max_access_count").Find(&exhaustedShares)
	assert.Equal(t, 0, len(exhaustedShares), "Unlimited shares should not be exhausted")

	t.Logf("✅ Unlimited share accessed 5 times, still active")
}

// TestMaxAccessCountOmittedUnlimited tests that default max_access_count is 0 (unlimited)
func TestMaxAccessCountOmittedUnlimited(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "default_user", "password123")
	noteID := createTestNote(t, userID, "Default Test")
	db := database.GetDB()

	// Create share without specifying max_access_count (should default to 0)
	shareLink := models.SharedLink{
		NoteID:          noteID,
		UserID:          userID,
		ShareToken:      fmt.Sprintf("share_default_%d", time.Now().Unix()),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		// MaxAccessCount not set, should default to 0
		AccessCount:     0,
		RequirePassword: false,
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create share with default max_access_count")

	// Verify default is 0
	var retrieved models.SharedLink
	db.First(&retrieved, shareLink.ID)
	assert.Equal(t, 0, retrieved.MaxAccessCount, "Default max_access_count should be 0")

	// Access many times
	retrieved.AccessCount = 10
	db.Save(&retrieved)

	// Should NOT be in exhausted shares
	var exhaustedShares []models.SharedLink
	db.Where("max_access_count > 0 AND access_count >= max_access_count").Find(&exhaustedShares)
	assert.Equal(t, 0, len(exhaustedShares), "Default shares should be unlimited")

	t.Logf("✅ Default share (max_access_count=0) is unlimited")
}
