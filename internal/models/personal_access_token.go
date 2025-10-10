package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PersonalAccessToken represents a personal access token for API authentication
// @Description Personal access token for secure API authentication without user credentials
type PersonalAccessToken struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`             // Unique identifier for the token
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`     // ID of the user who owns this token
	Name       string     `gorm:"size:255;not null" json:"name" validate:"required,min=1,max=255" example:"MCP Client Token"` // Human-readable name for the token
	TokenHash  string     `gorm:"size:255;not null" json:"-"`                                                                 // Bcrypt hash of the token (never exposed in JSON)
	Prefix     string     `gorm:"size:20;not null;default:'mcp_pat_'" json:"prefix" example:"mcp_pat_"`                       // Token prefix for identification
	Scopes     string     `gorm:"type:text;default:'[\"full_access\"]'" json:"scopes" example:"[\"full_access\"]"`            // JSON array of scopes/permissions
	ExpiresAt  *time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`                                                  // Optional expiration timestamp
	LastUsedAt *time.Time `json:"last_used_at" example:"2023-01-15T10:30:00Z"`                                                // Timestamp when token was last used
	CreatedAt  time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`                                                  // Timestamp when token was created
	UpdatedAt  time.Time  `json:"updated_at" example:"2023-01-02T12:30:00Z"`                                                  // Timestamp when token was last updated

	// Associations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"` // Associated user (populated when needed)
}

// BeforeCreate sets the ID if not already set and initializes default values
func (pat *PersonalAccessToken) BeforeCreate(tx *gorm.DB) error {
	if pat.ID == uuid.Nil {
		pat.ID = uuid.New()
	}

	// Set default prefix if not provided
	if pat.Prefix == "" {
		pat.Prefix = "mcp_pat_"
	}

	// Set default scopes if not provided
	if pat.Scopes == "" {
		pat.Scopes = `["full_access"]`
	}

	return nil
}

// TableName returns the table name for the PersonalAccessToken model
func (PersonalAccessToken) TableName() string {
	return "personal_access_tokens"
}

// IsExpired checks if the token has expired
func (pat *PersonalAccessToken) IsExpired() bool {
	if pat.ExpiresAt == nil {
		return true // expired by default
	}
	return time.Now().After(*pat.ExpiresAt)
}
