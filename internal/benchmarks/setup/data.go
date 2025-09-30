package setup

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"product-requirements-management/internal/models"
)

// DataGenerator provides utilities for creating realistic test datasets
type DataGenerator struct {
	DB   *gorm.DB
	rand *rand.Rand
}

// NewDataGenerator creates a new DataGenerator instance
func NewDataGenerator(db *gorm.DB) *DataGenerator {
	return &DataGenerator{
		DB:   db,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// DataSetConfig defines the size of entities to generate
type DataSetConfig struct {
	Users              int
	Epics              int
	UserStoriesPerEpic int
	RequirementsPerUS  int
	AcceptanceCriteria int
	Comments           int
}

// GetSmallDataSet returns configuration for a small dataset (development)
func GetSmallDataSet() DataSetConfig {
	return DataSetConfig{
		Users:              10,
		Epics:              25,
		UserStoriesPerEpic: 4,
		RequirementsPerUS:  3,
		AcceptanceCriteria: 50,
		Comments:           100,
	}
}

// GetMediumDataSet returns configuration for a medium dataset (CI/CD)
func GetMediumDataSet() DataSetConfig {
	return DataSetConfig{
		Users:              50,
		Epics:              100,
		UserStoriesPerEpic: 5,
		RequirementsPerUS:  3,
		AcceptanceCriteria: 250,
		Comments:           500,
	}
}

// GetLargeDataSet returns configuration for a large dataset (performance analysis)
func GetLargeDataSet() DataSetConfig {
	return DataSetConfig{
		Users:              200,
		Epics:              500,
		UserStoriesPerEpic: 4,
		RequirementsPerUS:  3,
		AcceptanceCriteria: 1000,
		Comments:           2000,
	}
}

// CreateUsers generates the specified number of users
func (dg *DataGenerator) CreateUsers(count int) ([]*models.User, error) {
	users := make([]*models.User, 0, count)

	// Create password hash once for all users (for performance)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("benchmark123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password hash: %w", err)
	}

	for i := 0; i < count; i++ {
		user := &models.User{
			ID:           uuid.New(),
			Username:     fmt.Sprintf("user%d", i+1),
			Email:        fmt.Sprintf("user%d@benchmark.test", i+1),
			PasswordHash: string(passwordHash),
			Role:         dg.randomUserRole(),
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		users = append(users, user)
	}

	// Bulk insert users
	if err := dg.DB.CreateInBatches(users, 100).Error; err != nil {
		return nil, fmt.Errorf("failed to create users: %w", err)
	}

	return users, nil
}

// CreateEpics generates the specified number of epics
func (dg *DataGenerator) CreateEpics(count int, users []*models.User) ([]*models.Epic, error) {
	if len(users) == 0 {
		return nil, fmt.Errorf("no users provided for epic creation")
	}

	epics := make([]*models.Epic, 0, count)

	for i := 0; i < count; i++ {
		creator := users[dg.rand.Intn(len(users))]
		assignee := users[dg.rand.Intn(len(users))]

		epic := &models.Epic{
			ID:          uuid.New(),
			ReferenceID: fmt.Sprintf("EP-%03d", i+1),
			CreatorID:   creator.ID,
			AssigneeID:  assignee.ID,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Priority:    dg.randomPriority(),
			Status:      dg.randomEpicStatus(),
			Title:       fmt.Sprintf("Epic %d: %s", i+1, dg.randomEpicTitle()),
			Description: dg.stringPtr(dg.randomEpicDescription()),
		}
		epics = append(epics, epic)
	}

	// Bulk insert epics
	if err := dg.DB.CreateInBatches(epics, 100).Error; err != nil {
		return nil, fmt.Errorf("failed to create epics: %w", err)
	}

	return epics, nil
}

// CreateUserStories generates user stories for the provided epics
func (dg *DataGenerator) CreateUserStories(storiesPerEpic int, epics []*models.Epic, users []*models.User) ([]*models.UserStory, error) {
	if len(epics) == 0 || len(users) == 0 {
		return nil, fmt.Errorf("no epics or users provided for user story creation")
	}

	userStories := make([]*models.UserStory, 0, len(epics)*storiesPerEpic)
	storyCounter := 1

	for _, epic := range epics {
		for j := 0; j < storiesPerEpic; j++ {
			creator := users[dg.rand.Intn(len(users))]
			assignee := users[dg.rand.Intn(len(users))]

			userStory := &models.UserStory{
				ID:          uuid.New(),
				ReferenceID: fmt.Sprintf("US-%03d", storyCounter),
				EpicID:      epic.ID,
				CreatorID:   creator.ID,
				AssigneeID:  assignee.ID,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Priority:    dg.randomPriority(),
				Status:      dg.randomUserStoryStatus(),
				Title:       fmt.Sprintf("User Story %d: %s", storyCounter, dg.randomUserStoryTitle()),
				Description: dg.stringPtr(dg.randomUserStoryDescription()),
			}
			userStories = append(userStories, userStory)
			storyCounter++
		}
	}

	// Bulk insert user stories
	if err := dg.DB.CreateInBatches(userStories, 100).Error; err != nil {
		return nil, fmt.Errorf("failed to create user stories: %w", err)
	}

	return userStories, nil
}

// CreateRequirements generates requirements for the provided user stories
func (dg *DataGenerator) CreateRequirements(requirementsPerUS int, userStories []*models.UserStory, users []*models.User) ([]*models.Requirement, error) {
	if len(userStories) == 0 || len(users) == 0 {
		return nil, fmt.Errorf("no user stories or users provided for requirement creation")
	}

	// Get requirement types
	var reqTypes []models.RequirementType
	if err := dg.DB.Find(&reqTypes).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch requirement types: %w", err)
	}
	if len(reqTypes) == 0 {
		return nil, fmt.Errorf("no requirement types found in database")
	}

	requirements := make([]*models.Requirement, 0, len(userStories)*requirementsPerUS)
	reqCounter := 1

	for _, userStory := range userStories {
		for j := 0; j < requirementsPerUS; j++ {
			creator := users[dg.rand.Intn(len(users))]
			assignee := users[dg.rand.Intn(len(users))]
			reqType := reqTypes[dg.rand.Intn(len(reqTypes))]

			requirement := &models.Requirement{
				ID:          uuid.New(),
				ReferenceID: fmt.Sprintf("REQ-%03d", reqCounter),
				UserStoryID: userStory.ID,
				CreatorID:   creator.ID,
				AssigneeID:  assignee.ID,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Priority:    dg.randomPriority(),
				Status:      dg.randomRequirementStatus(),
				TypeID:      reqType.ID,
				Title:       fmt.Sprintf("Requirement %d: %s", reqCounter, dg.randomRequirementTitle()),
				Description: dg.stringPtr(dg.randomRequirementDescription()),
			}
			requirements = append(requirements, requirement)
			reqCounter++
		}
	}

	// Bulk insert requirements
	if err := dg.DB.CreateInBatches(requirements, 100).Error; err != nil {
		return nil, fmt.Errorf("failed to create requirements: %w", err)
	}

	return requirements, nil
}

// CreateAcceptanceCriteria generates acceptance criteria for the provided user stories
func (dg *DataGenerator) CreateAcceptanceCriteria(count int, userStories []*models.UserStory, users []*models.User) ([]*models.AcceptanceCriteria, error) {
	if len(userStories) == 0 || len(users) == 0 {
		return nil, fmt.Errorf("no user stories or users provided for acceptance criteria creation")
	}

	acceptanceCriteria := make([]*models.AcceptanceCriteria, 0, count)

	for i := 0; i < count; i++ {
		userStory := userStories[dg.rand.Intn(len(userStories))]
		author := users[dg.rand.Intn(len(users))]

		ac := &models.AcceptanceCriteria{
			ID:          uuid.New(),
			ReferenceID: fmt.Sprintf("AC-%03d", i+1),
			UserStoryID: userStory.ID,
			AuthorID:    author.ID,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Description: dg.randomAcceptanceCriteriaDescription(),
		}
		acceptanceCriteria = append(acceptanceCriteria, ac)
	}

	// Bulk insert acceptance criteria
	if err := dg.DB.CreateInBatches(acceptanceCriteria, 100).Error; err != nil {
		return nil, fmt.Errorf("failed to create acceptance criteria: %w", err)
	}

	return acceptanceCriteria, nil
}

// CreateComments generates comments for various entities
func (dg *DataGenerator) CreateComments(count int, users []*models.User, epics []*models.Epic, userStories []*models.UserStory, requirements []*models.Requirement, acceptanceCriteria []*models.AcceptanceCriteria) ([]*models.Comment, error) {
	if len(users) == 0 {
		return nil, fmt.Errorf("no users provided for comment creation")
	}

	// Collect all entities that can be commented on
	var entities []struct {
		ID   uuid.UUID
		Type models.EntityType
	}

	for _, epic := range epics {
		entities = append(entities, struct {
			ID   uuid.UUID
			Type models.EntityType
		}{epic.ID, models.EntityTypeEpic})
	}

	for _, us := range userStories {
		entities = append(entities, struct {
			ID   uuid.UUID
			Type models.EntityType
		}{us.ID, models.EntityTypeUserStory})
	}

	for _, req := range requirements {
		entities = append(entities, struct {
			ID   uuid.UUID
			Type models.EntityType
		}{req.ID, models.EntityTypeRequirement})
	}

	for _, ac := range acceptanceCriteria {
		entities = append(entities, struct {
			ID   uuid.UUID
			Type models.EntityType
		}{ac.ID, models.EntityTypeAcceptanceCriteria})
	}

	if len(entities) == 0 {
		return nil, fmt.Errorf("no entities available for commenting")
	}

	comments := make([]*models.Comment, 0, count)

	for i := 0; i < count; i++ {
		author := users[dg.rand.Intn(len(users))]
		entity := entities[dg.rand.Intn(len(entities))]

		comment := &models.Comment{
			ID:         uuid.New(),
			EntityType: entity.Type,
			EntityID:   entity.ID,
			AuthorID:   author.ID,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
			Content:    dg.randomCommentContent(),
			IsResolved: dg.rand.Float32() < 0.3, // 30% chance of being resolved
		}

		// 20% chance of being an inline comment
		if dg.rand.Float32() < 0.2 {
			comment.LinkedText = dg.stringPtr("sample linked text")
			start := dg.rand.Intn(100)
			end := start + dg.rand.Intn(50) + 1
			comment.TextPositionStart = &start
			comment.TextPositionEnd = &end
		}

		comments = append(comments, comment)
	}

	// Bulk insert comments
	if err := dg.DB.CreateInBatches(comments, 100).Error; err != nil {
		return nil, fmt.Errorf("failed to create comments: %w", err)
	}

	return comments, nil
}

// GenerateFullDataSet creates a complete dataset with all entity types
func (dg *DataGenerator) GenerateFullDataSet(config DataSetConfig) error {
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
	userStories, err := dg.CreateUserStories(config.UserStoriesPerEpic, epics, users)
	if err != nil {
		return fmt.Errorf("failed to create user stories: %w", err)
	}

	// Create requirements
	requirements, err := dg.CreateRequirements(config.RequirementsPerUS, userStories, users)
	if err != nil {
		return fmt.Errorf("failed to create requirements: %w", err)
	}

	// Create acceptance criteria
	acceptanceCriteria, err := dg.CreateAcceptanceCriteria(config.AcceptanceCriteria, userStories, users)
	if err != nil {
		return fmt.Errorf("failed to create acceptance criteria: %w", err)
	}

	// Create comments
	_, err = dg.CreateComments(config.Comments, users, epics, userStories, requirements, acceptanceCriteria)
	if err != nil {
		return fmt.Errorf("failed to create comments: %w", err)
	}

	return nil
}

// CleanupDatabase removes all test data from the database
func (dg *DataGenerator) CleanupDatabase() error {
	// Delete in reverse order of dependencies
	tables := []string{
		"comments",
		"requirement_relationships",
		"requirements",
		"acceptance_criteria",
		"user_stories",
		"epics",
		"users",
	}

	for _, table := range tables {
		// Check if table exists before attempting to delete
		var exists bool
		query := fmt.Sprintf("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = '%s')", table)
		if err := dg.DB.Raw(query).Scan(&exists).Error; err != nil {
			return fmt.Errorf("failed to check if table %s exists: %w", table, err)
		}

		if exists {
			if err := dg.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
				return fmt.Errorf("failed to cleanup table %s: %w", table, err)
			}
		}
	}

	return nil
}

// ResetDatabase drops and recreates all tables, then seeds default data
func (dg *DataGenerator) ResetDatabase() error {
	// Drop all tables
	if err := dg.DB.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;").Error; err != nil {
		return fmt.Errorf("failed to reset database schema: %w", err)
	}

	// Re-run migrations
	if err := models.AutoMigrate(dg.DB); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed default data
	if err := models.SeedDefaultData(dg.DB); err != nil {
		return fmt.Errorf("failed to seed default data: %w", err)
	}

	// Verify tables were created successfully
	requiredTables := []string{"users", "epics", "user_stories", "requirements", "acceptance_criteria"}
	for _, table := range requiredTables {
		var exists bool
		query := fmt.Sprintf("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = '%s')", table)
		if err := dg.DB.Raw(query).Scan(&exists).Error; err != nil {
			return fmt.Errorf("failed to verify table %s exists after reset: %w", table, err)
		}
		if !exists {
			return fmt.Errorf("table %s was not created after database reset", table)
		}
	}

	return nil
}

// Helper functions for generating random data

func (dg *DataGenerator) randomUserRole() models.UserRole {
	roles := []models.UserRole{
		models.RoleAdministrator,
		models.RoleUser,
		models.RoleCommenter,
	}
	return roles[dg.rand.Intn(len(roles))]
}

func (dg *DataGenerator) randomPriority() models.Priority {
	priorities := []models.Priority{
		models.PriorityCritical,
		models.PriorityHigh,
		models.PriorityMedium,
		models.PriorityLow,
	}
	return priorities[dg.rand.Intn(len(priorities))]
}

func (dg *DataGenerator) randomEpicStatus() models.EpicStatus {
	statuses := []models.EpicStatus{
		models.EpicStatusBacklog,
		models.EpicStatusDraft,
		models.EpicStatusInProgress,
		models.EpicStatusDone,
		models.EpicStatusCancelled,
	}
	return statuses[dg.rand.Intn(len(statuses))]
}

func (dg *DataGenerator) randomUserStoryStatus() models.UserStoryStatus {
	statuses := []models.UserStoryStatus{
		models.UserStoryStatusBacklog,
		models.UserStoryStatusDraft,
		models.UserStoryStatusInProgress,
		models.UserStoryStatusDone,
		models.UserStoryStatusCancelled,
	}
	return statuses[dg.rand.Intn(len(statuses))]
}

func (dg *DataGenerator) randomRequirementStatus() models.RequirementStatus {
	statuses := []models.RequirementStatus{
		models.RequirementStatusDraft,
		models.RequirementStatusActive,
		models.RequirementStatusObsolete,
	}
	return statuses[dg.rand.Intn(len(statuses))]
}

func (dg *DataGenerator) randomEpicTitle() string {
	titles := []string{
		"User Authentication System",
		"Payment Processing Module",
		"Reporting Dashboard",
		"Mobile Application Support",
		"API Integration Layer",
		"Data Analytics Platform",
		"Security Enhancement",
		"Performance Optimization",
		"User Interface Redesign",
		"Third-party Integration",
	}
	return titles[dg.rand.Intn(len(titles))]
}

func (dg *DataGenerator) randomEpicDescription() string {
	descriptions := []string{
		"This epic focuses on implementing a comprehensive user authentication system with multi-factor authentication support.",
		"Development of a secure and scalable payment processing module that supports multiple payment gateways.",
		"Creation of an interactive reporting dashboard with real-time data visualization capabilities.",
		"Implementation of mobile application support with responsive design and offline capabilities.",
		"Building a robust API integration layer to connect with external services and third-party systems.",
		"Development of a data analytics platform for business intelligence and reporting needs.",
		"Enhancement of system security with advanced threat detection and prevention mechanisms.",
		"Optimization of system performance through caching, database tuning, and code improvements.",
		"Complete redesign of the user interface for better user experience and accessibility.",
		"Integration with third-party services to extend system functionality and capabilities.",
	}
	return descriptions[dg.rand.Intn(len(descriptions))]
}

func (dg *DataGenerator) randomUserStoryTitle() string {
	titles := []string{
		"Login with Email and Password",
		"Reset Password Functionality",
		"View Transaction History",
		"Generate Monthly Reports",
		"Upload Profile Picture",
		"Search and Filter Data",
		"Export Data to CSV",
		"Receive Email Notifications",
		"Manage User Permissions",
		"Configure System Settings",
	}
	return titles[dg.rand.Intn(len(titles))]
}

func (dg *DataGenerator) randomUserStoryDescription() string {
	descriptions := []string{
		"As a user, I want to login with my email and password, so that I can access my account securely.",
		"As a user, I want to reset my password, so that I can regain access to my account if I forget it.",
		"As a customer, I want to view my transaction history, so that I can track my purchases and payments.",
		"As a manager, I want to generate monthly reports, so that I can analyze business performance.",
		"As a user, I want to upload a profile picture, so that I can personalize my account.",
		"As a user, I want to search and filter data, so that I can find relevant information quickly.",
		"As an analyst, I want to export data to CSV, so that I can perform external analysis.",
		"As a user, I want to receive email notifications, so that I stay informed about important updates.",
		"As an administrator, I want to manage user permissions, so that I can control access to system features.",
		"As an administrator, I want to configure system settings, so that I can customize the application behavior.",
	}
	return descriptions[dg.rand.Intn(len(descriptions))]
}

func (dg *DataGenerator) randomRequirementTitle() string {
	titles := []string{
		"Password Complexity Validation",
		"Session Timeout Configuration",
		"Data Encryption Requirements",
		"API Rate Limiting",
		"Audit Log Generation",
		"Input Validation Rules",
		"Error Handling Standards",
		"Performance Benchmarks",
		"Accessibility Compliance",
		"Browser Compatibility",
	}
	return titles[dg.rand.Intn(len(titles))]
}

func (dg *DataGenerator) randomRequirementDescription() string {
	descriptions := []string{
		"The system shall enforce password complexity requirements including minimum length, special characters, and mixed case.",
		"The system shall automatically log out users after 30 minutes of inactivity to ensure security.",
		"All sensitive data shall be encrypted using AES-256 encryption both in transit and at rest.",
		"The API shall implement rate limiting to prevent abuse and ensure fair usage across all clients.",
		"The system shall generate comprehensive audit logs for all user actions and system events.",
		"All user inputs shall be validated and sanitized to prevent injection attacks and data corruption.",
		"The system shall implement standardized error handling with appropriate user-friendly messages.",
		"All API endpoints shall respond within 200ms for 95% of requests under normal load conditions.",
		"The application shall comply with WCAG 2.1 AA accessibility standards for all user interfaces.",
		"The application shall be compatible with the latest versions of Chrome, Firefox, Safari, and Edge browsers.",
	}
	return descriptions[dg.rand.Intn(len(descriptions))]
}

func (dg *DataGenerator) randomAcceptanceCriteriaDescription() string {
	descriptions := []string{
		"WHEN a user enters an invalid password THEN the system SHALL display a specific error message about password requirements.",
		"WHEN a user is inactive for 30 minutes THEN the system SHALL automatically log them out and redirect to the login page.",
		"WHEN sensitive data is transmitted THEN the system SHALL use HTTPS with TLS 1.2 or higher encryption.",
		"WHEN an API client exceeds the rate limit THEN the system SHALL return a 429 status code with retry-after header.",
		"WHEN a user performs any action THEN the system SHALL log the action with timestamp, user ID, and action details.",
		"WHEN a user submits a form THEN the system SHALL validate all inputs before processing the request.",
		"WHEN an error occurs THEN the system SHALL display a user-friendly message and log technical details separately.",
		"WHEN an API request is made THEN the system SHALL respond within the specified performance thresholds.",
		"WHEN a user navigates the interface THEN all elements SHALL be accessible via keyboard and screen readers.",
		"WHEN the application loads THEN it SHALL function correctly across all supported browser versions.",
	}
	return descriptions[dg.rand.Intn(len(descriptions))]
}

func (dg *DataGenerator) randomCommentContent() string {
	comments := []string{
		"This looks good to me. Ready for implementation.",
		"I have some concerns about the performance implications of this approach.",
		"Could we add more detail about the error handling scenarios?",
		"This requirement needs clarification on the business rules.",
		"Great work! This covers all the necessary acceptance criteria.",
		"I suggest we break this down into smaller, more manageable requirements.",
		"The implementation approach seems solid. Let's proceed.",
		"We should consider the impact on existing functionality.",
		"This aligns well with our architectural guidelines.",
		"I recommend adding security considerations to this requirement.",
	}
	return comments[dg.rand.Intn(len(comments))]
}

func (dg *DataGenerator) stringPtr(s string) *string {
	return &s
}
