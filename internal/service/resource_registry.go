package service

import (
	"context"
	"sort"

	"github.com/sirupsen/logrus"
)

// ResourceRegistryImpl implements the ResourceRegistry interface
type ResourceRegistryImpl struct {
	providers []ResourceProvider
	logger    *logrus.Logger
}

// NewResourceRegistry creates a new ResourceRegistry instance
func NewResourceRegistry(logger *logrus.Logger) ResourceRegistry {
	return &ResourceRegistryImpl{
		providers: make([]ResourceProvider, 0),
		logger:    logger,
	}
}

// RegisterProvider implements ResourceRegistry.RegisterProvider
func (r *ResourceRegistryImpl) RegisterProvider(provider ResourceProvider) {
	r.providers = append(r.providers, provider)
	r.logger.WithField("provider", provider.GetProviderName()).Info("Resource provider registered")
}

// GetAllResources implements ResourceRegistry.GetAllResources
func (r *ResourceRegistryImpl) GetAllResources(ctx context.Context) ([]ResourceDescriptor, error) {
	var allResources []ResourceDescriptor

	r.logger.WithContext(ctx).WithField("provider_count", len(r.providers)).Debug("Collecting resources from all providers")

	for _, provider := range r.providers {
		providerLogger := r.logger.WithContext(ctx).WithField("provider", provider.GetProviderName())
		providerLogger.Debug("Getting resources from provider")

		resources, err := provider.GetResourceDescriptors(ctx)
		if err != nil {
			// Log error but continue with other providers (graceful degradation)
			providerLogger.WithError(err).Error("Failed to get resources from provider")
			continue
		}

		providerLogger.WithField("resource_count", len(resources)).Debug("Successfully got resources from provider")
		allResources = append(allResources, resources...)
	}

	// Sort resources by URI for consistent ordering
	sort.Slice(allResources, func(i, j int) bool {
		return allResources[i].URI < allResources[j].URI
	})

	r.logger.WithContext(ctx).WithField("total_resources", len(allResources)).Debug("Successfully collected resources from all providers")
	return allResources, nil
}
