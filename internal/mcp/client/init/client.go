package init

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// NetworkClient handles all HTTP communication with the backend API.
type NetworkClient struct {
	httpClient *http.Client
	baseURL    string
}

// AuthResponse represents the response from authentication endpoint.
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

// PATResponse represents the response from PAT creation endpoint.
type PATResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Name      string    `json:"name"`
}

// User represents user information from authentication.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// NewNetworkClient creates a new network client for the specified base URL.
func NewNetworkClient(baseURL string) *NetworkClient {
	return &NetworkClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// NewSecureNetworkClient creates a new network client with proper HTTPS certificate validation.
func NewSecureNetworkClient(baseURL string) *NetworkClient {
	client := CreateSecureHTTPClient()
	client.Timeout = 30 * time.Second

	return &NetworkClient{
		httpClient: client,
		baseURL:    baseURL,
	}
}

// TestConnectivity tests connection to the server by checking the /ready endpoint.
func (c *NetworkClient) TestConnectivity() error {
	// Construct the ready endpoint URL
	readyURL := c.baseURL + "/ready"

	// Make GET request to the ready endpoint
	resp, err := c.httpClient.Get(readyURL)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response status indicates the server is ready
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server not ready, status: %d", resp.StatusCode)
	}

	return nil
}

// LoginRequest represents the request body for authentication.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Authenticate performs username/password authentication and returns JWT token.
func (c *NetworkClient) Authenticate(username, password string) (*AuthResponse, error) {
	// Construct the login endpoint URL
	loginURL := c.baseURL + "/auth/login"

	// Create login request body
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	// Set content type header
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make login request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	// Parse response body
	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to parse authentication response: %w", err)
	}

	return &authResp, nil
}

// CreatePATRequest represents the request body for PAT creation.
type CreatePATRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at"`
	Scopes    []string   `json:"scopes"`
}

// CreatePATResponse represents the response from PAT creation.
type CreatePATResponse struct {
	Token string `json:"token"`
	PAT   struct {
		ID        string     `json:"id"`
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expires_at"`
		CreatedAt time.Time  `json:"created_at"`
	} `json:"pat"`
}

// CreatePAT creates a Personal Access Token using the provided JWT token.
func (c *NetworkClient) CreatePAT(jwtToken string) (*PATResponse, error) {
	// Construct the PAT creation endpoint URL
	patURL := c.baseURL + "/api/v1/pats"

	// Get hostname for token naming
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Create PAT request with 1-year expiration
	expiresAt := time.Now().AddDate(1, 0, 0) // 1 year from now
	patReq := CreatePATRequest{
		Name:      fmt.Sprintf("MCP Server - %s - %s", hostname, time.Now().Format("2006-01-02")),
		ExpiresAt: &expiresAt,
		Scopes:    []string{"full_access"},
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(patReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal PAT request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", patURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create PAT request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make PAT request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("PAT creation failed with status: %d", resp.StatusCode)
	}

	// Parse response body
	var createResp CreatePATResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return nil, fmt.Errorf("failed to parse PAT creation response: %w", err)
	}

	// Convert to our expected response format
	patResp := &PATResponse{
		Token:     createResp.Token,
		ExpiresAt: *createResp.PAT.ExpiresAt,
		Name:      createResp.PAT.Name,
	}

	return patResp, nil
}

// MCPRequest represents a JSON-RPC 2.0 request for MCP.
type MCPRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response from MCP.
type MCPResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      int       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC 2.0 error.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// InitializeParams represents the parameters for the initialize method.
type InitializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      ClientInfo     `json:"clientInfo"`
}

// ClientInfo represents client information for initialization.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ValidatePAT validates the generated PAT token by making an MCP initialize request.
func (c *NetworkClient) ValidatePAT(patToken string) error {
	// Construct the MCP endpoint URL
	mcpURL := c.baseURL + "/api/v1/mcp"

	// Create MCP initialize request
	initRequest := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: "2025-06-18",
			Capabilities:    map[string]any{},
			ClientInfo: ClientInfo{
				Name:    "mcp-init-client",
				Version: "1.0.0",
			},
		},
	}

	// Marshal request to JSON
	reqBody, err := json.Marshal(initRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal MCP request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", mcpURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create MCP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+patToken)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make MCP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == 401 {
		return fmt.Errorf("PAT validation failed: invalid or expired token")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("PAT validation failed with status: %d", resp.StatusCode)
	}

	// Parse response body
	var mcpResp MCPResponse
	if err := json.NewDecoder(resp.Body).Decode(&mcpResp); err != nil {
		return fmt.Errorf("failed to parse MCP response: %w", err)
	}

	// Check for JSON-RPC errors
	if mcpResp.Error != nil {
		return fmt.Errorf("MCP error: %s (code: %d)", mcpResp.Error.Message, mcpResp.Error.Code)
	}

	// Verify we got a valid response
	if mcpResp.Result == nil {
		return fmt.Errorf("invalid MCP response: missing result")
	}

	return nil
}
