package handlers

import (
	"encoding/json"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/models"
	"log"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

// UpdatePublicKeyHandler allows user to register/update their DH public key
func UpdatePublicKeyHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse request
	var req models.UpdatePublicKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.DHPublicKey == "" {
		RespondWithError(w, http.StatusBadRequest, "DH public key is required")
		return
	}

	db := database.GetDB()

	// Update user's public key
	if err := db.Model(&models.User{}).Where("id = ?", claims.UserID).Update("dh_public_key", req.DHPublicKey).Error; err != nil {
		log.Printf("Error updating public key: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to update public key")
		return
	}

	log.Printf("✅ User %d updated DH public key", claims.UserID)

	RespondWithJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Public key updated successfully",
	})
}

// GetPublicKeyHandler retrieves a user's DH public key by username
func GetPublicKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Authenticate user
	_, err := AuthenticateRequest(r)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Extract username from URL path: /api/users/:username/publickey
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		RespondWithError(w, http.StatusBadRequest, "Username is required")
		return
	}

	username := pathParts[len(pathParts)-2]

	db := database.GetDB()

	// Get user's public key
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("Error fetching user: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user")
		return
	}

	if user.DHPublicKey == "" {
		RespondWithError(w, http.StatusNotFound, "User has not registered a public key")
		return
	}

	RespondWithJSON(w, http.StatusOK, models.GetPublicKeyResponse{
		Username:    user.Username,
		DHPublicKey: user.DHPublicKey,
	})
}
