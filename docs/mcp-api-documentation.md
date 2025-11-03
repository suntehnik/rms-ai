# MCP API Documentation

## Overview

The Model Context Protocol (MCP) API provides a standardized JSON-RPC 2.0 interface for AI applications to interact with the Product Requirements Management System. This API enables AI models to read entity data, perform CRUD operations, and search across the requirements management system.

## Table of Contents

1. [Architecture](#architecture)
2. [Authentication](#authentication)
3. [URI Schemes](#uri-schemes)
4. [Available Tools](#available-tools)
5. [Resource Operations](#resource-operations)
6. [Error Handling](#error-handling)
7. [Examples](#examples)
8. [Integration Guide](#integration-guide)

## Architecture

The MCP API follows the Model Context Protocol specification and implements:

- **JSON-RPC 2.0**: All communication uses JSON-RPC 2.0 protocol
- **RESTful Endpoint**: Single endpoint `/api/v1/mcp` for all MCP operations
- **PAT Authentication**: Personal Access Token authentication for secure access
- **Three Core Primitives**:
  - **Tools**: Executable functions for CRUD operations and search
  - **Resources**: Read-only access to entity data via URI schemes
  - **Prompts**: (Optional) AI-assisted workflow templates

### System Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   MCP Client    │───▶│   MCP Handler    │───▶│  Service Layer  │
│ (AI Application)│    │  (JSON-RPC 2.0)  │    │   (Business)    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌──────────────────┐
                       │   URI Parser     │
                       │ (Resource Access)│
                       └──────────────────┘
```

## Authentication

### Personal Access Token (PAT)

All MCP requests require authentication using Personal Access Tokens:

```http
POST /api/v1/mcp
Authorization: Bearer mcp_pat_your_token_here
Content-Type: application/json
```

### Token Format

PAT tokens follow the format: `mcp_pat_[base64_encoded_secret]`

- **Prefix**: `mcp_pat_` (identifies token type)
- **Secret**: Base64-encoded cryptographically secure random string
- **Storage**: Only hashed versions are stored in the database

## URI Schemes

The MCP API supports four URI schemes for accessing entity resources:

### Supported Schemes

| Scheme | Prefix | Description | Example |
|--------|--------|-------------|---------|
| `epic://` | EP | Epic entities | `epic://EP-001` |
| `user-story://` | US | User Story entities | `user-story://US-001` |
| `requirement://` | REQ | Requirement entities | `requirement://REQ-001` |
| `acceptance-criteria://` | AC | Acceptance Criteria entities | `acceptance-criteria://AC-001` |

### URI Structure

```
scheme://reference-id[/sub-path][?parameters]
```

**Components:**
- **scheme**: Entity type identifier
- **reference-id**: Human-readable entity identifier (e.g., EP-001)
- **sub-path**: Optional hierarchical data access
- **parameters**: Optional query parameters

### Supported Sub-paths

#### Epic Sub-paths
- `epic://EP-001/hierarchy` - Epic with all user stories
- `epic://EP-001/user-stories` - Just the user stories

#### User Story Sub-paths
- `user-story://US-001/requirements` - Requirements for the user story
- `user-story://US-001/acceptance-criteria` - Acceptance criteria for the user story

#### Requirement Sub-paths
- `requirement://REQ-001/relationships` - Requirement with all relationships

#### Acceptance Criteria Sub-paths
- No sub-paths currently supported

### URI Validation

The URI parser validates:
- **Scheme Format**: Must match supported schemes
- **Reference ID Pattern**: Must follow `PREFIX-NUMBER` format
- **Scheme-Prefix Matching**: Reference ID prefix must match scheme
- **Sub-path Support**: Sub-path must be supported for the scheme

## Available Tools

The MCP API provides 9 tools for CRUD operations and search:

### CRUD Tools

#### 1. create_epic
Create a new epic in the system.

**Required Parameters:**
- `title` (string): Epic title (max 500 chars)
- `priority` (integer): Priority level (1-4)

**Optional Parameters:**
- `description` (string): Epic description (max 50,000 chars)
- `assignee_id` (string): UUID of assignee

#### 2. update_epic
Update an existing epic.

**Required Parameters:**
- `epic_id` (string): UUID or reference ID (EP-XXX)

**Optional Parameters:**
- `title`, `description`, `priority`, `assignee_id`, `status`

**Status Values:**
- `Backlog`, `Draft`, `In Progress`, `Done`, `Cancelled`

#### 3. create_user_story
Create a new user story within an epic.

**Required Parameters:**
- `title` (string): User story title
- `epic_id` (string): UUID or reference ID of parent epic
- `priority` (integer): Priority level (1-4)

**Optional Parameters:**
- `description` (string): User story description
- `assignee_id` (string): UUID of assignee

#### 4. update_user_story
Update an existing user story.

**Required Parameters:**
- `user_story_id` (string): UUID or reference ID (US-XXX)

**Optional Parameters:**
- `title`, `description`, `priority`, `assignee_id`, `status`

**Status Values:**
- `Backlog`, `Draft`, `In Progress`, `Done`, `Cancelled`

#### 5. create_requirement
Create a new requirement within a user story.

**Required Parameters:**
- `title` (string): Requirement title
- `user_story_id` (string): UUID or reference ID of parent user story
- `type_id` (string): UUID of requirement type
- `priority` (integer): Priority level (1-4)

**Optional Parameters:**
- `description` (string): Requirement description
- `acceptance_criteria_id` (string): UUID of linked acceptance criteria
- `assignee_id` (string): UUID of assignee

#### 6. update_requirement
Update an existing requirement.

**Required Parameters:**
- `requirement_id` (string): UUID or reference ID (REQ-XXX)

**Optional Parameters:**
- `title`, `description`, `priority`, `assignee_id`, `status`

**Status Values:**
- `Draft`, `Active`, `Obsolete`

#### 7. create_relationship
Create a relationship between two requirements.

**Required Parameters:**
- `source_requirement_id` (string): UUID or reference ID of source requirement
- `target_requirement_id` (string): UUID or reference ID of target requirement
- `relationship_type_id` (string): UUID of relationship type

### Search Tools

#### 8. search_global
Search across all entities in the system.

**Required Parameters:**
- `query` (string): Search query string

**Optional Parameters:**
- `entity_types` (array): Entity types to search (epic, user_story, acceptance_criteria, requirement)
- `limit` (integer): Max results (1-100, default 50)
- `offset` (integer): Pagination offset (default 0)

#### 9. search_requirements
Search specifically within requirements.

**Required Parameters:**
- `query` (string): Search query string

## Resource Operations

### resources/read Method

Access entity data using URI schemes:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "resources/read",
  "params": {
    "uri": "epic://EP-001/hierarchy"
  }
}
```

### Response Format

All resource responses follow this structure:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "uri": "epic://EP-001",
    "name": "Epic EP-001: User Authentication System",
    "description": "Epic EP-001 with status In Progress and priority 1",
    "mimeType": "application/json",
    "contents": {
      "id": "uuid-here",
      "reference_id": "EP-001",
      "title": "User Authentication System",
      "description": "Implement comprehensive user authentication",
      "status": "In Progress",
      "priority": 1,
      "creator_id": "creator-uuid",
      "assignee_id": "assignee-uuid",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

## Error Handling

### JSON-RPC Error Codes

The MCP API uses standard JSON-RPC 2.0 error codes plus custom codes:

| Code | Name | Description |
|------|------|-------------|
| -32700 | Parse Error | Invalid JSON |
| -32600 | Invalid Request | Invalid JSON-RPC request |
| -32601 | Method Not Found | Unknown method |
| -32602 | Invalid Params | Invalid parameters |
| -32603 | Internal Error | Server error |
| -32002 | Not Found | Entity not found |

### Error Response Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Missing or invalid 'title' argument"
  }
}
```

### Common Error Scenarios

1. **Authentication Errors**: Invalid or expired PAT token
2. **Validation Errors**: Missing required parameters or invalid formats
3. **Not Found Errors**: Referenced entities don't exist
4. **Permission Errors**: Insufficient permissions for operation

## Examples

### Complete Workflow Example

```bash
# 1. Initialize MCP connection
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-06-18",
      "capabilities": {"elicitation": {}},
      "clientInfo": {"name": "example-client", "version": "1.0.0"}
    }
  }'

# 2. List available tools
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }'

# 3. Create an epic
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "create_epic",
      "arguments": {
        "title": "User Authentication System",
        "description": "Implement comprehensive user authentication",
        "priority": 1
      }
    }
  }'

# 4. Read the epic resource
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "resources/read",
    "params": {
      "uri": "epic://EP-001"
    }
  }'

# 5. Update epic status to In Progress
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "update_epic",
      "arguments": {
        "epic_id": "'$EPIC_ID'",
        "status": "In Progress"
      }
    }
  }'

# 6. Search for authentication-related items
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "tools/call",
    "params": {
      "name": "search_global",
      "arguments": {
        "query": "authentication",
        "limit": 10
      }
    }
  }'
```

## Integration Guide

### MCP Client Libraries

The MCP API is compatible with standard MCP client libraries:

- **Python**: `mcp` package
- **TypeScript**: `@modelcontextprotocol/sdk`
- **Go**: Custom JSON-RPC client

### Example Python Integration

```python
from mcp import ClientSession, StdioServerParameters
import asyncio

async def main():
    # Configure server parameters
    server_params = StdioServerParameters(
        command="your-mcp-server",
        args=["--config", "config.json"]
    )
    
    async with ClientSession(server_params) as session:
        # Initialize
        await session.initialize()
        
        # List tools
        tools = await session.list_tools()
        print(f"Available tools: {[tool.name for tool in tools.tools]}")
        
        # Call a tool
        result = await session.call_tool("create_epic", {
            "title": "New Epic",
            "priority": 2
        })
        print(f"Created epic: {result.content}")
        
        # Read a resource
        resource = await session.read_resource("epic://EP-001")
        print(f"Epic data: {resource.contents}")

if __name__ == "__main__":
    asyncio.run(main())
```

### Best Practices

1. **Error Handling**: Always handle JSON-RPC errors appropriately
2. **Authentication**: Securely store and rotate PAT tokens
3. **Rate Limiting**: Implement client-side rate limiting for bulk operations
4. **Logging**: Log all MCP operations for audit and debugging
5. **Validation**: Validate parameters before sending requests

### Performance Considerations

- **Batch Operations**: Use multiple requests for bulk operations
- **Caching**: Cache frequently accessed resources
- **Pagination**: Use limit/offset for large search results
- **Connection Reuse**: Maintain persistent connections when possible

## Troubleshooting

### Common Issues

1. **401 Unauthorized**: Check PAT token validity and format
2. **Invalid URI**: Verify URI scheme and reference ID format
3. **Method Not Found**: Ensure method name is correct and case-sensitive
4. **Invalid Params**: Check required parameters and data types

### Debug Tips

1. Use `jq` for JSON formatting: `curl ... | jq '.'`
2. Check server logs for detailed error messages
3. Validate JSON-RPC format before sending
4. Test with simple operations first (initialize, tools/list)

### Support

For additional support:
- Check the [MCP Status Management Guide](mcp-status-management-guide.md) for detailed status workflow information
- Check the [MCP Testing Guide](mcp-testing-guide.md)
- Review server logs for detailed error information
- Validate your JSON-RPC requests against the specification