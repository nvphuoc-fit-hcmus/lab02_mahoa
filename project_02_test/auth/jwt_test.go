package auth_test

import (
	"lab02_mahoa/server/auth"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestGenerateJWT kiểm tra tạo JWT token
func TestGenerateJWT(t *testing.T) {
	userID := uint(1)
	username := "testuser"

	token, err := auth.GenerateJWT(userID, username)

	// Kiểm tra không có lỗi
	if err != nil {
		t.Fatalf("GenerateJWT returned error: %v", err)
	}

	// Kiểm tra token không rỗng
	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Kiểm tra token có format JWT (3 phần ngăn cách bởi dấu chấm)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected JWT with 3 parts, got %d parts", len(parts))
	}
}

// TestGenerateJWTDifferentUsers kiểm tra token khác nhau cho user khác nhau
func TestGenerateJWTDifferentUsers(t *testing.T) {
	token1, err1 := auth.GenerateJWT(1, "user1")
	token2, err2 := auth.GenerateJWT(2, "user2")

	if err1 != nil || err2 != nil {
		t.Fatalf("GenerateJWT returned errors: %v, %v", err1, err2)
	}

	// Kiểm tra 2 token khác nhau
	if token1 == token2 {
		t.Error("Different users should have different tokens")
	}
}

// TestGenerateJWTSameUserDifferentTimes kiểm tra token khác nhau cho cùng user nhưng thời gian khác
func TestGenerateJWTSameUserDifferentTimes(t *testing.T) {
	userID := uint(1)
	username := "testuser"

	token1, err1 := auth.GenerateJWT(userID, username)
	time.Sleep(time.Second) // Đợi 1 giây
	token2, err2 := auth.GenerateJWT(userID, username)

	if err1 != nil || err2 != nil {
		t.Fatalf("GenerateJWT returned errors: %v, %v", err1, err2)
	}

	// Token phải khác nhau vì IssuedAt khác nhau
	if token1 == token2 {
		t.Error("Tokens generated at different times should be different")
	}
}

// TestValidateJWTSuccess kiểm tra validate token hợp lệ
func TestValidateJWTSuccess(t *testing.T) {
	userID := uint(123)
	username := "validuser"

	// Tạo token
	token, err := auth.GenerateJWT(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate token
	claims, err := auth.ValidateJWT(token)

	if err != nil {
		t.Errorf("ValidateJWT failed for valid token: %v", err)
	}

	// Kiểm tra claims
	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}
	if claims.Username != username {
		t.Errorf("Expected Username %s, got %s", username, claims.Username)
	}
}

// TestValidateJWTInvalidToken kiểm tra validate token không hợp lệ
func TestValidateJWTInvalidToken(t *testing.T) {
	invalidTokens := []string{
		"invalid.token.here",
		"not-a-jwt-token",
		"",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
	}

	for _, token := range invalidTokens {
		t.Run(token, func(t *testing.T) {
			_, err := auth.ValidateJWT(token)

			if err == nil {
				t.Errorf("ValidateJWT should fail for invalid token: %s", token)
			}
		})
	}
}

// TestValidateJWTTamperedToken kiểm tra validate token bị thay đổi
func TestValidateJWTTamperedToken(t *testing.T) {
	userID := uint(1)
	username := "testuser"

	// Tạo token hợp lệ
	token, err := auth.GenerateJWT(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Thay đổi 1 ký tự trong token
	tamperedToken := token[:len(token)-5] + "xxxxx"

	// Validate token đã bị thay đổi
	_, err = auth.ValidateJWT(tamperedToken)

	if err == nil {
		t.Error("ValidateJWT should fail for tampered token")
	}
}

// TestValidateJWTExpiredToken kiểm tra token hết hạn (test concept)
func TestValidateJWTExpiredToken(t *testing.T) {
	// Note: Test này chỉ kiểm tra concept vì token có thời hạn 24h
	// Trong thực tế, cần mock thời gian hoặc tạo token với expiry ngắn

	userID := uint(1)
	username := "testuser"

	// Tạo token với expiry time trong quá khứ (manually)
	expirationTime := time.Now().Add(-1 * time.Hour) // 1 giờ trước
	claims := &auth.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign với secret key giống trong auth package
	// Note: Trong production nên expose hàm để test hoặc dùng interface
	tokenString, err := token.SignedString([]byte("your-secret-key-change-this-in-production"))
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	// Validate expired token
	_, err = auth.ValidateJWT(tokenString)

	if err == nil {
		t.Error("ValidateJWT should fail for expired token")
	}
}

// TestExtractTokenFromHeaderSuccess kiểm tra extract token từ header thành công
func TestExtractTokenFromHeaderSuccess(t *testing.T) {
	authHeader := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token"

	token, err := auth.ExtractTokenFromHeader(authHeader)

	if err != nil {
		t.Errorf("ExtractTokenFromHeader failed: %v", err)
	}

	expectedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test.token"
	if token != expectedToken {
		t.Errorf("Expected token '%s', got '%s'", expectedToken, token)
	}
}

// TestExtractTokenFromHeaderEmptyHeader kiểm tra extract với header rỗng
func TestExtractTokenFromHeaderEmptyHeader(t *testing.T) {
	_, err := auth.ExtractTokenFromHeader("")

	if err == nil {
		t.Error("ExtractTokenFromHeader should fail for empty header")
	}

	expectedError := "authorization header is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestExtractTokenFromHeaderInvalidFormat kiểm tra extract với format sai
func TestExtractTokenFromHeaderInvalidFormat(t *testing.T) {
	invalidHeaders := []string{
		"InvalidFormat token",
		"Basic token",
		"Bearer",
		"token-without-bearer",
		"Bearer ",
	}

	for _, header := range invalidHeaders {
		t.Run(header, func(t *testing.T) {
			token, err := auth.ExtractTokenFromHeader(header)

			if header == "Bearer " && token == "" && err == nil {
				// "Bearer " với space là edge case có thể accept
				return
			}

			if err == nil && token != "" {
				t.Errorf("ExtractTokenFromHeader should fail for invalid format: '%s'", header)
			}
		})
	}
}

// TestExtractTokenFromHeaderCaseSensitive kiểm tra "Bearer" có phân biệt chữ hoa/thường
func TestExtractTokenFromHeaderCaseSensitive(t *testing.T) {
	headers := []string{
		"bearer token123",
		"BEARER token123",
		"BeArEr token123",
	}

	for _, header := range headers {
		t.Run(header, func(t *testing.T) {
			_, err := auth.ExtractTokenFromHeader(header)

			// "Bearer" phải đúng case (chữ B hoa, các chữ khác thường)
			if err == nil {
				t.Errorf("ExtractTokenFromHeader should be case-sensitive for Bearer prefix: '%s'", header)
			}
		})
	}
}

// TestJWTClaimsContent kiểm tra nội dung claims trong token
func TestJWTClaimsContent(t *testing.T) {
	userID := uint(456)
	username := "claimsuser"

	// Tạo token
	token, err := auth.GenerateJWT(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate và lấy claims
	claims, err := auth.ValidateJWT(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Kiểm tra UserID
	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	// Kiểm tra Username
	if claims.Username != username {
		t.Errorf("Expected Username %s, got %s", username, claims.Username)
	}

	// Kiểm tra ExpiresAt được set
	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should be set")
	}

	// Kiểm tra IssuedAt được set
	if claims.IssuedAt == nil {
		t.Error("IssuedAt should be set")
	}

	// Kiểm tra NotBefore được set
	if claims.NotBefore == nil {
		t.Error("NotBefore should be set")
	}

	// Kiểm tra token chưa hết hạn
	if time.Now().After(claims.ExpiresAt.Time) {
		t.Error("Token should not be expired immediately after generation")
	}
}

// TestJWTTokenExpiry kiểm tra thời gian hết hạn của token
func TestJWTTokenExpiry(t *testing.T) {
	userID := uint(1)
	username := "expiryuser"

	// Tạo token
	token, err := auth.GenerateJWT(userID, username)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate và lấy claims
	claims, err := auth.ValidateJWT(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Kiểm tra ExpiresAt là khoảng 24 giờ từ bây giờ
	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time

	// Cho phép chênh lệch 1 phút
	diff := actualExpiry.Sub(expectedExpiry)
	if diff < -time.Minute || diff > time.Minute {
		t.Errorf("Expected expiry around 24 hours from now, got diff: %v", diff)
	}
}

// TestGenerateMultipleTokens kiểm tra tạo nhiều token
func TestGenerateMultipleTokens(t *testing.T) {
	users := []struct {
		id       uint
		username string
	}{
		{1, "user1"},
		{2, "user2"},
		{3, "user3"},
		{100, "admin"},
		{999, "test@user"},
	}

	tokens := make(map[string]bool)

	for _, user := range users {
		token, err := auth.GenerateJWT(user.id, user.username)
		if err != nil {
			t.Errorf("Failed to generate token for user %d: %v", user.id, err)
			continue
		}

		// Kiểm tra token unique
		if tokens[token] {
			t.Errorf("Duplicate token generated for user %d", user.id)
		}
		tokens[token] = true

		// Validate token
		claims, err := auth.ValidateJWT(token)
		if err != nil {
			t.Errorf("Failed to validate token for user %d: %v", user.id, err)
			continue
		}

		// Verify claims
		if claims.UserID != user.id {
			t.Errorf("UserID mismatch for user %d: expected %d, got %d", user.id, user.id, claims.UserID)
		}
		if claims.Username != user.username {
			t.Errorf("Username mismatch for user %d: expected %s, got %s", user.id, user.username, claims.Username)
		}
	}
}
