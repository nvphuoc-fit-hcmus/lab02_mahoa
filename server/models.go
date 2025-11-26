package main

import "time"

// User represents a user in the system
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// Note represents an encrypted note
type Note struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserID           uint      `gorm:"not null;index" json:"user_id"`
	Title            string    `gorm:"not null" json:"title"`
	EncryptedContent string    `gorm:"type:text;not null" json:"encrypted_content"`
	IV               string    `gorm:"not null" json:"iv"` // Initialization Vector
	CreatedAt        time.Time `json:"created_at"`
	User             User      `gorm:"foreignKey:UserID" json:"-"`
}

// SharedLink represents a time-limited sharing link
type SharedLink struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	NoteID     uint      `gorm:"not null;index" json:"note_id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	ShareToken string    `gorm:"uniqueIndex;not null" json:"share_token"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	Note       Note      `gorm:"foreignKey:NoteID" json:"-"`
	User       User      `gorm:"foreignKey:UserID" json:"-"`
}

// Request/Response Models

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

// CreateNoteRequest for creating a new note
type CreateNoteRequest struct {
	Title            string `json:"title" binding:"required"`
	EncryptedContent string `json:"encrypted_content" binding:"required"`
	IV               string `json:"iv" binding:"required"`
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
