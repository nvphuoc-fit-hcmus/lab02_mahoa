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
	Title            string `json:"title" binding:"required"`
	EncryptedContent string `json:"encrypted_content" binding:"required"`
	IV               string `json:"iv" binding:"required"`
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
	IV               string    `json:"iv"`
	CreatedAt        time.Time `json:"created_at"`
}

// ListNotesResponse for returning list of notes
type ListNotesResponse struct {
	Notes []NoteResponse `json:"notes"`
	Count int            `json:"count"`
}
