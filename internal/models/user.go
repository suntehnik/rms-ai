package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents the role of a user in the system
// @Description Role that determines user permissions and access levels in the system
// @Example "User"
type UserRole string

const (
	RoleAdministrator UserRole = "Administrator" // Administrator - full system access including user and configuration management
	RoleUser          UserRole = "User"          // User - can create, edit, and delete entities
	RoleCommenter     UserRole = "Commenter"     // Commenter - can only add comments, limited editing capabilities
)

// User represents a system user
// @Description A user account in the system with authentication and role-based permissions
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`             // Unique identifier for the user
	Username     string    `gorm:"uniqueIndex;not null" json:"username" validate:"required,min=3,max=50" example:"john_doe"`   // Unique username for login
	Email        string    `gorm:"uniqueIndex;not null" json:"email" validate:"required,email" example:"john.doe@example.com"` // Unique email address for login and notifications
	PasswordHash string    `gorm:"not null" json:"-"`                                                                          // Hashed password (never exposed in JSON responses)
	Role         UserRole  `gorm:"not null" json:"role" validate:"required" example:"User"`                                    // User role determining permissions
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`                                                  // Timestamp when the user account was created
	UpdatedAt    time.Time `json:"updated_at" example:"2023-01-02T12:30:00Z"`                                                  // Timestamp when the user account was last updated

	// Relationships (excluded from JSON to prevent circular references and reduce payload size)
	CreatedEpics               []Epic                    `gorm:"foreignKey:CreatorID" json:"-"`  // Epics created by this user
	AssignedEpics              []Epic                    `gorm:"foreignKey:AssigneeID" json:"-"` // Epics assigned to this user
	CreatedUserStories         []UserStory               `gorm:"foreignKey:CreatorID" json:"-"`  // User stories created by this user
	AssignedUserStories        []UserStory               `gorm:"foreignKey:AssigneeID" json:"-"` // User stories assigned to this user
	AuthoredAcceptanceCriteria []AcceptanceCriteria      `gorm:"foreignKey:AuthorID" json:"-"`   // Acceptance criteria authored by this user
	CreatedRequirements        []Requirement             `gorm:"foreignKey:CreatorID" json:"-"`  // Requirements created by this user
	AssignedRequirements       []Requirement             `gorm:"foreignKey:AssigneeID" json:"-"` // Requirements assigned to this user
	Comments                   []Comment                 `gorm:"foreignKey:AuthorID" json:"-"`   // Comments authored by this user
	CreatedRelationships       []RequirementRelationship `gorm:"foreignKey:CreatedBy" json:"-"`  // Requirement relationships created by this user
}

// BeforeCreate sets the ID if not already set
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// IsAdministrator checks if the user has administrator role
func (u *User) IsAdministrator() bool {
	return u.Role == RoleAdministrator
}

// IsUser checks if the user has user role
func (u *User) IsUser() bool {
	return u.Role == RoleUser
}

// IsCommenter checks if the user has commenter role
func (u *User) IsCommenter() bool {
	return u.Role == RoleCommenter
}

// CanEdit checks if the user can edit entities (Administrator or User roles)
func (u *User) CanEdit() bool {
	return u.Role == RoleAdministrator || u.Role == RoleUser
}

// CanDelete checks if the user can delete entities (Administrator or User roles)
func (u *User) CanDelete() bool {
	return u.Role == RoleAdministrator || u.Role == RoleUser
}

// CanManageUsers checks if the user can manage other users (Administrator role only)
func (u *User) CanManageUsers() bool {
	return u.Role == RoleAdministrator
}

// CanManageConfig checks if the user can manage system configuration (Administrator role only)
func (u *User) CanManageConfig() bool {
	return u.Role == RoleAdministrator
}
