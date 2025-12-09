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
	IV               string `json:"iv"`
	EncryptedKey     string `json:"encrypted_key"`
	EncryptedKeyIV   string `json:"encrypted_key_iv"`
}

// CreateShareRequest for creating a share link
type CreateShareRequest struct {
	DurationHours   int     `json:"duration_hours"`         // How many hours the link is valid (default 24)
	DurationMinutes int     `json:"duration_minutes"`       // Alternative: duration in minutes (for testing)
	Password        *string `json:"password,omitempty"`     // Optional password protection
	MaxAccessCount  *int    `json:"max_access_count,omitempty"` // Optional max access limit (0 or nil = unlimited)
}

// AccessShareRequest for accessing a password-protected share
type AccessShareRequest struct {
	Password string `json:"password"` // Password for protected share
}

// Response Models

// LoginResponse returns JWT token
type LoginResponse struct {
	Token       string `json:"token"`
	Username    string `json:"username"`
	DHPublicKey string `json:"dh_public_key,omitempty"`
	Message     string `json:"message"`
}

// UpdatePublicKeyRequest for updating user's DH public key
type UpdatePublicKeyRequest struct {
	DHPublicKey string `json:"dh_public_key"`
}

// GetPublicKeyResponse for getting user's public key
type GetPublicKeyResponse struct {
	Username    string `json:"username"`
	DHPublicKey string `json:"dh_public_key"`
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
	EncryptedKey     string    `json:"encrypted_key"`
	EncryptedKeyIV   string    `json:"encrypted_key_iv"`
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
	Success         bool      `json:"success"`
	ShareToken      string    `json:"share_token"`
	ShareURL        string    `json:"share_url"`
	ExpiresAt       time.Time `json:"expires_at"`
	MaxAccessCount  int       `json:"max_access_count,omitempty"`  // If set
	RequirePassword bool      `json:"require_password"`            // If password is set
	Message         string    `json:"message"`
}

// CreateE2EEShareRequest for creating an E2EE share with specific user
type CreateE2EEShareRequest struct {
	RecipientUsername string `json:"recipient_username"`        // Username of recipient
	SenderPublicKey   string `json:"sender_public_key"`         // Sender's DH public key (base64)
	EncryptedContent  string `json:"encrypted_content"`         // Content encrypted with DH shared secret
	ContentIV         string `json:"content_iv"`                // IV for encrypted content
	DurationHours     int    `json:"duration_hours,omitempty"`  // Optional: default 24 hours
}

// E2EEShareResponse for returning E2EE share info
type E2EEShareResponse struct {
	Success           bool      `json:"success"`
	ShareID           uint      `json:"share_id"`
	RecipientUsername string    `json:"recipient_username"`
	ExpiresAt         time.Time `json:"expires_at"`
	Message           string    `json:"message"`
}

// E2EEShareDetailResponse for recipient to get share details
type E2EEShareDetailResponse struct {
	ID               uint      `json:"id"`
	NoteTitle        string    `json:"note_title"`
	SenderUsername   string    `json:"sender_username"`
	SenderPublicKey  string    `json:"sender_public_key"`  // Sender's DH public key
	EncryptedContent string    `json:"encrypted_content"`  // Content encrypted with shared secret
	ContentIV        string    `json:"content_iv"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// ListE2EESharesResponse for listing received E2EE shares
type ListE2EESharesResponse struct {
	Shares []E2EEShareDetailResponse `json:"shares"`
	Count  int                       `json:"count"`
}
