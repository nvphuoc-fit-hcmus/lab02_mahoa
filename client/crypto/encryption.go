package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"golang.org/x/crypto/pbkdf2"
)

// GenerateKey generates a random AES-256 key
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// EncryptAES encrypts plaintext using AES-256-GCM
func EncryptAES(plaintext string, key []byte) (ciphertext string, iv string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", err
	}

	ciphertextBytes := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertextBytes),
		base64.StdEncoding.EncodeToString(nonce),
		nil
}

// DecryptAES decrypts ciphertext using AES-256-GCM
func DecryptAES(ciphertext string, ivStr string, key []byte) (string, error) {
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	nonce, err := base64.StdEncoding.DecodeString(ivStr)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(nonce) != gcm.NonceSize() {
		return "", fmt.Errorf("invalid nonce size")
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func DeriveKeyFromPassword(password, salt string) []byte {
	// Sử dụng PBKDF2 với SHA-256, lặp 4096 lần để chống Brute-force
	return pbkdf2.Key([]byte(password), []byte(salt), 4096, 32, sha256.New)
}

func WrapKey(fileKey []byte, masterKey []byte) (string, error) {
	// 1. Chuyển FileKey (bytes) sang String (Base64) để dùng được hàm EncryptAES cũ
	fileKeyStr := base64.StdEncoding.EncodeToString(fileKey)

	// 2. Tận dụng hàm EncryptAES đã có
	ciphertext, iv, err := EncryptAES(fileKeyStr, masterKey)
	if err != nil {
		return "", err
	}

	// 3. Gộp IV và Ciphertext ngăn cách bởi dấu ":" để lưu vào 1 cột DB
	return iv + ":" + ciphertext, nil
}

// UnwrapKey dùng để giải mã lấy lại FileKey gốc từ chuỗi "IV:Ciphertext"
func UnwrapKey(wrappedKey string, masterKey []byte) ([]byte, error) {
	// 1. Tách chuỗi để lấy IV và Ciphertext
	parts := strings.Split(wrappedKey, ":")
	if len(parts) != 2 {
		return nil, errors.New("invalid wrapped key format")
	}
	iv := parts[0]
	ciphertext := parts[1]

	// 2. Tận dụng hàm DecryptAES đã có
	decryptedStr, err := DecryptAES(ciphertext, iv, masterKey)
	if err != nil {
		return nil, err
	}

	// 3. Chuyển String (Base64) ngược lại thành Bytes để dùng làm Key giải mã file
	return base64.StdEncoding.DecodeString(decryptedStr)
}