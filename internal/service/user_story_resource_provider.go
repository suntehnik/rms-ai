package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"product-requirements-management/internal/repository"
)

// UserStoryResourceProvider implements ResourceProvider for UserStory entities
// Provides MCP resource descriptors for individual user stories and user story collections
type UserStoryResourceProvider struct {
	userStoryRepo repository.UserStoryRepository
	logger        *logrus.Logger
}

// NewUserStoryResourceProvider creates a new UserStoryResourceProvider instance
func NewUserStoryResourceProvider(userStoryRepo repository.UserStoryRepository, logger *logrus.Logger) ResourceProvider {
	return &UserStoryResourceProvider{
		userStoryRepo: userStoryRepo,
		logger:        logger,
	}
}

// GetResourceDescriptors implements ResourceProvider.GetResourceDescriptors
// Returns resource descriptors for all user stories and the user story collection
func (p *UserStoryResourceProvider) GetResourceDescriptors(ctx context.Context) ([]ResourceDescriptor, error) {
	p.logger.WithContext(ctx).Debug("Getting user story resource descriptors")

	// Get user stories with reasonable limit (1000 items max as per design)
	userStories, err := p.userStoryRepo.List(nil, "created_at ASC", 1000, 0)
	if err != nil {
		p.logger.WithContext(ctx).WithError(err).Error("Failed to get user stories for resource descriptors")
		return nil, fmt.Errorf("failed to get user stories: %w", err)
	}

	var resources []ResourceDescriptor

	// Add individual user story resources
	for _, userStory := range userStories {
		resources = append(resources, ResourceDescriptor{
			URI:         fmt.Sprintf("requirements://user-stories/%s", userStory.ID),
			Name:        fmt.Sprintf("User Story: %s", userStory.Title),
			Description: fmt.Sprintf("User Story %s: %s", userStory.ReferenceID, userStory.Title),
			MimeType:    "application/json",
		})
	}

	// Add user stories collection resource
	resources = append(resources, ResourceDescriptor{
		URI:         "requirements://user-stories",
		Name:        "All User Stories",
		Description: "Complete list of all user stories in the system",
		MimeType:    "application/json",
	})

	p.logger.WithContext(ctx).WithField("resource_count", len(resources)).Debug("Successfully generated user story resource descriptors")
	return resources, nil
}

// GetProviderName implements ResourceProvider.GetProviderName
// Returns a unique name for this provider for logging and debugging
func (p *UserStoryResourceProvider) GetProviderName() string {
	return "user_story_provider"
}
