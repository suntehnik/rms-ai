package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/sirupsen/logrus"
)

// ResourceDescriptor represents a resource available through MCP
// Complies with MCP specification for resource metadata
type ResourceDescriptor struct {
	URI         string `json:"uri"`                   // Unique resource identifier
	Name        string `json:"name"`                  // Human-readable name
	Description string `json:"description,omitempty"` // Optional description
	MimeType    string `json:"mimeType"`              // Content type (application/json)
}

// ResourceService defines the main interface for resource management
// Provides the primary entry point for MCP resource operations
type ResourceService interface {
	// GetResourceList returns all available resources from registered providers
	GetResourceList(ctx context.Context) ([]ResourceDescriptor, error)
}

// ResourceProvider defines the interface for pluggable resource providers
// Each provider is responsible for generating resources for a specific domain
type ResourceProvider interface {
	// GetResourceDescriptors returns resource descriptors from this provider
	GetResourceDescriptors(ctx context.Context) ([]ResourceDescriptor, error)

	// GetProviderName returns a unique name for this provider (for logging/debugging)
	GetProviderName() string
}

// ResourceRegistry manages multiple resource providers
// Coordinates resource collection from all registered providers
type ResourceRegistry interface {
	// GetAllResources aggregates resources from all registered providers
	GetAllResources(ctx context.Context) ([]ResourceDescriptor, error)

	// RegisterProvider adds a new resource provider to the registry
	RegisterProvider(provider ResourceProvider)
}

// ResourceServiceImpl implements the ResourceService interface
type ResourceServiceImpl struct {
	registry ResourceRegistry
	logger   *logrus.Logger
}

// NewResourceService creates a new ResourceService instance
func NewResourceService(registry ResourceRegistry, logger *logrus.Logger) ResourceService {
	logger.WithFields(logrus.Fields{
		"service":   "ResourceService",
		"operation": "NewResourceService",
	}).Info("Creating new ResourceService instance")

	return &ResourceServiceImpl{
		registry: registry,
		logger:   logger,
	}
}

// GetResourceList implements ResourceService.GetResourceList
func (s *ResourceServiceImpl) GetResourceList(ctx context.Context) ([]ResourceDescriptor, error) {
	logger := s.logger.WithContext(ctx).WithFields(logrus.Fields{
		"operation": "GetResourceList",
		"service":   "ResourceService",
	})

	logger.Info("Starting resource list retrieval")

	resources, err := s.registry.GetAllResources(ctx)
	if err != nil {
		logger.WithError(err).Error("Failed to get resources from registry")
		return nil, fmt.Errorf("failed to get resources from registry: %w", err)
	}

	// Sort resources by URI for consistent ordering (defensive programming)
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].URI < resources[j].URI
	})

	logger.WithFields(logrus.Fields{
		"resource_count": len(resources),
		"status":         "success",
		"sorted":         true,
	}).Info("Successfully retrieved and sorted resource list")

	return resources, nil
}
