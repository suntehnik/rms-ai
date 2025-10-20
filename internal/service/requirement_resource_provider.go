package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"product-requirements-management/internal/repository"
)

// RequirementResourceProvider implements ResourceProvider for Requirement entities
// Provides MCP resource descriptors for individual requirements and requirement collections
type RequirementResourceProvider struct {
	requirementRepo repository.RequirementRepository
	logger          *logrus.Logger
}

// NewRequirementResourceProvider creates a new RequirementResourceProvider instance
func NewRequirementResourceProvider(requirementRepo repository.RequirementRepository, logger *logrus.Logger) ResourceProvider {
	return &RequirementResourceProvider{
		requirementRepo: requirementRepo,
		logger:          logger,
	}
}

// GetResourceDescriptors implements ResourceProvider.GetResourceDescriptors
// Returns resource descriptors for all requirements and the requirement collection
func (p *RequirementResourceProvider) GetResourceDescriptors(ctx context.Context) ([]ResourceDescriptor, error) {
	p.logger.WithContext(ctx).Debug("Getting requirement resource descriptors")

	// Get requirements with reasonable limit (1000 items max as per design)
	requirements, err := p.requirementRepo.List(nil, "created_at ASC", 1000, 0)
	if err != nil {
		p.logger.WithContext(ctx).WithError(err).Error("Failed to get requirements for resource descriptors")
		return nil, fmt.Errorf("failed to get requirements: %w", err)
	}

	var resources []ResourceDescriptor

	// Add individual requirement resources
	for _, requirement := range requirements {
		resources = append(resources, ResourceDescriptor{
			URI:         fmt.Sprintf("requirements://requirements/%s", requirement.ID),
			Name:        fmt.Sprintf("Requirement: %s", requirement.Title),
			Description: fmt.Sprintf("Requirement %s: %s", requirement.ReferenceID, requirement.Title),
			MimeType:    "application/json",
		})
	}

	// Add requirements collection resource
	resources = append(resources, ResourceDescriptor{
		URI:         "requirements://requirements",
		Name:        "All Requirements",
		Description: "Complete list of all requirements in the system",
		MimeType:    "application/json",
	})

	p.logger.WithContext(ctx).WithField("resource_count", len(resources)).Debug("Successfully generated requirement resource descriptors")
	return resources, nil
}

// GetProviderName implements ResourceProvider.GetProviderName
// Returns a unique name for this provider for logging and debugging
func (p *RequirementResourceProvider) GetProviderName() string {
	return "requirement_provider"
}
