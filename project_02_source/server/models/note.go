package models

import "time"

// Note represents an encrypted note
type Note struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserID           uint      `gorm:"not null;index" json:"user_id"`
	Title            string    `gorm:"not null" json:"title"`
	EncryptedContent string    `gorm:"type:text;not null" json:"encrypted_content"`
	IV               string    `gorm:"not null" json:"iv"` // Initialization Vector for content
	CreatedAt        time.Time `json:"created_at"`
	User             User      `gorm:"foreignKey:UserID" json:"-"`
	EncryptedKey     string    `gorm:"type:text;not null" json:"encrypted_key"`
	EncryptedKeyIV   string    `gorm:"type:text" json:"encrypted_key_iv"` // IV for encrypted key (nullable for backward compatibility)
}

// SharedLink represents a time-limited sharing link
type SharedLink struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	NoteID          uint      `gorm:"not null;index" json:"note_id"`
	UserID          uint      `gorm:"not null;index" json:"user_id"`
	ShareToken      string    `gorm:"uniqueIndex;not null" json:"share_token"`
	ExpiresAt       time.Time `gorm:"not null" json:"expires_at"`
	MaxAccessCount  int       `gorm:"default:0" json:"max_access_count"`          // 0 = unlimited
	AccessCount     int       `gorm:"default:0" json:"access_count"`
	RequirePassword bool      `gorm:"default:false" json:"require_password"`
	PasswordHash    string    `gorm:"type:text" json:"-"`                         // Bcrypt hash, not exposed in JSON
	CreatedAt       time.Time `json:"created_at"`
	Note            Note      `gorm:"foreignKey:NoteID" json:"-"`
	User            User      `gorm:"foreignKey:UserID" json:"-"`
}

// E2EEShare represents an end-to-end encrypted share between two specific users
// Uses Diffie-Hellman key exchange for secure session key generation
type E2EEShare struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	NoteID             uint      `gorm:"not null;index" json:"note_id"`
	SenderID           uint      `gorm:"not null;index" json:"sender_id"`
	RecipientID        uint      `gorm:"not null;index" json:"recipient_id"`
	SenderPublicKey    string    `gorm:"type:text;not null" json:"sender_public_key"`    // Sender's DH public key (base64)
	EncryptedContent   string    `gorm:"type:text;not null" json:"encrypted_content"`    // Content encrypted with DH shared secret
	ContentIV          string    `gorm:"not null" json:"content_iv"`                     // IV for encrypted content
	ExpiresAt          time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt          time.Time `json:"created_at"`
	Note               Note      `gorm:"foreignKey:NoteID" json:"-"`
	Sender             User      `gorm:"foreignKey:SenderID" json:"-"`
	Recipient          User      `gorm:"foreignKey:RecipientID" json:"-"`
}
