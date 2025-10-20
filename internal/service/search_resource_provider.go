package service

import (
	"context"

	"github.com/sirupsen/logrus"
)

// SearchResourceProvider implements ResourceProvider for search functionality
// Provides MCP resource descriptors for search template resources
// This provider does not require database dependencies as it provides static search templates
type SearchResourceProvider struct {
	logger *logrus.Logger
}

// NewSearchResourceProvider creates a new SearchResourceProvider instance
func NewSearchResourceProvider(logger *logrus.Logger) ResourceProvider {
	return &SearchResourceProvider{
		logger: logger,
	}
}

// GetResourceDescriptors implements ResourceProvider.GetResourceDescriptors
// Returns search template resource descriptors for MCP clients
// Provides a parameterized search URI that clients can use to search across all entities
func (p *SearchResourceProvider) GetResourceDescriptors(ctx context.Context) ([]ResourceDescriptor, error) {
	p.logger.WithContext(ctx).Debug("Getting search resource descriptors")

	// Return search template resource as specified in the design
	// URI format: requirements://search/{query} allows clients to perform searches
	resources := []ResourceDescriptor{
		{
			URI:         "requirements://search/{query}",
			Name:        "Search Requirements",
			Description: "Search across all epics, user stories, requirements, and acceptance criteria. Replace {query} with your search terms to find relevant entities in the requirements management system.",
			MimeType:    "application/json",
		},
	}

	p.logger.WithContext(ctx).WithField("resource_count", len(resources)).Debug("Successfully generated search resource descriptors")
	return resources, nil
}

// GetProviderName implements ResourceProvider.GetProviderName
// Returns a unique name for this provider for logging and debugging
func (p *SearchResourceProvider) GetProviderName() string {
	return "search_provider"
}
