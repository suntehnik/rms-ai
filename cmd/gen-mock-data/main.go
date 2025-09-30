package main

import (
	"fmt"
	"log"
	"math/rand"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/models"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("Starting mock data generation...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Seed default data (skip auto-migration as tables already exist)
	if err := models.SeedDefaultData(db); err != nil {
		log.Fatalf("Failed to seed default data: %v", err)
	}

	// Initialize random seed (using new approach for Go 1.20+)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	generator := &MockDataGenerator{db: db, rng: rng}

	// Generate mock data
	if err := generator.GenerateAll(); err != nil {
		log.Fatalf("Failed to generate mock data: %v", err)
	}

	fmt.Println("Mock data generation completed successfully!")
}

type MockDataGenerator struct {
	db                *gorm.DB
	rng               *rand.Rand
	users             []models.User
	requirementTypes  []models.RequirementType
	relationshipTypes []models.RelationshipType
}

func (g *MockDataGenerator) GenerateAll() error {
	fmt.Println("Creating users...")
	if err := g.createUsers(); err != nil {
		return fmt.Errorf("failed to create users: %w", err)
	}

	fmt.Println("Loading requirement types...")
	if err := g.loadRequirementTypes(); err != nil {
		return fmt.Errorf("failed to load requirement types: %w", err)
	}

	fmt.Println("Loading relationship types...")
	if err := g.loadRelationshipTypes(); err != nil {
		return fmt.Errorf("failed to load relationship types: %w", err)
	}

	fmt.Println("Creating epics with user stories and requirements...")
	if err := g.createEpicsWithContent(); err != nil {
		return fmt.Errorf("failed to create epics: %w", err)
	}

	return nil
}

// createUsers creates mock users (skips existing ones)
func (g *MockDataGenerator) createUsers() error {
	// First, load existing users
	var existingUsers []models.User
	if err := g.db.Find(&existingUsers).Error; err != nil {
		return fmt.Errorf("failed to load existing users: %w", err)
	}

	// Create a map of existing usernames for quick lookup
	existingUsernames := make(map[string]bool)
	for _, user := range existingUsers {
		existingUsernames[user.Username] = true
		g.users = append(g.users, user) // Add existing users to our list
	}
	userNames := []string{
		"admin", "john_doe", "jane_smith", "bob_wilson", "alice_johnson",
		"mike_brown", "sarah_davis", "tom_miller", "lisa_garcia", "david_martinez",
		"emma_rodriguez", "james_hernandez", "olivia_lopez", "william_gonzalez", "sophia_wilson",
		"benjamin_anderson", "isabella_thomas", "lucas_taylor", "mia_moore", "henry_jackson",
	}

	emails := []string{
		"admin@example.com", "john.doe@example.com", "jane.smith@example.com", "bob.wilson@example.com", "alice.johnson@example.com",
		"mike.brown@example.com", "sarah.davis@example.com", "tom.miller@example.com", "lisa.garcia@example.com", "david.martinez@example.com",
		"emma.rodriguez@example.com", "james.hernandez@example.com", "olivia.lopez@example.com", "william.gonzalez@example.com", "sophia.wilson@example.com",
		"benjamin.anderson@example.com", "isabella.thomas@example.com", "lucas.taylor@example.com", "mia.moore@example.com", "henry.jackson@example.com",
	}

	roles := []models.UserRole{
		models.RoleAdministrator, models.RoleUser, models.RoleUser, models.RoleUser, models.RoleUser,
		models.RoleUser, models.RoleUser, models.RoleCommenter, models.RoleUser, models.RoleUser,
		models.RoleUser, models.RoleCommenter, models.RoleUser, models.RoleUser, models.RoleUser,
		models.RoleUser, models.RoleUser, models.RoleUser, models.RoleCommenter, models.RoleUser,
	}

	// Hash password for all users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	for i, username := range userNames {
		// Skip if user already exists
		if existingUsernames[username] {
			fmt.Printf("User %s already exists, skipping...\n", username)
			continue
		}

		user := models.User{
			ID:           uuid.New(),
			Username:     username,
			Email:        emails[i],
			PasswordHash: string(hashedPassword),
			Role:         roles[i],
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		if err := g.db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", username, err)
		}

		g.users = append(g.users, user)
	}

	fmt.Printf("Created %d users\n", len(g.users))
	return nil
}

// loadRequirementTypes loads existing requirement types
func (g *MockDataGenerator) loadRequirementTypes() error {
	if err := g.db.Find(&g.requirementTypes).Error; err != nil {
		return fmt.Errorf("failed to load requirement types: %w", err)
	}
	fmt.Printf("Loaded %d requirement types\n", len(g.requirementTypes))
	return nil
}

// loadRelationshipTypes loads existing relationship types
func (g *MockDataGenerator) loadRelationshipTypes() error {
	if err := g.db.Find(&g.relationshipTypes).Error; err != nil {
		return fmt.Errorf("failed to load relationship types: %w", err)
	}
	fmt.Printf("Loaded %d relationship types\n", len(g.relationshipTypes))
	return nil
}

// createEpicsWithContent creates 200 epics with user stories and requirements
func (g *MockDataGenerator) createEpicsWithContent() error {
	epicTitles := g.generateEpicTitles()

	for i, title := range epicTitles {
		fmt.Printf("Creating epic %d/%d: %s\n", i+1, len(epicTitles), title)

		epic, err := g.createEpic(title)
		if err != nil {
			return fmt.Errorf("failed to create epic %s: %w", title, err)
		}

		// Create 3-7 user stories per epic
		numUserStories := g.rng.Intn(5) + 3
		for j := 0; j < numUserStories; j++ {
			userStory, err := g.createUserStory(epic.ID, title)
			if err != nil {
				return fmt.Errorf("failed to create user story for epic %s: %w", title, err)
			}

			// Create 2-5 acceptance criteria per user story
			numAcceptanceCriteria := g.rng.Intn(4) + 2
			for k := 0; k < numAcceptanceCriteria; k++ {
				if _, err := g.createAcceptanceCriteria(userStory.ID); err != nil {
					return fmt.Errorf("failed to create acceptance criteria: %w", err)
				}
			}

			// Create 3-8 requirements per user story
			numRequirements := g.rng.Intn(6) + 3
			for k := 0; k < numRequirements; k++ {
				requirement, err := g.createRequirement(userStory.ID)
				if err != nil {
					return fmt.Errorf("failed to create requirement: %w", err)
				}

				// Create 1-3 comments per requirement
				numComments := g.rng.Intn(3) + 1
				for l := 0; l < numComments; l++ {
					if err := g.createComment(models.EntityTypeRequirement, requirement.ID); err != nil {
						return fmt.Errorf("failed to create comment: %w", err)
					}
				}
			}

			// Create 1-2 comments per user story
			numComments := g.rng.Intn(2) + 1
			for k := 0; k < numComments; k++ {
				if err := g.createComment(models.EntityTypeUserStory, userStory.ID); err != nil {
					return fmt.Errorf("failed to create comment: %w", err)
				}
			}
		}

		// Create 1-2 comments per epic
		numComments := g.rng.Intn(2) + 1
		for j := 0; j < numComments; j++ {
			if err := g.createComment(models.EntityTypeEpic, epic.ID); err != nil {
				return fmt.Errorf("failed to create comment: %w", err)
			}
		}
	}

	return nil
}

// generateEpicTitles generates 200 meaningful epic titles
func (g *MockDataGenerator) generateEpicTitles() []string {
	domains := []string{
		"User Authentication", "Payment Processing", "Data Analytics", "Mobile App", "API Gateway",
		"Content Management", "Search Engine", "Notification System", "File Storage", "User Profile",
		"Shopping Cart", "Order Management", "Inventory System", "Customer Support", "Reporting Dashboard",
		"Security Framework", "Integration Platform", "Workflow Engine", "Document Management", "Communication Hub",
		"Performance Monitoring", "Backup System", "Configuration Management", "Audit Trail", "Multi-tenant Architecture",
		"Real-time Messaging", "Video Streaming", "Image Processing", "Machine Learning", "Data Warehouse",
		"Social Features", "Recommendation Engine", "Geolocation Services", "Calendar Integration", "Email Marketing",
		"A/B Testing", "Feature Flags", "Load Balancing", "Caching Layer", "Database Optimization",
	}

	features := []string{
		"Enhancement", "Redesign", "Migration", "Integration", "Optimization", "Modernization", "Expansion",
		"Refactoring", "Implementation", "Upgrade", "Automation", "Standardization", "Consolidation",
		"Personalization", "Localization", "Scalability", "Security", "Performance", "Accessibility", "Compliance",
	}

	versions := []string{
		"v2.0", "v3.0", "Next Gen", "Advanced", "Pro", "Enterprise", "Cloud", "Mobile", "Web", "API",
	}

	var titles []string
	for len(titles) < 200 {
		domain := domains[g.rng.Intn(len(domains))]
		feature := features[g.rng.Intn(len(features))]

		var title string
		if g.rng.Float32() < 0.3 { // 30% chance to include version
			version := versions[g.rng.Intn(len(versions))]
			title = fmt.Sprintf("%s %s %s", domain, feature, version)
		} else {
			title = fmt.Sprintf("%s %s", domain, feature)
		}

		// Avoid duplicates
		duplicate := false
		for _, existing := range titles {
			if existing == title {
				duplicate = true
				break
			}
		}

		if !duplicate {
			titles = append(titles, title)
		}
	}

	return titles
}

// createEpic creates a single epic
func (g *MockDataGenerator) createEpic(title string) (*models.Epic, error) {
	creator := g.getRandomUser()
	assignee := g.getRandomUser()

	description := g.generateEpicDescription(title)

	epic := &models.Epic{
		ID:          uuid.New(),
		CreatorID:   creator.ID,
		AssigneeID:  assignee.ID,
		CreatedAt:   g.getRandomPastTime(),
		UpdatedAt:   time.Now().UTC(),
		Priority:    g.getRandomPriority(),
		Status:      g.getRandomEpicStatus(),
		Title:       title,
		Description: &description,
	}

	if err := g.db.Create(epic).Error; err != nil {
		return nil, err
	}

	return epic, nil
}

// createUserStory creates a user story for an epic
func (g *MockDataGenerator) createUserStory(epicID uuid.UUID, epicTitle string) (*models.UserStory, error) {
	creator := g.getRandomUser()
	assignee := g.getRandomUser()

	title := g.generateUserStoryTitle(epicTitle)
	description := g.generateUserStoryDescription(title)

	userStory := &models.UserStory{
		ID:          uuid.New(),
		EpicID:      epicID,
		CreatorID:   creator.ID,
		AssigneeID:  assignee.ID,
		CreatedAt:   g.getRandomPastTime(),
		UpdatedAt:   time.Now().UTC(),
		Priority:    g.getRandomPriority(),
		Status:      g.getRandomUserStoryStatus(),
		Title:       title,
		Description: &description,
	}

	if err := g.db.Create(userStory).Error; err != nil {
		return nil, err
	}

	return userStory, nil
}

// createAcceptanceCriteria creates acceptance criteria for a user story
func (g *MockDataGenerator) createAcceptanceCriteria(userStoryID uuid.UUID) (*models.AcceptanceCriteria, error) {
	author := g.getRandomUser()
	description := g.generateAcceptanceCriteriaDescription()

	ac := &models.AcceptanceCriteria{
		ID:          uuid.New(),
		UserStoryID: userStoryID,
		AuthorID:    author.ID,
		CreatedAt:   g.getRandomPastTime(),
		Description: description,
	}

	if err := g.db.Create(ac).Error; err != nil {
		return nil, err
	}

	return ac, nil
}

// createRequirement creates a requirement for a user story
func (g *MockDataGenerator) createRequirement(userStoryID uuid.UUID) (*models.Requirement, error) {
	creator := g.getRandomUser()
	assignee := g.getRandomUser()
	reqType := g.getRandomRequirementType()

	title := g.generateRequirementTitle()
	description := g.generateRequirementDescription(title)

	requirement := &models.Requirement{
		ID:          uuid.New(),
		UserStoryID: userStoryID,
		CreatorID:   creator.ID,
		AssigneeID:  assignee.ID,
		CreatedAt:   g.getRandomPastTime(),
		UpdatedAt:   time.Now().UTC(),
		Priority:    g.getRandomPriority(),
		Status:      g.getRandomRequirementStatus(),
		TypeID:      reqType.ID,
		Title:       title,
		Description: &description,
	}

	if err := g.db.Create(requirement).Error; err != nil {
		return nil, err
	}

	return requirement, nil
}

// createComment creates a comment for an entity
func (g *MockDataGenerator) createComment(entityType models.EntityType, entityID uuid.UUID) error {
	author := g.getRandomUser()
	content := g.generateCommentContent()

	comment := &models.Comment{
		ID:         uuid.New(),
		EntityType: entityType,
		EntityID:   entityID,
		AuthorID:   author.ID,
		CreatedAt:  g.getRandomPastTime(),
		UpdatedAt:  time.Now().UTC(),
		Content:    content,
		IsResolved: g.rng.Float32() < 0.3, // 30% chance to be resolved
	}

	if err := g.db.Create(comment).Error; err != nil {
		return err
	}

	return nil
}

// Helper methods for generating content

func (g *MockDataGenerator) getRandomUser() models.User {
	return g.users[g.rng.Intn(len(g.users))]
}

func (g *MockDataGenerator) getRandomRequirementType() models.RequirementType {
	return g.requirementTypes[g.rng.Intn(len(g.requirementTypes))]
}

func (g *MockDataGenerator) getRandomPriority() models.Priority {
	priorities := []models.Priority{
		models.PriorityCritical, models.PriorityHigh, models.PriorityMedium, models.PriorityLow,
	}
	return priorities[g.rng.Intn(len(priorities))]
}

func (g *MockDataGenerator) getRandomEpicStatus() models.EpicStatus {
	statuses := []models.EpicStatus{
		models.EpicStatusBacklog, models.EpicStatusDraft, models.EpicStatusInProgress,
		models.EpicStatusDone, models.EpicStatusCancelled,
	}
	return statuses[g.rng.Intn(len(statuses))]
}

func (g *MockDataGenerator) getRandomUserStoryStatus() models.UserStoryStatus {
	statuses := []models.UserStoryStatus{
		models.UserStoryStatusBacklog, models.UserStoryStatusDraft, models.UserStoryStatusInProgress,
		models.UserStoryStatusDone, models.UserStoryStatusCancelled,
	}
	return statuses[g.rng.Intn(len(statuses))]
}

func (g *MockDataGenerator) getRandomRequirementStatus() models.RequirementStatus {
	statuses := []models.RequirementStatus{
		models.RequirementStatusDraft, models.RequirementStatusActive, models.RequirementStatusObsolete,
	}
	return statuses[g.rng.Intn(len(statuses))]
}

func (g *MockDataGenerator) getRandomPastTime() time.Time {
	// Random time in the past 90 days
	days := g.rng.Intn(90)
	hours := g.rng.Intn(24)
	minutes := g.rng.Intn(60)
	return time.Now().UTC().AddDate(0, 0, -days).Add(-time.Duration(hours)*time.Hour - time.Duration(minutes)*time.Minute)
}

func (g *MockDataGenerator) generateEpicDescription(title string) string {
	templates := []string{
		"This epic focuses on %s to improve user experience and system performance. The implementation will involve multiple phases including analysis, design, development, and testing.",
		"The %s initiative aims to modernize our platform and provide enhanced capabilities for our users. This comprehensive effort will span several months and require coordination across multiple teams.",
		"As part of our strategic roadmap, the %s epic will deliver significant value to our customers by introducing new features and improving existing functionality.",
		"The %s project represents a major milestone in our product evolution, incorporating user feedback and market research to deliver a superior solution.",
		"This %s epic encompasses the development of critical features that will position our product as a market leader while ensuring scalability and maintainability.",
	}

	template := templates[g.rng.Intn(len(templates))]
	return fmt.Sprintf(template, title)
}

func (g *MockDataGenerator) generateUserStoryTitle(epicTitle string) string {
	actions := []string{
		"Login", "Register", "View", "Edit", "Delete", "Create", "Search", "Filter", "Sort", "Export",
		"Import", "Configure", "Manage", "Monitor", "Analyze", "Report", "Integrate", "Sync", "Backup", "Restore",
	}

	objects := []string{
		"Profile", "Dashboard", "Settings", "Data", "Reports", "Notifications", "Messages", "Files", "Documents",
		"Accounts", "Permissions", "Logs", "Metrics", "Analytics", "Workflows", "Templates", "Categories", "Tags",
	}

	action := actions[g.rng.Intn(len(actions))]
	object := objects[g.rng.Intn(len(objects))]

	return fmt.Sprintf("%s %s", action, object)
}

func (g *MockDataGenerator) generateUserStoryDescription(title string) string {
	roles := []string{
		"registered user", "administrator", "manager", "analyst", "customer", "developer", "tester", "support agent",
	}

	goals := []string{
		"improve productivity", "save time", "reduce errors", "enhance security", "increase efficiency",
		"better understand data", "make informed decisions", "streamline workflows", "improve user experience",
		"ensure compliance", "maintain data integrity", "optimize performance",
	}

	role := roles[g.rng.Intn(len(roles))]
	goal := goals[g.rng.Intn(len(goals))]

	return fmt.Sprintf("As a %s, I want to %s, so that I can %s.", role, title, goal)
}

func (g *MockDataGenerator) generateAcceptanceCriteriaDescription() string {
	conditions := []string{
		"user enters valid credentials",
		"user clicks the submit button",
		"system validates the input",
		"data is successfully saved",
		"user has appropriate permissions",
		"network connection is available",
		"required fields are completed",
		"file format is supported",
	}

	actions := []string{
		"system SHALL authenticate the user",
		"system SHALL display a success message",
		"system SHALL redirect to the dashboard",
		"system SHALL save the data",
		"system SHALL send a notification",
		"system SHALL update the display",
		"system SHALL log the action",
		"system SHALL validate the input",
	}

	condition := conditions[g.rng.Intn(len(conditions))]
	action := actions[g.rng.Intn(len(actions))]

	return fmt.Sprintf("WHEN %s THEN %s", condition, action)
}

func (g *MockDataGenerator) generateRequirementTitle() string {
	subjects := []string{
		"Authentication system", "Data validation", "User interface", "API endpoint", "Database schema",
		"Security protocol", "Performance optimization", "Error handling", "Logging mechanism", "Configuration management",
		"Integration service", "Notification system", "File processing", "Data synchronization", "Backup procedure",
	}

	actions := []string{
		"must support", "shall implement", "should provide", "must validate", "shall ensure",
		"must handle", "should optimize", "shall maintain", "must comply with", "should integrate with",
	}

	requirements := []string{
		"OAuth 2.0 authentication", "input sanitization", "responsive design", "RESTful standards", "ACID compliance",
		"encryption at rest", "sub-second response times", "graceful error recovery", "structured logging", "environment-based configuration",
		"third-party APIs", "real-time updates", "multiple file formats", "data consistency", "automated backups",
	}

	subject := subjects[g.rng.Intn(len(subjects))]
	action := actions[g.rng.Intn(len(actions))]
	requirement := requirements[g.rng.Intn(len(requirements))]

	return fmt.Sprintf("%s %s %s", subject, action, requirement)
}

func (g *MockDataGenerator) generateRequirementDescription(title string) string {
	templates := []string{
		"The %s requirement ensures that the system meets industry standards and provides a secure, reliable experience for all users. Implementation must follow established best practices and include comprehensive testing.",
		"This requirement for %s is critical for maintaining system integrity and user trust. The implementation should be thoroughly documented and include appropriate error handling and logging.",
		"The %s specification defines the technical approach and constraints necessary for successful implementation. All edge cases must be considered and handled appropriately.",
		"This %s requirement addresses both functional and non-functional aspects of the system, ensuring scalability, maintainability, and performance under various load conditions.",
		"The implementation of %s must consider security implications, performance impact, and integration requirements with existing system components.",
	}

	template := templates[g.rng.Intn(len(templates))]
	return fmt.Sprintf(template, title)
}

func (g *MockDataGenerator) generateCommentContent() string {
	comments := []string{
		"This looks good to me. I think we should proceed with this approach.",
		"I have some concerns about the performance implications. Can we discuss alternatives?",
		"Great work! This addresses all the requirements we discussed.",
		"We need to consider the security aspects more carefully before implementation.",
		"This requirement might conflict with the existing architecture. Let's review.",
		"I suggest we add more detailed acceptance criteria for this user story.",
		"The implementation timeline seems aggressive. Can we break this down further?",
		"This aligns well with our strategic objectives. I approve this direction.",
		"We should involve the security team in reviewing this requirement.",
		"The user experience could be improved with some additional considerations.",
		"This requirement is well-defined and ready for implementation.",
		"I recommend adding error handling scenarios to this specification.",
		"The business value is clear, but we need to validate the technical feasibility.",
		"This change might impact other components. Let's do an impact analysis.",
		"The acceptance criteria are comprehensive and testable. Well done!",
		"We should consider the mobile experience when implementing this feature.",
		"This requirement addresses an important user pain point. Priority should be high.",
		"The technical approach is sound, but we need to consider scalability.",
		"I suggest we create a prototype to validate this concept before full implementation.",
		"This requirement is complete and ready for the development team.",
	}

	return comments[g.rng.Intn(len(comments))]
}
