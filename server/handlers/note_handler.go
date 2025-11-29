package handlers

import (
	"encoding/json"
	"fmt"
	"lab02_mahoa/server/auth"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CreateNoteHandler handles creating a new encrypted note
func CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	claims, err := AuthenticateRequest(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req models.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Title == "" || req.EncryptedContent == "" || req.IV == "" || req.EncryptedKey == "" { // <--- CẬP NHẬT DÒNG NÀY
        RespondWithError(w, http.StatusBadRequest, "Title, content, IV and encrypted key are required")
        return
    }

	db := database.GetDB()

	// Create note
	note := models.Note{
		UserID:           claims.UserID,
		Title:            req.Title,
		EncryptedContent: req.EncryptedContent,
		IV:               req.IV,
		EncryptedKey:     req.EncryptedKey,
	}

	if err := db.Create(&note).Error; err != nil {
		log.Printf("Error creating note: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create note")
		return
	}

	RespondWithJSON(w, http.StatusCreated, models.NoteResponse{
		ID:               note.ID,
		Title:            note.Title,
		EncryptedContent: note.EncryptedContent,
		EncryptedKey:     note.EncryptedKey,
		IV:               note.IV,
		CreatedAt:        note.CreatedAt,
	})
}

// ListNotesHandler returns all notes for the authenticated user
func ListNotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	claims, err := AuthenticateRequest(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	db := database.GetDB()

	// Get all notes for user
	var notes []models.Note
	if err := db.Where("user_id = ?", claims.UserID).Order("created_at DESC").Find(&notes).Error; err != nil {
		log.Printf("Error fetching notes: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch notes")
		return
	}

	// Convert to response format
	noteResponses := make([]models.NoteResponse, len(notes))
	for i, note := range notes {
		// Check if note has active shared links
		var shareCount int64
		db.Model(&models.SharedLink{}).Where("note_id = ? AND expires_at > ?", note.ID, time.Now()).Count(&shareCount)

		noteResponses[i] = models.NoteResponse{
			ID:               note.ID,
			Title:            note.Title,
			EncryptedContent: note.EncryptedContent,
			EncryptedKey:     note.EncryptedKey,
			IV:               note.IV,
			CreatedAt:        note.CreatedAt,
			IsShared:         shareCount > 0,
		}
	}

	RespondWithJSON(w, http.StatusOK, models.ListNotesResponse{
		Notes: noteResponses,
		Count: len(noteResponses),
	})
}

// GetNoteHandler returns a specific note by ID
func GetNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	claims, err := AuthenticateRequest(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Extract note ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		RespondWithError(w, http.StatusBadRequest, "Note ID is required")
		return
	}

	noteID, err := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 32)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	db := database.GetDB()

	// Get note
	var note models.Note
	if err := db.Where("id = ? AND user_id = ?", noteID, claims.UserID).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			RespondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		log.Printf("Error fetching note: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch note")
		return
	}

	RespondWithJSON(w, http.StatusOK, models.NoteResponse{
		ID:               note.ID,
		Title:            note.Title,
		EncryptedContent: note.EncryptedContent,
		EncryptedKey:     note.EncryptedKey,
		IV:               note.IV,
		CreatedAt:        note.CreatedAt,
	})
}

// DeleteNoteHandler deletes a note by ID
func DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	claims, err := AuthenticateRequest(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Extract note ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		RespondWithError(w, http.StatusBadRequest, "Note ID is required")
		return
	}

	noteID, err := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 32)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	db := database.GetDB()

	// Delete note (only if it belongs to the user)
	result := db.Where("id = ? AND user_id = ?", noteID, claims.UserID).Delete(&models.Note{})
	if result.Error != nil {
		log.Printf("Error deleting note: %v", result.Error)
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	if result.RowsAffected == 0 {
		RespondWithError(w, http.StatusNotFound, "Note not found")
		return
	}

	RespondWithJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Note deleted successfully",
	})
}

// RevokeShareHandler revokes all sharing links for a note
func RevokeShareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	claims, err := AuthenticateRequest(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Extract note ID from URL path: /api/notes/:id/revoke
	pathParts := strings.Split(r.URL.Path, "/")
	// pathParts = ["", "api", "notes", "id", "revoke"]
	if len(pathParts) < 4 {
		RespondWithError(w, http.StatusBadRequest, "Note ID is required")
		return
	}

	// The ID is at position len(pathParts)-2 (before "revoke")
	noteID, err := strconv.ParseUint(pathParts[len(pathParts)-2], 10, 32)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	db := database.GetDB()

	// Verify note belongs to user
	var note models.Note
	if err := db.Where("id = ? AND user_id = ?", noteID, claims.UserID).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			RespondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		log.Printf("Error fetching note: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch note")
		return
	}

	// Delete all SharedLinks for this note
	if err := db.Where("note_id = ?", noteID).Delete(&models.SharedLink{}).Error; err != nil {
		log.Printf("Error revoking shares: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to revoke shares")
		return
	}

	RespondWithJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Sharing revoked successfully",
	})
}

// Helper Functions

// AuthenticateRequest validates JWT token from Authorization header
func AuthenticateRequest(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	tokenString, err := auth.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return nil, err
	}

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}

// CreateShareHandler creates a share link for a note
func CreateShareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	claims, err := AuthenticateRequest(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Extract note ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		RespondWithError(w, http.StatusBadRequest, "Note ID is required")
		return
	}

	noteID, err := strconv.ParseUint(pathParts[len(pathParts)-2], 10, 32)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid note ID")
		return
	}

	var req models.CreateShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.DurationHours = 24 // Default 24 hours
	}

	db := database.GetDB()

	// Verify note belongs to user
	var note models.Note
	if err := db.Where("id = ? AND user_id = ?", noteID, claims.UserID).First(&note).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			RespondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		log.Printf("Error fetching note: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch note")
		return
	}

	// Generate share token
	shareToken := fmt.Sprintf("share_%d_%d", noteID, time.Now().Unix())

	// Create share link
	shareLink := models.SharedLink{
		NoteID:     uint(noteID),
		UserID:     claims.UserID,
		ShareToken: shareToken,
		ExpiresAt:  time.Now().Add(time.Hour * time.Duration(req.DurationHours)),
	}

	if err := db.Create(&shareLink).Error; err != nil {
		log.Printf("Error creating share link: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create share link")
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"success":     true,
		"share_token": shareToken,
		"expires_at":  shareLink.ExpiresAt,
		"message":     "Share link created successfully",
	})
}
