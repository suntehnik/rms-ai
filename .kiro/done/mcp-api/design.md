# Design Document

## Overview

This document outlines the design for implementing a JSON-RPC API handler in the Product Requirements Management System. The handler will process Model Context Protocol (MCP) requests, providing access to requirements management capabilities through a standardized JSON-RPC 2.0 interface. The design integrates with existing authentication infrastructure and service layer architecture while maintaining security and performance standards.

## Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph "External MCP Clients"
        MCP1[MCP Server 1]
        MCP2[MCP Server 2]
        MCP3[Claude Desktop]
    end
    
    subgraph "API Gateway Layer"
        Router[Gin Router]
        Auth[PAT Middleware]
        CORS[CORS Middleware]
    end
    
    subgraph "JSON-RPC Handler"
        Endpoint[/api/v1/mcp]
        Parser[JSON-RPC Parser]
        Router2[Method Router]
    end
    
    subgraph "MCP Handlers"
        ResourceH[Resources Handler]
        ToolsH[Tools Handler]
        PromptsH[Prompts Handler]
    end
    
    subgraph "Service Layer"
        EpicSvc[Epic Service]
        UserStorySvc[User Story Service]
        ReqSvc[Requirement Service]
        SearchSvc[Search Service]
    end
    
    subgraph "Data Layer"
        DB[(PostgreSQL)]
        Redis[(Redis Cache)]
    end
    
    MCP1 --> Router
    MCP2 --> Router
    MCP3 --> Router
    
    Router --> Auth
    Auth --> CORS
    CORS --> Endpoint
    
    Endpoint --> Parser
    Parser --> Router2
    
    Router2 --> ResourceH
    Router2 --> ToolsH
    Router2 --> PromptsH
    
    ResourceH --> EpicSvc
    ResourceH --> UserStorySvc
    ResourceH --> ReqSvc
    
    ToolsH --> EpicSvc
    ToolsH --> UserStorySvc
    ToolsH --> ReqSvc
    ToolsH --> SearchSvc
    
    PromptsH --> EpicSvc
    PromptsH --> UserStorySvc
    PromptsH --> ReqSvc
    
    EpicSvc --> DB
    UserStorySvc --> DB
    ReqSvc --> DB
    SearchSvc --> DB
    SearchSvc --> Redis
```

### Request Flow

1. **Authentication**: PAT validation using existing middleware
2. **JSON-RPC Parsing**: Parse and validate JSON-RPC 2.0 message structure
3. **Method Routing**: Route to appropriate handler based on method name
4. **Handler Execution**: Execute business logic using service layer
5. **Response Formatting**: Format response according to JSON-RPC 2.0 specification

## Components and Interfaces

### Core Components

#### 1. MCP Handler (`internal/handlers/mcp_handler.go`)

```go
type MCPHandler struct {
    resourceHandler *ResourceHandler
    toolsHandler    *ToolsHandler
    promptsHandler  *PromptsHandler
    logger          *logrus.Logger
}

type JSONRPCRequest struct {
    JSONRPC string      `json:"jsonrpc" validate:"required,eq=2.0"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method" validate:"required"`
    Params  interface{} `json:"params"`
}

type JSONRPCResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

#### 2. Resource Handler (`internal/handlers/mcp_resource_handler.go`)

```go
type ResourceHandler struct {
    epicService           *service.EpicService
    userStoryService      *service.UserStoryService
    requirementService    *service.RequirementService
    acceptanceCriteriaService *service.AcceptanceCriteriaService
    uriParser             *URIParser
}

type ResourceReadRequest struct {
    URI string `json:"uri" validate:"required"`
}

type ResourceResponse struct {
    URI         string      `json:"uri"`
    Name        string      `json:"name"`
    Description string      `json:"description"`
    MimeType    string      `json:"mimeType"`
    Contents    interface{} `json:"contents"`
}

// Service integration methods for reference ID resolution
func (rh *ResourceHandler) getEntityByReferenceID(ctx context.Context, scheme, referenceID string) (interface{}, error) {
    switch scheme {
    case "epic":
        return rh.epicService.GetByReferenceID(ctx, referenceID)
    case "user-story":
        return rh.userStoryService.GetByReferenceID(ctx, referenceID)
    case "requirement":
        return rh.requirementService.GetByReferenceID(ctx, referenceID)
    case "acceptance-criteria":
        return rh.acceptanceCriteriaService.GetByReferenceID(ctx, referenceID)
    default:
        return nil, fmt.Errorf("unsupported entity scheme: %s", scheme)
    }
}
```

#### 3. Tools Handler (`internal/handlers/mcp_tools_handler.go`)

```go
type ToolsHandler struct {
    epicService        *service.EpicService
    userStoryService   *service.UserStoryService
    requirementService *service.RequirementService
    searchService      *service.SearchService
}

type ToolCallRequest struct {
    Name      string                 `json:"name" validate:"required"`
    Arguments map[string]interface{} `json:"arguments"`
}

type ToolResponse struct {
    Content []ContentItem `json:"content"`
}

type ContentItem struct {
    Type string      `json:"type"`
    Text string      `json:"text,omitempty"`
    Data interface{} `json:"data,omitempty"`
}
```

#### 4. Prompts Handler (`internal/handlers/mcp_prompts_handler.go`)

```go
type PromptsHandler struct {
    promptService             *service.PromptService
    epicService               *service.EpicService
    userStoryService          *service.UserStoryService
    requirementService        *service.RequirementService
    acceptanceCriteriaService *service.AcceptanceCriteriaService
    logger                    *logrus.Logger
}

// MCP protocol request/response structures
type PromptListRequest struct {
    // No parameters needed for listing prompts
}

type PromptGetRequest struct {
    Name string `json:"name" validate:"required"`
}

type PromptListResponse struct {
    Prompts []*MCPPromptDescriptor `json:"prompts"`
}

type PromptGetResponse struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Messages    []PromptMessage `json:"messages"`
}

// MCP Tools for prompt management
type CreatePromptRequest struct {
    Name        string `json:"name" validate:"required"`
    Title       string `json:"title" validate:"required"`
    Description string `json:"description"`
    Content     string `json:"content" validate:"required"`
}

type UpdatePromptRequest struct {
    PromptID    string `json:"prompt_id" validate:"required"`
    Title       string `json:"title,omitempty"`
    Description string `json:"description,omitempty"`
    Content     string `json:"content,omitempty"`
}

type ActivatePromptRequest struct {
    PromptID string `json:"prompt_id" validate:"required"`
}

// Context loading methods
func (ph *PromptsHandler) loadEntityContext(ctx context.Context, entityType, referenceID string) (interface{}, error) {
    switch entityType {
    case "epic":
        return ph.epicService.GetByReferenceID(ctx, referenceID)
    case "user_story":
        return ph.userStoryService.GetByReferenceID(ctx, referenceID)
    case "requirement":
        return ph.requirementService.GetByReferenceID(ctx, referenceID)
    case "acceptance_criteria":
        return ph.acceptanceCriteriaService.GetByReferenceID(ctx, referenceID)
    default:
        return nil, fmt.Errorf("unsupported entity type: %s", entityType)
    }
}

// Template generation methods
func (ph *PromptsHandler) generatePromptMessages(ctx context.Context, prompt *Prompt) ([]PromptMessage, error) {
    // Generate structured prompt messages from prompt content
    // Return formatted messages for MCP protocol
    return []PromptMessage{
        {
            Role:    "system",
            Content: prompt.Content,
        },
    }, nil
}
```

#### 5. Prompt Service (`internal/service/prompt_service.go`)

```go
type PromptService struct {
    db     *gorm.DB
    logger *logrus.Logger
}

func NewPromptService(db *gorm.DB, logger *logrus.Logger) *PromptService {
    return &PromptService{
        db:     db,
        logger: logger,
    }
}

// CRUD operations
func (ps *PromptService) Create(ctx context.Context, req *CreatePromptRequest, creatorID uuid.UUID) (*Prompt, error) {
    prompt := &Prompt{
        Name:        req.Name,
        Title:       req.Title,
        Description: req.Description,
        Content:     req.Content,
        CreatorID:   creatorID,
        IsActive:    false, // New prompts are not active by default
    }
    
    if err := ps.db.WithContext(ctx).Create(prompt).Error; err != nil {
        return nil, err
    }
    
    return prompt, nil
}

func (ps *PromptService) GetByReferenceID(ctx context.Context, referenceID string) (*Prompt, error) {
    var prompt Prompt
    if err := ps.db.WithContext(ctx).Where("reference_id = ?", referenceID).First(&prompt).Error; err != nil {
        return nil, err
    }
    return &prompt, nil
}

func (ps *PromptService) GetByName(ctx context.Context, name string) (*Prompt, error) {
    var prompt Prompt
    if err := ps.db.WithContext(ctx).Where("name = ?", name).First(&prompt).Error; err != nil {
        return nil, err
    }
    return &prompt, nil
}

func (ps *PromptService) GetActive(ctx context.Context) (*Prompt, error) {
    var prompt Prompt
    if err := ps.db.WithContext(ctx).Where("is_active = ?", true).First(&prompt).Error; err != nil {
        return nil, err
    }
    return &prompt, nil
}

func (ps *PromptService) List(ctx context.Context, limit, offset int) ([]*Prompt, int64, error) {
    var prompts []*Prompt
    var total int64
    
    if err := ps.db.WithContext(ctx).Model(&Prompt{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    if err := ps.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&prompts).Error; err != nil {
        return nil, 0, err
    }
    
    return prompts, total, nil
}

func (ps *PromptService) Update(ctx context.Context, id uuid.UUID, req *UpdatePromptRequest) (*Prompt, error) {
    var prompt Prompt
    if err := ps.db.WithContext(ctx).Where("id = ?", id).First(&prompt).Error; err != nil {
        return nil, err
    }
    
    // Update fields
    if req.Title != "" {
        prompt.Title = req.Title
    }
    if req.Description != "" {
        prompt.Description = req.Description
    }
    if req.Content != "" {
        prompt.Content = req.Content
    }
    
    if err := ps.db.WithContext(ctx).Save(&prompt).Error; err != nil {
        return nil, err
    }
    
    return &prompt, nil
}

func (ps *PromptService) Activate(ctx context.Context, id uuid.UUID) error {
    return ps.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // Deactivate all prompts
        if err := tx.Model(&Prompt{}).Update("is_active", false).Error; err != nil {
            return err
        }
        
        // Activate selected prompt
        if err := tx.Model(&Prompt{}).Where("id = ?", id).Update("is_active", true).Error; err != nil {
            return err
        }
        
        return nil
    })
}

func (ps *PromptService) Delete(ctx context.Context, id uuid.UUID) error {
    return ps.db.WithContext(ctx).Where("id = ?", id).Delete(&Prompt{}).Error
}

// Helper methods for prompt management
```

#### 6. REST API Handler (`internal/handlers/prompt_handler.go`)

```go
type PromptHandler struct {
    promptService *service.PromptService
    logger        *logrus.Logger
}

func NewPromptHandler(promptService *service.PromptService, logger *logrus.Logger) *PromptHandler {
    return &PromptHandler{
        promptService: promptService,
        logger:        logger,
    }
}

// REST API endpoints
func (ph *PromptHandler) CreatePrompt(c *gin.Context) {
    // POST /api/v1/prompts
    // Requires Administrator role
}

func (ph *PromptHandler) ListPrompts(c *gin.Context) {
    // GET /api/v1/prompts
    // Supports pagination with limit/offset
}

func (ph *PromptHandler) GetPrompt(c *gin.Context) {
    // GET /api/v1/prompts/:id
    // Accepts both UUID and reference ID
}

func (ph *PromptHandler) UpdatePrompt(c *gin.Context) {
    // PUT /api/v1/prompts/:id
    // Requires Administrator role
}

func (ph *PromptHandler) DeletePrompt(c *gin.Context) {
    // DELETE /api/v1/prompts/:id
    // Requires Administrator role
}

func (ph *PromptHandler) ActivatePrompt(c *gin.Context) {
    // PATCH /api/v1/prompts/:id/activate
    // Requires Administrator role
}

func (ph *PromptHandler) GetActivePrompt(c *gin.Context) {
    // GET /api/v1/prompts/active
    // Returns currently active prompt
}
```

### Database Migration

#### Migration: `000007_add_prompts_table.up.sql`

```sql
-- Add prompts table for system prompt management
CREATE SEQUENCE prompt_ref_seq START 1;

CREATE TABLE prompts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reference_id VARCHAR(20) UNIQUE NOT NULL DEFAULT ('PROMPT-' || LPAD(nextval('prompt_ref_seq')::TEXT, 3, '0')),
    name VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for prompts
CREATE INDEX idx_prompts_reference ON prompts(reference_id);
CREATE INDEX idx_prompts_name ON prompts(name);
CREATE INDEX idx_prompts_is_active ON prompts(is_active);
CREATE INDEX idx_prompts_creator ON prompts(creator_id);
CREATE INDEX idx_prompts_created_at ON prompts(created_at);

-- Ensure only one prompt can be active at a time
CREATE UNIQUE INDEX idx_prompts_single_active ON prompts(is_active) WHERE is_active = true;

-- Add trigger for updated_at
CREATE TRIGGER update_prompts_updated_at BEFORE UPDATE ON prompts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default system prompt
INSERT INTO prompts (name, title, description, content, is_active, creator_id) 
SELECT 
    'requirements-analyst',
    'Requirements Analyst Assistant',
    'AI assistant specialized in requirements analysis and management',
    'You are an expert requirements analyst working with a Product Requirements Management System. Your role is to help users create, analyze, and manage requirements through a hierarchical structure: Epics (high-level features), User Stories (specific user needs), Acceptance Criteria (testable conditions), and Requirements (detailed specifications). You have access to tools for CRUD operations and can analyze requirement quality, suggest improvements, and identify dependencies. Always focus on clarity, testability, and traceability.',
    true,
    (SELECT id FROM users WHERE role = 'Administrator' LIMIT 1);
```

#### Migration: `000007_add_prompts_table.down.sql`

```sql
-- Remove prompts table
DROP TABLE IF EXISTS prompts;
DROP SEQUENCE IF EXISTS prompt_ref_seq;
```

### URI Schemes and Resource Mapping

#### Supported URI Schemes

```go
const (
    EpicURIScheme              = "epic"
    UserStoryURIScheme         = "user-story"
    RequirementURIScheme       = "requirement"
    AcceptanceCriteriaURIScheme = "acceptance-criteria"
)

// URI Patterns (using reference ID format only):
// epic://EP-001
// epic://EP-001/hierarchy
// user-story://US-001
// user-story://US-001/requirements
// requirement://REQ-001
// requirement://REQ-001/relationships
// acceptance-criteria://AC-001

// Reference ID Patterns:
// EP-001, EP-002, ... (Epics)
// US-001, US-002, ... (User Stories)
// REQ-001, REQ-002, ... (Requirements)
// AC-001, AC-002, ... (Acceptance Criteria)
```

#### URI Parser

```go
type URIParser struct{}

type ParsedURI struct {
    Scheme      string
    ReferenceID string    // Reference ID format (EP-001, US-001, etc.)
    SubPath     string
    Parameters  map[string]string
}

func (p *URIParser) Parse(uri string) (*ParsedURI, error) {
    // Parse URI components
    // Validate scheme
    // Extract reference ID
    // Validate reference ID format
    // Parse sub-paths and parameters
}

func (p *URIParser) isValidReferenceID(id string) bool {
    // Check if ID matches reference ID patterns:
    // EP-001, US-001, REQ-001, AC-001
    referenceIDPattern := regexp.MustCompile(`^(EP|US|REQ|AC)-\d+$`)
    return referenceIDPattern.MatchString(id)
}

func (p *URIParser) validateSchemeAndReferenceID(scheme, referenceID string) error {
    // Validate that reference ID prefix matches the scheme
    expectedPrefixes := map[string]string{
        "epic":                "EP",
        "user-story":          "US", 
        "requirement":         "REQ",
        "acceptance-criteria": "AC",
    }
    
    expectedPrefix, exists := expectedPrefixes[scheme]
    if !exists {
        return fmt.Errorf("unsupported scheme: %s", scheme)
    }
    
    if !strings.HasPrefix(referenceID, expectedPrefix+"-") {
        return fmt.Errorf("reference ID %s does not match scheme %s (expected prefix: %s)", referenceID, scheme, expectedPrefix)
    }
    
    return nil
}
```

### Tool Definitions

#### CRUD Tools

```go
var SupportedTools = map[string]ToolDefinition{
    "create_epic": {
        Name:        "create_epic",
        Description: "Create a new epic",
        InputSchema: epicCreateSchema,
    },
    "update_epic": {
        Name:        "update_epic", 
        Description: "Update an existing epic",
        InputSchema: epicUpdateSchema,
    },
    "create_user_story": {
        Name:        "create_user_story",
        Description: "Create a new user story",
        InputSchema: userStoryCreateSchema,
    },
    "update_user_story": {
        Name:        "update_user_story",
        Description: "Update an existing user story", 
        InputSchema: userStoryUpdateSchema,
    },
    "create_requirement": {
        Name:        "create_requirement",
        Description: "Create a new requirement",
        InputSchema: requirementCreateSchema,
    },
    "update_requirement": {
        Name:        "update_requirement",
        Description: "Update an existing requirement",
        InputSchema: requirementUpdateSchema,
    },
    "create_relationship": {
        Name:        "create_relationship",
        Description: "Create a relationship between requirements",
        InputSchema: relationshipCreateSchema,
    },
    "search_global": {
        Name:        "search_global",
        Description: "Search across all entities",
        InputSchema: globalSearchSchema,
    },
    "search_requirements": {
        Name:        "search_requirements", 
        Description: "Search requirements with filters",
        InputSchema: requirementSearchSchema,
    },
}
```

### System Prompts Database Model

```go
// Database model for system prompts
type Prompt struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    ReferenceID string    `gorm:"type:varchar(20);unique;not null" json:"reference_id"`
    Name        string    `gorm:"type:varchar(255);unique;not null" json:"name"`
    Title       string    `gorm:"type:varchar(500);not null" json:"title"`
    Description string    `gorm:"type:text" json:"description"`
    Content     string    `gorm:"type:text;not null" json:"content"`
    IsActive    bool      `gorm:"default:false" json:"is_active"`
    CreatorID   uuid.UUID `gorm:"type:uuid;not null" json:"creator_id"`
    CreatedAt   time.Time `gorm:"default:now()" json:"created_at"`
    UpdatedAt   time.Time `gorm:"default:now()" json:"updated_at"`
    
    // Relationships
    Creator *User `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}

// MCP Protocol structures
type MCPPromptDescriptor struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type MCPPromptDefinition struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Messages    []PromptMessage `json:"messages"`
}

type PromptMessage struct {
    Role    string `json:"role"`    // "system", "user", "assistant"
    Content string `json:"content"` // Simple text content
}

// Database constraints and hooks
func (p *Prompt) BeforeCreate(tx *gorm.DB) error {
    if p.ReferenceID == "" {
        referenceID, err := promptGenerator.Generate(tx, &Prompt{})
        if err != nil {
            return err
        }
        p.ReferenceID = referenceID
    }
    return nil
}

func (p *Prompt) BeforeUpdate(tx *gorm.DB) error {
    // If setting this prompt as active, deactivate all others
    if p.IsActive {
        if err := tx.Model(&Prompt{}).Where("id != ?", p.ID).Update("is_active", false).Error; err != nil {
            return err
        }
    }
    return nil
}

// Reference ID generator for prompts
var promptGenerator = NewPostgreSQLReferenceIDGenerator(2147483643, "PROMPT")
```

## Data Models

### JSON-RPC Message Models

```go
// Standard JSON-RPC 2.0 structures
type JSONRPCRequest struct {
    JSONRPC string      `json:"jsonrpc" validate:"required,eq=2.0"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method" validate:"required"`
    Params  interface{} `json:"params"`
}

type JSONRPCResponse struct {
    JSONRPC string        `json:"jsonrpc"`
    ID      interface{}   `json:"id"`
    Result  interface{}   `json:"result,omitempty"`
    Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

### MCP-Specific Models

```go
// Resource models
type Resource struct {
    URI         string `json:"uri"`
    Name        string `json:"name"`
    Description string `json:"description"`
    MimeType    string `json:"mimeType"`
}

type ResourceContents struct {
    URI      string      `json:"uri"`
    MimeType string      `json:"mimeType"`
    Contents interface{} `json:"contents"`
}

// Tool models
type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema interface{} `json:"inputSchema"`
}

type ToolResult struct {
    Content []ContentItem `json:"content"`
}

// Prompt models
type Prompt struct {
    Name        string           `json:"name"`
    Description string           `json:"description"`
    Arguments   []PromptArgument `json:"arguments"`
}

type PromptArgument struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Type        string `json:"type"`
    Required    bool   `json:"required"`
}
```

### Error Code Mapping

```go
const (
    // Standard JSON-RPC error codes
    ParseError     = -32700
    InvalidRequest = -32600
    MethodNotFound = -32601
    InvalidParams  = -32602
    InternalError  = -32603
    
    // Custom MCP error codes
    ResourceNotFound    = -32001
    UnauthorizedAccess  = -32002
    ValidationError     = -32003
    ServiceUnavailable  = -32004
    RateLimitExceeded   = -32005
)

var ErrorMessages = map[int]string{
    ParseError:          "Parse error",
    InvalidRequest:      "Invalid Request", 
    MethodNotFound:      "Method not found",
    InvalidParams:       "Invalid params",
    InternalError:       "Internal error",
    ResourceNotFound:    "Resource not found",
    UnauthorizedAccess:  "Unauthorized access",
    ValidationError:     "Validation error",
    ServiceUnavailable:  "Service unavailable",
    RateLimitExceeded:   "Rate limit exceeded",
}
```

## Error Handling

### Error Mapping Strategy

```go
type ErrorMapper struct{}

func (em *ErrorMapper) MapServiceError(err error) *JSONRPCError {
    switch {
    case errors.Is(err, service.ErrNotFound):
        return &JSONRPCError{
            Code:    ResourceNotFound,
            Message: "Resource not found",
            Data:    err.Error(),
        }
    case errors.Is(err, service.ErrUnauthorized):
        return &JSONRPCError{
            Code:    UnauthorizedAccess,
            Message: "Unauthorized access",
            Data:    err.Error(),
        }
    case errors.Is(err, service.ErrValidation):
        return &JSONRPCError{
            Code:    ValidationError,
            Message: "Validation error",
            Data:    err.Error(),
        }
    default:
        return &JSONRPCError{
            Code:    InternalError,
            Message: "Internal server error",
            Data:    "An unexpected error occurred",
        }
    }
}
```

### Logging Strategy

```go
type MCPLogger struct {
    logger *logrus.Logger
}

func (ml *MCPLogger) LogRequest(ctx context.Context, req *JSONRPCRequest) {
    ml.logger.WithFields(logrus.Fields{
        "request_id": ctx.Value("request_id"),
        "method":     req.Method,
        "user_id":    ctx.Value("user_id"),
    }).Info("Processing JSON-RPC request")
}

func (ml *MCPLogger) LogError(ctx context.Context, err error, method string) {
    ml.logger.WithFields(logrus.Fields{
        "request_id": ctx.Value("request_id"),
        "method":     method,
        "error":      err.Error(),
        "user_id":    ctx.Value("user_id"),
    }).Error("JSON-RPC request failed")
}
```

## Testing Strategy

### Unit Testing

#### Handler Testing
```go
func TestMCPHandler_HandleRequest(t *testing.T) {
    tests := []struct {
        name           string
        request        *JSONRPCRequest
        expectedResult interface{}
        expectedError  *JSONRPCError
    }{
        {
            name: "valid resources/read request",
            request: &JSONRPCRequest{
                JSONRPC: "2.0",
                ID:      1,
                Method:  "resources/read",
                Params:  map[string]interface{}{"uri": "epic://EP-123"},
            },
            expectedResult: &ResourceContents{},
        },
        {
            name: "invalid JSON-RPC version",
            request: &JSONRPCRequest{
                JSONRPC: "1.0",
                ID:      1,
                Method:  "resources/read",
            },
            expectedError: &JSONRPCError{Code: InvalidRequest},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### URI Parser Testing
```go
func TestURIParser_Parse(t *testing.T) {
    parser := &URIParser{}
    
    tests := []struct {
        uri      string
        expected *ParsedURI
        hasError bool
        {
            uri: "epic://EP-001",
            expected: &ParsedURI{
                Scheme:      "epic",
                ReferenceID: "EP-001",
            },
        },
        {
            uri: "epic://EP-001/hierarchy",
            expected: &ParsedURI{
                Scheme:      "epic", 
                ReferenceID: "EP-001",
                SubPath:     "hierarchy",
            },
        },
        {
            uri: "user-story://US-042",
            expected: &ParsedURI{
                Scheme:      "user-story",
                ReferenceID: "US-042",
            },
        },
        {
            uri: "requirement://REQ-123/relationships",
            expected: &ParsedURI{
                Scheme:      "requirement",
                ReferenceID: "REQ-123",
                SubPath:     "relationships",
            },
        },
        {
            uri: "acceptance-criteria://AC-005",
            expected: &ParsedURI{
                Scheme:      "acceptance-criteria",
                ReferenceID: "AC-005",
            },
        },
        // Invalid cases
        {
            uri:      "epic://US-001", // Wrong prefix for scheme
            hasError: true,
        },
        {
            uri:      "user-story://EP-001", // Wrong prefix for scheme
            hasError: true,
        },
    }
    
    for _, tt := range tests {
        result, err := parser.Parse(tt.uri)
        if tt.hasError {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        }
    }
}
```

### Integration Testing

#### Full Request Flow Testing
```go
func TestMCPEndpoint_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Setup test server
    server := setupTestServer(t, db)
    defer server.Close()
    
    // Test cases
    tests := []struct {
        name           string
        method         string
        body           string
        expectedStatus int
        expectedBody   string
    }{
        {
            name:   "create epic via tools/call",
            method: "POST",
            body: `{
                "jsonrpc": "2.0",
                "id": 1,
                "method": "tools/call",
                "params": {
                    "name": "create_epic",
                    "arguments": {
                        "title": "Test Epic",
                        "description": "Test Description",
                        "priority": 1
                    }
                }
            }`,
            expectedStatus: 200,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Execute request and validate response
        })
    }
}
```

### Performance Testing

#### Concurrent Request Testing
```go
func TestMCPHandler_ConcurrentRequests(t *testing.T) {
    handler := setupMCPHandler(t)
    
    const numRequests = 100
    const numWorkers = 10
    
    var wg sync.WaitGroup
    requests := make(chan *JSONRPCRequest, numRequests)
    
    // Start workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for req := range requests {
                _, err := handler.HandleRequest(context.Background(), req)
                assert.NoError(t, err)
            }
        }()
    }
    
    // Send requests with reference IDs
    for i := 0; i < numRequests; i++ {
        requests <- &JSONRPCRequest{
            JSONRPC: "2.0",
            ID:      i,
            Method:  "resources/read",
            Params:  map[string]interface{}{"uri": fmt.Sprintf("epic://EP-%03d", i+1)},
        }
    }
    
    close(requests)
    wg.Wait()
}
        }()
    }
    
    // Send requests
    for i := 0; i < numRequests; i++ {
        requests <- &JSONRPCRequest{
            JSONRPC: "2.0",
            ID:      i,
            Method:  "resources/read",
            Params:  map[string]interface{}{"uri": fmt.Sprintf("epic://EP-%d", i)},
        }
    }
    
    close(requests)
    wg.Wait()
}
```

## Security Considerations

### Authentication Flow
1. Extract PAT from Authorization header
2. Validate PAT using existing middleware
3. Extract user context from validated PAT
4. Pass user context to service layer
5. Apply user-based permissions in service methods

### Input Validation
- JSON-RPC message structure validation
- Method name validation against allowed methods
- Parameter validation using JSON Schema
- URI format validation for resource requests
- Tool argument validation against tool schemas

### Rate Limiting
- Apply existing rate limiting middleware
- Consider separate limits for different method types
- Implement per-user rate limiting based on PAT

### Audit Logging
- Log all JSON-RPC requests with user attribution
- Redact sensitive information (PAT tokens)
- Include correlation IDs for request tracing
- Log security events (authentication failures, unauthorized access)

## Performance Optimizations

### Caching Strategy
- Cache frequently accessed metadata (requirement types, statuses)
- Cache user permissions for PAT tokens
- Implement response caching for read-only operations
- Use Redis for distributed caching

### Connection Management
- Reuse database connections from existing pool
- Implement connection timeouts
- Monitor connection usage and performance

### Response Optimization
- Implement pagination for large result sets
- Use streaming for large responses where applicable
- Compress responses when appropriate
- Minimize data transfer with selective field inclusion

## Deployment Considerations

### Configuration
- No additional environment variables required
- Uses existing database and Redis connections
- Integrates with existing logging configuration
- Follows existing security and middleware patterns

### Monitoring
- Add JSON-RPC specific metrics
- Monitor request/response times by method
- Track error rates by error code
- Monitor authentication success/failure rates

### Scalability
- Stateless design enables horizontal scaling
- Uses existing database connection pooling
- Compatible with existing load balancing
- Supports concurrent request processing

This design provides a comprehensive foundation for implementing the JSON-RPC API handler while maintaining consistency with existing system architecture and patterns.