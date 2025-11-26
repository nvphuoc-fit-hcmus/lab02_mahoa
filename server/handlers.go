package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if len(req.Username) < 3 {
		respondWithError(w, http.StatusBadRequest, "Username must be at least 3 characters")
		return
	}
	if len(req.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	// Check if username already exists
	var existingUser User
	if err := DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		respondWithError(w, http.StatusConflict, "Username already exists")
		return
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Create new user
	user := User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
	}

	if err := DB.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondWithJSON(w, http.StatusCreated, SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("User '%s' registered successfully", user.Username),
	})
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Find user by username
	var user User
	if err := DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}
		log.Printf("Error finding user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Verify password
	if err := CheckPassword(req.Password, user.PasswordHash); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusOK, LoginResponse{
		Token:    token,
		Username: user.Username,
		Message:  "Login successful",
	})
}

// LogoutHandler handles user logout (client-side token removal)
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	respondWithJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Logout successful. Please remove your token from the client.",
	})
}

// CreateNoteHandler handles creating a new encrypted note
func CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	claims, err := authenticateRequest(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Title == "" || req.EncryptedContent == "" || req.IV == "" {
		respondWithError(w, http.StatusBadRequest, "Title, encrypted content, and IV are required")
		return
	}

	// Create note
	note := Note{
		UserID:           claims.UserID,
		Title:            req.Title,
		EncryptedContent: req.EncryptedContent,
		IV:               req.IV,
	}

	if err := DB.Create(&note).Error; err != nil {
		log.Printf("Error creating note: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create note")
		return
	}

	respondWithJSON(w, http.StatusCreated, NoteResponse{
		ID:               note.ID,
		Title:            note.Title,
		EncryptedContent: note.EncryptedContent,
		IV:               note.IV,
		CreatedAt:        note.CreatedAt,
	})
}

// ListNotesHandler returns all notes for the authenticated user
func ListNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	claims, err := authenticateRequest(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Get all notes for user
	var notes []Note
	if err := DB.Where("user_id = ?", claims.UserID).Order("created_at DESC").Find(&notes).Error; err != nil {
		log.Printf("Error fetching notes: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch notes")
		return
	}

	// Convert to response format
	noteResponses := make([]NoteResponse, len(notes))
	for i, note := range notes {
		noteResponses[i] = NoteResponse{
			ID:               note.ID,
			Title:            note.Title,
			EncryptedContent: note.EncryptedContent,
			IV:               note.IV,
			CreatedAt:        note.CreatedAt,
		}
	}

	respondWithJSON(w, http.StatusOK, ListNotesResponse{
		Notes: noteResponses,
		Count: len(noteResponses),
	})
}

// GetNoteHandler returns a specific note by ID
func GetNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	claims, err := authenticateRequest(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Extract note ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		respondWithError(w, http.StatusBadRequest, "Note ID is required")
		return
	}

	noteID, err := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	// Get note
	var note Note
	if err := DB.Where("id = ? AND user_id = ?", noteID, claims.UserID).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		log.Printf("Error fetching note: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch note")
		return
	}

	respondWithJSON(w, http.StatusOK, NoteResponse{
		ID:               note.ID,
		Title:            note.Title,
		EncryptedContent: note.EncryptedContent,
		IV:               note.IV,
		CreatedAt:        note.CreatedAt,
	})
}

// DeleteNoteHandler deletes a note by ID
func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	claims, err := authenticateRequest(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Extract note ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		respondWithError(w, http.StatusBadRequest, "Note ID is required")
		return
	}

	noteID, err := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 32)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	// Delete note (only if it belongs to the user)
	result := DB.Where("id = ? AND user_id = ?", noteID, claims.UserID).Delete(&Note{})
	if result.Error != nil {
		log.Printf("Error deleting note: %v", result.Error)
		respondWithError(w, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	if result.RowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Note not found")
		return
	}

	respondWithJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Note deleted successfully",
	})
}

// Helper Functions

// authenticateRequest validates JWT token from Authorization header
func authenticateRequest(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	tokenString, err := ExtractTokenFromHeader(authHeader)
	if err != nil {
		return nil, err
	}

	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}

// respondWithJSON writes JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// respondWithError writes error JSON response
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})
}
