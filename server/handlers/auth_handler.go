package handlers

import (
	"encoding/json"
	"fmt"
	"lab02_mahoa/server/auth"
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/models"
	"log"
	"net/http"

	"gorm.io/gorm"
)

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if len(req.Username) < 3 {
		RespondWithError(w, http.StatusBadRequest, "Username must be at least 3 characters")
		return
	}
	if len(req.Password) < 6 {
		RespondWithError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	db := database.GetDB()

	// Check if username already exists
	var existingUser models.User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		RespondWithError(w, http.StatusConflict, "Username already exists")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Create new user
	user := models.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	RespondWithJSON(w, http.StatusCreated, models.SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("User '%s' registered successfully", user.Username),
	})
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	db := database.GetDB()

	// Find user by username
	var user models.User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}
		log.Printf("Error finding user: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Verify password
	if err := auth.CheckPassword(req.Password, user.PasswordHash); err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID, user.Username)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		RespondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	RespondWithJSON(w, http.StatusOK, models.LoginResponse{
		Token:    token,
		Username: user.Username,
		Message:  "Login successful",
	})
}

// LogoutHandler handles user logout (client-side token removal)
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	RespondWithJSON(w, http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Logout successful. Please remove your token from the client.",
	})
}
