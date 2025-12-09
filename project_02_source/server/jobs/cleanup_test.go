package jobs

import (
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/models"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.User{}, &models.Note{}, &models.SharedLink{}, &models.E2EEShare{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestCleanupExpiredSharedLinks(t *testing.T) {
	db := setupTestDB(t)

	// Create a test user and note
	user := models.User{Username: "testuser", PasswordHash: "hash"}
	db.Create(&user)

	note := models.Note{
		UserID:           user.ID,
		Title:            "Test Note",
		EncryptedContent: "encrypted",
		IV:               "iv",
		EncryptedKey:     "key",
		EncryptedKeyIV:   "keyiv",
	}
	db.Create(&note)

	// Create expired shared link
	expiredLink := models.SharedLink{
		NoteID:     note.ID,
		UserID:     user.ID,
		ShareToken: "expired-token",
		ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	db.Create(&expiredLink)

	// Create valid shared link
	validLink := models.SharedLink{
		NoteID:     note.ID,
		UserID:     user.ID,
		ShareToken: "valid-token",
		ExpiresAt:  time.Now().Add(1 * time.Hour), // Expires in 1 hour
	}
	db.Create(&validLink)

	// Run cleanup
	cleanupExpiredData(db)

	// Verify expired link was deleted
	var links []models.SharedLink
	db.Find(&links)

	if len(links) != 1 {
		t.Errorf("Expected 1 link remaining, got %d", len(links))
	}

	if links[0].ShareToken != "valid-token" {
		t.Errorf("Expected valid-token to remain, got %s", links[0].ShareToken)
	}

	t.Log("✅ Expired shared links cleaned up successfully")
}

func TestCleanupExpiredE2EEShares(t *testing.T) {
	db := setupTestDB(t)

	// Create test users and note
	sender := models.User{Username: "sender", PasswordHash: "hash"}
	db.Create(&sender)

	recipient := models.User{Username: "recipient", PasswordHash: "hash"}
	db.Create(&recipient)

	note := models.Note{
		UserID:           sender.ID,
		Title:            "Test Note",
		EncryptedContent: "encrypted",
		IV:               "iv",
		EncryptedKey:     "key",
		EncryptedKeyIV:   "keyiv",
	}
	db.Create(&note)

	// Create expired E2EE share
	expiredShare := models.E2EEShare{
		NoteID:           note.ID,
		SenderID:         sender.ID,
		RecipientID:      recipient.ID,
		SenderPublicKey:  "pubkey",
		EncryptedContent: "encrypted",
		ContentIV:        "iv",
		ExpiresAt:        time.Now().Add(-1 * time.Hour),
	}
	db.Create(&expiredShare)

	// Create valid E2EE share
	validShare := models.E2EEShare{
		NoteID:           note.ID,
		SenderID:         sender.ID,
		RecipientID:      recipient.ID,
		SenderPublicKey:  "pubkey2",
		EncryptedContent: "encrypted2",
		ContentIV:        "iv2",
		ExpiresAt:        time.Now().Add(1 * time.Hour),
	}
	db.Create(&validShare)

	// Run cleanup
	cleanupExpiredData(db)

	// Verify expired share was deleted
	var shares []models.E2EEShare
	db.Find(&shares)

	if len(shares) != 1 {
		t.Errorf("Expected 1 share remaining, got %d", len(shares))
	}

	if shares[0].SenderPublicKey != "pubkey2" {
		t.Errorf("Expected pubkey2 to remain, got %s", shares[0].SenderPublicKey)
	}

	t.Log("✅ Expired E2EE shares cleaned up successfully")
}

// Note: Max access count feature can be added later by extending SharedLink model
// with MaxAccessCount and AccessCount fields

func TestCleanupJobIntegration(t *testing.T) {
	// Test that cleanup job can be started without errors
	db := setupTestDB(t)

	// Set database for the jobs package
	database.DB = db

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Cleanup job panicked: %v", r)
		}
	}()

	// Start cleanup job (will run in background)
	go StartCleanupJob(db)

	// Wait a bit to ensure job started
	time.Sleep(100 * time.Millisecond)

	t.Log("✅ Cleanup job started successfully")
}
