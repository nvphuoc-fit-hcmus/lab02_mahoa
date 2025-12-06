package auth_test

import (
	"lab02_mahoa/server/auth"
	"strings"
	"testing"
)

// TestHashPassword kiểm tra hàm hash password
func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := auth.HashPassword(password)

	// Kiểm tra không có lỗi
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	// Kiểm tra hash không rỗng
	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Kiểm tra hash khác với password gốc
	if hash == password {
		t.Error("Hash should not equal original password")
	}

	// Kiểm tra hash có format bcrypt (bắt đầu với $2a$ hoặc $2b$)
	if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
		t.Errorf("Hash should have bcrypt format, got: %s", hash[:10])
	}
}

// TestHashPasswordDifferentHashes kiểm tra cùng password tạo ra hash khác nhau
func TestHashPasswordDifferentHashes(t *testing.T) {
	password := "samePassword123"

	// Hash cùng password 2 lần
	hash1, err1 := auth.HashPassword(password)
	hash2, err2 := auth.HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("HashPassword returned errors: %v, %v", err1, err2)
	}

	// Kiểm tra 2 hash khác nhau (do bcrypt sử dụng random salt)
	if hash1 == hash2 {
		t.Error("Same password should produce different hashes due to random salt")
	}
}

// TestHashPasswordEmptyString kiểm tra hash empty string
func TestHashPasswordEmptyString(t *testing.T) {
	hash, err := auth.HashPassword("")

	// Bcrypt có thể hash empty string
	if err != nil {
		t.Errorf("HashPassword should handle empty string, got error: %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash even for empty password")
	}
}

// TestHashPasswordLongPassword kiểm tra hash password dài
func TestHashPasswordLongPassword(t *testing.T) {
	// Tạo password rất dài (72+ characters)
	// Bcrypt giới hạn 72 bytes, nên password dài hơn sẽ bị reject
	longPassword := strings.Repeat("a", 100)

	hash, err := auth.HashPassword(longPassword)

	// Bcrypt sẽ trả về error với password > 72 bytes
	if err == nil {
		t.Error("Expected error for password longer than 72 bytes")
	}

	if hash != "" {
		t.Error("Expected empty hash for password that exceeds bcrypt limit")
	}
}

// TestCheckPasswordSuccess kiểm tra verify password đúng
func TestCheckPasswordSuccess(t *testing.T) {
	password := "correctPassword123"

	// Hash password
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify password
	err = auth.CheckPassword(password, hash)

	if err != nil {
		t.Errorf("CheckPassword should succeed for correct password, got error: %v", err)
	}
}

// TestCheckPasswordWrong kiểm tra verify password sai
func TestCheckPasswordWrong(t *testing.T) {
	correctPassword := "correctPassword123"
	wrongPassword := "wrongPassword456"

	// Hash correct password
	hash, err := auth.HashPassword(correctPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify với wrong password
	err = auth.CheckPassword(wrongPassword, hash)

	if err == nil {
		t.Error("CheckPassword should fail for wrong password")
	}
}

// TestCheckPasswordCaseSensitive kiểm tra password có phân biệt chữ hoa/thường
func TestCheckPasswordCaseSensitive(t *testing.T) {
	password := "MyPassword123"
	wrongCasePassword := "mypassword123"

	// Hash password
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify với case khác
	err = auth.CheckPassword(wrongCasePassword, hash)

	if err == nil {
		t.Error("CheckPassword should be case-sensitive")
	}
}

// TestCheckPasswordEmptyHash kiểm tra verify với hash rỗng
func TestCheckPasswordEmptyHash(t *testing.T) {
	password := "somePassword123"

	err := auth.CheckPassword(password, "")

	if err == nil {
		t.Error("CheckPassword should fail for empty hash")
	}
}

// TestCheckPasswordInvalidHash kiểm tra verify với hash không hợp lệ
func TestCheckPasswordInvalidHash(t *testing.T) {
	password := "somePassword123"
	invalidHash := "not-a-valid-bcrypt-hash"

	err := auth.CheckPassword(password, invalidHash)

	if err == nil {
		t.Error("CheckPassword should fail for invalid hash format")
	}
}

// TestCheckPasswordEmptyPassword kiểm tra verify với password rỗng
func TestCheckPasswordEmptyPassword(t *testing.T) {
	originalPassword := "actualPassword123"

	// Hash original password
	hash, err := auth.HashPassword(originalPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify với empty password
	err = auth.CheckPassword("", hash)

	if err == nil {
		t.Error("CheckPassword should fail for empty password when hash is for non-empty password")
	}
}

// TestPasswordHashConsistency kiểm tra tính nhất quán của hash/verify
func TestPasswordHashConsistency(t *testing.T) {
	testCases := []string{
		"simple",
		"complex!@#$%^&*()",
		"with spaces in it",
		"unicode 密碼 パスワード",
		"12345678901234567890",
		"MixedCasePassword123!",
	}

	for _, password := range testCases {
		t.Run(password, func(t *testing.T) {
			// Hash password
			hash, err := auth.HashPassword(password)
			if err != nil {
				t.Fatalf("Failed to hash password '%s': %v", password, err)
			}

			// Verify với correct password
			err = auth.CheckPassword(password, hash)
			if err != nil {
				t.Errorf("CheckPassword failed for correct password '%s': %v", password, err)
			}

			// Verify với wrong password
			wrongPassword := password + "_wrong"
			err = auth.CheckPassword(wrongPassword, hash)
			if err == nil {
				t.Errorf("CheckPassword should fail for wrong password '%s'", wrongPassword)
			}
		})
	}
}

// TestHashPasswordMultipleTimes kiểm tra hash nhiều lần
func TestHashPasswordMultipleTimes(t *testing.T) {
	password := "testPassword123"
	hashes := make(map[string]bool)

	// Hash 10 lần
	for i := 0; i < 10; i++ {
		hash, err := auth.HashPassword(password)
		if err != nil {
			t.Fatalf("Iteration %d: HashPassword failed: %v", i, err)
		}

		// Kiểm tra hash unique
		if hashes[hash] {
			t.Errorf("Iteration %d: Duplicate hash generated", i)
		}
		hashes[hash] = true

		// Verify hash
		err = auth.CheckPassword(password, hash)
		if err != nil {
			t.Errorf("Iteration %d: CheckPassword failed: %v", i, err)
		}
	}

	// Kiểm tra đã tạo đủ 10 hash khác nhau
	if len(hashes) != 10 {
		t.Errorf("Expected 10 unique hashes, got %d", len(hashes))
	}
}

// TestCheckPasswordSpecialCharacters kiểm tra password với ký tự đặc biệt
func TestCheckPasswordSpecialCharacters(t *testing.T) {
	specialPasswords := []string{
		"pass@word!",
		"p@$$w0rd#123",
		"test<>\"'&password",
		"pwd\t\n\r",
		"pwd with\nnewline",
	}

	for _, password := range specialPasswords {
		t.Run(password, func(t *testing.T) {
			hash, err := auth.HashPassword(password)
			if err != nil {
				t.Fatalf("Failed to hash password: %v", err)
			}

			err = auth.CheckPassword(password, hash)
			if err != nil {
				t.Errorf("CheckPassword failed: %v", err)
			}
		})
	}
}
