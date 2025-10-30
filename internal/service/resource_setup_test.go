package service

import (
	"testing"

	"product-requirements-management/internal/repository"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSetupResourceService(t *testing.T) {
	// Create mock repositories
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockRequirementRepo := &MockRequirementRepository{}

	// Test with nil logger (should handle gracefully)
	resourceService := SetupResourceService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockRequirementRepo,
		nil, // mockRequirementTypeRepo
		nil, // Test nil logger handling
	)

	// Verify that the service was created
	assert.NotNil(t, resourceService, "Resource service should not be nil")

	// Verify that the service is of the correct type
	serviceImpl, ok := resourceService.(*ResourceServiceImpl)
	assert.True(t, ok, "Resource service should be of type ResourceServiceImpl")
	assert.NotNil(t, serviceImpl.registry, "Resource service should have a registry")
	assert.NotNil(t, serviceImpl.logger, "Resource service should have a logger")
}

// TestSetupResourceService_Integration tests the complete setup with provider registration
func TestSetupResourceService_ProviderRegistration(t *testing.T) {
	// Create mock repositories
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockRequirementRepo := &MockRequirementRepository{}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	// Test the setup function
	resourceService := SetupResourceService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockRequirementRepo,
		nil, // mockRequirementTypeRepo
		logger,
	)

	// Verify that the service was created and has the correct structure
	serviceImpl, ok := resourceService.(*ResourceServiceImpl)
	assert.True(t, ok, "Resource service should be of type ResourceServiceImpl")

	// Verify that the registry was created and has providers
	registryImpl, ok := serviceImpl.registry.(*ResourceRegistryImpl)
	assert.True(t, ok, "Registry should be of type ResourceRegistryImpl")

	// Verify that all 4 providers were registered (epic, user story, requirement, search)
	assert.Len(t, registryImpl.providers, 5, "Should have 4 providers registered")

	// Verify provider names
	providerNames := make([]string, len(registryImpl.providers))
	for i, provider := range registryImpl.providers {
		providerNames[i] = provider.GetProviderName()
	}

	expectedProviders := []string{"epic_provider", "user_story_provider", "requirement_provider", "requirement_type_provider", "search_provider"}
	assert.ElementsMatch(t, expectedProviders, providerNames, "Should have all expected providers")
}

func TestSetupResourceServiceForMCPHandler(t *testing.T) {
	// Create mock repositories
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockRequirementRepo := &MockRequirementRepository{}

	// Create mock repositories struct
	repos := &repository.Repositories{
		Epic:        mockEpicRepo,
		UserStory:   mockUserStoryRepo,
		Requirement: mockRequirementRepo,
	}

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	// Test the MCP handler setup function
	resourceService := SetupResourceServiceForMCPHandler(repos, logger)

	// Verify that the service was created
	assert.NotNil(t, resourceService, "Resource service should not be nil")

	// Verify that the service is of the correct type
	serviceImpl, ok := resourceService.(*ResourceServiceImpl)
	assert.True(t, ok, "Resource service should be of type ResourceServiceImpl")
	assert.NotNil(t, serviceImpl.registry, "Resource service should have a registry")
	assert.NotNil(t, serviceImpl.logger, "Resource service should have a logger")
}
func TestSetupResourceService_NilLoggerHandling(t *testing.T) {
	// Create mock repositories
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockRequirementRepo := &MockRequirementRepository{}

	// Test with nil logger (should handle gracefully)
	resourceService := SetupResourceService(
		mockEpicRepo,
		mockUserStoryRepo,
		mockRequirementRepo,
		nil,
		nil, // Explicitly test nil logger
	)

	// Verify that the service was created successfully
	assert.NotNil(t, resourceService, "Resource service should not be nil even with nil logger")

	// Verify that the service has a logger (should be created internally)
	serviceImpl, ok := resourceService.(*ResourceServiceImpl)
	assert.True(t, ok, "Resource service should be of type ResourceServiceImpl")
	assert.NotNil(t, serviceImpl.logger, "Resource service should have created a logger internally")
}

func TestSetupResourceServiceForMCPHandler_NilLoggerHandling(t *testing.T) {
	// Create mock repositories
	mockEpicRepo := &MockEpicRepository{}
	mockUserStoryRepo := &MockUserStoryRepository{}
	mockRequirementRepo := &MockRequirementRepository{}

	// Create mock repositories struct
	repos := &repository.Repositories{
		Epic:        mockEpicRepo,
		UserStory:   mockUserStoryRepo,
		Requirement: mockRequirementRepo,
	}

	// Test with nil logger (should handle gracefully)
	resourceService := SetupResourceServiceForMCPHandler(repos, nil)

	// Verify that the service was created successfully
	assert.NotNil(t, resourceService, "Resource service should not be nil even with nil logger")

	// Verify that the service has a logger (should be created internally)
	serviceImpl, ok := resourceService.(*ResourceServiceImpl)
	assert.True(t, ok, "Resource service should be of type ResourceServiceImpl")
	assert.NotNil(t, serviceImpl.logger, "Resource service should have created a logger internally")
}
