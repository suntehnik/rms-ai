package docs

import (
	"time"

	"github.com/google/uuid"
)

// ExampleData contains comprehensive examples for all entity types
// These examples are used in Swagger documentation and interactive testing
type ExampleData struct {
	Epic               ExampleEpic               `json:"epic"`
	UserStory          ExampleUserStory          `json:"user_story"`
	AcceptanceCriteria ExampleAcceptanceCriteria `json:"acceptance_criteria"`
	Requirement        ExampleRequirement        `json:"requirement"`
	Comment            ExampleComment            `json:"comment"`
	User               ExampleUser               `json:"user"`
	RequirementType    ExampleRequirementType    `json:"requirement_type"`
	RelationshipType   ExampleRelationshipType   `json:"relationship_type"`
}

// ExampleEpic provides realistic example data for Epic entities
type ExampleEpic struct {
	ID          string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	ReferenceID string    `json:"reference_id" example:"EP-001"`
	Title       string    `json:"title" example:"User Authentication System"`
	Description string    `json:"description" example:"Implement comprehensive user authentication and authorization system with JWT tokens, role-based access control, and secure session management"`
	Priority    int       `json:"priority" example:"1"`
	Status      string    `json:"status" example:"in_progress"`
	CreatorID   string    `json:"creator_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	AssigneeID  *string   `json:"assignee_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174002"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-15T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2023-01-20T14:45:00Z"`
}

// ExampleUserStory provides realistic example data for UserStory entities
type ExampleUserStory struct {
	ID          string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174010"`
	ReferenceID string    `json:"reference_id" example:"US-001"`
	Title       string    `json:"title" example:"User Login with JWT Token"`
	Description string    `json:"description" example:"As a user, I want to log in with my credentials and receive a JWT token, so that I can access protected resources securely"`
	Priority    int       `json:"priority" example:"1"`
	Status      string    `json:"status" example:"ready"`
	EpicID      string    `json:"epic_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CreatorID   string    `json:"creator_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	AssigneeID  *string   `json:"assignee_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174003"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-16T09:15:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2023-01-18T16:20:00Z"`
}

// ExampleAcceptanceCriteria provides realistic example data for AcceptanceCriteria entities
type ExampleAcceptanceCriteria struct {
	ID          string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174020"`
	ReferenceID string    `json:"reference_id" example:"AC-001"`
	UserStoryID string    `json:"user_story_id" example:"123e4567-e89b-12d3-a456-426614174010"`
	AuthorID    string    `json:"author_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	Description string    `json:"description" example:"WHEN user enters valid credentials THEN system SHALL authenticate user and return JWT token with 1-hour expiration"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-16T11:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2023-01-16T11:30:00Z"`
}

// ExampleRequirement provides realistic example data for Requirement entities
type ExampleRequirement struct {
	ID                   string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174030"`
	ReferenceID          string    `json:"reference_id" example:"REQ-001"`
	Title                string    `json:"title" example:"JWT Token Generation"`
	Description          string    `json:"description" example:"System must generate JWT tokens using HS256 algorithm with configurable secret key and expiration time"`
	Priority             int       `json:"priority" example:"1"`
	Status               string    `json:"status" example:"approved"`
	UserStoryID          *string   `json:"user_story_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174010"`
	AcceptanceCriteriaID *string   `json:"acceptance_criteria_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174020"`
	RequirementTypeID    string    `json:"requirement_type_id" example:"123e4567-e89b-12d3-a456-426614174040"`
	CreatorID            string    `json:"creator_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	AssigneeID           *string   `json:"assignee_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174004"`
	CreatedAt            time.Time `json:"created_at" example:"2023-01-17T08:45:00Z"`
	UpdatedAt            time.Time `json:"updated_at" example:"2023-01-19T13:15:00Z"`
}

// ExampleComment provides realistic example data for Comment entities
type ExampleComment struct {
	ID               string     `json:"id" example:"123e4567-e89b-12d3-a456-426614174050"`
	EntityType       string     `json:"entity_type" example:"epic"`
	EntityID         string     `json:"entity_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	AuthorID         string     `json:"author_id" example:"123e4567-e89b-12d3-a456-426614174001"`
	Content          string     `json:"content" example:"This epic looks comprehensive. Should we consider adding two-factor authentication as well?"`
	IsInline         bool       `json:"is_inline" example:"false"`
	InlinePosition   *int       `json:"inline_position,omitempty" example:"45"`
	LinkedText       *string    `json:"linked_text,omitempty" example:"authentication system"`
	ParentCommentID  *string    `json:"parent_comment_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174051"`
	Status           string     `json:"status" example:"open"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty" example:"2023-01-21T10:00:00Z"`
	ResolvedByUserID *string    `json:"resolved_by_user_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174002"`
	CreatedAt        time.Time  `json:"created_at" example:"2023-01-18T14:30:00Z"`
	UpdatedAt        time.Time  `json:"updated_at" example:"2023-01-18T14:30:00Z"`
}

// ExampleUser provides realistic example data for User entities
type ExampleUser struct {
	ID        string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174001"`
	Username  string    `json:"username" example:"john.doe"`
	Email     string    `json:"email" example:"john.doe@company.com"`
	FullName  string    `json:"full_name" example:"John Doe"`
	Role      string    `json:"role" example:"user"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-10T09:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-15T12:30:00Z"`
}

// ExampleRequirementType provides realistic example data for RequirementType entities
type ExampleRequirementType struct {
	ID          string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174040"`
	Name        string    `json:"name" example:"Functional Requirement"`
	Description string    `json:"description" example:"Requirements that specify what the system should do - the functionality and behavior"`
	Color       string    `json:"color" example:"#2563eb"`
	IsActive    bool      `json:"is_active" example:"true"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// ExampleRelationshipType provides realistic example data for RelationshipType entities
type ExampleRelationshipType struct {
	ID          string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174060"`
	Name        string    `json:"name" example:"depends_on"`
	Description string    `json:"description" example:"Indicates that one requirement depends on another requirement being completed first"`
	IsActive    bool      `json:"is_active" example:"true"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// GetExampleData returns comprehensive example data for all entity types
func GetExampleData() *ExampleData {
	now := time.Now()
	assigneeID := uuid.New().String()

	return &ExampleData{
		Epic: ExampleEpic{
			ID:          uuid.New().String(),
			ReferenceID: "EP-001",
			Title:       "User Authentication System",
			Description: "Implement comprehensive user authentication and authorization system with JWT tokens, role-based access control, and secure session management",
			Priority:    1,
			Status:      "in_progress",
			CreatorID:   uuid.New().String(),
			AssigneeID:  &assigneeID,
			CreatedAt:   now.Add(-5 * 24 * time.Hour),
			UpdatedAt:   now.Add(-1 * 24 * time.Hour),
		},
		UserStory: ExampleUserStory{
			ID:          uuid.New().String(),
			ReferenceID: "US-001",
			Title:       "User Login with JWT Token",
			Description: "As a user, I want to log in with my credentials and receive a JWT token, so that I can access protected resources securely",
			Priority:    1,
			Status:      "ready",
			EpicID:      uuid.New().String(),
			CreatorID:   uuid.New().String(),
			AssigneeID:  &assigneeID,
			CreatedAt:   now.Add(-4 * 24 * time.Hour),
			UpdatedAt:   now.Add(-2 * 24 * time.Hour),
		},
		AcceptanceCriteria: ExampleAcceptanceCriteria{
			ID:          uuid.New().String(),
			ReferenceID: "AC-001",
			UserStoryID: uuid.New().String(),
			AuthorID:    uuid.New().String(),
			Description: "WHEN user enters valid credentials THEN system SHALL authenticate user and return JWT token with 1-hour expiration",
			CreatedAt:   now.Add(-3 * 24 * time.Hour),
			UpdatedAt:   now.Add(-3 * 24 * time.Hour),
		},
		Requirement: ExampleRequirement{
			ID:                   uuid.New().String(),
			ReferenceID:          "REQ-001",
			Title:                "JWT Token Generation",
			Description:          "System must generate JWT tokens using HS256 algorithm with configurable secret key and expiration time",
			Priority:             1,
			Status:               "approved",
			UserStoryID:          &assigneeID,
			AcceptanceCriteriaID: &assigneeID,
			RequirementTypeID:    uuid.New().String(),
			CreatorID:            uuid.New().String(),
			AssigneeID:           &assigneeID,
			CreatedAt:            now.Add(-2 * 24 * time.Hour),
			UpdatedAt:            now.Add(-6 * time.Hour),
		},
		Comment: ExampleComment{
			ID:         uuid.New().String(),
			EntityType: "epic",
			EntityID:   uuid.New().String(),
			AuthorID:   uuid.New().String(),
			Content:    "This epic looks comprehensive. Should we consider adding two-factor authentication as well?",
			IsInline:   false,
			Status:     "open",
			CreatedAt:  now.Add(-1 * 24 * time.Hour),
			UpdatedAt:  now.Add(-1 * 24 * time.Hour),
		},
		User: ExampleUser{
			ID:        uuid.New().String(),
			Username:  "john.doe",
			Email:     "john.doe@company.com",
			FullName:  "John Doe",
			Role:      "user",
			IsActive:  true,
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now.Add(-5 * 24 * time.Hour),
		},
		RequirementType: ExampleRequirementType{
			ID:          uuid.New().String(),
			Name:        "Functional Requirement",
			Description: "Requirements that specify what the system should do - the functionality and behavior",
			Color:       "#2563eb",
			IsActive:    true,
			CreatedAt:   now.Add(-60 * 24 * time.Hour),
			UpdatedAt:   now.Add(-60 * 24 * time.Hour),
		},
		RelationshipType: ExampleRelationshipType{
			ID:          uuid.New().String(),
			Name:        "depends_on",
			Description: "Indicates that one requirement depends on another requirement being completed first",
			IsActive:    true,
			CreatedAt:   now.Add(-60 * 24 * time.Hour),
			UpdatedAt:   now.Add(-60 * 24 * time.Hour),
		},
	}
}

// GetExampleRequestBodies returns example request bodies for different operations
func GetExampleRequestBodies() map[string]interface{} {
	return map[string]interface{}{
		"create_epic": map[string]interface{}{
			"title":       "User Authentication System",
			"description": "Implement comprehensive user authentication and authorization system",
			"priority":    1,
			"creator_id":  "123e4567-e89b-12d3-a456-426614174001",
		},
		"update_epic": map[string]interface{}{
			"title":       "Enhanced User Authentication System",
			"description": "Implement comprehensive user authentication with two-factor authentication",
			"priority":    1,
			"status":      "in_progress",
			"assignee_id": "123e4567-e89b-12d3-a456-426614174002",
		},
		"create_user_story": map[string]interface{}{
			"title":       "User Login with JWT Token",
			"description": "As a user, I want to log in with my credentials and receive a JWT token",
			"priority":    1,
			"epic_id":     "123e4567-e89b-12d3-a456-426614174000",
			"creator_id":  "123e4567-e89b-12d3-a456-426614174001",
		},
		"create_acceptance_criteria": map[string]interface{}{
			"user_story_id": "123e4567-e89b-12d3-a456-426614174010",
			"author_id":     "123e4567-e89b-12d3-a456-426614174001",
			"description":   "WHEN user enters valid credentials THEN system SHALL authenticate user and return JWT token",
		},
		"create_requirement": map[string]interface{}{
			"title":                  "JWT Token Generation",
			"description":            "System must generate JWT tokens using HS256 algorithm",
			"priority":               1,
			"user_story_id":          "123e4567-e89b-12d3-a456-426614174010",
			"acceptance_criteria_id": "123e4567-e89b-12d3-a456-426614174020",
			"requirement_type_id":    "123e4567-e89b-12d3-a456-426614174040",
			"creator_id":             "123e4567-e89b-12d3-a456-426614174001",
		},
		"create_comment": map[string]interface{}{
			"entity_type": "epic",
			"entity_id":   "123e4567-e89b-12d3-a456-426614174000",
			"author_id":   "123e4567-e89b-12d3-a456-426614174001",
			"content":     "This epic looks comprehensive. Should we consider adding two-factor authentication?",
		},
		"create_inline_comment": map[string]interface{}{
			"entity_type":     "user_story",
			"entity_id":       "123e4567-e89b-12d3-a456-426614174010",
			"author_id":       "123e4567-e89b-12d3-a456-426614174001",
			"content":         "Consider adding password strength requirements here",
			"is_inline":       true,
			"inline_position": 45,
			"linked_text":     "credentials",
		},
		"create_requirement_relationship": map[string]interface{}{
			"source_requirement_id": "123e4567-e89b-12d3-a456-426614174030",
			"target_requirement_id": "123e4567-e89b-12d3-a456-426614174031",
			"relationship_type_id":  "123e4567-e89b-12d3-a456-426614174060",
			"description":           "JWT generation depends on user authentication being implemented first",
		},
	}
}

// GetExampleQueryParameters returns example query parameters for different endpoints
func GetExampleQueryParameters() map[string]map[string]string {
	return map[string]map[string]string{
		"search": {
			"query":      "authentication",
			"limit":      "20",
			"offset":     "0",
			"sort_by":    "relevance",
			"sort_order": "desc",
			"creator_id": "123e4567-e89b-12d3-a456-426614174001",
			"priority":   "1",
			"status":     "in_progress",
		},
		"list_epics": {
			"limit":       "25",
			"offset":      "0",
			"order_by":    "created_at DESC",
			"creator_id":  "123e4567-e89b-12d3-a456-426614174001",
			"assignee_id": "123e4567-e89b-12d3-a456-426614174002",
			"status":      "in_progress",
			"priority":    "1",
		},
		"list_user_stories": {
			"limit":    "30",
			"offset":   "0",
			"epic_id":  "123e4567-e89b-12d3-a456-426614174000",
			"status":   "ready",
			"priority": "1",
		},
		"list_requirements": {
			"limit":               "40",
			"offset":              "0",
			"user_story_id":       "123e4567-e89b-12d3-a456-426614174010",
			"requirement_type_id": "123e4567-e89b-12d3-a456-426614174040",
			"status":              "approved",
		},
	}
}
