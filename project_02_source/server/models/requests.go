package models

import "time"

// Request Models

// RegisterRequest for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginRequest for user login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateNoteRequest for creating a new note
type CreateNoteRequest struct {
	Title            string `json:"title"`
	EncryptedContent string `json:"encrypted_content"`
	
	EncryptedKey     string `json:"encrypted_key"`
	IV               string `json:"iv"`
}

// CreateShareRequest for creating a share link
type CreateShareRequest struct {
	DurationHours   int `json:"duration_hours"`   // How many hours the link is valid (default 24)
	DurationMinutes int `json:"duration_minutes"` // Alternative: duration in minutes (for testing)
}

// Response Models

// LoginResponse returns JWT token
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// ErrorResponse for API errors
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse for general success messages
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NoteResponse for returning note data
type NoteResponse struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	EncryptedContent string    `json:"encrypted_content"`
	EncryptedKey     string    `json:"encrypted_key"`
	IV               string    `json:"iv"`
	CreatedAt        time.Time `json:"created_at"`
	IsShared         bool      `json:"is_shared"` // Track if note has active shares
}

// ListNotesResponse for returning list of notes
type ListNotesResponse struct {
	Notes []NoteResponse `json:"notes"`
	Count int            `json:"count"`
}

// SharedNoteResponse for returning shared note data
type SharedNoteResponse struct {
	ID               uint      `json:"id"`
	Title            string    `json:"title"`
	EncryptedContent string    `json:"encrypted_content"`
	IV               string    `json:"iv"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	OwnerUsername    string    `json:"owner_username"`
}

// ShareLinkResponse for returning share link info
type ShareLinkResponse struct {
	Success    bool      `json:"success"`
	ShareToken string    `json:"share_token"`
	ShareURL   string    `json:"share_url"`
	ExpiresAt  time.Time `json:"expires_at"`
	Message    string    `json:"message"`
}
