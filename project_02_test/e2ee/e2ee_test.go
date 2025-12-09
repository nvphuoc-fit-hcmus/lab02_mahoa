package e2ee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lab02_mahoa/client/crypto"
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
)

// setupTestDB initializes a test database
func setupTestDB(t *testing.T) {
	err := database.InitTestDB(&models.User{}, &models.Note{}, &models.SharedLink{}, &models.E2EEShare{})
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
		t.Fatalf("Failed to create user: %v", err)
	}

	return user.ID
}

// createTestNote creates a note for testing
func createTestNote(t *testing.T, userID uint, title string) uint {
	db := database.GetDB()

	note := models.Note{
		UserID:           userID,
		Title:            title,
		EncryptedContent: "encrypted_content",
		IV:               "test_iv",
		EncryptedKey:     "encrypted_key",
		EncryptedKeyIV:   "encrypted_key_iv",
	}

	if err := db.Create(&note).Error; err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	return note.ID
}

// getJWTToken generates a JWT token for testing
func getJWTToken(t *testing.T, userID uint, username string) string {
	token, err := auth.GenerateJWT(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}
	return token
}

// getExpirationTime returns expiration time for testing
func getExpirationTime(hours int) time.Time {
	return time.Now().Add(time.Duration(hours) * time.Hour)
}

func TestDiffieHellmanKeyExchange(t *testing.T) {
	// Test basic DH key exchange
	aliceKeyPair, err := crypto.GenerateDHKeyPair()
	assert.NoError(t, err, "Alice should generate key pair")

	bobKeyPair, err := crypto.GenerateDHKeyPair()
	assert.NoError(t, err, "Bob should generate key pair")

	// Compute shared secrets
	aliceShared, err := crypto.ComputeSharedSecret(aliceKeyPair.PrivateKey, bobKeyPair.PublicKey)
	assert.NoError(t, err, "Alice should compute shared secret")

	bobShared, err := crypto.ComputeSharedSecret(bobKeyPair.PrivateKey, aliceKeyPair.PublicKey)
	assert.NoError(t, err, "Bob should compute shared secret")

	// Verify shared secrets match
	assert.Equal(t, aliceShared, bobShared, "Shared secrets should match")
}

func TestE2EEEncryptionDecryption(t *testing.T) {
	// Generate DH key pairs
	aliceKeyPair, _ := crypto.GenerateDHKeyPair()
	bobKeyPair, _ := crypto.GenerateDHKeyPair()

	// Compute shared secret
	sharedSecret, _ := crypto.ComputeSharedSecret(aliceKeyPair.PrivateKey, bobKeyPair.PublicKey)

	// Encrypt message
	message := "Secret E2EE message 🔐"
	ciphertext, iv, err := crypto.EncryptWithSharedSecret(message, sharedSecret)
	assert.NoError(t, err, "Encryption should succeed")
	assert.NotEmpty(t, ciphertext, "Ciphertext should not be empty")
	assert.NotEmpty(t, iv, "IV should not be empty")

	// Decrypt message
	decrypted, err := crypto.DecryptWithSharedSecret(ciphertext, iv, sharedSecret)
	assert.NoError(t, err, "Decryption should succeed")
	assert.Equal(t, message, decrypted, "Decrypted message should match original")

	// Test with wrong key
	wrongKeyPair, _ := crypto.GenerateDHKeyPair()
	wrongShared, _ := crypto.ComputeSharedSecret(wrongKeyPair.PrivateKey, bobKeyPair.PublicKey)
	_, err = crypto.DecryptWithSharedSecret(ciphertext, iv, wrongShared)
	assert.Error(t, err, "Decryption with wrong key should fail")
}

func TestCreateE2EEShare(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create sender and recipient
	senderID := createTestUser(t, "alice", "password123")
	_ = createTestUser(t, "bob", "password123") // Create bob but don't need ID here

	// Create note
	noteID := createTestNote(t, senderID, "Test Note")

	// Generate DH key pair for sender
	senderKeyPair, _ := crypto.GenerateDHKeyPair()
	senderPubKey := crypto.PublicKeyToBase64(senderKeyPair.PublicKey)

	// Encrypt content with a mock shared secret
	mockSecret := []byte("mock32bytesecretkeymock32bytesec")
	encryptedContent, contentIV, _ := crypto.EncryptWithSharedSecret("Secret content", mockSecret)

	// Create request
	reqBody := models.CreateE2EEShareRequest{
		RecipientUsername: "bob",
		SenderPublicKey:   senderPubKey,
		EncryptedContent:  encryptedContent,
		ContentIV:         contentIV,
		DurationHours:     24,
	}

	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/notes/%d/e2ee", noteID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+getJWTToken(t, senderID, "alice"))

	w := httptest.NewRecorder()
	handlers.CreateE2EEShareHandler(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Should create E2EE share successfully")

	var response models.E2EEShareResponse
	json.NewDecoder(w.Body).Decode(&response)

	assert.True(t, response.Success, "Response should indicate success")
	assert.Equal(t, "bob", response.RecipientUsername, "Recipient should be bob")
	assert.NotZero(t, response.ShareID, "Share ID should be set")
}

func TestListE2EEShares(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create users
	senderID := createTestUser(t, "alice", "password123")
	recipientID := createTestUser(t, "bob", "password123")

	// Create note
	noteID := createTestNote(t, senderID, "Test Note")

	// Create E2EE share directly in database
	db := database.GetDB()
	share := models.E2EEShare{
		NoteID:           noteID,
		SenderID:         senderID,
		RecipientID:      recipientID,
		SenderPublicKey:  "mock_public_key",
		EncryptedContent: "encrypted_content",
		ContentIV:        "content_iv",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}
	db.Create(&share)

	// List shares as Bob (recipient)
	req := httptest.NewRequest("GET", "/api/e2ee", nil)
	req.Header.Set("Authorization", "Bearer "+getJWTToken(t, recipientID, "bob"))

	w := httptest.NewRecorder()
	handlers.ListE2EESharesHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should list shares successfully")

	var response models.ListE2EESharesResponse
	json.NewDecoder(w.Body).Decode(&response)

	assert.Equal(t, 1, response.Count, "Should have 1 share")
	assert.Equal(t, "Test Note", response.Shares[0].NoteTitle, "Note title should match")
	assert.Equal(t, "alice", response.Shares[0].SenderUsername, "Sender should be alice")
}

func TestE2EEShareExpiration(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create users
	senderID := createTestUser(t, "alice", "password123")
	recipientID := createTestUser(t, "bob", "password123")

	// Create note
	noteID := createTestNote(t, senderID, "Test Note")

	// Create expired E2EE share
	db := database.GetDB()
	expiredShare := models.E2EEShare{
		NoteID:           noteID,
		SenderID:         senderID,
		RecipientID:      recipientID,
		SenderPublicKey:  "mock_public_key",
		EncryptedContent: "encrypted_content",
		ContentIV:        "content_iv",
		ExpiresAt:        time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	db.Create(&expiredShare)

	// Try to access expired share
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/e2ee/%d", expiredShare.ID), nil)
	req.Header.Set("Authorization", "Bearer "+getJWTToken(t, recipientID, "bob"))

	w := httptest.NewRecorder()
	handlers.GetE2EEShareHandler(w, req)

	assert.Equal(t, http.StatusGone, w.Code, "Should return 410 Gone for expired share")
}

func TestE2EEShareAccessControl(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create users
	senderID := createTestUser(t, "alice", "password123")
	recipientID := createTestUser(t, "bob", "password123")
	eveID := createTestUser(t, "eve", "password123") // Unauthorized user

	// Create note
	noteID := createTestNote(t, senderID, "Test Note")

	// Create E2EE share
	db := database.GetDB()
	share := models.E2EEShare{
		NoteID:           noteID,
		SenderID:         senderID,
		RecipientID:      recipientID,
		SenderPublicKey:  "mock_public_key",
		EncryptedContent: "encrypted_content",
		ContentIV:        "content_iv",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}
	db.Create(&share)

	// Try to access as Eve (unauthorized)
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/e2ee/%d", share.ID), nil)
	req.Header.Set("Authorization", "Bearer "+getJWTToken(t, eveID, "eve"))

	w := httptest.NewRecorder()
	handlers.GetE2EEShareHandler(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code, "Should deny access to unauthorized user")
}

func TestDeleteE2EEShare(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create users
	senderID := createTestUser(t, "alice", "password123")
	recipientID := createTestUser(t, "bob", "password123")

	// Create note
	noteID := createTestNote(t, senderID, "Test Note")

	// Create E2EE share
	db := database.GetDB()
	share := models.E2EEShare{
		NoteID:           noteID,
		SenderID:         senderID,
		RecipientID:      recipientID,
		SenderPublicKey:  "mock_public_key",
		EncryptedContent: "encrypted_content",
		ContentIV:        "content_iv",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	}
	db.Create(&share)

	// Delete share as sender
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/e2ee/%d", share.ID), nil)
	req.Header.Set("Authorization", "Bearer "+getJWTToken(t, senderID, "alice"))

	w := httptest.NewRecorder()
	handlers.DeleteE2EEShareHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should delete share successfully")

	// Verify share is deleted
	var count int64
	db.Model(&models.E2EEShare{}).Where("id = ?", share.ID).Count(&count)
	assert.Equal(t, int64(0), count, "Share should be deleted from database")
}
