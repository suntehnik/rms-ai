package init

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

// InputHandler manages all user interactions and input collection during initialization.
type InputHandler struct {
	reader *bufio.Reader
}

// ServerConfig holds server connection configuration.
type ServerConfig struct {
	URL      string
	Username string
	Password string
}

// NewInputHandler creates a new input handler for user interactions.
func NewInputHandler() *InputHandler {
	return &InputHandler{
		reader: bufio.NewReader(os.Stdin),
	}
}

// DisplayWelcome displays the welcome message and explains the initialization process.
func (h *InputHandler) DisplayWelcome() {
	fmt.Println("🚀 MCP Server Initialization")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("Welcome to the MCP Server interactive setup!")
	fmt.Println("This process will guide you through configuring your MCP server")
	fmt.Println("by connecting to your backend API and generating the necessary")
	fmt.Println("authentication tokens.")
	fmt.Println()
	fmt.Println("📋 What this process will do:")
	fmt.Println("   1. 🌐 Collect and validate your backend API server URL")
	fmt.Println("   2. 🔗 Test connectivity to ensure the server is reachable")
	fmt.Println("   3. 🔐 Securely collect your authentication credentials")
	fmt.Println("   4. 🔑 Authenticate with the server and obtain a JWT token")
	fmt.Println("   5. 🎟️  Generate a Personal Access Token (PAT) with 1-year expiration")
	fmt.Println("   6. 📝 Create and save your configuration file")
	fmt.Println("   7. 🔍 Validate the configuration to ensure everything works")
	fmt.Println()
	fmt.Println("🔒 Security notes:")
	fmt.Println("   • Your password will not be displayed as you type")
	fmt.Println("   • Credentials are used only for token generation and not stored")
	fmt.Println("   • Configuration files are created with secure permissions")
	fmt.Println("   • All network communication uses HTTPS when available")
	fmt.Println()
}

// CollectServerURL prompts for and validates the backend API URL.
func (h *InputHandler) CollectServerURL() (string, error) {
	fmt.Println("Step 1: Backend API Configuration")
	fmt.Println("---------------------------------")
	fmt.Println()
	fmt.Println("Please enter your backend API URL.")
	fmt.Println("Examples:")
	fmt.Println("  https://api.example.com")
	fmt.Println("  http://localhost:8080")
	fmt.Println("  https://requirements.company.com")
	fmt.Println()

	for {
		fmt.Print("Backend API URL: ")
		input, err := h.readLine()
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		if input == "" {
			fmt.Println("❌ URL cannot be empty. Please enter a valid URL.")
			fmt.Println()
			continue
		}

		// Validate URL format
		parsedURL, err := url.Parse(input)
		if err != nil {
			fmt.Printf("❌ Invalid URL format: %v\n", err)
			fmt.Println("Please enter a valid URL (e.g., https://api.example.com)")
			fmt.Println()
			continue
		}

		if parsedURL.Scheme == "" {
			fmt.Println("❌ URL must include a scheme (http:// or https://)")
			fmt.Println("Please enter a valid URL (e.g., https://api.example.com)")
			fmt.Println()
			continue
		}

		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			fmt.Println("❌ URL must use http:// or https://")
			fmt.Println("Please enter a valid URL (e.g., https://api.example.com)")
			fmt.Println()
			continue
		}

		if parsedURL.Host == "" {
			fmt.Println("❌ URL must include a host (e.g., api.example.com)")
			fmt.Println("Please enter a valid URL (e.g., https://api.example.com)")
			fmt.Println()
			continue
		}

		// Test connectivity
		fmt.Printf("🔍 Testing connectivity to %s...\n", input)
		if err := h.testConnectivity(input); err != nil {
			fmt.Printf("❌ Connection test failed: %v\n", err)
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("1. Check that the server is running and accessible")
			fmt.Println("2. Verify the URL is correct")
			fmt.Println("3. Check your network connection")
			fmt.Println()

			if !h.askRetry("Would you like to try a different URL?") {
				return "", fmt.Errorf("user cancelled URL configuration")
			}
			continue
		}

		fmt.Println("✅ Connection successful!")
		fmt.Println()
		return input, nil
	}
}

// CollectCredentials prompts for username and password with secure input.
func (h *InputHandler) CollectCredentials() (username, password string, err error) {
	fmt.Println("Step 2: Authentication Credentials")
	fmt.Println("----------------------------------")
	fmt.Println()
	fmt.Println("Please provide your login credentials for the backend API.")
	fmt.Println("These will be used to authenticate and generate a Personal Access Token.")
	fmt.Println()

	// Collect username
	for {
		fmt.Print("Username: ")
		username, err = h.readLine()
		if err != nil {
			return "", "", fmt.Errorf("failed to read username: %w", err)
		}

		username = strings.TrimSpace(username)
		if username == "" {
			fmt.Println("❌ Username cannot be empty. Please enter your username.")
			fmt.Println()
			continue
		}

		break
	}

	// Collect password with secure input
	for {
		fmt.Print("Password: ")
		passwordBytes, err := h.readPasswordSecurely()
		if err != nil {
			return "", "", fmt.Errorf("failed to read password: %w", err)
		}

		password = string(passwordBytes)
		if password == "" {
			fmt.Println("❌ Password cannot be empty. Please enter your password.")
			fmt.Println()
			continue
		}

		break
	}

	fmt.Println("✅ Credentials collected successfully!")
	fmt.Println()
	return username, password, nil
}

// ConfirmOverwrite asks user for confirmation before overwriting existing config.
func (h *InputHandler) ConfirmOverwrite(existingPath string) (bool, error) {
	fmt.Println("⚠️  Existing Configuration Detected")
	fmt.Println("===================================")
	fmt.Println()
	fmt.Printf("A configuration file already exists at: %s\n", existingPath)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("1. Overwrite - Replace the existing configuration (a backup will be created)")
	fmt.Println("2. Cancel - Exit without making changes")
	fmt.Println()

	for {
		fmt.Print("Do you want to overwrite the existing configuration? (y/n): ")
		input, err := h.readLine()
		if err != nil {
			return false, fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.ToLower(strings.TrimSpace(input))
		switch input {
		case "y", "yes":
			fmt.Println("✅ Proceeding with configuration overwrite...")
			fmt.Println("📁 A backup of the existing configuration will be created.")
			fmt.Println()
			return true, nil
		case "n", "no":
			fmt.Println("❌ Configuration cancelled by user.")
			return false, nil
		default:
			fmt.Println("Please enter 'y' for yes or 'n' for no.")
		}
	}
}

// DisplaySuccess displays the success message with next steps.
func (h *InputHandler) DisplaySuccess(configPath string) {
	fmt.Println()
	fmt.Println("🎉 MCP Server initialization completed successfully!")
	fmt.Println("=====================================================")
	fmt.Println()
	fmt.Printf("📁 Configuration saved to: %s\n", configPath)
	fmt.Println()
	fmt.Println("🚀 Next steps:")
	fmt.Println("   1. You can now run the MCP server normally without the -i flag")
	fmt.Println("   2. The server will use the generated configuration automatically")
	fmt.Println("   3. Your PAT token is valid for 1 year from today")
	fmt.Println()
	fmt.Println("💻 To start the MCP server:")
	fmt.Printf("   mcp-server -config %s\n", configPath)
	fmt.Println()
	fmt.Println("📚 Additional information:")
	fmt.Println("   • Configuration file permissions are set to 600 (owner-only)")
	fmt.Println("   • PAT token provides secure access to the backend API")
	fmt.Println("   • You can regenerate the PAT token anytime through the web interface")
	fmt.Println("   • Backup of any existing configuration was created automatically")
	fmt.Println()
	fmt.Println("🔧 Troubleshooting:")
	fmt.Println("   • If the server fails to start, check the configuration file syntax")
	fmt.Println("   • Ensure the backend API server is running and accessible")
	fmt.Println("   • Verify network connectivity and firewall settings")
	fmt.Println()
}

// readLine reads a line of input from the user.
func (h *InputHandler) readLine() (string, error) {
	line, err := h.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// testConnectivity tests if the server is reachable by making a GET request to /ready endpoint.
func (h *InputHandler) testConnectivity(baseURL string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Try the /ready endpoint first
	readyURL := strings.TrimSuffix(baseURL, "/") + "/ready"
	resp, err := client.Get(readyURL)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d, expected 200", resp.StatusCode)
	}

	return nil
}

// askRetry prompts the user for a yes/no confirmation.
func (h *InputHandler) askRetry(question string) bool {
	for {
		fmt.Printf("%s (y/n): ", question)
		input, err := h.readLine()
		if err != nil {
			fmt.Println("Error reading input, assuming 'no'")
			return false
		}

		input = strings.ToLower(input)
		switch input {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please enter 'y' for yes or 'n' for no.")
		}
	}
}

// readPasswordSecurely reads a password from stdin without echoing characters to the terminal.
func (h *InputHandler) readPasswordSecurely() ([]byte, error) {
	// Get the file descriptor for stdin
	fd := int(syscall.Stdin)

	// Check if stdin is a terminal
	if !term.IsTerminal(fd) {
		// If not a terminal (e.g., input is piped), read normally but warn user
		fmt.Println("Warning: Input is not from a terminal. Password will not be hidden.")
		password, err := h.readLine()
		if err != nil {
			return nil, err
		}
		return []byte(password), nil
	}

	// Read password with terminal echo disabled
	passwordBytes, err := term.ReadPassword(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to read password securely: %w", err)
	}

	// Print a newline since ReadPassword doesn't echo the Enter key
	fmt.Println()

	return passwordBytes, nil
}

// AskRetryWithOptions prompts the user with multiple retry options.
func (h *InputHandler) AskRetryWithOptions(operation string, err error) (bool, error) {
	fmt.Printf("❌ %s failed: %v\n", operation, err)
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("1. Retry - Try the operation again")
	fmt.Println("2. Cancel - Exit the initialization process")
	fmt.Println()

	for {
		fmt.Print("Would you like to retry? (y/n): ")
		input, readErr := h.readLine()
		if readErr != nil {
			return false, fmt.Errorf("failed to read input: %w", readErr)
		}

		input = strings.ToLower(strings.TrimSpace(input))
		switch input {
		case "y", "yes":
			fmt.Println("🔄 Retrying operation...")
			fmt.Println()
			return true, nil
		case "n", "no":
			fmt.Println("❌ Operation cancelled by user.")
			return false, nil
		default:
			fmt.Println("Please enter 'y' for yes or 'n' for no.")
		}
	}
}
