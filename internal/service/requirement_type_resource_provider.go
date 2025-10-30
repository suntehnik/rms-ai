package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"product-requirements-management/internal/repository"
)

// RequirementTypeResourceProvider implements ResourceProvider for RequirementType entities
// Provides MCP resource descriptors for requirement types collection
// Requirements: REQ-038 - Типы требований должны отображаться в виде ресурса requirements://requirements-types
type RequirementTypeResourceProvider struct {
	requirementTypeRepo repository.RequirementTypeRepository
	logger              *logrus.Logger
}

// NewRequirementTypeResourceProvider creates a new RequirementTypeResourceProvider instance
func NewRequirementTypeResourceProvider(requirementTypeRepo repository.RequirementTypeRepository, logger *logrus.Logger) ResourceProvider {
	return &RequirementTypeResourceProvider{
		requirementTypeRepo: requirementTypeRepo,
		logger:              logger,
	}
}

// GetResourceDescriptors implements ResourceProvider.GetResourceDescriptors
// Returns resource descriptors for requirement types collection
// Requirements: REQ-038 - Типы требований должны отображаться в виде ресурса requirements://requirements-types
func (p *RequirementTypeResourceProvider) GetResourceDescriptors(ctx context.Context) ([]ResourceDescriptor, error) {
	p.logger.WithContext(ctx).Debug("Getting requirement type resource descriptors")

	// Add requirement types collection resource
	// This provides the requirements://requirements-types resource as specified in REQ-038
	resources := []ResourceDescriptor{
		{
			URI:         "requirements://requirements-types",
			Name:        "Requirement Types",
			Description: "List of all supported requirement types in the system",
			MimeType:    "application/json",
		},
	}

	p.logger.WithContext(ctx).WithField("resource_count", len(resources)).Debug("Successfully generated requirement type resource descriptors")
	return resources, nil
}

// GetProviderName implements ResourceProvider.GetProviderName
// Returns a unique name for this provider for logging and debugging
func (p *RequirementTypeResourceProvider) GetProviderName() string {
	return "requirement_type_provider"
}
