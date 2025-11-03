# MCP Tools Reference

## Overview

The MCP API provides 9 tools for performing CRUD operations and search across the Product Requirements Management System. This document provides comprehensive reference information for all available tools, including schemas, examples, and usage patterns.

## Tool Categories

### CRUD Operations (7 tools)
- Epic Management: `create_epic`, `update_epic`
- User Story Management: `create_user_story`, `update_user_story`
- Requirement Management: `create_requirement`, `update_requirement`
- Relationship Management: `create_relationship`

### Search Operations (2 tools)
- Global Search: `search_global`
- Requirement Search: `search_requirements`

## Tool Reference

### 1. create_epic

Create a new epic in the requirements management system.

**Method**: `tools/call`
**Tool Name**: `create_epic`

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | string | ✓ | Epic title (max 500 characters) |
| `priority` | integer | ✓ | Priority level (1=Critical, 2=High, 3=Medium, 4=Low) |
| `description` | string | ✗ | Detailed description (max 50,000 characters) |
| `assignee_id` | string | ✗ | UUID of the user to assign the epic to |

#### JSON Schema

```json
{
  "type": "object",
  "properties": {
    "title": {
      "type": "string",
      "description": "Title of the epic (required, max 500 characters)",
      "maxLength": 500
    },
    "description": {
      "type": "string",
      "description": "Detailed description of the epic (optional, max 50000 characters)",
      "maxLength": 50000
    },
    "priority": {
      "type": "integer",
      "description": "Priority level (1=Critical, 2=High, 3=Medium, 4=Low)",
      "minimum": 1,
      "maximum": 4
    },
    "assignee_id": {
      "type": "string",
      "description": "UUID of the user to assign the epic to (optional)",
      "format": "uuid"
    }
  },
  "required": ["title", "priority"]
}
```

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "create_epic",
    "arguments": {
      "title": "User Authentication System",
      "description": "Implement comprehensive user authentication and authorization system with multi-factor authentication support",
      "priority": 1,
      "assignee_id": "123e4567-e89b-12d3-a456-426614174000"
    }
  }
}
```

#### Example Response

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Successfully created epic EP-001: User Authentication System"
      },
      {
        "type": "data",
        "data": {
          "id": "epic-uuid-here",
          "reference_id": "EP-001",
          "title": "User Authentication System",
          "description": "Implement comprehensive user authentication and authorization system",
          "status": "Backlog",
          "priority": 1,
          "creator_id": "current-user-uuid",
          "assignee_id": "123e4567-e89b-12d3-a456-426614174000",
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T00:00:00Z"
        }
      }
    ]
  }
}
```

### 2. update_epic

Update an existing epic in the requirements management system.

**Method**: `tools/call`
**Tool Name**: `update_epic`

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `epic_id` | string | ✓ | UUID or reference ID (EP-XXX) of the epic to update |
| `title` | string | ✗ | New title (max 500 characters) |
| `description` | string | ✗ | New description (max 50,000 characters) |
| `priority` | integer | ✗ | New priority level (1-4) |
| `assignee_id` | string | ✗ | UUID of assignee (empty string to unassign) |
| `status` | string | ✗ | New status (Backlog, Draft, In Progress, Done, Cancelled) |

#### Status Values

Epic status can be updated to one of the following values:
- `Backlog` - Initial state for new epics
- `Draft` - Epic is being planned and refined
- `In Progress` - Epic is actively being worked on
- `Done` - Epic is completed
- `Cancelled` - Epic is cancelled and will not be completed

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "update_epic",
    "arguments": {
      "epic_id": "EP-001",
      "title": "Enhanced User Authentication System",
      "description": "Implement comprehensive user authentication with OAuth2 and SAML support",
      "priority": 1,
      "status": "In Progress"
    }
  }
}
```

#### Example Response

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Successfully updated epic EP-001: Enhanced User Authentication System"
      },
      {
        "type": "data",
        "data": {
          "id": "epic-uuid-here",
          "reference_id": "EP-001",
          "title": "Enhanced User Authentication System",
          "description": "Implement comprehensive user authentication with OAuth2 and SAML support",
          "status": "In Progress",
          "priority": 1,
          "creator_id": "current-user-uuid",
          "assignee_id": "123e4567-e89b-12d3-a456-426614174000",
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T12:30:00Z"
        }
      }
    ]
  }
}
```

### 3. create_user_story

Create a new user story within an epic.

**Method**: `tools/call`
**Tool Name**: `create_user_story`

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | string | ✓ | User story title (max 500 characters) |
| `epic_id` | string | ✓ | UUID or reference ID (EP-XXX) of the parent epic |
| `priority` | integer | ✓ | Priority level (1-4) |
| `description` | string | ✗ | Description, preferably in "As [role], I want [function], so that [goal]" format |
| `assignee_id` | string | ✗ | UUID of the user to assign the user story to |

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "create_user_story",
    "arguments": {
      "title": "User Login with Email and Password",
      "description": "As a registered user, I want to log in with my email and password, so that I can access my personalized dashboard and account features",
      "priority": 2,
      "epic_id": "EP-001",
      "assignee_id": "123e4567-e89b-12d3-a456-426614174000"
    }
  }
}
```

### 4. update_user_story

Update an existing user story.

**Method**: `tools/call`
**Tool Name**: `update_user_story`

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `user_story_id` | string | ✓ | UUID or reference ID (US-XXX) of the user story to update |
| `title` | string | ✗ | New title (max 500 characters) |
| `description` | string | ✗ | New description (max 50,000 characters) |
| `priority` | integer | ✗ | New priority level (1-4) |
| `assignee_id` | string | ✗ | UUID of assignee (empty string to unassign) |
| `status` | string | ✗ | New status (Backlog, Draft, In Progress, Done, Cancelled) |

#### Status Values

User story status can be updated to one of the following values:
- `Backlog` - Initial state for new user stories
- `Draft` - User story is being refined
- `In Progress` - User story is actively being developed
- `Done` - User story is completed
- `Cancelled` - User story is cancelled

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "update_user_story",
    "arguments": {
      "user_story_id": "US-001",
      "title": "Secure User Login with Multi-Factor Authentication",
      "description": "As a registered user, I want to securely log in with my email, password, and optional MFA, so that I can access my account with confidence",
      "priority": 1,
      "status": "In Progress"
    }
  }
}
```

#### Example Response

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Successfully updated user story US-001: Secure User Login with Multi-Factor Authentication"
      },
      {
        "type": "data",
        "data": {
          "id": "user-story-uuid-here",
          "reference_id": "US-001",
          "title": "Secure User Login with Multi-Factor Authentication",
          "description": "As a registered user, I want to securely log in with my email, password, and optional MFA, so that I can access my account with confidence",
          "status": "In Progress",
          "priority": 1,
          "epic_id": "epic-uuid",
          "creator_id": "current-user-uuid",
          "assignee_id": "123e4567-e89b-12d3-a456-426614174000",
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T12:30:00Z"
        }
      }
    ]
  }
}
```

### 5. create_requirement

Create a new requirement within a user story.

**Method**: `tools/call`
**Tool Name**: `create_requirement`

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `title` | string | ✓ | Requirement title (max 500 characters) |
| `user_story_id` | string | ✓ | UUID or reference ID (US-XXX) of the parent user story |
| `type_id` | string | ✓ | UUID of the requirement type (Functional, Non-Functional, etc.) |
| `priority` | integer | ✓ | Priority level (1-4) |
| `description` | string | ✗ | Detailed description (max 50,000 characters) |
| `acceptance_criteria_id` | string | ✗ | UUID of linked acceptance criteria |
| `assignee_id` | string | ✗ | UUID of the user to assign the requirement to |

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/call",
  "params": {
    "name": "create_requirement",
    "arguments": {
      "title": "Password validation must enforce strong password policy",
      "description": "The system shall enforce a strong password policy requiring minimum 12 characters, including uppercase, lowercase, numbers, and special characters. Common dictionary words and personal information should be rejected.",
      "priority": 1,
      "user_story_id": "US-001",
      "type_id": "123e4567-e89b-12d3-a456-426614174010",
      "assignee_id": "123e4567-e89b-12d3-a456-426614174000"
    }
  }
}
```

### 6. update_requirement

Update an existing requirement.

**Method**: `tools/call`
**Tool Name**: `update_requirement`

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `requirement_id` | string | ✓ | UUID or reference ID (REQ-XXX) of the requirement to update |
| `title` | string | ✗ | New title (max 500 characters) |
| `description` | string | ✗ | New description (max 50,000 characters) |
| `priority` | integer | ✗ | New priority level (1-4) |
| `assignee_id` | string | ✗ | UUID of assignee (empty string to unassign) |
| `status` | string | ✗ | New status (Draft, Active, Obsolete) |

#### Status Values

Requirement status can be updated to one of the following values:
- `Draft` - Initial state for new requirements
- `Active` - Requirement is approved and active
- `Obsolete` - Requirement is no longer valid

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "tools/call",
  "params": {
    "name": "update_requirement",
    "arguments": {
      "requirement_id": "REQ-001",
      "title": "Enhanced password validation with complexity scoring",
      "description": "The system shall enforce a comprehensive password policy with complexity scoring, requiring minimum 12 characters, character diversity, and rejection of common patterns, dictionary words, and personal information.",
      "priority": 1,
      "status": "Active"
    }
  }
}
```

#### Example Response

```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Successfully updated requirement REQ-001: Enhanced password validation with complexity scoring"
      },
      {
        "type": "data",
        "data": {
          "id": "requirement-uuid-here",
          "reference_id": "REQ-001",
          "title": "Enhanced password validation with complexity scoring",
          "description": "The system shall enforce a comprehensive password policy with complexity scoring, requiring minimum 12 characters, character diversity, and rejection of common patterns, dictionary words, and personal information.",
          "status": "Active",
          "priority": 1,
          "user_story_id": "user-story-uuid",
          "type_id": "functional-type-uuid",
          "creator_id": "current-user-uuid",
          "assignee_id": "123e4567-e89b-12d3-a456-426614174000",
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T12:30:00Z"
        }
      }
    ]
  }
}
```

### 7. create_relationship

Create a relationship between two requirements.

**Method**: `tools/call`
**Tool Name**: `create_relationship`

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source_requirement_id` | string | ✓ | UUID or reference ID (REQ-XXX) of the source requirement |
| `target_requirement_id` | string | ✓ | UUID or reference ID (REQ-XXX) of the target requirement |
| `relationship_type_id` | string | ✓ | UUID of the relationship type (depends_on, blocks, relates_to, etc.) |

#### Common Relationship Types

- **depends_on**: Source requirement depends on target requirement
- **blocks**: Source requirement blocks target requirement
- **relates_to**: Source requirement is related to target requirement
- **conflicts_with**: Source requirement conflicts with target requirement
- **derives_from**: Source requirement derives from target requirement

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 7,
  "method": "tools/call",
  "params": {
    "name": "create_relationship",
    "arguments": {
      "source_requirement_id": "REQ-001",
      "target_requirement_id": "REQ-002",
      "relationship_type_id": "123e4567-e89b-12d3-a456-426614174020"
    }
  }
}
```

### 8. search_global

Search across all entities in the requirements management system.

**Method**: `tools/call`
**Tool Name**: `search_global`

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | ✓ | Search query string (minimum 1 character) |
| `entity_types` | array | ✗ | Entity types to search (epic, user_story, acceptance_criteria, requirement) |
| `limit` | integer | ✗ | Maximum results to return (1-100, default 50) |
| `offset` | integer | ✗ | Number of results to skip for pagination (default 0) |

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 8,
  "method": "tools/call",
  "params": {
    "name": "search_global",
    "arguments": {
      "query": "authentication password security",
      "entity_types": ["epic", "user_story", "requirement"],
      "limit": 20,
      "offset": 0
    }
  }
}
```

#### Example Response

```json
{
  "jsonrpc": "2.0",
  "id": 8,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Found 15 results for query 'authentication password security'"
      },
      {
        "type": "data",
        "data": {
          "results": [
            {
              "entity_type": "epic",
              "entity_id": "epic-uuid",
              "reference_id": "EP-001",
              "title": "User Authentication System",
              "description": "Implement comprehensive user authentication...",
              "highlight": "...comprehensive user <em>authentication</em> and authorization...",
              "rank": 0.95
            },
            {
              "entity_type": "requirement",
              "entity_id": "req-uuid",
              "reference_id": "REQ-001",
              "title": "Password validation policy",
              "description": "System shall enforce strong password policy...",
              "highlight": "...enforce strong <em>password</em> policy with <em>security</em>...",
              "rank": 0.87
            }
          ],
          "total_count": 15,
          "query": "authentication password security",
          "entity_types": ["epic", "user_story", "requirement"],
          "limit": 20,
          "offset": 0
        }
      }
    ]
  }
}
```

### 9. search_requirements

Search specifically within requirements.

**Method**: `tools/call`
**Tool Name**: `search_requirements`

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | ✓ | Search query string for requirements (minimum 1 character) |

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 9,
  "method": "tools/call",
  "params": {
    "name": "search_requirements",
    "arguments": {
      "query": "password validation complexity"
    }
  }
}
```

#### Example Response

```json
{
  "jsonrpc": "2.0",
  "id": 9,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Found 3 requirements matching query 'password validation complexity'"
      },
      {
        "type": "data",
        "data": {
          "requirements": [
            {
              "id": "req-uuid-1",
              "reference_id": "REQ-001",
              "title": "Password validation policy",
              "description": "System shall enforce password complexity requirements",
              "status": "Active",
              "priority": 1,
              "user_story_id": "us-uuid",
              "type_id": "functional-type-uuid"
            },
            {
              "id": "req-uuid-2",
              "reference_id": "REQ-005",
              "title": "Password strength meter",
              "description": "Display real-time password complexity feedback",
              "status": "Active",
              "priority": 2,
              "user_story_id": "us-uuid",
              "type_id": "functional-type-uuid"
            }
          ],
          "query": "password validation complexity",
          "count": 2
        }
      }
    ]
  }
}
```

## Usage Patterns

### Creating a Complete Feature

```bash
# 1. Create Epic
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "create_epic",
      "arguments": {
        "title": "User Authentication System",
        "description": "Complete authentication system with login, registration, and password management",
        "priority": 1
      }
    }
  }'

# 2. Create User Story
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "create_user_story",
      "arguments": {
        "title": "User Login",
        "description": "As a registered user, I want to log in with email and password",
        "priority": 2,
        "epic_id": "EP-001"
      }
    }
  }'

# 3. Create Requirement
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "create_requirement",
      "arguments": {
        "title": "Email validation",
        "description": "System must validate email format according to RFC 5322",
        "priority": 1,
        "user_story_id": "US-001",
        "type_id": "functional-type-uuid"
      }
    }
  }'

# 4. Update Epic Status to In Progress
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "update_epic",
      "arguments": {
        "epic_id": "EP-001",
        "status": "In Progress"
      }
    }
  }'

# 5. Update User Story Status to In Progress
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "update_user_story",
      "arguments": {
        "user_story_id": "US-001",
        "status": "In Progress"
      }
    }
  }'

# 6. Update Requirement Status to Active
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "tools/call",
    "params": {
      "name": "update_requirement",
      "arguments": {
        "requirement_id": "REQ-001",
        "status": "Active"
      }
    }
  }'
```

### Searching and Analysis

```bash
# Search for authentication-related items
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "search_global",
      "arguments": {
        "query": "authentication login security",
        "entity_types": ["epic", "user_story", "requirement"],
        "limit": 10
      }
    }
  }'

# Search specifically in requirements
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "search_requirements",
      "arguments": {
        "query": "password validation"
      }
    }
  }'
```

## Status Management

### Overview

The MCP API supports status management for epics, user stories, and requirements through the update tools. Status changes enable workflow management and lifecycle tracking for all entity types.

### Status Workflows

#### Epic Status Workflow
```
Backlog → Draft → In Progress → Done
    ↓         ↓         ↓
Cancelled ← Cancelled ← Cancelled
```

#### User Story Status Workflow
```
Backlog → Draft → In Progress → Done
    ↓         ↓         ↓
Cancelled ← Cancelled ← Cancelled
```

#### Requirement Status Workflow
```
Draft → Active → Obsolete
```

### Status Update Examples

#### Update Epic Status

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "update_epic",
    "arguments": {
      "epic_id": "EP-001",
      "status": "In Progress"
    }
  }
}
```

#### Update User Story Status

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "update_user_story",
    "arguments": {
      "user_story_id": "US-001",
      "status": "Done"
    }
  }
}
```

#### Update Requirement Status

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "update_requirement",
    "arguments": {
      "requirement_id": "REQ-001",
      "status": "Active"
    }
  }
}
```

### Status Validation

The API validates status values against allowed enums for each entity type:

- **Epic**: Backlog, Draft, In Progress, Done, Cancelled
- **User Story**: Backlog, Draft, In Progress, Done, Cancelled  
- **Requirement**: Draft, Active, Obsolete

Invalid status values will result in validation errors with helpful messages indicating the valid options.

## Error Handling

### Common Error Scenarios

#### Missing Required Parameters

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

#### Invalid Tool Name

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": "Unknown tool: invalid_tool_name"
  }
}
```

#### Invalid UUID Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Invalid 'assignee_id' format"
  }
}
```

#### Entity Not Found

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "Failed to create user story: epic not found"
  }
}
```

#### Invalid Status Value

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Invalid status 'InvalidStatus' for epic. Valid statuses are: Backlog, Draft, In Progress, Done, Cancelled"
  }
}
```

## Best Practices

### Parameter Validation

1. **Required Fields**: Always provide all required parameters
2. **Data Types**: Ensure correct data types (string, integer, array)
3. **Format Validation**: Validate UUIDs and reference IDs before sending
4. **Length Limits**: Respect maximum length constraints

### ID Handling

1. **Reference IDs**: Use human-readable reference IDs (EP-001, US-001) when possible
2. **UUID Fallback**: Tools accept both UUIDs and reference IDs
3. **Consistency**: Use the same ID format throughout a workflow

### Error Recovery

1. **Validation**: Validate parameters client-side before sending
2. **Retry Logic**: Implement retry logic for transient errors
3. **Error Logging**: Log all errors for debugging and monitoring
4. **User Feedback**: Provide meaningful error messages to users

### Performance Optimization

1. **Batch Operations**: Group related operations when possible
2. **Search Pagination**: Use limit/offset for large search results
3. **Caching**: Cache frequently accessed data
4. **Connection Reuse**: Maintain persistent connections for multiple operations

## Integration Examples

### Python Integration

```python
import json
import requests

class MCPClient:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.token = token
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    
    def call_tool(self, tool_name, arguments, request_id=1):
        payload = {
            "jsonrpc": "2.0",
            "id": request_id,
            "method": "tools/call",
            "params": {
                "name": tool_name,
                "arguments": arguments
            }
        }
        
        response = requests.post(
            f"{self.base_url}/api/v1/mcp",
            headers=self.headers,
            json=payload
        )
        
        return response.json()
    
    def create_epic(self, title, priority, description=None, assignee_id=None):
        args = {
            "title": title,
            "priority": priority
        }
        if description:
            args["description"] = description
        if assignee_id:
            args["assignee_id"] = assignee_id
            
        return self.call_tool("create_epic", args)

# Usage
client = MCPClient("http://localhost:8080", "mcp_pat_your_token")
result = client.create_epic(
    title="User Authentication System",
    priority=1,
    description="Complete authentication system"
)
print(json.dumps(result, indent=2))
```

### JavaScript Integration

```javascript
class MCPClient {
    constructor(baseUrl, token) {
        this.baseUrl = baseUrl;
        this.token = token;
    }
    
    async callTool(toolName, arguments, requestId = 1) {
        const payload = {
            jsonrpc: "2.0",
            id: requestId,
            method: "tools/call",
            params: {
                name: toolName,
                arguments: arguments
            }
        };
        
        const response = await fetch(`${this.baseUrl}/api/v1/mcp`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });
        
        return await response.json();
    }
    
    async searchGlobal(query, entityTypes = null, limit = 50, offset = 0) {
        const args = { query, limit, offset };
        if (entityTypes) {
            args.entity_types = entityTypes;
        }
        
        return await this.callTool("search_global", args);
    }
}

// Usage
const client = new MCPClient("http://localhost:8080", "mcp_pat_your_token");
const results = await client.searchGlobal("authentication", ["epic", "user_story"], 10);
console.log(JSON.stringify(results, null, 2));
```