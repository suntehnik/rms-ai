package init

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"product-requirements-management/internal/mcp"
)

// TestConfigGenerationAndValidation tests that generated configs work with normal server startup
func TestConfigGenerationAndValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directory for test config
	tempDir, err := os.MkdirTemp("", "mcp-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	t.Run("generated_config_structure_validation", func(t *testing.T) {
		// Test config generation
		generator := NewConfigGenerator()
		config := generator.GenerateConfig("https://api.example.com", "test-pat-token")

		// Verify config structure
		assert.Equal(t, "https://api.example.com", config.BackendAPIURL)
		assert.Equal(t, "test-pat-token", config.PATToken)
		assert.Equal(t, "30s", config.RequestTimeout)
		assert.Equal(t, "info", config.LogLevel)

		// Test config validation
		err := generator.ValidateConfig(config)
		assert.NoError(t, err)
	})

	t.Run("config_file_operations", func(t *testing.T) {
		// Test file manager operations
		fileManager := NewFileManager()

		// Test directory creation
		err := fileManager.EnsureConfigDirectory(configPath)
		assert.NoError(t, err)

		// Verify directory exists
		dir := filepath.Dir(configPath)
		info, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		// Test config writing
		config := &GeneratedConfig{
			BackendAPIURL:  "https://test.example.com",
			PATToken:       "test-token-123",
			RequestTimeout: "30s",
			LogLevel:       "info",
		}

		configData, err := json.MarshalIndent(config, "", "  ")
		require.NoError(t, err)

		err = fileManager.WriteConfig(configPath, configData)
		assert.NoError(t, err)

		// Verify file exists and has correct permissions
		info, err = os.Stat(configPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

		// Verify file content
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var readConfig GeneratedConfig
		err = json.Unmarshal(data, &readConfig)
		require.NoError(t, err)
		assert.Equal(t, config.BackendAPIURL, readConfig.BackendAPIURL)
		assert.Equal(t, config.PATToken, readConfig.PATToken)
	})

	t.Run("config_compatibility_with_mcp_server", func(t *testing.T) {
		// Create a config using the generator
		generator := NewConfigGenerator()
		config := generator.GenerateConfig("https://api.example.com", "test-pat-token-456")

		// Write config to file
		fileManager := NewFileManager()
		configData, err := json.MarshalIndent(config, "", "  ")
		require.NoError(t, err)

		err = fileManager.WriteConfig(configPath, configData)
		require.NoError(t, err)

		// Test that the config can be loaded by the MCP server
		mcpConfig, err := mcp.LoadConfigFromPath(configPath)
		require.NoError(t, err)

		// Verify config validation passes
		err = mcpConfig.Validate()
		assert.NoError(t, err)

		// Verify all fields are correctly mapped
		assert.Equal(t, config.BackendAPIURL, mcpConfig.BackendAPIURL)
		assert.Equal(t, config.PATToken, mcpConfig.PATToken)
		assert.Equal(t, config.RequestTimeout, mcpConfig.RequestTimeout)
		assert.Equal(t, config.LogLevel, mcpConfig.LogLevel)

		// Test timeout parsing
		timeout := mcpConfig.GetRequestTimeout()
		assert.Equal(t, 30*time.Second, timeout)
	})

	t.Run("backup_functionality", func(t *testing.T) {
		// Create initial config
		initialConfig := &GeneratedConfig{
			BackendAPIURL:  "https://old.example.com",
			PATToken:       "old-token",
			RequestTimeout: "60s",
			LogLevel:       "debug",
		}

		fileManager := NewFileManager()
		configData, err := json.MarshalIndent(initialConfig, "", "  ")
		require.NoError(t, err)

		err = fileManager.WriteConfig(configPath, configData)
		require.NoError(t, err)

		// Test backup creation
		backupPath, err := fileManager.BackupExistingConfig(configPath)
		require.NoError(t, err)
		assert.NotEmpty(t, backupPath)

		// Verify backup file exists
		assert.FileExists(t, backupPath)

		// Verify backup contains original content
		backupData, err := os.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Contains(t, string(backupData), "old.example.com")

		// Verify original file still exists
		assert.FileExists(t, configPath)
	})
}

// TestNetworkClientIntegration tests network client functionality with mock server
func TestNetworkClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("connectivity_testing", func(t *testing.T) {
		// Setup mock server
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/ready" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer mockServer.Close()

		// Test connectivity
		client := NewNetworkClient(mockServer.URL)
		err := client.TestConnectivity()
		assert.NoError(t, err)
	})

	t.Run("authentication_flow", func(t *testing.T) {
		// Setup mock server with auth endpoints
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/ready":
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

			case "/auth/login":
				var loginReq map[string]string
				json.NewDecoder(r.Body).Decode(&loginReq)

				if loginReq["username"] == "testuser" && loginReq["password"] == "testpass" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"token":      "mock-jwt-token",
						"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
						"user": map[string]interface{}{
							"id":       uuid.New().String(),
							"username": "testuser",
							"email":    "test@example.com",
						},
					})
				} else {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
				}

			case "/api/v1/pats":
				if r.Header.Get("Authorization") != "Bearer mock-jwt-token" {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
					return
				}

				expiresAt := time.Now().Add(365 * 24 * time.Hour)
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"token": "mock-pat-token-" + uuid.New().String()[:8],
					"pat": map[string]interface{}{
						"id":         uuid.New().String(),
						"name":       "MCP Server - test - " + time.Now().Format("2006-01-02"),
						"expires_at": expiresAt.Format(time.RFC3339),
						"created_at": time.Now().Format(time.RFC3339),
					},
				})

			case "/auth/profile":
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer mock-pat-token") {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
					return
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":       uuid.New().String(),
					"username": "testuser",
					"email":    "test@example.com",
					"role":     "User",
				})

			case "/api/v1/mcp":
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer mock-pat-token") {
					w.WriteHeader(http.StatusUnauthorized)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"jsonrpc": "2.0",
						"id":      1,
						"error": map[string]interface{}{
							"code":    -32001,
							"message": "Unauthorized",
						},
					})
					return
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      1,
					"result": map[string]interface{}{
						"protocolVersion": "2025-06-18",
						"capabilities":    map[string]interface{}{},
						"serverInfo": map[string]interface{}{
							"name":    "mock-mcp-server",
							"version": "1.0.0",
						},
					},
				})

			default:
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
			}
		}))
		defer mockServer.Close()

		client := NewNetworkClient(mockServer.URL)

		// Test authentication
		authResponse, err := client.Authenticate("testuser", "testpass")
		require.NoError(t, err)
		assert.Equal(t, "mock-jwt-token", authResponse.Token)
		assert.NotEmpty(t, authResponse.User.Username)

		// Test PAT creation
		patResponse, err := client.CreatePAT(authResponse.Token)
		require.NoError(t, err)
		assert.NotEmpty(t, patResponse.Token)
		assert.Contains(t, patResponse.Name, "MCP Server")

		// Test PAT validation
		err = client.ValidatePAT(patResponse.Token)
		assert.NoError(t, err)
	})

	t.Run("error_handling", func(t *testing.T) {
		// Test with unreachable server
		client := NewNetworkClient("http://localhost:99999")
		err := client.TestConnectivity()
		assert.Error(t, err)

		// Test with server that returns errors
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Server error"})
		}))
		defer errorServer.Close()

		client = NewNetworkClient(errorServer.URL)
		err = client.TestConnectivity()
		assert.Error(t, err)
	})
}

// TestErrorHandling tests error scenarios and recovery
func TestErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("init_error_types", func(t *testing.T) {
		// Test network error creation
		networkErr := NewNetworkError("Connection failed", nil)
		assert.Equal(t, ErrorTypeNetwork, networkErr.Type)
		assert.Contains(t, networkErr.Message, "Connection failed")
		assert.NotEmpty(t, networkErr.UserGuidance)

		// Test auth error creation
		authErr := NewAuthError("Authentication failed", nil, true)
		assert.Equal(t, ErrorTypeAuth, authErr.Type)
		assert.Contains(t, authErr.Message, "Authentication failed")
		assert.NotEmpty(t, authErr.UserGuidance)

		// Test filesystem error creation
		fsErr := NewFileSystemError("Permission denied", nil, "/test/path")
		assert.Equal(t, ErrorTypeFileSystem, fsErr.Type)
		assert.Contains(t, fsErr.Message, "Permission denied")
		assert.NotEmpty(t, fsErr.UserGuidance)

		// Test validation error creation
		validationErr := NewValidationError("Invalid config", nil)
		assert.Equal(t, ErrorTypeValidation, validationErr.Type)
		assert.Contains(t, validationErr.Message, "Invalid config")
		assert.NotEmpty(t, validationErr.UserGuidance)
	})

	t.Run("filesystem_permission_errors", func(t *testing.T) {
		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "mcp-perm-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		// Create read-only directory
		readOnlyDir := filepath.Join(tempDir, "readonly")
		err = os.Mkdir(readOnlyDir, 0444)
		require.NoError(t, err)
		defer os.Chmod(readOnlyDir, 0755) // Restore permissions for cleanup

		readOnlyConfigPath := filepath.Join(readOnlyDir, "config.json")

		// Test file manager with permission error
		fileManager := NewFileManager()
		config := &GeneratedConfig{
			BackendAPIURL:  "https://test.example.com",
			PATToken:       "test-token",
			RequestTimeout: "30s",
			LogLevel:       "info",
		}

		configData, err := json.MarshalIndent(config, "", "  ")
		require.NoError(t, err)

		err = fileManager.WriteConfig(readOnlyConfigPath, configData)
		assert.Error(t, err)
	})
}

// TestProgressTracking tests progress tracking functionality
func TestProgressTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("progress_tracker_functionality", func(t *testing.T) {
		tracker := NewProgressTracker()

		// Test step completion
		tracker.CompleteStep("collect_url")
		tracker.CompleteStep("test_connectivity")

		// Test progress calculation
		progress := tracker.GetOverallProgress()
		assert.GreaterOrEqual(t, progress, 0.0)
		assert.LessOrEqual(t, progress, 100.0)

		// Test display functionality
		tracker.DisplaySummary()
	})

	t.Run("progress_indicator_functionality", func(t *testing.T) {
		indicator := NewProgressIndicator()

		// Test that indicator can be started and stopped without panicking
		indicator.Start("Testing connectivity...")

		// Brief delay to simulate work
		time.Sleep(100 * time.Millisecond)

		indicator.Stop()

		// Test message update
		indicator.UpdateMessage("Connection successful!")
	})
}

// TestSecurityFeatures tests security-related functionality
func TestSecurityFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("secure_cleanup", func(t *testing.T) {
		cleanup := NewSecureCleanup()

		// Test cleanup function registration
		cleanupCalled := false
		cleanup.AddCleanup(func() {
			cleanupCalled = true
		})

		// Cleanup should execute registered functions
		cleanup.Cleanup()

		// Verify cleanup was called
		assert.True(t, cleanupCalled)
	})

	t.Run("secure_logger", func(t *testing.T) {
		logger := NewSecureLogger()

		// Test that sensitive data is filtered
		// Note: We can't easily capture log output in tests,
		// but we can verify the logger doesn't panic with sensitive data
		logger.Info("User logged in with password: secret123")
		logger.Warn("Token validation failed: jwt-token-abc123")
		logger.Error("PAT creation error: pat_token_xyz789")
	})
}
