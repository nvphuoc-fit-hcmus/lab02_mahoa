package auth_test

import (
	"bytes"
	"encoding/json"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/handlers"
	"lab02_mahoa/server/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

// setupTestDB khởi tạo database test
func setupTestDB(t *testing.T) {
	err := database.InitTestDB(&models.User{})
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	// Xóa dữ liệu cũ để test độc lập
	db := database.GetDB()
	db.Exec("DELETE FROM users")
}

// TestRegisterSuccess kiểm tra đăng ký thành công
func TestRegisterSuccess(t *testing.T) {
	setupTestDB(t)

	// Tạo request body
	reqBody := models.RegisterRequest{
		Username: "testuser123",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	// Tạo HTTP request
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Gọi handler
	handlers.RegisterHandler(w, req)

	// Kiểm tra status code
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Kiểm tra response body
	var response models.SuccessResponse
	json.NewDecoder(w.Body).Decode(&response)

	if !response.Success {
		t.Error("Expected success to be true")
	}
	if response.Message == "" {
		t.Error("Expected success message")
	}
}

// TestRegisterDuplicateUsername kiểm tra đăng ký với username đã tồn tại
func TestRegisterDuplicateUsername(t *testing.T) {
	setupTestDB(t)

	username := "duplicateuser"

	// Đăng ký lần đầu
	reqBody := models.RegisterRequest{
		Username: username,
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handlers.RegisterHandler(w, req)

	// Đăng ký lần thứ 2 với cùng username
	req2 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	handlers.RegisterHandler(w2, req2)

	// Kiểm tra status code phải là Conflict
	if w2.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w2.Code)
	}

	// Kiểm tra error message
	var response models.ErrorResponse
	json.NewDecoder(w2.Body).Decode(&response)

	if response.Message != "Username already exists" {
		t.Errorf("Expected 'Username already exists', got '%s'", response.Message)
	}
}

// TestRegisterShortUsername kiểm tra username quá ngắn
func TestRegisterShortUsername(t *testing.T) {
	setupTestDB(t)

	reqBody := models.RegisterRequest{
		Username: "ab", // Chỉ 2 ký tự
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.RegisterHandler(w, req)

	// Kiểm tra status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Kiểm tra error message
	var response models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Message != "Username must be at least 3 characters" {
		t.Errorf("Expected username length error, got '%s'", response.Message)
	}
}

// TestRegisterShortPassword kiểm tra password quá ngắn
func TestRegisterShortPassword(t *testing.T) {
	setupTestDB(t)

	reqBody := models.RegisterRequest{
		Username: "validuser",
		Password: "12345", // Chỉ 5 ký tự
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.RegisterHandler(w, req)

	// Kiểm tra status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	// Kiểm tra error message
	var response models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Message != "Password must be at least 6 characters" {
		t.Errorf("Expected password length error, got '%s'", response.Message)
	}
}

// TestRegisterInvalidJSON kiểm tra request body không hợp lệ
func TestRegisterInvalidJSON(t *testing.T) {
	setupTestDB(t)

	// Tạo invalid JSON
	invalidJSON := []byte(`{"username": "test", "password":`)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.RegisterHandler(w, req)

	// Kiểm tra status code
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestRegisterMethodNotAllowed kiểm tra sai HTTP method
func TestRegisterMethodNotAllowed(t *testing.T) {
	setupTestDB(t)

	// Sử dụng GET thay vì POST
	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	w := httptest.NewRecorder()

	handlers.RegisterHandler(w, req)

	// Kiểm tra status code
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

// TestRegisterEmptyFields kiểm tra các trường rỗng
func TestRegisterEmptyFields(t *testing.T) {
	setupTestDB(t)

	tests := []struct {
		name     string
		username string
		password string
		wantCode int
	}{
		{"Empty username", "", "password123", http.StatusBadRequest},
		{"Empty password", "testuser", "", http.StatusBadRequest},
		{"Both empty", "", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := models.RegisterRequest{
				Username: tt.username,
				Password: tt.password,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handlers.RegisterHandler(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("Expected status %d, got %d", tt.wantCode, w.Code)
			}
		})
	}
}
