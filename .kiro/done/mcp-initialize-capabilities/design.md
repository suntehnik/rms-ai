# Design Document

## Overview

This document outlines the design for implementing the MCP initialize method in the spexus requirements management system. The implementation will provide a complete MCP server initialization response that includes protocol version, server information, capabilities declaration, and system instructions for AI agents.

The design follows the Model Context Protocol specification and integrates with the existing Go-based spexus architecture using JSON-RPC 2.0 as the transport layer.

**Related User Story:** US-044 - MCP initialize: получение списка возможностей spexus  
**Related Requirements:** REQ-027, REQ-028, REQ-029, REQ-030, REQ-031, REQ-032, REQ-033, REQ-034, REQ-035

## Architecture

### High-Level Architecture

```
┌─────────────────┐    JSON-RPC 2.0     ┌──────────────────────┐
│   MCP Client    │ ◄─────────────────► │   spexus MCP Server  │
│                 │                     │                      │
│ - AI Agent      │                     │ - Initialize Handler │
│ - IDE           │                     │ - Capabilities Mgr   │
│ - CLI Tool      │                     │ - Resource Provider  │
└─────────────────┘                     └──────────────────────┘
                                                    │
                                                    ▼
                                        ┌──────────────────────┐
                                        │  spexus Core System  │
                                        │                      │
                                        │ - Requirements DB    │
                                        │ - User Management    │
                                        │ - Configuration      │
                                        └──────────────────────┘
```

### MCP Protocol Layer

The MCP server will implement the initialize method as part of the JSON-RPC 2.0 protocol handler. The initialize method is the first method called by clients to establish the connection and discover capabilities.

**Request Flow:**
1. Client sends initialize request with protocol version and client info
2. Server validates request format and protocol compatibility
3. Server generates capabilities based on current system state
4. Server returns complete initialize response with all required fields

## Components and Interfaces

### 1. MCP Initialize Handler

**Location:** `internal/handlers/mcp_initialize.go`

```go
type InitializeHandler struct {
    capabilitiesManager *CapabilitiesManager
    systemPromptProvider *SystemPromptProvider
    logger *logrus.Logger
}

type InitializeRequest struct {
    JSONRPC         string      `json:"jsonrpc"`
    ID              interface{} `json:"id"`
    Method          string      `json:"method"`
    Params          InitializeParams `json:"params"`
}

type InitializeParams struct {
    ProtocolVersion string     `json:"protocolVersion"`
    Capabilities    ClientCapabilities `json:"capabilities"`
    ClientInfo      ClientInfo `json:"clientInfo"`
}

type InitializeResponse struct {
    JSONRPC string           `json:"jsonrpc"`
    ID      interface{}      `json:"id"`
    Result  InitializeResult `json:"result"`
}

type InitializeResult struct {
    ProtocolVersion string           `json:"protocolVersion"`
    Capabilities    ServerCapabilities `json:"capabilities"`
    ServerInfo      ServerInfo       `json:"serverInfo"`
    Instructions    string           `json:"instructions"`
}
```

### 2. Capabilities Manager

**Location:** `internal/mcp/capabilities.go`

```go
type CapabilitiesManager struct {
    toolProvider     *ToolProvider
    promptProvider   *PromptProvider
}

type ServerCapabilities struct {
    Prompts   PromptsCapability   `json:"prompts"`
    Resources ResourcesCapability `json:"resources"`
    Tools     ToolsCapability     `json:"tools"`
}

type PromptsCapability struct {
    ListChanged bool `json:"listChanged"`
}

type ResourcesCapability struct {
    ListChanged bool `json:"listChanged"`
    Subscribe   bool `json:"subscribe,omitempty"`
}

type ToolsCapability struct {
    ListChanged bool `json:"listChanged"`
}


```

### 3. System Prompt Provider

**Location:** `internal/mcp/system_prompt.go`

```go
type SystemPromptProvider struct {
    promptService *services.PromptService
}

func (spp *SystemPromptProvider) GetInstructions(ctx context.Context) (string, error) {
    // Get active system prompt using PromptService.GetActive()
    // Return prompt content and description as instxt) method for reference
}   // Handle case when no active prompt exists
}
```



## Data Models

### Initialize Response Structure

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2025-03-26",
    "capabilities": {
      "prompts": {
        "listChanged": true
      },
      "resources": {
        "listChanged": true,
        "subscribe": true
      },
      "tools": {
        "listChanged": true
      }
    },
    "serverInfo": {
      "name": "spexus mcp",
      "title": "MCP server for requirements management system",
      "version": "1.0.0"
    },
    "instructions": "You are an AI assistant working with the spexus requirements management system..."
  }
}
```

### Server Information Constants

```go
const (
    ServerName    = "spexus mcp"
    ServerTitle   = "MCP server for requirements management system"
    ServerVersion = "1.0.0"
    ProtocolVersion = "2025-03-26"
)
```

### System Instructions Integration

The system instructions will be obtained from the active system prompt using `PromptService.GetActive()`. This ensures that the MCP initialize method returns the current active prompt content and description as system instructions, maintaining consistency with the prompt management system.
If `PromptService.GetActive()` return error or nothing - use empty sting for "instructions"

## Error Handling


### Validation Errors

```go
type InitializeError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Error codes
const (
    InvalidProtocolVersion = "INVALID_PROTOCOL_VERSION"
    MalformedRequest      = "MALFORMED_REQUEST"
    InternalError         = "INTERNAL_ERROR"
)
```

### Error Response Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": "INVALID_PROTOCOL_VERSION",
    "message": "Unsupported protocol version. Expected: 2025-03-26",
    "data": {
      "supported_versions": ["2025-03-26"],
      "received_version": "2024-01-01"
    }
  }
}
```

## Testing Strategy

### Unit Tests

1. **Initialize Handler Tests**
   - Valid request processing
   - Invalid protocol version handling
   - Malformed request handling
   - Response format validation

2. **Capabilities Manager Tests**
   - Capability generation accuracy
   - Dynamic capability updates
   - Resource availability reflection

3. **System Prompt Provider Tests**
   - Instructions content generation
   - Template rendering
   - Context-specific customization

### Integration Tests

1. **End-to-End Initialize Flow**
   - Complete request-response cycle
   - Database integration
   - Configuration loading

2. **MCP Protocol Compliance**
   - JSON-RPC 2.0 specification adherence
   - MCP protocol specification compliance
   - Cross-client compatibility

### Test Data

```go
// Test cases for initialize method
var initializeTestCases = []struct {
    name           string
    request        InitializeRequest
    expectedResult InitializeResult
    expectedError  *InitializeError
}{
    {
        name: "valid_initialize_request",
        request: InitializeRequest{
            JSONRPC: "2.0",
            ID:      1,
            Method:  "initialize",
            Params: InitializeParams{
                ProtocolVersion: "2025-03-26",
                ClientInfo: ClientInfo{
                    Name:    "test-client",
                    Version: "1.0.0",
                },
            },
        },
        expectedResult: InitializeResult{
            ProtocolVersion: "2025-03-26",
            ServerInfo: ServerInfo{
                Name:    "spexus mcp",
                Title:   "MCP server for requirements management system",
                Version: "1.0.0",
            },
            // ... capabilities and instructions
        },
    },
}
```

## Implementation Notes

### Configuration Integration

The initialize handler will integrate with the existing spexus configuration system to:
- Load server metadata from environment variables
- Access database connection for resource queries
- Retrieve system-specific customizations

### Performance Considerations

- Cache system instructions to avoid regeneration on each request
- Pre-compute capabilities structure during server startup
- Minimize database queries during initialize method execution

### Security Considerations

- Validate all input parameters to prevent injection attacks
- Sanitize client information before logging
- Ensure no sensitive information is exposed in instructions
- Implement rate limiting for initialize requests

### Monitoring and Logging

- Log all initialize requests with client information
- Track protocol version usage statistics
- Monitor response times and error rates
- Alert on unsupported protocol version attempts