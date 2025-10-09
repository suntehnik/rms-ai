# Design Document

## Overview

This design document outlines the implementation of Model Context Protocol (MCP) server integration for the Product Requirements Management System. The solution consists of two main components:

1. **MCP Server** - A standalone Go console application that implements the MCP protocol and communicates via STDIO
2. **Backend API Handler** - An HTTP endpoint in the existing backend that processes MCP requests from the server

The design leverages existing infrastructure including PAT authentication, service layer architecture, and database models while providing a standardized interface for AI agents to interact with the requirements management system.

## Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph "AI Host Environment"
        Claude[Claude Desktop]
        ClaudeConfig[claude_desktop_config.json]
    end
    
    subgraph "MCP Server Process"
        MCPServer[MCP Server<br/>Go Console App]
        Config[~/.requirements-mcp/config.json]
        STDIO[STDIO Transport]
        HTTPClient[HTTP Client]
    end
    
    subgraph "Backend API"
        MCPEndpoint[/api/v1/mcp]
        PATAuth[PAT Authentication]
        JSONRPCHandler[JSON-RPC Handler]
        MCPHandler[MCP Request Handler]
        Services[Service Layer]
        Database[(PostgreSQL)]
    end
    
    Claude -->|Launch Process| MCPServer
    ClaudeConfig -->|Process Launch| MCPServer
    MCPServer -->|Read Config| Config
    Claude <-->|JSON-RPC Messages| STDIO
    STDIO -->|Raw JSON| HTTPClient
    HTTPClient -->|HTTP POST + PAT| MCPEndpoint
    MCPEndpoint --> PATAuth
    PATAuth --> JSONRPCHandler
    JSONRPCHandler --> MCPHandler
    MCPHandler --> Services
    Services --> Database
    
    MCPEndpoint -->|JSON Response| HTTPClient
    HTTPClient -->|Success| STDIO
    HTTPClient -->|Errors| STDERR[STDERR Stream]
```

### Component Interaction Flow

1. **Process Launch**: Claude Desktop launches MCP Server console app based on `claude_desktop_config.json`
2. **Configuration Loading**: MCP Server reads its own config from `~/.requirements-mcp/config.json`
3. **Message Relay**: MCP Server acts as transport layer, forwarding JSON-RPC messages from STDIN
4. **Backend Processing**: All JSON-RPC logic handled in `/api/v1/mcp` endpoint with PAT authentication
5. **Response Relay**: MCP Server forwards backend responses to STDOUT
6. **Error Handling**: Connection/auth errors sent to STDERR stream

## Components and Interfaces

### MCP Server Component (Console App)

#### Core Structure
```go
// cmd/mcp-server/main.go
type MCPServer struct {
    config     *Config
    httpClient *http.Client
    logger     *logrus.Logger
}

type Config struct {
    BackendAPIURL  string `json:"backend_api_url"`
    PATToken       string `json:"pat_token"`
    RequestTimeout string `json:"request_timeout"`
    LogLevel       string `json:"log_level"`
}

func (s *MCPServer) Run() error {
    // Read from STDIN, forward to backend, write to STDOUT/STDERR
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        message := scanner.Bytes()
        response, err := s.forwardToBackend(message)
        if err != nil {
            s.writeError(err)
            continue
        }
        s.writeResponse(response)
    }
    return scanner.Err()
}
```

#### Configuration Loading
```go
// internal/mcp/config.go
func LoadConfig() (*Config, error) {
    configPath := filepath.Join(os.Getenv("HOME"), ".requirements-mcp", "config.json")
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    
    return &config, nil
}
```

#### Message Forwarding
```go
// internal/mcp/transport.go
func (s *MCPServer) forwardToBackend(message []byte) ([]byte, error) {
    // Create HTTP request with raw JSON message
    req, err := http.NewRequest("POST", s.config.BackendAPIURL+"/api/v1/mcp", bytes.NewReader(message))
    if err != nil {
        return nil, err
    }
    
    // Add required headers - Content-Type and Authorization only
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+s.config.PATToken)
    
    // Send request to backend
    resp, err := s.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("backend connection failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Return raw response body without modification
    return io.ReadAll(resp.Body)
}

func (s *MCPServer) writeResponse(data []byte) {
    // Write response to STDOUT as-is
    os.Stdout.Write(data)
}

func (s *MCPServer) writeError(err error) {
    // Write errors to STDERR
    fmt.Fprintf(os.Stderr, "MCP Server Error: %v\n", err)
}
```

#### Configuration File Format
```json
// ~/.requirements-mcp/config.json
{
    "backend_api_url": "http://localhost:8080",
    "pat_token": "mcp_pat_xxxxxxxxxxxxx",
    "request_timeout": "30s",
    "log_level": "info"
}
```

### Backend API Handler Component

The backend MCP handler is already implemented in the existing codebase. The console application will send JSON-RPC messages to the existing `/api/v1/mcp` endpoint which handles all MCP protocol logic including:

- JSON-RPC 2.0 message processing
- Resource operations (resources/list, resources/read, resources/templates/list)
- Tool operations (tools/list, tools/call)
- Prompt operations (prompts/list, prompts/get)
- Authentication and authorization via PAT middleware
- Error handling and response formatting

The console app only needs to:
1. Read JSON from STDIN
2. Add `Content-Type: application/json` header
3. Add `Authorization: Bearer <pat_token>` header
4. POST to `/api/v1/mcp` endpoint
5. Forward response to STDOUT or errors to STDERR

### URI Schemes (Backend Implementation)

The backend already implements URI parsing and routing for MCP resources. The console app doesn't need to understand these schemes - it just forwards the JSON messages. For reference, the supported URI schemes are:

```
// Direct Resources
epic://EP-001
epic://EP-001/hierarchy
user-story://US-001
requirement://REQ-001
acceptance-criteria://AC-001

// Resource Templates
epics://list?status={status}&priority={priority}
user-stories://list?epic_id={epic_id}&status={status}
requirements://search?query={query}&type={type}
```

## Data Models

### Console App Configuration Model

The console application only needs a simple configuration model:

```go
type Config struct {
    BackendAPIURL  string `json:"backend_api_url"`
    PATToken       string `json:"pat_token"`
    RequestTimeout string `json:"request_timeout"`
    LogLevel       string `json:"log_level"`
}
```

### MCP Protocol Models

All MCP protocol models (ServerCapabilities, ResourceContent, ToolResult, etc.) are implemented in the backend. The console app treats all JSON-RPC messages as opaque byte arrays and forwards them without parsing or modification.

## Error Handling

### Error Classification

#### JSON-RPC Standard Errors
```go
const (
    ParseError     = -32700 // Invalid JSON
    InvalidRequest = -32600 // Invalid Request object
    MethodNotFound = -32601 // Method not found
    InvalidParams  = -32602 // Invalid method parameters
    InternalError  = -32603 // Internal JSON-RPC error
)
```

#### Custom MCP Errors
```go
const (
    ResourceNotFound    = -32000 // Resource not found
    ToolExecutionError  = -32001 // Tool execution failed
    AuthenticationError = -32002 // Authentication failed
    AuthorizationError  = -32003 // Authorization failed
    BackendError        = -32004 // Backend API error
    ValidationError     = -32005 // Input validation error
)
```

#### Error Response Structure
```go
type JSONRPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

type ErrorDetails struct {
    RequestID   string            `json:"request_id,omitempty"`
    Timestamp   time.Time         `json:"timestamp"`
    Context     map[string]string `json:"context,omitempty"`
    Suggestions []string          `json:"suggestions,omitempty"`
}
```

### Error Handling Strategy

1. **Input Validation**: Validate all JSON-RPC messages and parameters before processing
2. **Authentication Errors**: Return appropriate codes for PAT validation failures
3. **Backend Errors**: Map HTTP status codes to JSON-RPC error codes
4. **Resource Errors**: Handle entity not found and access denied scenarios
5. **Tool Errors**: Validate tool arguments and handle execution failures
6. **Logging**: Log all errors with context for debugging and monitoring

## Testing Strategy

### Unit Testing

#### MCP Server Tests
```go
// internal/mcp/processor_test.go
func TestJSONRPCProcessor_Initialize(t *testing.T)
func TestJSONRPCProcessor_HandleResourcesList(t *testing.T)
func TestJSONRPCProcessor_HandleToolCall(t *testing.T)
func TestJSONRPCProcessor_HandlePromptGet(t *testing.T)

// internal/mcp/uri_parser_test.go
func TestURIParser_ParseEpicURI(t *testing.T)
func TestURIParser_ParseResourceTemplate(t *testing.T)
func TestURIParser_ValidateURI(t *testing.T)
```

#### Backend Handler Tests
```go
// internal/handlers/mcp_handler_test.go
func TestMCPHandler_HandleResourceRead(t *testing.T)
func TestMCPHandler_HandleToolCall(t *testing.T)
func TestMCPHandler_AuthenticationValidation(t *testing.T)

// internal/handlers/mcp_resource_handler_test.go
func TestMCPResourceHandler_GetEpicResource(t *testing.T)
func TestMCPResourceHandler_GetHierarchyResource(t *testing.T)
```

### Integration Testing

#### End-to-End Flow Tests
```go
// tests/integration/mcp_integration_test.go
func TestMCPServer_FullWorkflow(t *testing.T) {
    // Test complete flow from STDIO to database
    // 1. Start MCP server process
    // 2. Send initialize request
    // 3. Execute resource/tool/prompt operations
    // 4. Verify database state changes
    // 5. Cleanup
}

func TestMCPServer_AuthenticationFlow(t *testing.T)
func TestMCPServer_ErrorHandling(t *testing.T)
```

#### STDIO Communication Tests
```go
func TestSTDIOTransport_MessageExchange(t *testing.T)
func TestSTDIOTransport_LargeMessages(t *testing.T)
func TestSTDIOTransport_ErrorRecovery(t *testing.T)
```

### Performance Testing

#### Load Testing
```go
func BenchmarkMCPServer_ResourceOperations(b *testing.B)
func BenchmarkMCPServer_ToolOperations(b *testing.B)
func BenchmarkMCPServer_ConcurrentRequests(b *testing.B)
```

#### Memory and Resource Testing
```go
func TestMCPServer_MemoryUsage(t *testing.T)
func TestMCPServer_ConnectionPooling(t *testing.T)
func TestMCPServer_GracefulShutdown(t *testing.T)
```

## Security Considerations

### Authentication and Authorization

1. **PAT Validation**: All backend requests must include valid PAT tokens
2. **User Context**: Operations execute with permissions of PAT owner
3. **Token Security**: PAT tokens are redacted from logs and error messages
4. **Rate Limiting**: Apply existing rate limits to MCP endpoints

### Input Validation

1. **JSON Schema Validation**: Validate tool arguments against defined schemas
2. **URI Validation**: Validate resource URIs and prevent path traversal
3. **Parameter Sanitization**: Sanitize all user inputs before database queries
4. **Size Limits**: Implement limits on request/response sizes

### Error Information Disclosure

1. **Error Sanitization**: Avoid exposing internal system details in error messages
2. **Logging Security**: Log security events for monitoring and auditing
3. **Debug Information**: Restrict debug information to development environments

## Performance Optimization

### Caching Strategy

#### Metadata Caching
```go
type MetadataCache struct {
    requirementTypes   map[uuid.UUID]*models.RequirementType
    relationshipTypes  map[uuid.UUID]*models.RelationshipType
    statusModels       map[string]*models.StatusModel
    ttl               time.Duration
    mutex             sync.RWMutex
}
```

#### Resource Caching
```go
type ResourceCache struct {
    resources map[string]*CachedResource
    ttl       time.Duration
    maxSize   int
    mutex     sync.RWMutex
}

type CachedResource struct {
    Content   *ResourceContent
    Timestamp time.Time
    ETag      string
}
```

### Connection Management

1. **HTTP Connection Pooling**: Reuse HTTP connections to backend API
2. **Database Connection Pooling**: Leverage existing GORM connection pool
3. **Request Batching**: Batch multiple resource requests when possible
4. **Streaming**: Use streaming for large resource responses

### Memory Management

1. **Request Size Limits**: Limit maximum request/response sizes
2. **Garbage Collection**: Optimize Go GC settings for server workload
3. **Resource Cleanup**: Implement proper cleanup for long-running operations
4. **Memory Monitoring**: Add metrics for memory usage tracking

## Deployment and Configuration

### Configuration Management

#### Configuration File Location
```bash
~/.requirements-mcp/config.json
```

#### Configuration File Format
```json
{
    "backend_api_url": "http://localhost:8080",
    "pat_token": "mcp_pat_xxxxxxxxxxxxx",
    "request_timeout": "30s",
    "log_level": "info"
}
```

#### Claude Desktop Configuration
```json
{
  "mcpServers": {
    "requirements-mcp-server": {
      "command": "/usr/local/bin/requirements-mcp-server",
      "args": []
    }
  }
}
```

**Note**: Claude Desktop only knows about the executable path. The MCP Server console app is responsible for reading its own configuration from `~/.requirements-mcp/config.json` after being launched.

#### Configuration Validation
```go
func (c *Config) Validate() error {
    if c.BackendAPIURL == "" {
        return errors.New("backend_api_url is required")
    }
    if c.PATToken == "" {
        return errors.New("pat_token is required")
    }
    if _, err := url.Parse(c.BackendAPIURL); err != nil {
        return fmt.Errorf("invalid backend_api_url: %w", err)
    }
    return nil
}
```

### Build and Distribution

#### Build Configuration
```makefile
# Makefile targets for MCP server
build-mcp-server:
	go build -o bin/requirements-mcp-server cmd/mcp-server/main.go

install-mcp-server: build-mcp-server
	cp bin/requirements-mcp-server /usr/local/bin/

test-mcp-server:
	go test ./internal/mcp/... -v
	go test ./tests/integration/mcp_... -v
```

#### Distribution Package
```
requirements-mcp-server/
├── bin/
│   └── requirements-mcp-server
├── docs/
│   ├── setup-guide.md
│   ├── claude-desktop-config.md
│   └── troubleshooting.md
├── examples/
│   ├── claude_desktop_config.json
│   └── sample-workflows.md
└── README.md
```

## Monitoring and Observability

### Logging Strategy

#### Structured Logging
```go
type MCPLogger struct {
    logger *logrus.Logger
    fields logrus.Fields
}

func (l *MCPLogger) LogRequest(method string, params interface{}) {
    l.logger.WithFields(logrus.Fields{
        "component": "mcp-server",
        "method":    method,
        "timestamp": time.Now(),
    }).Info("Processing MCP request")
}
```

#### Log Levels and Categories
- **ERROR**: Authentication failures, backend errors, critical failures
- **WARN**: Retry attempts, deprecated features, performance issues
- **INFO**: Request processing, successful operations, lifecycle events
- **DEBUG**: Detailed request/response data, internal state changes

### Metrics Collection

#### Performance Metrics
```go
type MCPMetrics struct {
    RequestCount     prometheus.Counter
    RequestDuration  prometheus.Histogram
    ErrorCount       prometheus.Counter
    ActiveConnections prometheus.Gauge
    CacheHitRate     prometheus.Gauge
}
```

#### Health Checks
```go
func (s *MCPServer) HealthCheck() error {
    // Check backend API connectivity
    // Verify PAT token validity
    // Check resource availability
    // Validate configuration
}
```

### Troubleshooting Support

#### Debug Endpoints (Development Only)
```go
// Debug endpoints for development
GET /debug/mcp/status     // Server status and configuration
GET /debug/mcp/metrics    // Performance metrics
GET /debug/mcp/cache      // Cache status and statistics
POST /debug/mcp/test      // Test MCP operations
```

#### Common Issues and Solutions
1. **Authentication Failures**: PAT token validation and renewal
2. **Connection Issues**: Backend API connectivity and timeouts
3. **Performance Problems**: Caching and connection pooling
4. **Protocol Errors**: JSON-RPC message validation and formatting

This design provides a comprehensive foundation for implementing the MCP server integration while maintaining security, performance, and maintainability standards.