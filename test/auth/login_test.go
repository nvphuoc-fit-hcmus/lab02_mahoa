package auth_test

import (
	"bytes"
	"encoding/json"
	"lab02_mahoa/server/auth"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/handlers"
	"lab02_mahoa/server/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupLoginTest khởi tạo database và tạo user test
func setupLoginTest(t *testing.T) models.RegisterRequest {
	err := database.InitTestDB(&models.User{})
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	db := database.GetDB()
	db.Exec("DELETE FROM users")

	// Tạo user mẫu để test login
	testUser := models.RegisterRequest{
		Username: "logintest",
		Password: "password123",
	}

	hashedPassword, _ := auth.HashPassword(testUser.Password)
	user := models.User{
		Username:     testUser.Username,
		PasswordHash: hashedPassword,
	}
	db.Create(&user)

	return testUser
}

// TestLoginSuccess kiểm tra đăng nhập thành công
func TestLoginSuccess(t *testing.T) {
	testUser := setupLoginTest(t)

	// Tạo login request
	reqBody := models.LoginRequest{
		Username: testUser.Username,
		Password: testUser.Password,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.LoginHandler(w, req)

	// Kiểm tra status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Kiểm tra response body
	var response models.LoginResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Token == "" {
		t.Error("Expected token in response")
	}
	if response.Username != testUser.Username {
		t.Errorf("Expected username '%s', got '%s'", testUser.Username, response.Username)
	}
	if response.Message == "" {
		t.Error("Expected success message")
	}
}

// TestLoginWrongPassword kiểm tra đăng nhập với mật khẩu sai
func TestLoginWrongPassword(t *testing.T) {
	testUser := setupLoginTest(t)

	// Tạo login request với password sai
	reqBody := models.LoginRequest{
		Username: testUser.Username,
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.LoginHandler(w, req)

	// Kiểm tra status code phải là Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Kiểm tra error message
	var response models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Message != "Invalid username or password" {
		t.Errorf("Expected invalid credentials error, got '%s'", response.Message)
	}
}

// TestLoginUserNotFound kiểm tra đăng nhập với username không tồn tại
func TestLoginUserNotFound(t *testing.T) {
	setupLoginTest(t)

	// Tạo login request với username không tồn tại
	reqBody := models.LoginRequest{
		Username: "nonexistentuser",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.LoginHandler(w, req)

	// Kiểm tra status code phải là Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Kiểm tra error message
	var response models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Message != "Invalid username or password" {
		t.Errorf("Expected invalid credentials error, got '%s'", response.Message)
	}
}

// TestLoginInvalidJSON kiểm tra login với request body không hợp lệ
func TestLoginInvalidJSON(t *testing.T) {
	setupLoginTest(t)

	// Tạo invalid JSON
	invalidJSON := []byte(`{"username": "test"`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.LoginHandler(w, req)

	// Kiểm tra status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestLoginMethodNotAllowed kiểm tra sai HTTP method
func TestLoginMethodNotAllowed(t *testing.T) {
	setupLoginTest(t)

	// Sử dụng GET thay vì POST
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	handlers.LoginHandler(w, req)

	// Kiểm tra status code
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

// TestLoginEmptyCredentials kiểm tra login với thông tin rỗng
func TestLoginEmptyCredentials(t *testing.T) {
	setupLoginTest(t)

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"Empty username", "", "password123"},
		{"Empty password", "testuser", ""},
		{"Both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := models.LoginRequest{
				Username: tt.username,
				Password: tt.password,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handlers.LoginHandler(w, req)

			// Login với empty fields sẽ trả về Unauthorized
			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
			}
		})
	}
}

// TestLoginCaseSensitive kiểm tra login có phân biệt chữ hoa/thường
func TestLoginCaseSensitive(t *testing.T) {
	testUser := setupLoginTest(t)

	// Thử login với username khác case
	reqBody := models.LoginRequest{
		Username: "LOGINTEST", // uppercase
		Password: testUser.Password,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.LoginHandler(w, req)

	// Username case-sensitive nên phải fail
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for case-sensitive username, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestLoginMultipleAttempts kiểm tra nhiều lần login
func TestLoginMultipleAttempts(t *testing.T) {
	testUser := setupLoginTest(t)

	// Login thành công nhiều lần
	for i := 0; i < 3; i++ {
		reqBody := models.LoginRequest{
			Username: testUser.Username,
			Password: testUser.Password,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.LoginHandler(w, req)

		// Mỗi lần đều phải thành công
		if w.Code != http.StatusOK {
			t.Errorf("Attempt %d: Expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}

		// Mỗi lần phải trả về token mới
		var response models.LoginResponse
		json.NewDecoder(w.Body).Decode(&response)
		if response.Token == "" {
			t.Errorf("Attempt %d: Expected token in response", i+1)
		}
	}
}
