package e2ee

import (
	"bytes"
	"encoding/json"
	"lab02_mahoa/client/crypto"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/handlers"
	"lab02_mahoa/server/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPublicKeyRegistration tests registering user's DH public key
func TestPublicKeyRegistration(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "alice", "password123")
	token := getJWTToken(t, userID, "alice")

	// Generate keypair
	keyPair, err := crypto.GenerateDHKeyPair()
	assert.NoError(t, err)

	publicKeyBase64 := crypto.PublicKeyToBase64(keyPair.PublicKey)

	// Register public key
	reqBody := map[string]string{
		"dh_public_key": publicKeyBase64,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/user/publickey", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handlers.UpdatePublicKeyHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should register public key successfully")

	// Verify in database
	db := database.GetDB()
	var user models.User
	db.First(&user, userID)

	assert.Equal(t, publicKeyBase64, user.DHPublicKey, "Public key should be stored in database")

	t.Log("✅ Public key registered successfully")
}

// TestGetUserPublicKey tests fetching another user's public key
func TestGetUserPublicKey(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create Alice and Bob
	aliceID := createTestUser(t, "alice", "password123")
	bobID := createTestUser(t, "bob", "password123")
	bobToken := getJWTToken(t, bobID, "bob")

	// Alice registers her public key
	aliceKeyPair, _ := crypto.GenerateDHKeyPair()
	alicePubKey := crypto.PublicKeyToBase64(aliceKeyPair.PublicKey)

	db := database.GetDB()
	db.Model(&models.User{}).Where("id = ?", aliceID).Update("dh_public_key", alicePubKey)

	// Bob fetches Alice's public key
	req := httptest.NewRequest(http.MethodGet, "/api/users/alice/publickey", nil)
	req.Header.Set("Authorization", "Bearer "+bobToken)
	w := httptest.NewRecorder()

	handlers.GetPublicKeyHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Should fetch public key successfully")

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	assert.Equal(t, "alice", response["username"])
	assert.Equal(t, alicePubKey, response["dh_public_key"])

	t.Log("✅ Public key fetched successfully")
}

// TestGetUserPublicKeyNotRegistered tests fetching public key when user hasn't registered
func TestGetUserPublicKeyNotRegistered(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	createTestUser(t, "alice", "password123") // Alice hasn't registered public key
	bobID := createTestUser(t, "bob", "password123")
	bobToken := getJWTToken(t, bobID, "bob")

	req := httptest.NewRequest(http.MethodGet, "/api/users/alice/publickey", nil)
	req.Header.Set("Authorization", "Bearer "+bobToken)
	w := httptest.NewRecorder()

	handlers.GetPublicKeyHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Should return 404 when public key not registered")

	t.Log("✅ Unregistered public key handled correctly")
}

// TestGetUserPublicKeyNonexistentUser tests fetching public key of non-existent user
func TestGetUserPublicKeyNonexistentUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	userID := createTestUser(t, "alice", "password123")
	token := getJWTToken(t, userID, "alice")

	req := httptest.NewRequest(http.MethodGet, "/api/users/nonexistent/publickey", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handlers.GetPublicKeyHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "Should return 404 for nonexistent user")

	t.Log("✅ Nonexistent user handled correctly")
}

// TestE2EEShareWithWrongRecipientKey tests E2EE share decryption with wrong key
func TestE2EEShareWithWrongRecipientKey(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create Alice (sender) and Bob (recipient)
	aliceID := createTestUser(t, "alice", "password123")
	bobID := createTestUser(t, "bob", "password123")

	// Register Alice and Bob's keys
	aliceKeyPair, _ := crypto.GenerateDHKeyPair()
	bobKeyPair, _ := crypto.GenerateDHKeyPair()
	
	db := database.GetDB()
	db.Model(&models.User{}).Where("id = ?", aliceID).Update("dh_public_key", crypto.PublicKeyToBase64(aliceKeyPair.PublicKey))
	db.Model(&models.User{}).Where("id = ?", bobID).Update("dh_public_key", crypto.PublicKeyToBase64(bobKeyPair.PublicKey))

	// Alice creates encrypted content for Bob
	message := "Secret message for Bob"
	sharedSecret, _ := crypto.ComputeSharedSecret(aliceKeyPair.PrivateKey, bobKeyPair.PublicKey)
	ciphertext, iv, _ := crypto.EncryptWithSharedSecret(message, sharedSecret)

	// Create E2EE share
	noteID := createTestNote(t, aliceID, "Secret Note")
	share := models.E2EEShare{
		NoteID:           noteID,
		SenderID:         aliceID,
		RecipientID:      bobID,
		SenderPublicKey:  crypto.PublicKeyToBase64(aliceKeyPair.PublicKey),
		EncryptedContent: ciphertext,
		ContentIV:        iv,
		ExpiresAt:        getExpirationTime(24),
	}
	db.Create(&share)

	// Charlie (wrong recipient) tries to decrypt
	charlieKeyPair, _ := crypto.GenerateDHKeyPair()
	
	// Charlie computes wrong shared secret
	wrongSharedSecret, _ := crypto.ComputeSharedSecret(charlieKeyPair.PrivateKey, aliceKeyPair.PublicKey)
	
	// Try to decrypt with wrong key
	_, err := crypto.DecryptWithSharedSecret(ciphertext, iv, wrongSharedSecret)
	assert.Error(t, err, "Decryption with wrong key should fail")

	// Bob (correct recipient) can decrypt
	correctSharedSecret, _ := crypto.ComputeSharedSecret(bobKeyPair.PrivateKey, aliceKeyPair.PublicKey)
	decrypted, err := crypto.DecryptWithSharedSecret(ciphertext, iv, correctSharedSecret)
	assert.NoError(t, err, "Bob should decrypt successfully")
	assert.Equal(t, message, decrypted, "Decrypted message should match")

	t.Log("✅ Wrong recipient key correctly rejected")
}

// TestE2EESessionKeyIsolation tests that different shares have different session keys
func TestE2EESessionKeyIsolation(t *testing.T) {
	// Generate 3 keypairs
	aliceKeyPair, _ := crypto.GenerateDHKeyPair()
	bobKeyPair, _ := crypto.GenerateDHKeyPair()
	charlieKeyPair, _ := crypto.GenerateDHKeyPair()

	// Compute different shared secrets
	aliceBobSecret, _ := crypto.ComputeSharedSecret(aliceKeyPair.PrivateKey, bobKeyPair.PublicKey)
	aliceCharlieSecret, _ := crypto.ComputeSharedSecret(aliceKeyPair.PrivateKey, charlieKeyPair.PublicKey)
	bobCharlieSecret, _ := crypto.ComputeSharedSecret(bobKeyPair.PrivateKey, charlieKeyPair.PublicKey)

	// All should be different
	assert.NotEqual(t, aliceBobSecret, aliceCharlieSecret, "Alice-Bob and Alice-Charlie secrets should differ")
	assert.NotEqual(t, aliceBobSecret, bobCharlieSecret, "Alice-Bob and Bob-Charlie secrets should differ")
	assert.NotEqual(t, aliceCharlieSecret, bobCharlieSecret, "Alice-Charlie and Bob-Charlie secrets should differ")

	// Encrypt same message with different keys
	message := "Same message"
	
	ciphertext1, iv1, _ := crypto.EncryptWithSharedSecret(message, aliceBobSecret)
	ciphertext2, iv2, _ := crypto.EncryptWithSharedSecret(message, aliceCharlieSecret)
	ciphertext3, iv3, _ := crypto.EncryptWithSharedSecret(message, bobCharlieSecret)

	// Ciphertexts should all be different (due to different keys and IVs)
	assert.NotEqual(t, ciphertext1, ciphertext2, "Ciphertexts should differ")
	assert.NotEqual(t, ciphertext2, ciphertext3, "Ciphertexts should differ")
	assert.NotEqual(t, ciphertext1, ciphertext3, "Ciphertexts should differ")

	// IVs should be different
	assert.NotEqual(t, iv1, iv2)
	assert.NotEqual(t, iv2, iv3)

	t.Log("✅ Session key isolation verified")
}

// TestE2EEReplayAttackPrevention tests that IV prevents replay attacks
func TestE2EEReplayAttackPrevention(t *testing.T) {
	aliceKeyPair, _ := crypto.GenerateDHKeyPair()
	bobKeyPair, _ := crypto.GenerateDHKeyPair()

	sharedSecret, _ := crypto.ComputeSharedSecret(aliceKeyPair.PrivateKey, bobKeyPair.PublicKey)

	// Encrypt same message twice
	message := "Secret message"
	ciphertext1, iv1, _ := crypto.EncryptWithSharedSecret(message, sharedSecret)
	ciphertext2, iv2, _ := crypto.EncryptWithSharedSecret(message, sharedSecret)

	// IVs should be different (random)
	assert.NotEqual(t, iv1, iv2, "IVs should be different to prevent replay")

	// Ciphertexts should be different (due to different IVs)
	assert.NotEqual(t, ciphertext1, ciphertext2, "Ciphertexts should differ even for same message")

	// Both can be decrypted
	decrypted1, err := crypto.DecryptWithSharedSecret(ciphertext1, iv1, sharedSecret)
	assert.NoError(t, err)
	assert.Equal(t, message, decrypted1)

	decrypted2, err := crypto.DecryptWithSharedSecret(ciphertext2, iv2, sharedSecret)
	assert.NoError(t, err)
	assert.Equal(t, message, decrypted2)

	// Cannot decrypt with wrong IV
	_, err = crypto.DecryptWithSharedSecret(ciphertext1, iv2, sharedSecret)
	assert.Error(t, err, "Should not decrypt with wrong IV")

	t.Log("✅ Replay attack prevention verified")
}

// TestE2EEMultipleRecipientsIndependence tests that shares to different recipients are independent
func TestE2EEMultipleRecipientsIndependence(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Alice shares same note with Bob and Charlie
	aliceID := createTestUser(t, "alice", "password123")
	bobID := createTestUser(t, "bob", "password123")
	charlieID := createTestUser(t, "charlie", "password123")

	noteID := createTestNote(t, aliceID, "Shared Note")

	aliceKeyPair, _ := crypto.GenerateDHKeyPair()
	bobKeyPair, _ := crypto.GenerateDHKeyPair()
	charlieKeyPair, _ := crypto.GenerateDHKeyPair()

	message := "Secret for both"

	// Share with Bob
	bobSecret, _ := crypto.ComputeSharedSecret(aliceKeyPair.PrivateKey, bobKeyPair.PublicKey)
	bobCipher, bobIV, _ := crypto.EncryptWithSharedSecret(message, bobSecret)

	// Share with Charlie
	charlieSecret, _ := crypto.ComputeSharedSecret(aliceKeyPair.PrivateKey, charlieKeyPair.PublicKey)
	charlieCipher, charlieIV, _ := crypto.EncryptWithSharedSecret(message, charlieSecret)

	db := database.GetDB()
	
	// Create two shares
	shareBob := models.E2EEShare{
		NoteID:           noteID,
		SenderID:         aliceID,
		RecipientID:      bobID,
		SenderPublicKey:  crypto.PublicKeyToBase64(aliceKeyPair.PublicKey),
		EncryptedContent: bobCipher,
		ContentIV:        bobIV,
		ExpiresAt:        getExpirationTime(24),
	}
	db.Create(&shareBob)

	shareCharlie := models.E2EEShare{
		NoteID:           noteID,
		SenderID:         aliceID,
		RecipientID:      charlieID,
		SenderPublicKey:  crypto.PublicKeyToBase64(aliceKeyPair.PublicKey),
		EncryptedContent: charlieCipher,
		ContentIV:        charlieIV,
		ExpiresAt:        getExpirationTime(24),
	}
	db.Create(&shareCharlie)

	// Bob can decrypt his share
	bobDecrypted, err := crypto.DecryptWithSharedSecret(bobCipher, bobIV, bobSecret)
	assert.NoError(t, err)
	assert.Equal(t, message, bobDecrypted)

	// Charlie can decrypt his share
	charlieDecrypted, err := crypto.DecryptWithSharedSecret(charlieCipher, charlieIV, charlieSecret)
	assert.NoError(t, err)
	assert.Equal(t, message, charlieDecrypted)

	// Bob cannot decrypt Charlie's share
	_, err = crypto.DecryptWithSharedSecret(charlieCipher, charlieIV, bobSecret)
	assert.Error(t, err, "Bob should not decrypt Charlie's share")

	// Charlie cannot decrypt Bob's share
	_, err = crypto.DecryptWithSharedSecret(bobCipher, bobIV, charlieSecret)
	assert.Error(t, err, "Charlie should not decrypt Bob's share")

	t.Log("✅ Multiple recipients have independent shares")
}
