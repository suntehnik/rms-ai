package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"product-requirements-management/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPromptService is a mock implementation of PromptServiceInterface
type MockPromptService struct {
	mock.Mock
}

func (m *MockPromptService) GetActive(ctx context.Context) (*models.Prompt, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Prompt), args.Error(1)
}

func TestNewSystemPromptProvider(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	assert.NotNil(t, provider)
	assert.Equal(t, mockService, provider.promptService)
	assert.NotNil(t, provider.cache)
	assert.NotNil(t, provider.logger)
	assert.Equal(t, 5*time.Minute, provider.cache.ttl)
}

func TestSystemPromptProvider_GetInstructions_Success(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	description := "Test description"
	content := "Test content"
	activePrompt := &models.Prompt{
		ID:          uuid.New(),
		Name:        "test-prompt",
		Title:       "Test Prompt",
		Description: &description,
		Content:     content,
		IsActive:    true,
	}

	mockService.On("GetActive", mock.Anything).Return(activePrompt, nil)

	ctx := context.Background()
	instructions, err := provider.GetInstructions(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "Test description\n\nTest content", instructions)
	mockService.AssertExpectations(t)
}

func TestSystemPromptProvider_GetInstructions_ContentOnly(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	content := "Test content only"
	activePrompt := &models.Prompt{
		ID:          uuid.New(),
		Name:        "test-prompt",
		Title:       "Test Prompt",
		Description: nil,
		Content:     content,
		IsActive:    true,
	}

	mockService.On("GetActive", mock.Anything).Return(activePrompt, nil)

	ctx := context.Background()
	instructions, err := provider.GetInstructions(ctx)

	assert.NoError(t, err)
	assert.Equal(t, content, instructions)
	mockService.AssertExpectations(t)
}

func TestSystemPromptProvider_GetInstructions_EmptyDescription(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	emptyDescription := ""
	content := "Test content"
	activePrompt := &models.Prompt{
		ID:          uuid.New(),
		Name:        "test-prompt",
		Title:       "Test Prompt",
		Description: &emptyDescription,
		Content:     content,
		IsActive:    true,
	}

	mockService.On("GetActive", mock.Anything).Return(activePrompt, nil)

	ctx := context.Background()
	instructions, err := provider.GetInstructions(ctx)

	assert.NoError(t, err)
	assert.Equal(t, content, instructions)
	mockService.AssertExpectations(t)
}

func TestSystemPromptProvider_GetInstructions_NoActivePrompt(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	mockService.On("GetActive", mock.Anything).Return(nil, errors.New("no active prompt found"))

	ctx := context.Background()
	instructions, err := provider.GetInstructions(ctx)

	assert.NoError(t, err) // Should not return error, just empty string
	assert.Equal(t, "", instructions)
	mockService.AssertExpectations(t)
}

func TestSystemPromptProvider_GetInstructions_ServiceError(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	mockService.On("GetActive", mock.Anything).Return(nil, errors.New("database error"))

	ctx := context.Background()
	instructions, err := provider.GetInstructions(ctx)

	assert.NoError(t, err) // Should not return error, just empty string
	assert.Equal(t, "", instructions)
	mockService.AssertExpectations(t)
}

func TestSystemPromptProvider_GetInstructions_NilService(t *testing.T) {
	provider := NewSystemPromptProvider(nil)

	ctx := context.Background()
	instructions, err := provider.GetInstructions(ctx)

	assert.NoError(t, err) // Should not return error, just empty string
	assert.Equal(t, "", instructions)
}

func TestSystemPromptProvider_Caching(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	content := "Cached content"
	activePrompt := &models.Prompt{
		ID:       uuid.New(),
		Name:     "test-prompt",
		Title:    "Test Prompt",
		Content:  content,
		IsActive: true,
	}

	// First call should hit the service
	mockService.On("GetActive", mock.Anything).Return(activePrompt, nil).Once()

	ctx := context.Background()

	// First call
	instructions1, err1 := provider.GetInstructions(ctx)
	assert.NoError(t, err1)
	assert.Equal(t, content, instructions1)

	// Second call should use cache (no additional service call)
	instructions2, err2 := provider.GetInstructions(ctx)
	assert.NoError(t, err2)
	assert.Equal(t, content, instructions2)

	mockService.AssertExpectations(t)
}

func TestSystemPromptProvider_CacheExpiration(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)
	provider.SetCacheTTL(100 * time.Millisecond) // Very short TTL for testing

	content := "Test content"
	activePrompt := &models.Prompt{
		ID:       uuid.New(),
		Name:     "test-prompt",
		Title:    "Test Prompt",
		Content:  content,
		IsActive: true,
	}

	// Should be called twice due to cache expiration
	mockService.On("GetActive", mock.Anything).Return(activePrompt, nil).Twice()

	ctx := context.Background()

	// First call
	instructions1, err1 := provider.GetInstructions(ctx)
	assert.NoError(t, err1)
	assert.Equal(t, content, instructions1)

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Second call should hit service again
	instructions2, err2 := provider.GetInstructions(ctx)
	assert.NoError(t, err2)
	assert.Equal(t, content, instructions2)

	mockService.AssertExpectations(t)
}

func TestSystemPromptProvider_InvalidateCache(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	content := "Test content"
	activePrompt := &models.Prompt{
		ID:       uuid.New(),
		Name:     "test-prompt",
		Title:    "Test Prompt",
		Content:  content,
		IsActive: true,
	}

	// Should be called twice due to cache invalidation
	mockService.On("GetActive", mock.Anything).Return(activePrompt, nil).Twice()

	ctx := context.Background()

	// First call
	instructions1, err1 := provider.GetInstructions(ctx)
	assert.NoError(t, err1)
	assert.Equal(t, content, instructions1)

	// Invalidate cache
	provider.InvalidateCache()

	// Second call should hit service again
	instructions2, err2 := provider.GetInstructions(ctx)
	assert.NoError(t, err2)
	assert.Equal(t, content, instructions2)

	mockService.AssertExpectations(t)
}

func TestSystemPromptProvider_UpdateInstructions(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	content := "Updated content"
	activePrompt := &models.Prompt{
		ID:       uuid.New(),
		Name:     "test-prompt",
		Title:    "Test Prompt",
		Content:  content,
		IsActive: true,
	}

	mockService.On("GetActive", mock.Anything).Return(activePrompt, nil)

	ctx := context.Background()
	err := provider.UpdateInstructions(ctx)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestSystemPromptProvider_UpdateInstructions_Error(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	mockService.On("GetActive", mock.Anything).Return(nil, errors.New("service error"))

	ctx := context.Background()
	err := provider.UpdateInstructions(ctx)

	assert.NoError(t, err) // UpdateInstructions should not return error, it calls GetInstructions which handles errors
	mockService.AssertExpectations(t)
}

func TestSystemPromptProvider_SetCacheTTL(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	newTTL := 10 * time.Minute
	provider.SetCacheTTL(newTTL)

	assert.Equal(t, newTTL, provider.cache.ttl)
}

func TestSystemPromptProvider_GetCacheStatus(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	// Initially empty cache
	status := provider.GetCacheStatus()
	assert.False(t, status["has_cached_instructions"].(bool))
	assert.False(t, status["is_valid"].(bool))
	assert.Equal(t, 300.0, status["ttl_seconds"].(float64)) // 5 minutes

	// Add something to cache
	provider.setCachedInstructions("test instructions")

	status = provider.GetCacheStatus()
	assert.True(t, status["has_cached_instructions"].(bool))
	assert.True(t, status["is_valid"].(bool))
}

func TestSystemPromptProvider_ConcurrentAccess(t *testing.T) {
	mockService := &MockPromptService{}
	provider := NewSystemPromptProvider(mockService)

	content := "Concurrent test content"
	activePrompt := &models.Prompt{
		ID:       uuid.New(),
		Name:     "test-prompt",
		Title:    "Test Prompt",
		Content:  content,
		IsActive: true,
	}

	// Allow multiple calls since concurrent access might hit the service multiple times
	// before cache is populated
	mockService.On("GetActive", mock.Anything).Return(activePrompt, nil)

	ctx := context.Background()

	// Run multiple goroutines concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			instructions, err := provider.GetInstructions(ctx)
			assert.NoError(t, err)
			assert.Equal(t, content, instructions)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	mockService.AssertExpectations(t)
}
