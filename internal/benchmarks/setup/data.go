package setup

import (
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// DataGenerator handles test data creation for benchmarks
type DataGenerator struct {
	DB *gorm.DB
}

// NewDataGenerator creates a new data generator instance
func NewDataGenerator(db *gorm.DB) *DataGenerator {
	return &DataGenerator{DB: db}
}

// DataSetConfig defines the size of test datasets
type DataSetConfig struct {
	Users              int
	Epics              int
	UserStoriesPerEpic int
	RequirementsPerUS  int
	AcceptanceCriteria int
	Comments           int
}

// PredefinedDataSets contains common dataset configurations
var PredefinedDataSets = map[string]DataSetConfig{
	"small": {
		Users:              10,
		Epics:              25,
		UserStoriesPerEpic: 4,
		RequirementsPerUS:  3,
		AcceptanceCriteria: 100,
		Comments:           50,
	},
	"medium": {
		Users:              50,
		Epics:              100,
		UserStoriesPerEpic: 5,
		RequirementsPerUS:  3,
		AcceptanceCriteria: 500,
		Comments:           250,
	},
	"large": {
		Users:              200,
		Epics:              500,
		UserStoriesPerEpic: 4,
		RequirementsPerUS:  3,
		AcceptanceCriteria: 2000,
		Comments:           1000,
	},
}

// GenerateDataSet creates a complete dataset based on configuration
func (dg *DataGenerator) GenerateDataSet(config DataSetConfig) error {

	// Create users first
	users, err := dg.CreateUsers(config.Users)
	if err != nil {
		return fmt.Errorf("failed to create users: %w", err)
	}

	// Create epics
	epics, err := dg.CreateEpics(config.Epics, users)
	if err != nil {
		return fmt.Errorf("failed to create epics: %w", err)
	}

	// Create user stories
	var allUserStories []*models.UserStory
	for _, epic := range epics {
		userStories, err := dg.CreateUserStories(config.UserStoriesPerEpic, epic, users)
		if err != nil {
			return fmt.Errorf("failed to create user stories for epic %s: %w", epic.ReferenceID, err)
		}
		allUserStories = append(allUserStories, userStories...)
	}

	// Create requirements
	var allRequirements []*models.Requirement
	for _, userStory := range allUserStories {
		requirements, err := dg.CreateRequirements(config.RequirementsPerUS, userStory, users)
		if err != nil {
			return fmt.Errorf("failed to create requirements for user story %s: %w", userStory.ReferenceID, err)
		}
		allRequirements = append(allRequirements, requirements...)
	}

	// Create acceptance criteria
	if err := dg.CreateAcceptanceCriteria(config.AcceptanceCriteria, allRequirements, users); err != nil {
		return fmt.Errorf("failed to create acceptance criteria: %w", err)
	}

	// Create comments
	if err := dg.CreateComments(config.Comments, allRequirements, users); err != nil {
		return fmt.Errorf("failed to create comments: %w", err)
	}

	return nil
}

// CreateUsers generates test users
func (dg *DataGenerator) CreateUsers(count int) ([]*models.User, error) {
	users := make([]*models.User, count)
	
	for i := 0; i < count; i++ {
		user := &models.User{
			Username:     fmt.Sprintf("benchmark_user_%d", i+1),
			Email:        fmt.Sprintf("benchmark_user_%d@example.com", i+1),
			PasswordHash: "benchmark-password-hash", // In real usage, this would be properly hashed
			Role:         models.RoleUser,
		}
		
		if err := dg.DB.Create(user).Error; err != nil {
			return nil, fmt.Errorf("failed to create user %d: %w", i+1, err)
		}
		
		users[i] = user
	}
	
	return users, nil
}

// CreateEpics generates test epics
func (dg *DataGenerator) CreateEpics(count int, users []*models.User) ([]*models.Epic, error) {
	epics := make([]*models.Epic, count)
	
	for i := 0; i < count; i++ {
		description := fmt.Sprintf("This is a benchmark epic for performance testing purposes. Epic number %d contains multiple user stories and requirements for comprehensive testing.", i+1)
		epic := &models.Epic{
			Title:       fmt.Sprintf("Benchmark Epic %d", i+1),
			Description: &description,
			CreatorID:   users[rand.Intn(len(users))].ID,
			AssigneeID:  users[rand.Intn(len(users))].ID,
			Priority:    models.PriorityMedium,
			Status:      models.EpicStatusBacklog,
		}
		
		if err := dg.DB.Create(epic).Error; err != nil {
			return nil, fmt.Errorf("failed to create epic %d: %w", i+1, err)
		}
		
		epics[i] = epic
	}
	
	return epics, nil
}

// CreateUserStories generates test user stories for an epic
func (dg *DataGenerator) CreateUserStories(count int, epic *models.Epic, users []*models.User) ([]*models.UserStory, error) {
	userStories := make([]*models.UserStory, count)
	
	// Get the current count once at the beginning
	var baseCount int64
	dg.DB.Model(&models.UserStory{}).Count(&baseCount)
	
	for i := 0; i < count; i++ {
		description := fmt.Sprintf("As a user, I want to perform action %d so that I can achieve goal %d. This user story is part of epic %s for benchmark testing.", i+1, i+1, epic.ReferenceID)
		
		// Generate a unique reference ID for benchmarks
		referenceID := fmt.Sprintf("US-%03d", baseCount+int64(i)+1)
		
		userStory := &models.UserStory{
			ReferenceID: referenceID,
			Title:       fmt.Sprintf("User Story %d for %s", i+1, epic.Title),
			Description: &description,
			EpicID:      epic.ID,
			CreatorID:   users[rand.Intn(len(users))].ID,
			AssigneeID:  users[rand.Intn(len(users))].ID,
			Priority:    models.PriorityMedium,
			Status:      models.UserStoryStatusBacklog,
		}
		
		if err := dg.DB.Create(userStory).Error; err != nil {
			return nil, fmt.Errorf("failed to create user story %d: %w", i+1, err)
		}
		
		userStories[i] = userStory
	}
	
	return userStories, nil
}

// CreateRequirements generates test requirements for a user story
func (dg *DataGenerator) CreateRequirements(count int, userStory *models.UserStory, users []*models.User) ([]*models.Requirement, error) {
	requirements := make([]*models.Requirement, count)
	
	// Get the requirement type (should already exist)
	var reqType models.RequirementType
	if err := dg.DB.Where("name = ?", "Functional").First(&reqType).Error; err != nil {
		return nil, fmt.Errorf("failed to find requirement type: %w", err)
	}
	
	// Get the current count once at the beginning
	var baseCount int64
	dg.DB.Model(&models.Requirement{}).Count(&baseCount)
	
	for i := 0; i < count; i++ {
		description := fmt.Sprintf("The system shall implement functionality %d to support user story %s. This requirement includes detailed specifications for benchmark testing purposes.", i+1, userStory.ReferenceID)
		
		// Generate a unique reference ID for benchmarks
		referenceID := fmt.Sprintf("REQ-%03d", baseCount+int64(i)+1)
		
		requirement := &models.Requirement{
			ReferenceID: referenceID,
			Title:       fmt.Sprintf("Requirement %d for %s", i+1, userStory.Title),
			Description: &description,
			UserStoryID: userStory.ID,
			CreatorID:   users[rand.Intn(len(users))].ID,
			AssigneeID:  users[rand.Intn(len(users))].ID,
			Priority:    models.PriorityMedium,
			Status:      models.RequirementStatusDraft,
			TypeID:      reqType.ID,
		}
		
		if err := dg.DB.Create(requirement).Error; err != nil {
			return nil, fmt.Errorf("failed to create requirement %d: %w", i+1, err)
		}
		
		requirements[i] = requirement
	}
	
	return requirements, nil
}

// CreateAcceptanceCriteria generates test acceptance criteria
func (dg *DataGenerator) CreateAcceptanceCriteria(count int, requirements []*models.Requirement, users []*models.User) error {
	// Get the current count once at the beginning
	var baseCount int64
	dg.DB.Model(&models.AcceptanceCriteria{}).Count(&baseCount)
	
	for i := 0; i < count; i++ {
		requirement := requirements[rand.Intn(len(requirements))]
		
		// Generate a unique reference ID for benchmarks
		referenceID := fmt.Sprintf("AC-%03d", baseCount+int64(i)+1)
		
		ac := &models.AcceptanceCriteria{
			ReferenceID:  referenceID,
			Description:  fmt.Sprintf("GIVEN the system is in state X, WHEN action Y is performed, THEN result Z should occur. This is acceptance criteria %d for benchmark testing.", i+1),
			UserStoryID:  requirement.UserStoryID, // Link to the user story, not requirement
			AuthorID:     users[rand.Intn(len(users))].ID,
		}
		
		if err := dg.DB.Create(ac).Error; err != nil {
			return fmt.Errorf("failed to create acceptance criteria %d: %w", i+1, err)
		}
	}
	
	return nil
}

// CreateComments generates test comments
func (dg *DataGenerator) CreateComments(count int, requirements []*models.Requirement, users []*models.User) error {
	for i := 0; i < count; i++ {
		requirement := requirements[rand.Intn(len(requirements))]
		
		comment := &models.Comment{
			Content:    fmt.Sprintf("This is a benchmark comment %d. It provides feedback and discussion about the requirement for performance testing purposes.", i+1),
			EntityType: models.EntityTypeRequirement,
			EntityID:   requirement.ID,
			AuthorID:   users[rand.Intn(len(users))].ID,
		}
		
		if err := dg.DB.Create(comment).Error; err != nil {
			return fmt.Errorf("failed to create comment %d: %w", i+1, err)
		}
	}
	
	return nil
}

// CleanupData removes all test data from the database
func (dg *DataGenerator) CleanupData() error {
	// Delete in reverse order of dependencies
	tables := []string{
		"comments",
		"acceptance_criteria", 
		"requirements",
		"user_stories",
		"epics",
		"users",
	}
	
	for _, table := range tables {
		if err := dg.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			return fmt.Errorf("failed to cleanup table %s: %w", table, err)
		}
	}
	
	return nil
}

// init seeds the random number generator
func init() {
	rand.Seed(time.Now().UnixNano())
}