package models

import "time"

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
