package crypto_test

import (
	"lab02_mahoa/client/crypto"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestKeystoreSaveAndLoad tests saving and loading keypair from encrypted file
func TestKeystoreSaveAndLoad(t *testing.T) {
	username := "testuser"
	password := "password123"

	// Clean up before test
	defer crypto.DeleteDHKeyPair(username)

	// Generate keypair
	keyPair, err := crypto.GenerateDHKeyPair()
	assert.NoError(t, err, "Should generate keypair")

	originalPubKey := crypto.PublicKeyToBase64(keyPair.PublicKey)

	// Save to keystore
	err = crypto.SaveDHKeyPair(username, password, keyPair.PrivateKey)
	assert.NoError(t, err, "Should save keypair")

	// Verify file exists
	keystorePath := crypto.GetKeystorePath(username)
	_, err = os.Stat(keystorePath)
	assert.NoError(t, err, "Keystore file should exist")

	// Load from keystore
	loadedPrivKey, err := crypto.LoadDHKeyPair(username, password)
	assert.NoError(t, err, "Should load keypair")
	assert.NotNil(t, loadedPrivKey, "Loaded private key should not be nil")

	// Verify public keys match
	loadedPubKey := crypto.PublicKeyToBase64(loadedPrivKey.PublicKey())
	assert.Equal(t, originalPubKey, loadedPubKey, "Public keys should match")

	t.Log("✅ Keypair saved and loaded successfully")
}

// TestKeystoreWrongPassword tests loading with incorrect password
func TestKeystoreWrongPassword(t *testing.T) {
	username := "testuser2"
	password := "correctpassword"
	wrongPassword := "wrongpassword"

	defer crypto.DeleteDHKeyPair(username)

	// Generate and save keypair
	keyPair, _ := crypto.GenerateDHKeyPair()
	err := crypto.SaveDHKeyPair(username, password, keyPair.PrivateKey)
	assert.NoError(t, err)

	// Try to load with wrong password
	_, err = crypto.LoadDHKeyPair(username, wrongPassword)
	assert.Error(t, err, "Loading with wrong password should fail")
	assert.Contains(t, err.Error(), "failed to decrypt", "Error should mention decryption failure")

	t.Log("✅ Wrong password correctly rejected")
}

// TestKeystoreNonexistent tests loading non-existent keystore
func TestKeystoreNonexistent(t *testing.T) {
	username := "nonexistentuser"
	password := "anypassword"

	// Ensure keystore doesn't exist
	crypto.DeleteDHKeyPair(username)

	// Try to load
	privKey, err := crypto.LoadDHKeyPair(username, password)
	assert.NoError(t, err, "Loading nonexistent keystore should return nil without error")
	assert.Nil(t, privKey, "Private key should be nil for nonexistent keystore")

	t.Log("✅ Nonexistent keystore handled correctly")
}

// TestKeystoreDelete tests deleting keystore file
func TestKeystoreDelete(t *testing.T) {
	username := "deletetestuser"
	password := "password"

	// Create keystore
	keyPair, _ := crypto.GenerateDHKeyPair()
	crypto.SaveDHKeyPair(username, password, keyPair.PrivateKey)

	// Verify exists
	keystorePath := crypto.GetKeystorePath(username)
	_, err := os.Stat(keystorePath)
	assert.NoError(t, err, "Keystore should exist")

	// Delete
	err = crypto.DeleteDHKeyPair(username)
	assert.NoError(t, err, "Should delete keystore")

	// Verify deleted
	_, err = os.Stat(keystorePath)
	assert.Error(t, err, "Keystore should not exist after deletion")
	assert.True(t, os.IsNotExist(err), "Error should be 'not exist'")

	t.Log("✅ Keystore deleted successfully")
}

// TestKeystorePersistence tests that keypair remains same across sessions
func TestKeystorePersistence(t *testing.T) {
	username := "persistuser"
	password := "password"

	defer crypto.DeleteDHKeyPair(username)

	// Session 1: Generate and save
	keyPair1, _ := crypto.GenerateDHKeyPair()
	crypto.SaveDHKeyPair(username, password, keyPair1.PrivateKey)
	pubKey1 := crypto.PublicKeyToBase64(keyPair1.PublicKey)

	// Session 2: Load existing
	privKey2, err := crypto.LoadDHKeyPair(username, password)
	assert.NoError(t, err)
	pubKey2 := crypto.PublicKeyToBase64(privKey2.PublicKey())

	// Session 3: Load again
	privKey3, err := crypto.LoadDHKeyPair(username, password)
	assert.NoError(t, err)
	pubKey3 := crypto.PublicKeyToBase64(privKey3.PublicKey())

	// All should match
	assert.Equal(t, pubKey1, pubKey2, "Public key should persist in session 2")
	assert.Equal(t, pubKey1, pubKey3, "Public key should persist in session 3")

	t.Log("✅ Keypair persistence verified across multiple loads")
}

// TestKeystoreMultipleUsers tests keystore isolation between users
func TestKeystoreMultipleUsers(t *testing.T) {
	users := []struct {
		username string
		password string
	}{
		{"alice", "alicepass"},
		{"bob", "bobpass"},
		{"charlie", "charliepass"},
	}

	defer func() {
		for _, u := range users {
			crypto.DeleteDHKeyPair(u.username)
		}
	}()

	// Create keypairs for each user
	pubKeys := make(map[string]string)
	for _, u := range users {
		keyPair, _ := crypto.GenerateDHKeyPair()
		crypto.SaveDHKeyPair(u.username, u.password, keyPair.PrivateKey)
		pubKeys[u.username] = crypto.PublicKeyToBase64(keyPair.PublicKey)
	}

	// Verify each user has different keys
	assert.NotEqual(t, pubKeys["alice"], pubKeys["bob"], "Alice and Bob should have different keys")
	assert.NotEqual(t, pubKeys["bob"], pubKeys["charlie"], "Bob and Charlie should have different keys")
	assert.NotEqual(t, pubKeys["alice"], pubKeys["charlie"], "Alice and Charlie should have different keys")

	// Verify each user can load their own key
	for _, u := range users {
		privKey, err := crypto.LoadDHKeyPair(u.username, u.password)
		assert.NoError(t, err, "User %s should load their key", u.username)
		
		loadedPubKey := crypto.PublicKeyToBase64(privKey.PublicKey())
		assert.Equal(t, pubKeys[u.username], loadedPubKey, "User %s public key should match", u.username)
	}

	t.Log("✅ Multiple users have isolated keystores")
}

// TestKeystorePasswordChange tests changing user password
func TestKeystorePasswordChange(t *testing.T) {
	username := "changepassuser"
	oldPassword := "oldpass123"
	newPassword := "newpass456"

	defer crypto.DeleteDHKeyPair(username)

	// Create with old password
	keyPair, _ := crypto.GenerateDHKeyPair()
	crypto.SaveDHKeyPair(username, oldPassword, keyPair.PrivateKey)
	originalPubKey := crypto.PublicKeyToBase64(keyPair.PublicKey)

	// Load with old password and save with new password
	privKey, err := crypto.LoadDHKeyPair(username, oldPassword)
	assert.NoError(t, err, "Should load with old password")
	
	err = crypto.SaveDHKeyPair(username, newPassword, privKey)
	assert.NoError(t, err, "Should save with new password")

	// Old password should not work
	_, err = crypto.LoadDHKeyPair(username, oldPassword)
	assert.Error(t, err, "Old password should not work")

	// New password should work
	newPrivKey, err := crypto.LoadDHKeyPair(username, newPassword)
	assert.NoError(t, err, "New password should work")

	// Public key should remain the same
	newPubKey := crypto.PublicKeyToBase64(newPrivKey.PublicKey())
	assert.Equal(t, originalPubKey, newPubKey, "Public key should not change")

	t.Log("✅ Password change handled correctly")
}

// TestKeystoreSpecialCharactersInPassword tests passwords with special characters
func TestKeystoreSpecialCharactersInPassword(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"Symbols", "user1", "p@ssw0rd!#$%"},
		{"Unicode", "user2", "密碼パスワード"},
		{"Spaces", "user3", "pass word 123"},
		{"Quotes", "user4", `pa"ss'wo"rd`},
		{"Newline", "user5", "pass\nword"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer crypto.DeleteDHKeyPair(tc.username)

			// Generate and save
			keyPair, _ := crypto.GenerateDHKeyPair()
			err := crypto.SaveDHKeyPair(tc.username, tc.password, keyPair.PrivateKey)
			assert.NoError(t, err, "Should save with special password")

			// Load
			_, err = crypto.LoadDHKeyPair(tc.username, tc.password)
			assert.NoError(t, err, "Should load with special password")

			t.Logf("✅ Password with %s handled correctly", tc.name)
		})
	}
}
