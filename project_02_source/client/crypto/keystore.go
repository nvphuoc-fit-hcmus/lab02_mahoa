package crypto

import (
	"crypto/ecdh"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

// GetKeystorePath returns the path to the keystore file for a user
func GetKeystorePath(username string) string {
	homeDir, _ := os.UserHomeDir()
	keystoreDir := filepath.Join(homeDir, ".lab02_mahoa", "keys")
	os.MkdirAll(keystoreDir, 0700) // Create directory if not exists
	return filepath.Join(keystoreDir, username+".key")
}

// SaveDHKeyPair saves a DH keypair to encrypted file
func SaveDHKeyPair(username, password string, privateKey *ecdh.PrivateKey) error {
	// Serialize private key
	privateKeyBytes := privateKey.Bytes()
	
	// Derive encryption key from password
	kek := DeriveKeyFromPassword(password, []byte(username))
	
	// Encrypt private key
	encryptedKey, iv, err := EncryptAES(base64.StdEncoding.EncodeToString(privateKeyBytes), kek)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %w", err)
	}
	
	// Format: iv:encryptedKey (both base64)
	data := iv + ":" + encryptedKey
	
	// Save to file
	keystorePath := GetKeystorePath(username)
	if err := os.WriteFile(keystorePath, []byte(data), 0600); err != nil {
		return fmt.Errorf("failed to write keystore: %w", err)
	}
	
	return nil
}

// LoadDHKeyPair loads a DH keypair from encrypted file
func LoadDHKeyPair(username, password string) (*ecdh.PrivateKey, error) {
	keystorePath := GetKeystorePath(username)
	
	// Read file
	data, err := os.ReadFile(keystorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No keystore found, return nil
		}
		return nil, fmt.Errorf("failed to read keystore: %w", err)
	}
	
	// Parse format: iv:encryptedKey
	parts := []byte(data)
	var iv, encryptedKey string
	for i, b := range parts {
		if b == ':' {
			iv = string(parts[:i])
			encryptedKey = string(parts[i+1:])
			break
		}
	}
	
	if iv == "" || encryptedKey == "" {
		return nil, fmt.Errorf("invalid keystore format")
	}
	
	// Derive decryption key from password
	kek := DeriveKeyFromPassword(password, []byte(username))
	
	// Decrypt private key
	privateKeyBase64, err := DecryptAES(encryptedKey, iv, kek)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key (wrong password?): %w", err)
	}
	
	// Decode base64
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	
	// Reconstruct private key
	curve := ecdh.X25519()
	privateKey, err := curve.NewPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct private key: %w", err)
	}
	
	return privateKey, nil
}

// DeleteDHKeyPair deletes a user's keystore file
func DeleteDHKeyPair(username string) error {
	keystorePath := GetKeystorePath(username)
	if err := os.Remove(keystorePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
