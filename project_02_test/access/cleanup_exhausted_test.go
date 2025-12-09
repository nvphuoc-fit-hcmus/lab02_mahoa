package access

import (
	"fmt"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/jobs"
	"lab02_mahoa/server/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestCleanupJobExhaustedShares tests that cleanup job deletes exhausted shares
func TestCleanupJobExhaustedShares(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "cleanup_user", "password123")
	noteID := createTestNote(t, userID, "Cleanup Test")
	db := database.GetDB()

	// Create exhausted share (access_count >= max_access_count)
	shareLink := models.SharedLink{
		NoteID:          noteID,
		UserID:          userID,
		ShareToken:      fmt.Sprintf("share_exhausted_%d", time.Now().Unix()),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		MaxAccessCount:  1,
		AccessCount:     1, // Already exhausted
		RequirePassword: false,
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create share")
	t.Logf("✅ Created exhausted share (count=1, max=1)")

	// Verify it exists
	var beforeCleanup models.SharedLink
	err = db.First(&beforeCleanup, shareLink.ID).Error
	assert.NoError(t, err, "Share should exist before cleanup")

	// Trigger cleanup job
	t.Logf("🧹 Triggering cleanup job...")
	jobs.CleanupExpiredDataNow()
	time.Sleep(500 * time.Millisecond)

	// Verify it was deleted
	var afterCleanup models.SharedLink
	err = db.First(&afterCleanup, shareLink.ID).Error
	assert.Error(t, err, "Share should be deleted after cleanup")
	assert.Equal(t, gorm.ErrRecordNotFound, err, "Should return record not found")

	t.Logf("✅ Cleanup job deleted exhausted share")
}

// TestCleanupJobPreservesActiveShares tests that cleanup doesn't delete active shares
func TestCleanupJobPreservesActiveShares(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "preserve_user", "password123")
	noteID := createTestNote(t, userID, "Preserve Test")
	db := database.GetDB()

	// Create active share (not exhausted yet)
	shareLink := models.SharedLink{
		NoteID:          noteID,
		UserID:          userID,
		ShareToken:      fmt.Sprintf("share_active_%d", time.Now().Unix()),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		MaxAccessCount:  5,
		AccessCount:     2, // Not exhausted yet (2 < 5)
		RequirePassword: false,
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create share")
	t.Logf("✅ Created active share (count=2, max=5)")

	// Trigger cleanup
	t.Logf("🧹 Triggering cleanup job...")
	jobs.CleanupExpiredDataNow()
	time.Sleep(500 * time.Millisecond)

	// Verify share still exists
	var stillExists models.SharedLink
	err = db.First(&stillExists, shareLink.ID).Error
	assert.NoError(t, err, "Active share should not be deleted")
	assert.Equal(t, 2, stillExists.AccessCount, "Access count should be preserved")

	t.Logf("✅ Cleanup job preserved active share")
}

// TestCleanupJobUnlimitedShares tests that cleanup doesn't delete unlimited shares (max_access_count=0)
func TestCleanupJobUnlimitedShares(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "unlimited_cleanup_user", "password123")
	noteID := createTestNote(t, userID, "Unlimited Cleanup")
	db := database.GetDB()

	// Create unlimited share (max_access_count=0)
	shareLink := models.SharedLink{
		NoteID:          noteID,
		UserID:          userID,
		ShareToken:      fmt.Sprintf("share_unlimited_%d", time.Now().Unix()),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		MaxAccessCount:  0,           // Unlimited
		AccessCount:     10,          // Accessed many times
		RequirePassword: false,
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create share")
	t.Logf("✅ Created unlimited share (count=10, max=0)")

	// Trigger cleanup
	t.Logf("🧹 Triggering cleanup job...")
	jobs.CleanupExpiredDataNow()
	time.Sleep(500 * time.Millisecond)

	// Verify share still exists (unlimited should never be deleted by count)
	var stillExists models.SharedLink
	err = db.First(&stillExists, shareLink.ID).Error
	assert.NoError(t, err, "Unlimited share should not be deleted")
	assert.Equal(t, 0, stillExists.MaxAccessCount, "Max access count should be 0")
	assert.Equal(t, 10, stillExists.AccessCount, "Access count should be preserved")

	t.Logf("✅ Cleanup job preserved unlimited share")
}
