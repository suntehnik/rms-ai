# MCP URI Schemes Reference

## Overview

The MCP API uses custom URI schemes to provide structured access to entities in the Product Requirements Management System. This document provides a comprehensive reference for all supported URI patterns and their usage.

## URI Structure

```
scheme://reference-id[/sub-path][?parameters]
```

### Components

- **scheme**: Identifies the entity type (epic, user-story, requirement, acceptance-criteria)
- **reference-id**: Human-readable entity identifier (e.g., EP-001, US-001)
- **sub-path**: Optional path for accessing related data
- **parameters**: Optional query parameters for filtering or configuration

## Supported Schemes

### 1. Epic Scheme (`epic://`)

Access epic entities and their hierarchical data.

**Reference ID Pattern**: `EP-XXX` (e.g., EP-001, EP-042)

#### Basic Epic Access

```
epic://EP-001
```

**Returns**: Complete epic entity data

**Example Response**:
```json
{
  "uri": "epic://EP-001",
  "name": "Epic EP-001: User Authentication System",
  "description": "Epic EP-001 with status In Progress and priority 1",
  "mimeType": "application/json",
  "contents": {
    "id": "uuid-here",
    "reference_id": "EP-001",
    "title": "User Authentication System",
    "description": "Implement comprehensive user authentication and authorization",
    "status": "In Progress",
    "priority": 1,
    "creator_id": "creator-uuid",
    "assignee_id": "assignee-uuid",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Epic Hierarchy

```
epic://EP-001/hierarchy
```

**Returns**: Epic with all its user stories

**Use Case**: Get complete epic structure for analysis or display

**Example Response**:
```json
{
  "uri": "epic://EP-001/hierarchy",
  "name": "Epic EP-001 Hierarchy: User Authentication System",
  "description": "Epic EP-001 with all its user stories",
  "mimeType": "application/json",
  "contents": {
    "id": "epic-uuid",
    "reference_id": "EP-001",
    "title": "User Authentication System",
    "description": "Implement comprehensive user authentication",
    "status": "In Progress",
    "priority": 1,
    "user_stories": [
      {
        "id": "us-uuid-1",
        "reference_id": "US-001",
        "title": "User Login",
        "description": "As a user, I want to log in...",
        "status": "In Progress",
        "priority": 2
      },
      {
        "id": "us-uuid-2",
        "reference_id": "US-002",
        "title": "Password Reset",
        "description": "As a user, I want to reset my password...",
        "status": "Backlog",
        "priority": 3
      }
    ]
  }
}
```

#### Epic User Stories

```
epic://EP-001/user-stories
```

**Returns**: Just the user stories belonging to the epic

**Use Case**: Focus on user stories without epic metadata

### 2. User Story Scheme (`user-story://`)

Access user story entities and their related data.

**Reference ID Pattern**: `US-XXX` (e.g., US-001, US-042)

#### Basic User Story Access

```
user-story://US-001
```

**Returns**: Complete user story entity data

**Example Response**:
```json
{
  "uri": "user-story://US-001",
  "name": "User Story US-001: User Login",
  "description": "User Story US-001 with status In Progress and priority 2",
  "mimeType": "application/json",
  "contents": {
    "id": "us-uuid",
    "reference_id": "US-001",
    "title": "User Login",
    "description": "As a registered user, I want to log in with email and password",
    "status": "In Progress",
    "priority": 2,
    "epic_id": "epic-uuid",
    "creator_id": "creator-uuid",
    "assignee_id": "assignee-uuid",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### User Story Requirements

```
user-story://US-001/requirements
```

**Returns**: All requirements associated with the user story

**Use Case**: Analyze requirements for a specific user story

**Example Response**:
```json
{
  "uri": "user-story://US-001/requirements",
  "name": "Requirements",
  "description": "List of 3 requirements",
  "mimeType": "application/json",
  "contents": {
    "requirements": [
      {
        "id": "req-uuid-1",
        "reference_id": "REQ-001",
        "title": "Email validation",
        "description": "System must validate email format",
        "status": "Active",
        "priority": 1
      },
      {
        "id": "req-uuid-2",
        "reference_id": "REQ-002",
        "title": "Password strength",
        "description": "Password must meet complexity requirements",
        "status": "Active",
        "priority": 1
      }
    ],
    "count": 2
  }
}
```

#### User Story Acceptance Criteria

```
user-story://US-001/acceptance-criteria
```

**Returns**: All acceptance criteria for the user story

**Use Case**: Review acceptance criteria for testing or validation

### 3. Requirement Scheme (`requirement://`)

Access requirement entities and their relationships.

**Reference ID Pattern**: `REQ-XXX` (e.g., REQ-001, REQ-042)

#### Basic Requirement Access

```
requirement://REQ-001
```

**Returns**: Complete requirement entity data

**Example Response**:
```json
{
  "uri": "requirement://REQ-001",
  "name": "Requirement REQ-001: Email validation",
  "description": "Requirement REQ-001 with status Active and priority 1",
  "mimeType": "application/json",
  "contents": {
    "id": "req-uuid",
    "reference_id": "REQ-001",
    "title": "Email validation",
    "description": "System must validate email format according to RFC 5322",
    "status": "Active",
    "priority": 1,
    "user_story_id": "us-uuid",
    "acceptance_criteria_id": "ac-uuid",
    "type_id": "functional-type-uuid",
    "creator_id": "creator-uuid",
    "assignee_id": "assignee-uuid",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Requirement Relationships

```
requirement://REQ-001/relationships
```

**Returns**: Requirement with all its relationships (source and target)

**Use Case**: Analyze requirement dependencies and impacts

**Example Response**:
```json
{
  "uri": "requirement://REQ-001/relationships",
  "name": "Requirement REQ-001 Relationships: Email validation",
  "description": "Requirement REQ-001 with all its relationships",
  "mimeType": "application/json",
  "contents": {
    "id": "req-uuid",
    "reference_id": "REQ-001",
    "title": "Email validation",
    "description": "System must validate email format",
    "status": "Active",
    "priority": 1,
    "source_relationships": [
      {
        "id": "rel-uuid-1",
        "source_requirement_id": "req-uuid",
        "target_requirement_id": "other-req-uuid",
        "relationship_type_id": "depends-on-type-uuid",
        "created_by": "creator-uuid",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "target_relationships": [
      {
        "id": "rel-uuid-2",
        "source_requirement_id": "another-req-uuid",
        "target_requirement_id": "req-uuid",
        "relationship_type_id": "blocks-type-uuid",
        "created_by": "creator-uuid",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

### 4. Acceptance Criteria Scheme (`acceptance-criteria://`)

Access acceptance criteria entities.

**Reference ID Pattern**: `AC-XXX` (e.g., AC-001, AC-042)

#### Basic Acceptance Criteria Access

```
acceptance-criteria://AC-001
```

**Returns**: Complete acceptance criteria entity data

**Example Response**:
```json
{
  "uri": "acceptance-criteria://AC-001",
  "name": "Acceptance Criteria AC-001",
  "description": "Acceptance Criteria AC-001 for user story",
  "mimeType": "application/json",
  "contents": {
    "id": "ac-uuid",
    "reference_id": "AC-001",
    "description": "Given a user enters an email address, when the system validates it, then it should accept valid RFC 5322 format emails and reject invalid ones",
    "user_story_id": "us-uuid",
    "author_id": "author-uuid",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**Note**: Acceptance criteria currently do not support sub-paths.

## URI Validation Rules

### Reference ID Format

All reference IDs must follow the pattern: `PREFIX-NUMBER`

- **PREFIX**: 2-3 uppercase letters matching the scheme
- **SEPARATOR**: Single hyphen (-)
- **NUMBER**: One or more digits

**Valid Examples**:
- `EP-1`, `EP-001`, `EP-1234`
- `US-1`, `US-001`, `US-5678`
- `REQ-1`, `REQ-001`, `REQ-9999`
- `AC-1`, `AC-001`, `AC-4321`

**Invalid Examples**:
- `ep-001` (lowercase prefix)
- `EP_001` (underscore separator)
- `EP-` (missing number)
- `EP-01A` (non-numeric suffix)

### Scheme-Prefix Matching

The reference ID prefix must match the URI scheme:

| Scheme | Required Prefix |
|--------|----------------|
| `epic://` | `EP-` |
| `user-story://` | `US-` |
| `requirement://` | `REQ-` |
| `acceptance-criteria://` | `AC-` |

### Sub-path Validation

Sub-paths are validated against a whitelist for each scheme:

**Epic Sub-paths**:
- `hierarchy` ✓
- `user-stories` ✓
- `requirements` ✗ (not supported)

**User Story Sub-paths**:
- `requirements` ✓
- `acceptance-criteria` ✓
- `hierarchy` ✗ (not supported)

**Requirement Sub-paths**:
- `relationships` ✓
- `dependencies` ✗ (not supported)

**Acceptance Criteria Sub-paths**:
- None currently supported

## Usage Examples

### Reading Epic Hierarchy

```bash
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "resources/read",
    "params": {
      "uri": "epic://EP-001/hierarchy"
    }
  }'
```

### Reading User Story Requirements

```bash
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "resources/read",
    "params": {
      "uri": "user-story://US-001/requirements"
    }
  }'
```

### Reading Requirement Relationships

```bash
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "resources/read",
    "params": {
      "uri": "requirement://REQ-001/relationships"
    }
  }'
```

## Error Scenarios

### Invalid Scheme

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Invalid URI: unsupported URI scheme: invalid"
  }
}
```

### Invalid Reference ID Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Invalid URI: invalid reference ID format: ep-001"
  }
}
```

### Scheme-Prefix Mismatch

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Invalid URI: reference ID US-001 does not match scheme epic (expected prefix: EP)"
  }
}
```

### Unsupported Sub-path

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Unsupported sub-path for epic: invalid-path"
  }
}
```

### Entity Not Found

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32002,
    "message": "Epic not found"
  }
}
```

## Best Practices

### URI Construction

1. **Use Reference IDs**: Prefer reference IDs over UUIDs for readability
2. **Validate Format**: Always validate URI format before sending requests
3. **Handle Errors**: Implement proper error handling for invalid URIs
4. **Cache Results**: Cache frequently accessed resources to improve performance

### Sub-path Usage

1. **Hierarchy Access**: Use `/hierarchy` for complete entity trees
2. **Focused Queries**: Use specific sub-paths to get only needed data
3. **Performance**: Sub-paths can reduce response size and improve performance

### Integration Patterns

1. **Resource Discovery**: Start with basic entity access, then use sub-paths for details
2. **Relationship Analysis**: Use `/relationships` sub-path for dependency analysis
3. **Hierarchical Navigation**: Use `/hierarchy` and `/user-stories` for navigation UIs

## Future Extensions

### Planned Sub-paths

- `epic://EP-001/requirements` - All requirements in an epic
- `user-story://US-001/hierarchy` - User story with requirements and acceptance criteria
- `requirement://REQ-001/dependencies` - Direct dependencies only

### Query Parameters

Future versions may support query parameters for:
- Filtering: `?status=active`
- Sorting: `?sort=priority`
- Pagination: `?limit=10&offset=20`
- Field Selection: `?fields=id,title,status`

### Additional Schemes

Potential future schemes:
- `project://` - Project-level access
- `user://` - User entity access
- `comment://` - Comment entity access