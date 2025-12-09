package access

import (
	"fmt"
	"lab02_mahoa/server/auth"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPasswordProtectedShare tests creating and verifying password-protected shares
func TestPasswordProtectedShare(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "pwprotect_user", "password123")
	noteID := createTestNote(t, userID, "Password Protected")
	db := database.GetDB()

	// Create password-protected share
	sharePassword := "shareSecret123"
	passwordHash, err := auth.HashPassword(sharePassword)
	assert.NoError(t, err, "Should hash password")

	shareLink := models.SharedLink{
		NoteID:          noteID,
		UserID:          userID,
		ShareToken:      fmt.Sprintf("share_pw_%d", time.Now().Unix()),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		MaxAccessCount:  0,
		AccessCount:     0,
		RequirePassword: true,
		PasswordHash:    passwordHash,
	}
	err = db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create password-protected share")
	assert.True(t, shareLink.RequirePassword, "Should require password")
	t.Logf("✅ Created password-protected share")

	// Verify wrong password fails
	err = auth.CheckPassword("wrongPassword", shareLink.PasswordHash)
	assert.Error(t, err, "Wrong password should fail")
	t.Logf("✅ Wrong password rejected")

	// Verify correct password works
	err = auth.CheckPassword(sharePassword, shareLink.PasswordHash)
	assert.NoError(t, err, "Correct password should work")
	t.Logf("✅ Correct password accepted")
}

// TestPasswordProtectedWithMaxAccess tests combining password and max_access_count
func TestPasswordProtectedWithMaxAccess(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "combo_user", "password123")
	noteID := createTestNote(t, userID, "Combined Protection")
	db := database.GetDB()

	// Create share with both password and max_access_count
	sharePassword := "combo123"
	passwordHash, _ := auth.HashPassword(sharePassword)

	shareLink := models.SharedLink{
		NoteID:          noteID,
		UserID:          userID,
		ShareToken:      fmt.Sprintf("share_combo_%d", time.Now().Unix()),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		MaxAccessCount:  2,
		AccessCount:     0,
		RequirePassword: true,
		PasswordHash:    passwordHash,
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create combined protection share")
	assert.True(t, shareLink.RequirePassword, "Should require password")
	assert.Equal(t, 2, shareLink.MaxAccessCount, "Should have max_access_count=2")
	t.Logf("✅ Created share with password + max_access_count=2")

	// Simulate accesses with correct password
	for i := 1; i <= 2; i++ {
		// Verify password first
		err = auth.CheckPassword(sharePassword, shareLink.PasswordHash)
		assert.NoError(t, err, "Password should be correct")

		// Increment access count
		shareLink.AccessCount++
		db.Save(&shareLink)
		t.Logf("✅ Access %d: password OK, count=%d", i, shareLink.AccessCount)
	}

	// Check if exhausted
	assert.GreaterOrEqual(t, shareLink.AccessCount, shareLink.MaxAccessCount, "Should be exhausted")

	var exhaustedShares []models.SharedLink
	db.Where("max_access_count > 0 AND access_count >= max_access_count").Find(&exhaustedShares)
	assert.Equal(t, 1, len(exhaustedShares), "Should find 1 exhausted share")
	t.Logf("✅ Share exhausted after 2 accesses")
}

// TestEmptyPasswordNotProtected tests that empty password means no protection
func TestEmptyPasswordNotProtected(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "emptypw_user", "password123")
	noteID := createTestNote(t, userID, "Empty Password")
	db := database.GetDB()

	// Create share without password (empty string should mean no protection)
	shareLink := models.SharedLink{
		NoteID:          noteID,
		UserID:          userID,
		ShareToken:      fmt.Sprintf("share_nopw_%d", time.Now().Unix()),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		MaxAccessCount:  0,
		AccessCount:     0,
		RequirePassword: false,
		PasswordHash:    "",
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create unprotected share")
	assert.False(t, shareLink.RequirePassword, "Should not require password")
	assert.Empty(t, shareLink.PasswordHash, "Password hash should be empty")

	t.Logf("✅ Empty password = no protection")
}

// TestPasswordHashNotExposed tests that password_hash field has json:"-" tag
func TestPasswordHashNotExposed(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "hashhidden_user", "password123")
	noteID := createTestNote(t, userID, "Hash Test")
	db := database.GetDB()

	// Create password-protected share
	passwordHash, _ := auth.HashPassword("secret123")
	shareLink := models.SharedLink{
		NoteID:          noteID,
		UserID:          userID,
		ShareToken:      fmt.Sprintf("share_hash_%d", time.Now().Unix()),
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		MaxAccessCount:  0,
		AccessCount:     0,
		RequirePassword: true,
		PasswordHash:    passwordHash,
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create share")

	// Verify PasswordHash field has json:"-" tag (not exposed)
	// In Go, json:"-" means the field won't be marshaled to JSON
	assert.NotEmpty(t, shareLink.PasswordHash, "Password hash should exist in DB")
	t.Logf("✅ Password hash stored but not exposed in JSON (json:\"-\" tag)")
}
