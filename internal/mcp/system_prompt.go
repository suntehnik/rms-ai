package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"product-requirements-management/internal/models"

	"github.com/sirupsen/logrus"
)

// PromptServiceInterface defines the interface for prompt service operations
type PromptServiceInterface interface {
	GetActive(ctx context.Context) (*models.Prompt, error)
}

// SystemPromptProvider provides system instructions for MCP initialization
type SystemPromptProvider struct {
	promptService PromptServiceInterface
	cache         *instructionsCache
	logger        *logrus.Logger
}

// instructionsCache caches system instructions to avoid repeated database queries
type instructionsCache struct {
	instructions string
	lastUpdated  time.Time
	mutex        sync.RWMutex
	ttl          time.Duration
}

// NewSystemPromptProvider creates a new system prompt provider
func NewSystemPromptProvider(promptService PromptServiceInterface) *SystemPromptProvider {
	return &SystemPromptProvider{
		promptService: promptService,
		cache: &instructionsCache{
			ttl: 5 * time.Minute, // Cache for 5 minutes
		},
		logger: logrus.New(),
	}
}

// GetInstructions retrieves system instructions using PromptService.GetActive()
func (spp *SystemPromptProvider) GetInstructions(ctx context.Context) (string, error) {
	// Check cache first
	if instructions := spp.getCachedInstructions(); instructions != "" {
		return instructions, nil
	}

	// Get active prompt from service
	instructions, err := spp.fetchInstructions(ctx)
	if err != nil {
		spp.logger.WithError(err).Warn("Failed to fetch system instructions")
		// Return empty string as fallback when no active prompt exists
		return "", nil
	}

	// Cache the instructions
	spp.setCachedInstructions(instructions)

	return instructions, nil
}

// fetchInstructions fetches instructions from the prompt service
func (spp *SystemPromptProvider) fetchInstructions(ctx context.Context) (string, error) {
	if spp.promptService == nil {
		return "", fmt.Errorf("prompt service not available")
	}

	// Get active system prompt
	activePrompt, err := spp.promptService.GetActive(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get active prompt: %w", err)
	}

	if activePrompt == nil {
		return "", fmt.Errorf("no active prompt found")
	}

	// Combine content and description as system instructions
	instructions := activePrompt.Content
	if activePrompt.Description != nil && *activePrompt.Description != "" {
		if instructions != "" {
			instructions = fmt.Sprintf("%s\n\n%s", *activePrompt.Description, instructions)
		} else {
			instructions = *activePrompt.Description
		}
	}

	return instructions, nil
}

// getCachedInstructions retrieves cached instructions if still valid
func (spp *SystemPromptProvider) getCachedInstructions() string {
	spp.cache.mutex.RLock()
	defer spp.cache.mutex.RUnlock()

	// Check if cache is still valid
	if time.Since(spp.cache.lastUpdated) < spp.cache.ttl {
		return spp.cache.instructions
	}

	return ""
}

// setCachedInstructions caches the instructions with current timestamp
func (spp *SystemPromptProvider) setCachedInstructions(instructions string) {
	spp.cache.mutex.Lock()
	defer spp.cache.mutex.Unlock()

	spp.cache.instructions = instructions
	spp.cache.lastUpdated = time.Now()
}

// InvalidateCache invalidates the cached instructions
func (spp *SystemPromptProvider) InvalidateCache() {
	spp.cache.mutex.Lock()
	defer spp.cache.mutex.Unlock()

	spp.cache.instructions = ""
	spp.cache.lastUpdated = time.Time{}
}

// UpdateInstructions updates cached instructions when active prompt changes
func (spp *SystemPromptProvider) UpdateInstructions(ctx context.Context) error {
	// Invalidate cache first
	spp.InvalidateCache()

	// Fetch fresh instructions
	_, err := spp.GetInstructions(ctx)
	return err
}

// SetCacheTTL sets the cache time-to-live duration
func (spp *SystemPromptProvider) SetCacheTTL(ttl time.Duration) {
	spp.cache.mutex.Lock()
	defer spp.cache.mutex.Unlock()

	spp.cache.ttl = ttl
}

// GetCacheStatus returns information about the cache status
func (spp *SystemPromptProvider) GetCacheStatus() map[string]interface{} {
	spp.cache.mutex.RLock()
	defer spp.cache.mutex.RUnlock()

	return map[string]interface{}{
		"has_cached_instructions": spp.cache.instructions != "",
		"last_updated":            spp.cache.lastUpdated,
		"ttl_seconds":             spp.cache.ttl.Seconds(),
		"is_valid":                time.Since(spp.cache.lastUpdated) < spp.cache.ttl,
	}
}
