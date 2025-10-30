package service

import (
	"product-requirements-management/internal/repository"

	"github.com/sirupsen/logrus"
)

// SetupResourceService initializes all components for the MCP resource service
// This function provides centralized dependency injection for the resource system
// Requirements: REQ-c4588fc0 - JSON-RPC Handler Implementation
func SetupResourceService(
	epicRepo repository.EpicRepository,
	userStoryRepo repository.UserStoryRepository,
	requirementRepo repository.RequirementRepository,
	requirementTypeRepo repository.RequirementTypeRepository,
	logger *logrus.Logger,
) ResourceService {
	// Handle nil logger case (e.g., in tests where logger might not be initialized)
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	}

	logger.WithFields(logrus.Fields{
		"component": "resource_setup",
		"operation": "SetupResourceService",
	}).Info("Initializing MCP resource service with all providers")

	// Create resource registry
	registry := NewResourceRegistry(logger)

	// Register all resource providers in order
	// Epic resource provider - provides individual epic resources and epic collection
	epicProvider := NewEpicResourceProvider(epicRepo, logger)
	registry.RegisterProvider(epicProvider)
	logger.WithField("provider", epicProvider.GetProviderName()).Debug("Registered epic resource provider")

	// User story resource provider - provides individual user story resources and user story collection
	userStoryProvider := NewUserStoryResourceProvider(userStoryRepo, logger)
	registry.RegisterProvider(userStoryProvider)
	logger.WithField("provider", userStoryProvider.GetProviderName()).Debug("Registered user story resource provider")

	// Requirement resource provider - provides individual requirement resources and requirement collection
	requirementProvider := NewRequirementResourceProvider(requirementRepo, logger)
	registry.RegisterProvider(requirementProvider)
	logger.WithField("provider", requirementProvider.GetProviderName()).Debug("Registered requirement resource provider")

	// Requirement type resource provider - provides requirement types collection resource
	// Requirements: REQ-038 - Типы требований должны отображаться в виде ресурса requirements://requirements-types
	requirementTypeProvider := NewRequirementTypeResourceProvider(requirementTypeRepo, logger)
	registry.RegisterProvider(requirementTypeProvider)
	logger.WithField("provider", requirementTypeProvider.GetProviderName()).Debug("Registered requirement type resource provider")

	// Search resource provider - provides search template resources (no database dependency)
	searchProvider := NewSearchResourceProvider(logger)
	registry.RegisterProvider(searchProvider)
	logger.WithField("provider", searchProvider.GetProviderName()).Debug("Registered search resource provider")

	// Create resource service with registry and proper dependency injection
	resourceService := NewResourceService(registry, logger)

	logger.WithFields(logrus.Fields{
		"component":       "resource_setup",
		"operation":       "SetupResourceService",
		"providers_count": 5,
		"providers": []string{
			epicProvider.GetProviderName(),
			userStoryProvider.GetProviderName(),
			requirementProvider.GetProviderName(),
			requirementTypeProvider.GetProviderName(),
			searchProvider.GetProviderName(),
		},
	}).Info("Successfully initialized MCP resource service with all providers")

	return resourceService
}

// SetupResourceServiceForMCPHandler creates a resource service specifically configured for MCP handler
// This function ensures proper dependency injection throughout the chain
// Requirements: REQ-c4588fc0 - JSON-RPC Handler Implementation
func SetupResourceServiceForMCPHandler(repos *repository.Repositories, logger *logrus.Logger) ResourceService {
	// Handle nil logger case (e.g., in tests where logger might not be initialized)
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	}

	logger.WithFields(logrus.Fields{
		"component": "resource_setup",
		"operation": "SetupResourceServiceForMCPHandler",
	}).Info("Setting up resource service for MCP handler initialization")

	// Use the main setup function with repository dependencies
	resourceService := SetupResourceService(
		repos.Epic,
		repos.UserStory,
		repos.Requirement,
		repos.RequirementType,
		logger,
	)

	logger.WithFields(logrus.Fields{
		"component": "resource_setup",
		"operation": "SetupResourceServiceForMCPHandler",
		"status":    "complete",
	}).Info("Resource service successfully configured for MCP handler")

	return resourceService
}
