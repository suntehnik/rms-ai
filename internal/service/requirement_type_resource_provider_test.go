package service

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRequirementTypeResourceProvider_GetResourceDescriptors(t *testing.T) {
	// Setup
	mockRepo := new(MockRequirementTypeRepository)
	logger := logrus.New()
	provider := NewRequirementTypeResourceProvider(mockRepo, logger)

	// Execute
	ctx := context.Background()
	resources, err := provider.GetResourceDescriptors(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resources)
	assert.Len(t, resources, 1)

	// Verify the resource descriptor
	resource := resources[0]
	assert.Equal(t, "requirements://requirements-types", resource.URI)
	assert.Equal(t, "Requirement Types", resource.Name)
	assert.Equal(t, "List of all supported requirement types in the system", resource.Description)
	assert.Equal(t, "application/json", resource.MimeType)
}

func TestRequirementTypeResourceProvider_GetProviderName(t *testing.T) {
	// Setup
	mockRepo := new(MockRequirementTypeRepository)
	logger := logrus.New()
	provider := NewRequirementTypeResourceProvider(mockRepo, logger)

	// Execute
	name := provider.GetProviderName()

	// Assert
	assert.Equal(t, "requirement_type_provider", name)
}

func TestNewRequirementTypeResourceProvider(t *testing.T) {
	// Setup
	mockRepo := new(MockRequirementTypeRepository)
	logger := logrus.New()

	// Execute
	provider := NewRequirementTypeResourceProvider(mockRepo, logger)

	// Assert
	assert.NotNil(t, provider)
	assert.Equal(t, "requirement_type_provider", provider.GetProviderName())
}
