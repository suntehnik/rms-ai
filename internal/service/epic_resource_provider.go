package service

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"product-requirements-management/internal/repository"
)

// EpicResourceProvider implements ResourceProvider for Epic entities
// Provides MCP resource descriptors for individual epics and epic collections
type EpicResourceProvider struct {
	epicRepo repository.EpicRepository
	logger   *logrus.Logger
}

// NewEpicResourceProvider creates a new EpicResourceProvider instance
func NewEpicResourceProvider(epicRepo repository.EpicRepository, logger *logrus.Logger) ResourceProvider {
	return &EpicResourceProvider{
		epicRepo: epicRepo,
		logger:   logger,
	}
}

// GetResourceDescriptors implements ResourceProvider.GetResourceDescriptors
// Returns resource descriptors for all epics and the epic collection
func (p *EpicResourceProvider) GetResourceDescriptors(ctx context.Context) ([]ResourceDescriptor, error) {
	p.logger.WithContext(ctx).Debug("Getting epic resource descriptors")

	// Get epics with reasonable limit (1000 items max as per design)
	epics, err := p.epicRepo.List(nil, "created_at ASC", 1000, 0)
	if err != nil {
		p.logger.WithContext(ctx).WithError(err).Error("Failed to get epics for resource descriptors")
		return nil, fmt.Errorf("failed to get epics: %w", err)
	}

	var resources []ResourceDescriptor

	// Add individual epic resources - both UUID and reference ID variants
	for _, epic := range epics {
		// Add UUID-based resource
		resources = append(resources, ResourceDescriptor{
			URI:         fmt.Sprintf("requirements://epics/%s", epic.ID),
			Name:        fmt.Sprintf("Epic: %s", epic.Title),
			Description: fmt.Sprintf("Epic %s: %s", epic.ReferenceID, epic.Title),
			MimeType:    "application/json",
		})

		// Add reference ID-based resource
		resources = append(resources, ResourceDescriptor{
			URI:         fmt.Sprintf("requirements://epics/%s", epic.ReferenceID),
			Name:        fmt.Sprintf("Epic: %s", epic.Title),
			Description: fmt.Sprintf("Epic %s: %s", epic.ReferenceID, epic.Title),
			MimeType:    "application/json",
		})
	}

	// Add epics collection resource
	resources = append(resources, ResourceDescriptor{
		URI:         "requirements://epics",
		Name:        "All Epics",
		Description: "Complete list of all epics in the system",
		MimeType:    "application/json",
	})

	p.logger.WithContext(ctx).WithField("resource_count", len(resources)).Debug("Successfully generated epic resource descriptors")
	return resources, nil
}

// GetProviderName implements ResourceProvider.GetProviderName
// Returns a unique name for this provider for logging and debugging
func (p *EpicResourceProvider) GetProviderName() string {
	return "epic_provider"
}
