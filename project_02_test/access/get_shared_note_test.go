package access

import (
	"encoding/json"
	"fmt"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/handlers"
	"lab02_mahoa/server/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGetSharedNoteSuccess tests successfully accessing a valid shared note
func TestGetSharedNoteSuccess(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Shared Note")

	db := database.GetDB()

	// Create active share link
	shareToken := "valid_share_token_123"
	shareLink := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: shareToken,
		ExpiresAt:  time.Now().Add(1 * time.Hour), // Expires in 1 hour
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create share link")

	// Test accessing the shared note via API
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/shares/%s", shareToken), nil)
	rr := httptest.NewRecorder()
	
	handler := http.HandlerFunc(handlers.GetSharedNoteHandler)
	handler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code, "Should return 200 OK")

	var response models.SharedNoteResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "Should parse response")

	// Verify response data
	assert.Equal(t, noteID, response.ID, "Note ID should match")
	assert.Equal(t, "Shared Note", response.Title, "Title should match")
	assert.Equal(t, "encrypted_content_test", response.EncryptedContent, "Content should match")
	assert.Equal(t, "test_iv", response.IV, "IV should match")
	assert.Equal(t, "testuser", response.OwnerUsername, "Owner username should match")
	assert.True(t, response.ExpiresAt.After(time.Now()), "ExpiresAt should be in the future")
}

// TestGetSharedNoteExpired tests accessing an expired shared note
func TestGetSharedNoteExpired(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Expired Shared Note")

	db := database.GetDB()

	// Create expired share link
	shareToken := "expired_share_token_456"
	shareLink := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: shareToken,
		ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	err := db.Create(&shareLink).Error
	assert.NoError(t, err, "Should create expired share link")

	// Test accessing the expired shared note via API
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/shares/%s", shareToken), nil)
	rr := httptest.NewRecorder()
	
	handler := http.HandlerFunc(handlers.GetSharedNoteHandler)
	handler.ServeHTTP(rr, req)

	// Verify response returns 410 Gone (expired)
	assert.Equal(t, http.StatusGone, rr.Code, "Should return 410 Gone for expired link")

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "Should parse error response")
	
	// Check message field for expiration info
	if message, ok := response["message"].(string); ok {
		assert.Contains(t, message, "expired", "Error message should mention expiration")
	}

	// Verify the expired link is deleted from database
	var deletedLink models.SharedLink
	err = db.Where("share_token = ?", shareToken).First(&deletedLink).Error
	assert.Error(t, err, "Expired link should be deleted from database")
}

// TestGetSharedNoteNotFound tests accessing a non-existent shared note
func TestGetSharedNoteNotFound(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Test accessing a non-existent share token
	req, _ := http.NewRequest("GET", "/api/shares/nonexistent_token_999", nil)
	rr := httptest.NewRecorder()
	
	handler := http.HandlerFunc(handlers.GetSharedNoteHandler)
	handler.ServeHTTP(rr, req)

	// Verify response returns 404 Not Found
	assert.Equal(t, http.StatusNotFound, rr.Code, "Should return 404 Not Found")

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "Should parse error response")
	
	// Check message field for "not found" info
	if message, ok := response["message"].(string); ok {
		assert.Contains(t, message, "not found", "Error message should mention not found")
	}
}

// TestGetSharedNoteInvalidToken tests accessing with empty/invalid token
func TestGetSharedNoteInvalidToken(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Test with empty token
	req, _ := http.NewRequest("GET", "/api/shares/", nil)
	rr := httptest.NewRecorder()
	
	handler := http.HandlerFunc(handlers.GetSharedNoteHandler)
	handler.ServeHTTP(rr, req)

	// Verify response returns 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Should return 400 Bad Request")

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "Should parse error response")
}

// TestGetSharedNoteMethodNotAllowed tests using wrong HTTP method
func TestGetSharedNoteMethodNotAllowed(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Test with POST instead of GET
	req, _ := http.NewRequest("POST", "/api/shares/some_token", nil)
	rr := httptest.NewRecorder()
	
	handler := http.HandlerFunc(handlers.GetSharedNoteHandler)
	handler.ServeHTTP(rr, req)

	// Verify response returns 405 Method Not Allowed
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Should return 405 Method Not Allowed")
}

// TestGetSharedNoteMultipleAccess tests accessing the same shared note multiple times
func TestGetSharedNoteMultipleAccess(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Multi-access Note")

	db := database.GetDB()

	// Create active share link
	shareToken := "multi_access_token"
	shareLink := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: shareToken,
		ExpiresAt:  time.Now().Add(2 * time.Hour),
	}
	db.Create(&shareLink)

	// Access the shared note multiple times
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/shares/%s", shareToken), nil)
		rr := httptest.NewRecorder()
		
		handler := http.HandlerFunc(handlers.GetSharedNoteHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, fmt.Sprintf("Access %d should succeed", i+1))
	}

	// Verify share link still exists after multiple accesses
	var checkLink models.SharedLink
	err := db.Where("share_token = ?", shareToken).First(&checkLink).Error
	assert.NoError(t, err, "Share link should still exist after multiple accesses")
}

// TestGetSharedNoteExpirationBoundary tests accessing a note at exact expiration time
func TestGetSharedNoteExpirationBoundary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping time-based test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note
	userID := createTestUser(t, "testuser", "password123")
	noteID := createTestNote(t, userID, "Boundary Test Note")

	db := database.GetDB()

	// Create share link that expires in 2 seconds
	shareToken := "boundary_token"
	shareLink := models.SharedLink{
		NoteID:     noteID,
		UserID:     userID,
		ShareToken: shareToken,
		ExpiresAt:  time.Now().Add(2 * time.Second),
	}
	db.Create(&shareLink)

	// Access while still valid
	req1, _ := http.NewRequest("GET", fmt.Sprintf("/api/shares/%s", shareToken), nil)
	rr1 := httptest.NewRecorder()
	handlers.GetSharedNoteHandler(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code, "Should succeed before expiration")

	// Wait for expiration
	time.Sleep(3 * time.Second)

	// Access after expiration
	req2, _ := http.NewRequest("GET", fmt.Sprintf("/api/shares/%s", shareToken), nil)
	rr2 := httptest.NewRecorder()
	handlers.GetSharedNoteHandler(rr2, req2)
	assert.Equal(t, http.StatusGone, rr2.Code, "Should fail after expiration")
}

// TestGetSharedNoteWithDifferentUsers tests that shared notes can be accessed by anyone with the link
func TestGetSharedNoteWithDifferentUsers(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create owner user and note
	ownerID := createTestUser(t, "owner", "password123")
	noteID := createTestNote(t, ownerID, "Public Shared Note")

	// Create another user (recipient)
	createTestUser(t, "recipient", "password456")

	db := database.GetDB()

	// Owner creates share link
	shareToken := "public_share_token"
	shareLink := models.SharedLink{
		NoteID:     noteID,
		UserID:     ownerID,
		ShareToken: shareToken,
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}
	db.Create(&shareLink)

	// Anyone (including unauthenticated) should be able to access the shared note
	// This simulates public access via share link
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/shares/%s", shareToken), nil)
	// Note: No Authorization header - simulating public access
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.GetSharedNoteHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Public should be able to access shared note")

	var response models.SharedNoteResponse
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "owner", response.OwnerUsername, "Should show owner's username")
}

// TestGetSharedNoteDataIntegrity tests that shared note returns correct encrypted data
func TestGetSharedNoteDataIntegrity(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test user and note with specific data
	userID := createTestUser(t, "testuser", "password123")
	
	db := database.GetDB()
	
	// Create note with specific encrypted content
	specificNote := models.Note{
		UserID:           userID,
		Title:            "Integrity Test Note",
		EncryptedContent: "specific_encrypted_content_xyz",
		EncryptedKey:     "specific_encrypted_key_abc",
		IV:               "specific_iv_123",
	}
	db.Create(&specificNote)

	// Create share link
	shareToken := "integrity_token"
	shareLink := models.SharedLink{
		NoteID:     specificNote.ID,
		UserID:     userID,
		ShareToken: shareToken,
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}
	db.Create(&shareLink)

	// Access shared note
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/shares/%s", shareToken), nil)
	rr := httptest.NewRecorder()
	handlers.GetSharedNoteHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.SharedNoteResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	// Verify data integrity
	assert.Equal(t, "Integrity Test Note", response.Title)
	assert.Equal(t, "specific_encrypted_content_xyz", response.EncryptedContent)
	assert.Equal(t, "specific_iv_123", response.IV)
	
	// Note: EncryptedKey should NOT be in SharedNoteResponse (security measure)
	// The key should be provided via URL fragment by the client
}
