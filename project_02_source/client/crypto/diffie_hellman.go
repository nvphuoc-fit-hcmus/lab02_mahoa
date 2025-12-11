package crypto

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// DHKeyPair represents a Diffie-Hellman key pair
type DHKeyPair struct {
	PrivateKey *ecdh.PrivateKey
	PublicKey  *ecdh.PublicKey
}

// GenerateDHKeyPair generates a new ECDH key pair using X25519 curve
// X25519 is modern, fast, and secure alternative to traditional DH
func GenerateDHKeyPair() (*DHKeyPair, error) {
	// Use X25519 curve (Curve25519 for ECDH)
	curve := ecdh.X25519()
	
	privateKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	return &DHKeyPair{
		PrivateKey: privateKey,
		PublicKey:  privateKey.PublicKey(),
	}, nil
}

// ComputeSharedSecret computes the shared secret from our private key and their public key
// This implements the core Diffie-Hellman key exchange algorithm
func ComputeSharedSecret(ourPrivateKey *ecdh.PrivateKey, theirPublicKey *ecdh.PublicKey) ([]byte, error) {
	// Perform ECDH operation
	sharedSecret, err := ourPrivateKey.ECDH(theirPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to compute shared secret: %w", err)
	}

	// Derive a 32-byte key from the shared secret using SHA-256
	// This ensures we have a proper AES-256 key
	hash := sha256.Sum256(sharedSecret)
	return hash[:], nil
}

// PublicKeyToBase64 converts a public key to base64 string for transmission
func PublicKeyToBase64(pubKey *ecdh.PublicKey) string {
	return base64.StdEncoding.EncodeToString(pubKey.Bytes())
}

// PublicKeyFromBase64 reconstructs a public key from base64 string
func PublicKeyFromBase64(pubKeyBase64 string) (*ecdh.PublicKey, error) {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	curve := ecdh.X25519()
	pubKey, err := curve.NewPublicKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return pubKey, nil
}

// EncryptWithSharedSecret encrypts data using the DH shared secret as the key
func EncryptWithSharedSecret(plaintext string, sharedSecret []byte) (ciphertext, iv string, err error) {
	return EncryptAES(plaintext, sharedSecret)
}

// DecryptWithSharedSecret decrypts data using the DH shared secret as the key
func DecryptWithSharedSecret(ciphertext, iv string, sharedSecret []byte) (string, error) {
	return DecryptAES(ciphertext, iv, sharedSecret)
}
