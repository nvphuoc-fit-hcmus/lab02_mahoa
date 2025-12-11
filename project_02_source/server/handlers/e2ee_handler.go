package handlers

import (
	"encoding/json"
	"fmt"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/models"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CreateE2EEShareHandler creates an end-to-end encrypted share with a specific user
func CreateE2EEShareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate sender
	claims, err := AuthenticateRequest(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Extract note ID from URL path: /api/notes/:id/e2ee
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

	// Parse request body
	var req models.CreateE2EEShareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.RecipientUsername == "" {
		RespondWithError(w, http.StatusBadRequest, "Recipient username is required")
		return
	}
	if req.SenderPublicKey == "" {
		RespondWithError(w, http.StatusBadRequest, "Sender public key is required")
		return
	}
	if req.EncryptedContent == "" {
		RespondWithError(w, http.StatusBadRequest, "Encrypted content is required")
		return
	}
	if req.ContentIV == "" {
		RespondWithError(w, http.StatusBadRequest, "Content IV is required")
		return
	}

	db := database.GetDB()

	// Verify note exists and belongs to sender
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

	// Find recipient user
	var recipient models.User
	if err := db.Where("username = ?", req.RecipientUsername).First(&recipient).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			RespondWithError(w, http.StatusNotFound, "Recipient user not found")
			return
		}
		log.Printf("Error fetching recipient: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch recipient")
		return
	}

	// Prevent sharing with self
	if recipient.ID == claims.UserID {
		RespondWithError(w, http.StatusBadRequest, "Cannot share with yourself")
		return
	}

	// Set default duration if not specified
	durationHours := req.DurationHours
	if durationHours <= 0 {
		durationHours = 24 // Default 24 hours
	}

	expiresAt := time.Now().Add(time.Duration(durationHours) * time.Hour)

	// Create E2EE share
	e2eeShare := models.E2EEShare{
		NoteID:           uint(noteID),
		SenderID:         claims.UserID,
		RecipientID:      recipient.ID,
		SenderPublicKey:  req.SenderPublicKey,
		EncryptedContent: req.EncryptedContent,
		ContentIV:        req.ContentIV,
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
	}

	if err := db.Create(&e2eeShare).Error; err != nil {
		log.Printf("Error creating E2EE share: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create E2EE share")
		return
	}

	log.Printf("✅ E2EE share created: sender=%d, recipient=%d, note=%d, expires=%v",
		claims.UserID, recipient.ID, noteID, expiresAt)

	RespondWithJSON(w, http.StatusCreated, models.E2EEShareResponse{
		Success:           true,
		ShareID:           e2eeShare.ID,
		RecipientUsername: recipient.Username,
		ExpiresAt:         expiresAt,
		Message:           fmt.Sprintf("E2EE share created successfully with %s", recipient.Username),
	})
}

// ListE2EESharesHandler lists all E2EE shares received by the authenticated user
func ListE2EESharesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get all E2EE shares where user is recipient and not expired
	var shares []models.E2EEShare
	if err := db.Preload("Note").Preload("Sender").
		Where("recipient_id = ? AND expires_at > ?", claims.UserID, time.Now()).
		Order("created_at DESC").
		Find(&shares).Error; err != nil {
		log.Printf("Error fetching E2EE shares: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch E2EE shares")
		return
	}

	// Build response
	shareResponses := make([]models.E2EEShareDetailResponse, len(shares))
	for i, share := range shares {
		shareResponses[i] = models.E2EEShareDetailResponse{
			ID:               share.ID,
			NoteTitle:        share.Note.Title,
			SenderUsername:   share.Sender.Username,
			SenderPublicKey:  share.SenderPublicKey,
			EncryptedContent: share.EncryptedContent,
			ContentIV:        share.ContentIV,
			ExpiresAt:        share.ExpiresAt,
			CreatedAt:        share.CreatedAt,
		}
	}

	RespondWithJSON(w, http.StatusOK, models.ListE2EESharesResponse{
		Shares: shareResponses,
		Count:  len(shareResponses),
	})
}

// GetE2EEShareHandler retrieves a specific E2EE share by ID
func GetE2EEShareHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extract share ID from URL path: /api/e2ee/:id
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		RespondWithError(w, http.StatusBadRequest, "Share ID is required")
		return
	}

	shareID, err := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 32)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid share ID")
		return
	}

	db := database.GetDB()

	// Get E2EE share
	var share models.E2EEShare
	if err := db.Preload("Note").Preload("Sender").
		Where("id = ?", shareID).First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			RespondWithError(w, http.StatusNotFound, "E2EE share not found")
			return
		}
		log.Printf("Error fetching E2EE share: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch E2EE share")
		return
	}

	// Verify user is the recipient
	if share.RecipientID != claims.UserID {
		RespondWithError(w, http.StatusForbidden, "You don't have access to this share")
		return
	}

	// Check if share has expired
	if time.Now().After(share.ExpiresAt) {
		// Delete expired share
		db.Delete(&share)
		log.Printf("❌ E2EE share expired and deleted: id=%d", shareID)
		RespondWithError(w, http.StatusGone, "E2EE share has expired")
		return
	}

	log.Printf("✅ E2EE share accessed: id=%d, recipient=%d", shareID, claims.UserID)

	RespondWithJSON(w, http.StatusOK, models.E2EEShareDetailResponse{
		ID:               share.ID,
		NoteTitle:        share.Note.Title,
		SenderUsername:   share.Sender.Username,
		SenderPublicKey:  share.SenderPublicKey,
		EncryptedContent: share.EncryptedContent,
		ContentIV:        share.ContentIV,
		ExpiresAt:        share.ExpiresAt,
		CreatedAt:        share.CreatedAt,
	})
}

// DeleteE2EEShareHandler deletes an E2EE share (sender can revoke)
func DeleteE2EEShareHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extract share ID from URL path: /api/e2ee/:id
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		RespondWithError(w, http.StatusBadRequest, "Share ID is required")
		return
	}

	shareID, err := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 32)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid share ID")
		return
	}

	db := database.GetDB()

	// Find E2EE share
	var share models.E2EEShare
	if err := db.Where("id = ?", shareID).First(&share).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			RespondWithError(w, http.StatusNotFound, "E2EE share not found")
			return
		}
		log.Printf("Error fetching E2EE share: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch E2EE share")
		return
	}

	// Verify user is the sender
	if share.SenderID != claims.UserID {
		RespondWithError(w, http.StatusForbidden, "You can only delete shares you created")
		return
	}

	// Delete share
	if err := db.Delete(&share).Error; err != nil {
		log.Printf("Error deleting E2EE share: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete E2EE share")
		return
	}

	log.Printf("🗑️ E2EE share deleted: id=%d, sender=%d", shareID, claims.UserID)

	RespondWithJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "E2EE share deleted successfully",
	})
}
