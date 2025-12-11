package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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
	if req.Title == "" || req.EncryptedContent == "" || req.IV == "" || req.EncryptedKey == "" || req.EncryptedKeyIV == "" {
		RespondWithError(w, http.StatusBadRequest, "Title, content, IV, encrypted key and encrypted key IV are required")
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
		EncryptedKeyIV:   req.EncryptedKeyIV,
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
		IV:               note.IV,
		EncryptedKey:     note.EncryptedKey,
		EncryptedKeyIV:   note.EncryptedKeyIV,
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
			IV:               note.IV,
			EncryptedKey:     note.EncryptedKey,
			EncryptedKeyIV:   note.EncryptedKeyIV,
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
		IV:               note.IV,
		EncryptedKey:     note.EncryptedKey,
		EncryptedKeyIV:   note.EncryptedKeyIV,
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

	// Calculate expiration duration
	var duration time.Duration
	if req.DurationMinutes > 0 {
		// Use minutes if specified (for testing)
		duration = time.Minute * time.Duration(req.DurationMinutes)
	} else if req.DurationHours > 0 {
		// Use hours if specified
		duration = time.Hour * time.Duration(req.DurationHours)
	} else {
		// Default 24 hours
		duration = time.Hour * 24
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
	shareToken, err := generateSecureToken(32)
	if err != nil {
		log.Printf("Error generating share token: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate share token")
		return
	}

	// Create share link
	shareLink := models.SharedLink{
		NoteID:          uint(noteID),
		UserID:          claims.UserID,
		ShareToken:      shareToken,
		ExpiresAt:       time.Now().Add(duration),
		MaxAccessCount:  0, // Default: unlimited
		AccessCount:     0,
		RequirePassword: false,
		PasswordHash:    "",
	}

	// Handle optional password protection
	if req.Password != nil && *req.Password != "" {
		passwordHash, err := auth.HashPassword(*req.Password)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			RespondWithError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		shareLink.RequirePassword = true
		shareLink.PasswordHash = passwordHash
	}

	// Handle optional max access count
	if req.MaxAccessCount != nil && *req.MaxAccessCount > 0 {
		shareLink.MaxAccessCount = *req.MaxAccessCount
	}

	if err := db.Create(&shareLink).Error; err != nil {
		log.Printf("Error creating share link: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create share link")
		return
	}

	log.Printf("✅ Share link created: token=%s, expires_at=%v, duration=%v, max_access=%d, password_protected=%v", 
		shareToken[:10]+"...", shareLink.ExpiresAt, duration, shareLink.MaxAccessCount, shareLink.RequirePassword)

	// Create share URL (the encryption key should be added by client in fragment)
	shareURL := fmt.Sprintf("http://localhost:8080/share/%s", shareToken)

	RespondWithJSON(w, http.StatusCreated, models.ShareLinkResponse{
		Success:         true,
		ShareToken:      shareToken,
		ShareURL:        shareURL,
		ExpiresAt:       shareLink.ExpiresAt,
		MaxAccessCount:  shareLink.MaxAccessCount,
		RequirePassword: shareLink.RequirePassword,
		Message:         "Share link created successfully",
	})
}

// GetSharedNoteHandler retrieves a note via share token with time validation
func GetSharedNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract share token from URL path: /api/shares/:token
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/shares/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		RespondWithError(w, http.StatusBadRequest, "Share token is required")
		return
	}

	shareToken := pathParts[0]

	db := database.GetDB()

	// Find share link
	var shareLink models.SharedLink
	if err := db.Preload("Note").Preload("User").Where("share_token = ?", shareToken).First(&shareLink).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			RespondWithError(w, http.StatusNotFound, "Share link not found")
			return
		}
		log.Printf("Error fetching share link: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch share link")
		return
	}

	// Check if link has expired
	now := time.Now()
	log.Printf("🔍 Checking expiry: now=%v, expires_at=%v, expired=%v", 
		now, shareLink.ExpiresAt, now.After(shareLink.ExpiresAt))
	
	if now.After(shareLink.ExpiresAt) {
		// Delete expired link
		db.Delete(&shareLink)
		log.Printf("❌ Share link expired and deleted: token=%s", shareToken[:10]+"...")
		RespondWithError(w, http.StatusGone, "Share link has expired")
		return
	}

	// Check if max access count reached
	if shareLink.MaxAccessCount > 0 && shareLink.AccessCount >= shareLink.MaxAccessCount {
		// Delete exhausted link
		db.Delete(&shareLink)
		log.Printf("❌ Share link exhausted and deleted: token=%s, access_count=%d/%d", 
			shareToken[:10]+"...", shareLink.AccessCount, shareLink.MaxAccessCount)
		RespondWithError(w, http.StatusGone, "Share link has reached maximum access count")
		return
	}

	// Check password if required
	log.Printf("🔐 Password check: RequirePassword=%v, PasswordHash=%v", 
		shareLink.RequirePassword, shareLink.PasswordHash != "")
	
	if shareLink.RequirePassword {
		var req models.AccessShareRequest
		bodyBytes, _ := io.ReadAll(r.Body)
		log.Printf("📥 Request body received: %s", string(bodyBytes))
		
		// Restore body for json.Decode
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
			log.Printf("❌ Password required but not provided. Error: %v, Password empty: %v", err, req.Password == "")
			RespondWithError(w, http.StatusUnauthorized, "Password required to access this share")
			return
		}

		// Verify password
		if err := auth.CheckPassword(req.Password, shareLink.PasswordHash); err != nil {
			log.Printf("❌ Wrong password for share link: token=%s", shareToken[:10]+"...")
			RespondWithError(w, http.StatusUnauthorized, "Incorrect password")
			return
		}

		log.Printf("✅ Password verified for share link: token=%s", shareToken[:10]+"...")
	}

	// Increment access count
	shareLink.AccessCount++
	if err := db.Save(&shareLink).Error; err != nil {
		log.Printf("Error updating access count: %v", err)
		// Don't fail the request, just log the error
	}

	log.Printf("✅ Share link valid: token=%s, remaining=%v, access_count=%d/%d", 
		shareToken[:10]+"...", shareLink.ExpiresAt.Sub(now), shareLink.AccessCount, shareLink.MaxAccessCount)

	// Return shared note data (without encrypted key - key should be in URL fragment)
	RespondWithJSON(w, http.StatusOK, models.SharedNoteResponse{
		ID:               shareLink.Note.ID,
		Title:            shareLink.Note.Title,
		EncryptedContent: shareLink.Note.EncryptedContent,
		IV:               shareLink.Note.IV,
		CreatedAt:        shareLink.Note.CreatedAt,
		ExpiresAt:        shareLink.ExpiresAt,
		OwnerUsername:    shareLink.User.Username,
	})
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// URL-safe base64 encoding
	return base64.URLEncoding.EncodeToString(bytes), nil
}
