package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshToken represents a refresh token for JWT authentication
// @Description Refresh token for maintaining user sessions without re-authentication
type RefreshToken struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`         // Unique identifier for the refresh token
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"` // ID of the user who owns this token
	TokenHash  string     `gorm:"type:text;not null" json:"-"`                                                            // Bcrypt hash of the token (never exposed in JSON)
	CreatedAt  time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`                                              // Timestamp when token was created
	ExpiresAt  time.Time  `gorm:"not null;index" json:"expires_at" example:"2023-01-31T00:00:00Z"`                        // Timestamp when token expires (30 days from creation)
	LastUsedAt *time.Time `json:"last_used_at,omitempty" example:"2023-01-15T10:30:00Z"`                                  // Timestamp when token was last used for refresh

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"` // Associated user (cascade delete when user is deleted)
}

// BeforeCreate sets the ID if not already set
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the RefreshToken model
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsExpired checks if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}
