package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleAdministrator UserRole = "Administrator"
	RoleUser          UserRole = "User"
	RoleCommenter     UserRole = "Commenter"
)

// User represents a system user
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Role         UserRole  `gorm:"not null" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	CreatedEpics              []Epic              `gorm:"foreignKey:CreatorID" json:"-"`
	AssignedEpics             []Epic              `gorm:"foreignKey:AssigneeID" json:"-"`
	CreatedUserStories        []UserStory         `gorm:"foreignKey:CreatorID" json:"-"`
	AssignedUserStories       []UserStory         `gorm:"foreignKey:AssigneeID" json:"-"`
	AuthoredAcceptanceCriteria []AcceptanceCriteria `gorm:"foreignKey:AuthorID" json:"-"`
	CreatedRequirements       []Requirement       `gorm:"foreignKey:CreatorID" json:"-"`
	AssignedRequirements      []Requirement       `gorm:"foreignKey:AssigneeID" json:"-"`
	Comments                  []Comment           `gorm:"foreignKey:AuthorID" json:"-"`
	CreatedRelationships      []RequirementRelationship `gorm:"foreignKey:CreatedBy" json:"-"`
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