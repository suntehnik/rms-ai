package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

func main() {
	fmt.Println("Verifying GORM models...")

	// Create in-memory SQLite database for verification
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate all models
	fmt.Println("Running auto-migration...")
	err = models.AutoMigrate(db)
	if err != nil {
		log.Fatalf("Failed to migrate models: %v", err)
	}

	// Seed default data
	fmt.Println("Seeding default data...")
	err = models.SeedDefaultData(db)
	if err != nil {
		log.Fatalf("Failed to seed default data: %v", err)
	}

	// Verify default data was created
	var reqTypeCount int64
	db.Model(&models.RequirementType{}).Count(&reqTypeCount)
	fmt.Printf("Created %d requirement types\n", reqTypeCount)

	var relTypeCount int64
	db.Model(&models.RelationshipType{}).Count(&relTypeCount)
	fmt.Printf("Created %d relationship types\n", relTypeCount)

	// Test creating a complete hierarchy
	fmt.Println("Testing entity creation...")

	// Create user
	user := models.User{
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleUser,
	}
	err = db.Create(&user).Error
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("Created user with ID: %s\n", user.ID)

	// Create epic
	epic := models.Epic{
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityHigh,
		Status:      models.EpicStatusBacklog,
		Title:       "Test Epic",
		Description: stringPtr("Test epic description"),
	}
	err = db.Create(&epic).Error
	if err != nil {
		log.Fatalf("Failed to create epic: %v", err)
	}
	fmt.Printf("Created epic with ID: %s, Reference: %s\n", epic.ID, epic.ReferenceID)

	// Create user story
	userStory := models.UserStory{
		EpicID:      epic.ID,
		CreatorID:   user.ID,
		AssigneeID:  user.ID,
		Priority:    models.PriorityMedium,
		Status:      models.UserStoryStatusBacklog,
		Title:       "Test User Story",
		Description: stringPtr("As a user, I want to test, so that I can verify functionality"),
	}
	err = db.Create(&userStory).Error
	if err != nil {
		log.Fatalf("Failed to create user story: %v", err)
	}
	fmt.Printf("Created user story with ID: %s, Reference: %s\n", userStory.ID, userStory.ReferenceID)

	// Create acceptance criteria
	ac := models.AcceptanceCriteria{
		UserStoryID: userStory.ID,
		AuthorID:    user.ID,
		Description: "WHEN user clicks button THEN system SHALL display message",
	}
	err = db.Create(&ac).Error
	if err != nil {
		log.Fatalf("Failed to create acceptance criteria: %v", err)
	}
	fmt.Printf("Created acceptance criteria with ID: %s, Reference: %s\n", ac.ID, ac.ReferenceID)

	// Get requirement type
	var reqType models.RequirementType
	err = db.Where("name = ?", "Functional").First(&reqType).Error
	if err != nil {
		log.Fatalf("Failed to find requirement type: %v", err)
	}

	// Create requirement
	requirement := models.Requirement{
		UserStoryID:          userStory.ID,
		AcceptanceCriteriaID: &ac.ID,
		CreatorID:            user.ID,
		AssigneeID:           user.ID,
		Priority:             models.PriorityHigh,
		Status:               models.RequirementStatusActive,
		TypeID:               reqType.ID,
		Title:                "Test Requirement",
		Description:          stringPtr("Detailed requirement description"),
	}
	err = db.Create(&requirement).Error
	if err != nil {
		log.Fatalf("Failed to create requirement: %v", err)
	}
	fmt.Printf("Created requirement with ID: %s, Reference: %s\n", requirement.ID, requirement.ReferenceID)

	// Create comment
	comment := models.Comment{
		EntityType: models.EntityTypeEpic,
		EntityID:   epic.ID,
		AuthorID:   user.ID,
		Content:    "This is a test comment",
	}
	err = db.Create(&comment).Error
	if err != nil {
		log.Fatalf("Failed to create comment: %v", err)
	}
	fmt.Printf("Created comment with ID: %s\n", comment.ID)

	// Test relationships
	fmt.Println("Testing relationships...")
	var loadedEpic models.Epic
	err = db.Preload("UserStories.AcceptanceCriteria").Preload("UserStories.Requirements").Where("id = ?", epic.ID).First(&loadedEpic).Error
	if err != nil {
		log.Fatalf("Failed to load epic with relationships: %v", err)
	}

	fmt.Printf("Epic has %d user stories\n", len(loadedEpic.UserStories))
	if len(loadedEpic.UserStories) > 0 {
		fmt.Printf("User story has %d acceptance criteria\n", len(loadedEpic.UserStories[0].AcceptanceCriteria))
		fmt.Printf("User story has %d requirements\n", len(loadedEpic.UserStories[0].Requirements))
	}

	// Test validation methods
	fmt.Println("Testing validation methods...")
	fmt.Printf("User can edit: %t\n", user.CanEdit())
	fmt.Printf("Epic priority string: %s\n", epic.GetPriorityString())
	fmt.Printf("User story follows template: %t\n", userStory.IsUserStoryTemplate())
	fmt.Printf("Acceptance criteria is EARS format: %t\n", ac.IsEARSFormat())
	fmt.Printf("Comment is general comment: %t\n", comment.IsGeneralComment())

	fmt.Println("✅ All models verified successfully!")
}

func stringPtr(s string) *string {
	return &s
}