package init

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// InitController orchestrates the entire MCP server initialization process.
// It coordinates between different components to guide users through server
// connection setup, credential collection, PAT token generation, and
// configuration file creation.
type InitController struct {
	inputHandler      *InputHandler
	networkClient     *NetworkClient
	configGen         *ConfigGenerator
	fileManager       *FileManager
	logger            *SecureLogger
	cleanup           *SecureCleanup
	progressTracker   *ProgressTracker
	progressIndicator *ProgressIndicator
}

// InitError represents different types of errors that can occur during initialization.
// It provides user-friendly error messages and specific guidance for common failure scenarios.
type InitError struct {
	Type         ErrorType
	Message      string
	Cause        error
	UserGuidance []string
	Retryable    bool
}

// ErrorType defines the category of initialization error.
type ErrorType int

const (
	ErrorTypeNetwork ErrorType = iota
	ErrorTypeAuth
	ErrorTypeFileSystem
	ErrorTypeValidation
	ErrorTypeUserInput
)

// String returns a human-readable name for the error type.
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeNetwork:
		return "Network Error"
	case ErrorTypeAuth:
		return "Authentication Error"
	case ErrorTypeFileSystem:
		return "File System Error"
	case ErrorTypeValidation:
		return "Validation Error"
	case ErrorTypeUserInput:
		return "User Input Error"
	default:
		return "Unknown Error"
	}
}

// Error implements the error interface for InitError.
func (e *InitError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type.String(), e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type.String(), e.Message)
}

// DisplayError provides a user-friendly display of the error with guidance.
func (e *InitError) DisplayError() {
	fmt.Printf("\n❌ %s\n", e.Type.String())
	fmt.Printf("   %s\n", e.Message)

	if len(e.UserGuidance) > 0 {
		fmt.Println("\n💡 Troubleshooting suggestions:")
		for i, guidance := range e.UserGuidance {
			fmt.Printf("   %d. %s\n", i+1, guidance)
		}
	}

	if e.Retryable {
		fmt.Println("\n🔄 This operation can be retried.")
	}
	fmt.Println()
}

// NewNetworkError creates a network-related initialization error with appropriate guidance.
func NewNetworkError(message string, cause error) *InitError {
	// Sanitize the cause error to remove sensitive information
	var sanitizedCause error
	if cause != nil {
		sanitizedCause = fmt.Errorf("%s", SanitizeErrorMessage(cause))
	}

	return &InitError{
		Type:      ErrorTypeNetwork,
		Message:   message,
		Cause:     sanitizedCause,
		Retryable: true,
		UserGuidance: []string{
			"Check that the server is running and accessible",
			"Verify the URL is correct and includes the protocol (http:// or https://)",
			"Check your network connection and firewall settings",
			"Ensure the server's /ready endpoint is available",
			"Try accessing the URL in a web browser to verify connectivity",
		},
	}
}

// NewAuthError creates an authentication-related initialization error with appropriate guidance.
func NewAuthError(message string, cause error, isCredentialIssue bool) *InitError {
	// Sanitize the cause error to remove sensitive information
	var sanitizedCause error
	if cause != nil {
		sanitizedCause = fmt.Errorf("%s", SanitizeErrorMessage(cause))
	}

	guidance := []string{
		"Verify your username and password are correct",
		"Check if your account is active and not locked",
		"Ensure you have the necessary permissions to access the API",
		"Try logging in through the web interface to verify credentials",
	}

	if !isCredentialIssue {
		guidance = append(guidance,
			"Check if the authentication endpoint (/auth/login) is available",
			"Verify the server is configured to accept login requests",
		)
	}

	return &InitError{
		Type:         ErrorTypeAuth,
		Message:      message,
		Cause:        sanitizedCause,
		Retryable:    true,
		UserGuidance: guidance,
	}
}

// NewFileSystemError creates a file system-related initialization error with appropriate guidance.
func NewFileSystemError(message string, cause error, configPath string) *InitError {
	// Sanitize the cause error to remove sensitive information
	var sanitizedCause error
	if cause != nil {
		sanitizedCause = fmt.Errorf("%s", SanitizeErrorMessage(cause))
	}

	return &InitError{
		Type:      ErrorTypeFileSystem,
		Message:   message,
		Cause:     sanitizedCause,
		Retryable: true,
		UserGuidance: []string{
			fmt.Sprintf("Check that you have write permissions to: %s", filepath.Dir(configPath)),
			"Ensure the parent directory exists or can be created",
			"Verify there is sufficient disk space available",
			"Check if the file is not currently in use by another process",
			"Try running with elevated permissions if necessary",
		},
	}
}

// NewValidationError creates a validation-related initialization error with appropriate guidance.
func NewValidationError(message string, cause error) *InitError {
	// Sanitize the cause error to remove sensitive information
	var sanitizedCause error
	if cause != nil {
		sanitizedCause = fmt.Errorf("%s", SanitizeErrorMessage(cause))
	}

	return &InitError{
		Type:      ErrorTypeValidation,
		Message:   message,
		Cause:     sanitizedCause,
		Retryable: true,
		UserGuidance: []string{
			"Verify the generated PAT token is valid and not expired",
			"Check if the token has the necessary permissions",
			"Ensure the MCP endpoint (/api/v1/mcp) is available",
			"Try generating a new PAT token if the current one fails",
			"Contact your system administrator if validation continues to fail",
		},
	}
}

// NewUserInputError creates a user input-related initialization error with appropriate guidance.
func NewUserInputError(message string, cause error) *InitError {
	// Sanitize the cause error to remove sensitive information
	var sanitizedCause error
	if cause != nil {
		sanitizedCause = fmt.Errorf("%s", SanitizeErrorMessage(cause))
	}

	return &InitError{
		Type:      ErrorTypeUserInput,
		Message:   message,
		Cause:     sanitizedCause,
		Retryable: true,
		UserGuidance: []string{
			"Ensure all required fields are filled in correctly",
			"Check that URLs include the protocol (http:// or https://)",
			"Verify usernames and passwords don't contain invalid characters",
			"Make sure you're entering information in the correct format",
		},
	}
}

// NewInitController creates a new initialization controller with all required components.
func NewInitController() *InitController {
	return &InitController{
		inputHandler:      NewInputHandler(),
		networkClient:     nil, // Will be created after URL collection
		configGen:         NewConfigGenerator(),
		fileManager:       NewFileManager(),
		logger:            NewSecureLogger(),
		cleanup:           NewSecureCleanup(),
		progressTracker:   NewProgressTracker(),
		progressIndicator: NewProgressIndicator(),
	}
}

// RunInitialization orchestrates the complete initialization process.
// It handles error recovery and retry logic throughout the process.
// Ensures secure handling of credentials and proper cleanup.
func (c *InitController) RunInitialization(configPath string) error {
	c.logger.Info("Starting MCP Server initialization process")

	// Ensure cleanup happens regardless of success or failure
	defer c.cleanup.Cleanup()

	// Display welcome message and initial progress
	c.inputHandler.DisplayWelcome()
	c.progressTracker.displayProgress()

	// Resolve configuration path
	resolvedConfigPath, err := c.resolveConfigPath(configPath)
	if err != nil {
		return NewFileSystemError("Failed to resolve configuration path", err, configPath)
	}

	// Step 1: Collect server URL with retry logic
	c.progressTracker.StartStep("url_collection")
	DisplayOperationStart("url_collection")
	serverURL, err := c.collectServerURLWithRetry()
	if err != nil {
		c.progressTracker.FailStep("url_collection", err)
		DisplayOperationError("url_collection", err)
		return err
	}
	c.progressTracker.CompleteStep("url_collection")
	DisplayOperationSuccess("url_collection", fmt.Sprintf("Server URL: %s", serverURL))

	// Step 2: Initialize network client with secure HTTPS validation and test connectivity
	c.progressTracker.StartStep("connectivity_test")
	DisplayOperationStart("connectivity_test")
	c.networkClient = NewSecureNetworkClient(serverURL)
	if err := c.testConnectivityWithRetry(); err != nil {
		c.progressTracker.FailStep("connectivity_test", err)
		DisplayOperationError("connectivity_test", err)
		return err
	}
	c.progressTracker.CompleteStep("connectivity_test")
	DisplayOperationSuccess("connectivity_test", "Server is reachable and ready")

	// Step 3: Collect credentials with retry logic and secure handling
	c.progressTracker.StartStep("credential_collection")
	DisplayOperationStart("credential_collection")
	credentials, err := c.collectSecureCredentialsWithRetry()
	if err != nil {
		c.progressTracker.FailStep("credential_collection", err)
		DisplayOperationError("credential_collection", err)
		return err
	}
	// Register credentials for cleanup
	c.cleanup.AddSecureCredentials(credentials)
	c.progressTracker.CompleteStep("credential_collection")
	DisplayOperationSuccess("credential_collection", "Credentials collected securely")

	// Step 4: Authenticate and get JWT token
	c.progressTracker.StartStep("authentication")
	DisplayOperationStart("authentication")
	authResponse, err := c.authenticateWithRetry(credentials.Username.String(), credentials.Password.String())
	if err != nil {
		c.progressTracker.FailStep("authentication", err)
		DisplayOperationError("authentication", err)
		return err
	}

	// Create secure token for JWT and register for cleanup
	jwtToken := NewSecureToken(authResponse.Token)
	c.cleanup.AddSecureToken(jwtToken)
	c.progressTracker.CompleteStep("authentication")
	DisplayOperationSuccess("authentication", fmt.Sprintf("Authenticated as: %s", authResponse.User.Username))

	// Step 5: Generate PAT token
	c.progressTracker.StartStep("pat_generation")
	DisplayOperationStart("pat_generation")
	patResponse, err := c.generatePATWithRetry(jwtToken.Token.String())
	if err != nil {
		c.progressTracker.FailStep("pat_generation", err)
		DisplayOperationError("pat_generation", err)
		return err
	}

	// Create secure token for PAT and register for cleanup
	patToken := NewSecureToken(patResponse.Token)
	c.cleanup.AddSecureToken(patToken)
	c.progressTracker.CompleteStep("pat_generation")
	DisplayOperationSuccess("pat_generation",
		fmt.Sprintf("Token name: %s", patResponse.Name),
		fmt.Sprintf("Expires: %s", patResponse.ExpiresAt.Format("2006-01-02")))

	// Step 6: Handle existing configuration file
	if err := c.handleExistingConfig(resolvedConfigPath); err != nil {
		return err
	}

	// Step 7: Generate and write configuration (only store PAT token, never credentials)
	c.progressTracker.StartStep("config_generation")
	DisplayOperationStart("config_generation")
	config := c.configGen.GenerateConfig(serverURL, patToken.Token.String())
	if err := c.writeConfigWithRetry(resolvedConfigPath, config); err != nil {
		c.progressTracker.FailStep("config_generation", err)
		DisplayOperationError("config_generation", err)
		return err
	}
	c.progressTracker.CompleteStep("config_generation")
	DisplayOperationSuccess("config_generation", fmt.Sprintf("Configuration saved to: %s", resolvedConfigPath))

	// Step 8: Validate generated configuration
	c.progressTracker.StartStep("config_validation")
	DisplayOperationStart("config_validation")
	if err := c.validateConfigWithRetry(patToken.Token.String()); err != nil {
		c.progressTracker.FailStep("config_validation", err)
		DisplayOperationError("config_validation", err)
		return err
	}
	c.progressTracker.CompleteStep("config_validation")
	DisplayOperationSuccess("config_validation", "Configuration is valid and ready to use")

	// Display final summary and success message
	c.progressTracker.DisplaySummary()
	c.inputHandler.DisplaySuccess(resolvedConfigPath)
	c.logger.Info("MCP Server initialization completed successfully")

	return nil
}

// resolveConfigPath determines the final configuration file path.
// Uses the same default path as the main server for consistency.
func (c *InitController) resolveConfigPath(providedPath string) (string, error) {
	if providedPath != "" {
		return providedPath, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Use the same default path as main server: ~/.requirements-mcp/config.json
	return filepath.Join(homeDir, ".requirements-mcp", "config.json"), nil
}

// collectServerURLWithRetry collects server URL with retry logic.
func (c *InitController) collectServerURLWithRetry() (string, error) {
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		serverURL, err := c.inputHandler.CollectServerURL()
		if err == nil {
			return serverURL, nil
		}

		c.logger.WithError(err).Warnf("Failed to collect server URL (attempt %d/%d)", attempt, maxRetries)

		if attempt == maxRetries {
			return "", NewUserInputError("Failed to collect valid server URL after multiple attempts", err)
		}

		fmt.Println("Please try again...")
	}

	return "", NewUserInputError("Maximum retry attempts exceeded for server URL collection", nil)
}

// testConnectivityWithRetry tests server connectivity with retry logic.
func (c *InitController) testConnectivityWithRetry() error {
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		c.logger.Infof("Testing connectivity to server (attempt %d/%d)...", attempt, maxRetries)

		// Use progress indicator for network request
		message := fmt.Sprintf("🔗 Testing server connectivity (attempt %d/%d)...", attempt, maxRetries)
		err := c.progressIndicator.ShowProgressWithTimeout(message, 15*time.Second, func() error {
			return c.networkClient.TestConnectivity()
		})

		if err == nil {
			c.logger.Info("Server connectivity test successful")
			return nil
		} else {
			c.logger.WithError(err).Warnf("Connectivity test failed (attempt %d/%d)", attempt, maxRetries)

			if attempt == maxRetries {
				return NewNetworkError("Failed to connect to server after multiple attempts", err)
			}

			// Show retry delay with progress
			retryDelay := time.Duration(attempt) * time.Second
			retryMessage := fmt.Sprintf("⏳ Waiting %v before retry...", retryDelay)
			c.progressIndicator.ShowProgress(retryMessage, func() error {
				time.Sleep(retryDelay)
				return nil
			})
		}
	}

	return NewNetworkError("Maximum retry attempts exceeded for connectivity test", nil)
}

// collectCredentialsWithRetry collects user credentials with retry logic.
func (c *InitController) collectCredentialsWithRetry() (string, string, error) {
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		username, password, err := c.inputHandler.CollectCredentials()
		if err == nil {
			return username, password, nil
		}

		c.logger.WithError(err).Warnf("Failed to collect credentials (attempt %d/%d)", attempt, maxRetries)

		if attempt == maxRetries {
			return "", "", NewUserInputError("Failed to collect valid credentials after multiple attempts", err)
		}

		fmt.Println("Please try again...")
	}

	return "", "", NewUserInputError("Maximum retry attempts exceeded for credential collection", nil)
}

// collectSecureCredentialsWithRetry collects user credentials with secure handling and retry logic.
func (c *InitController) collectSecureCredentialsWithRetry() (*SecureCredentials, error) {
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		username, password, err := c.inputHandler.CollectCredentials()
		if err == nil {
			// Create secure credentials and immediately clear the plain strings
			credentials := NewSecureCredentials(username, password)

			// Securely clear the original strings from memory
			// Note: This is best-effort as Go's GC may have already moved the strings
			for i := range username {
				username = username[:i] + "0" + username[i+1:]
			}
			for i := range password {
				password = password[:i] + "0" + password[i+1:]
			}

			return credentials, nil
		}

		// Log error without sensitive information
		c.logger.Warnf("Failed to collect credentials (attempt %d/%d): %s", attempt, maxRetries, SanitizeErrorMessage(err))

		if attempt == maxRetries {
			return nil, NewUserInputError("Failed to collect valid credentials after multiple attempts", err)
		}

		fmt.Println("Please try again...")
	}

	return nil, NewUserInputError("Maximum retry attempts exceeded for credential collection", nil)
}

// authenticateWithRetry performs authentication with retry logic.
func (c *InitController) authenticateWithRetry(username, password string) (*AuthResponse, error) {
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		c.logger.Infof("Authenticating with server (attempt %d/%d)...", attempt, maxRetries)

		// Use progress indicator for authentication request
		message := fmt.Sprintf("🔑 Authenticating with server (attempt %d/%d)...", attempt, maxRetries)
		var authResponse *AuthResponse
		err := c.progressIndicator.ShowProgressWithTimeout(message, 30*time.Second, func() error {
			var authErr error
			authResponse, authErr = c.networkClient.Authenticate(username, password)
			return authErr
		})

		if err == nil {
			c.logger.Info("Authentication successful")
			return authResponse, nil
		}

		c.logger.WithError(err).Warnf("Authentication failed (attempt %d/%d)", attempt, maxRetries)

		if attempt == maxRetries {
			return nil, NewAuthError("Authentication failed after multiple attempts", err, true)
		}

		// For auth errors, ask for credentials again
		fmt.Println("Authentication failed. Please check your credentials and try again.")
		username, password, err = c.collectCredentialsWithRetry()
		if err != nil {
			return nil, err
		}
	}

	return nil, NewAuthError("Maximum retry attempts exceeded for authentication", nil, true)
}

// generatePATWithRetry generates PAT token with retry logic.
func (c *InitController) generatePATWithRetry(jwtToken string) (*PATResponse, error) {
	maxRetries := 2
	for attempt := 1; attempt <= maxRetries; attempt++ {
		c.logger.Infof("Generating PAT token (attempt %d/%d)...", attempt, maxRetries)

		// Use progress indicator for PAT generation request
		message := fmt.Sprintf("🎟️  Generating Personal Access Token (attempt %d/%d)...", attempt, maxRetries)
		var patResponse *PATResponse
		err := c.progressIndicator.ShowProgressWithTimeout(message, 30*time.Second, func() error {
			var patErr error
			patResponse, patErr = c.networkClient.CreatePAT(jwtToken)
			return patErr
		})

		if err == nil {
			c.logger.Info("PAT token generated successfully")
			return patResponse, nil
		}

		c.logger.WithError(err).Warnf("PAT generation failed (attempt %d/%d)", attempt, maxRetries)

		if attempt == maxRetries {
			return nil, NewAuthError("Failed to generate PAT token after multiple attempts", err, false)
		}

		// Show retry delay with progress
		retryDelay := 2 * time.Second
		retryMessage := fmt.Sprintf("⏳ Waiting %v before retry...", retryDelay)
		c.progressIndicator.ShowProgress(retryMessage, func() error {
			time.Sleep(retryDelay)
			return nil
		})
	}

	return nil, NewAuthError("Maximum retry attempts exceeded for PAT generation", nil, false)
}

// handleExistingConfig handles existing configuration files.
func (c *InitController) handleExistingConfig(configPath string) error {
	if !c.fileManager.ConfigExists(configPath) {
		return nil
	}

	c.logger.Info("Existing configuration file detected")

	overwrite, err := c.inputHandler.ConfirmOverwrite(configPath)
	if err != nil {
		return NewUserInputError("Failed to get user confirmation for overwrite", err)
	}

	if !overwrite {
		return NewUserInputError("User cancelled initialization due to existing configuration", nil)
	}

	// Create backup of existing configuration
	backupPath, err := c.fileManager.BackupExistingConfig(configPath)
	if err != nil {
		c.logger.WithError(err).Warn("Failed to create backup of existing configuration")
		// Continue without backup - not critical
	} else {
		c.logger.Infof("Created backup of existing configuration: %s", backupPath)
	}

	return nil
}

// writeConfigWithRetry writes configuration with retry logic.
func (c *InitController) writeConfigWithRetry(configPath string, config *GeneratedConfig) error {
	maxRetries := 2
	for attempt := 1; attempt <= maxRetries; attempt++ {
		c.logger.Infof("Writing configuration file (attempt %d/%d)...", attempt, maxRetries)

		// Use progress indicator for file operations
		message := fmt.Sprintf("📝 Writing configuration file (attempt %d/%d)...", attempt, maxRetries)
		err := c.progressIndicator.ShowProgress(message, func() error {
			// Ensure config directory exists
			if err := c.fileManager.EnsureConfigDirectory(configPath); err != nil {
				return err
			}

			// Convert config to JSON
			configJSON, err := c.configGen.ToJSON(config)
			if err != nil {
				return err
			}

			// Write configuration file
			return c.fileManager.WriteConfig(configPath, configJSON)
		})

		if err == nil {
			c.logger.Info("Configuration file written successfully")
			return nil
		}

		c.logger.WithError(err).Warnf("Failed to write config file (attempt %d/%d)", attempt, maxRetries)

		if attempt == maxRetries {
			return NewFileSystemError("Failed to write configuration file after multiple attempts", err, configPath)
		}
	}

	return NewFileSystemError("Maximum retry attempts exceeded for configuration file writing", nil, configPath)
}

// validateConfigWithRetry validates the generated configuration.
func (c *InitController) validateConfigWithRetry(patToken string) error {
	maxRetries := 2
	for attempt := 1; attempt <= maxRetries; attempt++ {
		c.logger.Infof("Validating PAT token (attempt %d/%d)...", attempt, maxRetries)

		// Use progress indicator for validation request
		message := fmt.Sprintf("🔍 Validating PAT token (attempt %d/%d)...", attempt, maxRetries)
		err := c.progressIndicator.ShowProgressWithTimeout(message, 30*time.Second, func() error {
			return c.networkClient.ValidatePAT(patToken)
		})

		if err == nil {
			c.logger.Info("PAT token validation successful")
			return nil
		} else {
			c.logger.WithError(err).Warnf("PAT validation failed (attempt %d/%d)", attempt, maxRetries)

			if attempt == maxRetries {
				return NewValidationError("PAT token validation failed after multiple attempts", err)
			}

			// Show retry delay with progress
			retryDelay := 2 * time.Second
			retryMessage := fmt.Sprintf("⏳ Waiting %v before retry...", retryDelay)
			c.progressIndicator.ShowProgress(retryMessage, func() error {
				time.Sleep(retryDelay)
				return nil
			})
		}
	}

	return NewValidationError("Maximum retry attempts exceeded for PAT validation", nil)
}
