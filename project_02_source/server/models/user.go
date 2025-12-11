package models

import "time"

// User represents a user in the system
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	DHPublicKey  string    `gorm:"type:text" json:"dh_public_key,omitempty"` // User's DH public key for E2EE
	CreatedAt    time.Time `json:"created_at"`
}
