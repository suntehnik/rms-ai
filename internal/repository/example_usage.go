package repository

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	
	"product-requirements-management/internal/models"
)

// ExampleUsage demonstrates how to use the repository layer
func ExampleUsage() {
	// Setup database (in real application, this would be PostgreSQL)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	if err := models.AutoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed default data
	if err := models.SeedDefaultData(db); err != nil {
		log.Fatal("Failed to seed default data:", err)
	}

	// Create repositories
	repos := NewRepositories(db)

	// Example workflow: Create User -> Epic -> UserStory -> AcceptanceCriteria -> Requirement
	err = repos.WithTransaction(func(txRepos *Repositories) error {
		// 1. Create a user
		user := &models.User{
			Username:     "john_doe",
			Email:        "john.doe@example.com",
			PasswordHash: "hashed_password_here",
			Role:         models.RoleUser,
		}
		if err := txRepos.User.Create(user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		fmt.Printf("Created user: %s (ID: %s)\n", user.Username, user.ID)

		// 2. Create an epic
		epic := &models.Epic{
			ReferenceID: "EP-001",
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityHigh,
			Status:      models.EpicStatusBacklog,
			Title:       "User Authentication System",
			Description: stringPtr("Implement comprehensive user authentication and authorization system"),
		}
		if err := txRepos.Epic.Create(epic); err != nil {
			return fmt.Errorf("failed to create epic: %w", err)
		}
		fmt.Printf("Created epic: %s (ID: %s)\n", epic.Title, epic.ID)

		// 3. Create a user story
		userStory := &models.UserStory{
			ReferenceID: "US-001",
			EpicID:      epic.ID,
			CreatorID:   user.ID,
			AssigneeID:  user.ID,
			Priority:    models.PriorityMedium,
			Status:      models.UserStoryStatusBacklog,
			Title:       "User Login",
			Description: stringPtr("As a user, I want to log in to the system, so that I can access my account"),
		}
		if err := txRepos.UserStory.Create(userStory); err != nil {
			return fmt.Errorf("failed to create user story: %w", err)
		}
		fmt.Printf("Created user story: %s (ID: %s)\n", userStory.Title, userStory.ID)

		// 4. Create acceptance criteria
		acceptanceCriteria := &models.AcceptanceCriteria{
			ReferenceID: "AC-001",
			UserStoryID: userStory.ID,
			AuthorID:    user.ID,
			Description: "WHEN user enters valid credentials THEN system SHALL authenticate user and redirect to dashboard",
		}
		if err := txRepos.AcceptanceCriteria.Create(acceptanceCriteria); err != nil {
			return fmt.Errorf("failed to create acceptance criteria: %w", err)
		}
		fmt.Printf("Created acceptance criteria: %s (ID: %s)\n", acceptanceCriteria.ReferenceID, acceptanceCriteria.ID)

		// 5. Get a requirement type
		reqTypes, err := txRepos.RequirementType.List(nil, "", 1, 0)
		if err != nil {
			return fmt.Errorf("failed to get requirement types: %w", err)
		}
		if len(reqTypes) == 0 {
			return fmt.Errorf("no requirement types found")
		}

		// 6. Create a requirement
		requirement := &models.Requirement{
			ReferenceID:          "REQ-001",
			UserStoryID:          userStory.ID,
			AcceptanceCriteriaID: &acceptanceCriteria.ID,
			CreatorID:            user.ID,
			AssigneeID:           user.ID,
			Priority:             models.PriorityLow,
			Status:               models.RequirementStatusDraft,
			TypeID:               reqTypes[0].ID,
			Title:                "Implement JWT Authentication",
			Description:          stringPtr("System must implement JWT-based authentication with secure token generation and validation"),
		}
		if err := txRepos.Requirement.Create(requirement); err != nil {
			return fmt.Errorf("failed to create requirement: %w", err)
		}
		fmt.Printf("Created requirement: %s (ID: %s)\n", requirement.Title, requirement.ID)

		return nil
	})

	if err != nil {
		log.Fatal("Transaction failed:", err)
	}

	// Demonstrate querying capabilities
	fmt.Println("\n--- Querying Examples ---")

	// Get all epics
	epics, err := repos.Epic.List(nil, "created_at DESC", 0, 0)
	if err != nil {
		log.Fatal("Failed to list epics:", err)
	}
	fmt.Printf("Found %d epics\n", len(epics))

	// Get epic with user stories
	if len(epics) > 0 {
		epicWithStories, err := repos.Epic.GetWithUserStories(epics[0].ID)
		if err != nil {
			log.Fatal("Failed to get epic with user stories:", err)
		}
		fmt.Printf("Epic '%s' has %d user stories\n", epicWithStories.Title, len(epicWithStories.UserStories))
	}

	// Search requirements by text
	requirements, err := repos.Requirement.SearchByText("JWT")
	if err != nil {
		log.Fatal("Failed to search requirements:", err)
	}
	fmt.Printf("Found %d requirements containing 'JWT'\n", len(requirements))

	// Get requirements by status
	draftRequirements, err := repos.Requirement.GetByStatus(models.RequirementStatusDraft)
	if err != nil {
		log.Fatal("Failed to get draft requirements:", err)
	}
	fmt.Printf("Found %d draft requirements\n", len(draftRequirements))

	fmt.Println("\nRepository layer example completed successfully!")
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}