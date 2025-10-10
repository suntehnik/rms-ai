# MCP (Model Context Protocol) Testing Guide

This guide provides comprehensive curl commands to test the MCP functionality of the Product Requirements Management API.

## Prerequisites

1. **Server Running**: Ensure the server is running on `http://localhost:8080`
2. **Authentication**: You need a valid JWT token or Personal Access Token (PAT)
3. **Test Data**: Some tests require existing entities (users, requirement types, etc.)

## Authentication Setup

### Option 1: Using JWT Token

First, login to get a JWT token:

```bash
# Login to get JWT token
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your_admin_password"
  }'
```



Save the token from the response and use it in subsequent requests:

```bash
export JWT="your_jwt_token_here"
```

### Option 2: Using Personal Access Token (PAT)

Create a PAT (requires JWT first):

```bash
# Create a PAT
curl -X POST http://localhost:8080/api/v1/pats \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MCP Testing Token",
    "description": "Token for testing MCP functionality",
    "scopes": ["full access"],
    "expires_at": "2024-12-31T23:59:59Z"
  }'
```

Save the PAT token:

```bash
export PAT_TOKEN="mcp_pat_your_token_here"
```

## MCP Protocol Testing

The MCP endpoint uses JSON-RPC 2.0 protocol. All requests go to `/api/v1/mcp` with POST method.

### 1. Initialize MCP Connection

```bash
# Initialize MCP connection
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-06-18",
      "capabilities": {
        "elicitation": {}
      },
      "clientInfo": {
        "name": "curl-test-client",
        "version": "1.0.0"
      }
    }
  }'
```

Expected response:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2025-06-18",
    "capabilities": {
      "tools": {
        "listChanged": true
      },
      "resources": {}
    },
    "serverInfo": {
      "name": "product-requirements-mcp-server",
      "version": "1.0.0"
    }
  }
}
```

### 2. List Available Tools

```bash
# List all available MCP tools
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }'
```

This will return all 9 available tools with their schemas.

## CRUD Operations via MCP Tools

### 3. Create Epic

```bash
# Create a new epic
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "create_epic",
      "arguments": {
        "title": "User Authentication System",
        "description": "Implement comprehensive user authentication and authorization system",
        "priority": 1,
        "creator_id": "123e4567-e89b-12d3-a456-426614174000"
      }
    }
  }'
```

### 4. Update Epic

```bash
# Update an existing epic (using reference ID)
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "update_epic",
      "arguments": {
        "epic_id": "EP-001",
        "title": "Enhanced User Authentication System",
        "description": "Implement comprehensive user authentication with multi-factor authentication",
        "priority": 1
      }
    }
  }'
```

### 5. Create User Story

```bash
# Create a user story within an epic
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "create_user_story",
      "arguments": {
        "title": "User Login with Email and Password",
        "description": "As a registered user, I want to log in with my email and password, so that I can access my personalized dashboard",
        "priority": 2,
        "epic_id": "EP-001",
        "creator_id": "123e4567-e89b-12d3-a456-426614174000"
      }
    }
  }'
```

### 6. Update User Story

```bash
# Update an existing user story
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "tools/call",
    "params": {
      "name": "update_user_story",
      "arguments": {
        "user_story_id": "US-001",
        "title": "Secure User Login with Email and Password",
        "description": "As a registered user, I want to securely log in with my email and password, so that I can access my personalized dashboard with confidence",
        "priority": 1
      }
    }
  }'
```

### 7. Create Requirement

```bash
# Create a requirement within a user story
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 7,
    "method": "tools/call",
    "params": {
      "name": "create_requirement",
      "arguments": {
        "title": "Password validation must enforce strong password policy",
        "description": "The system shall enforce a strong password policy requiring minimum 8 characters, including uppercase, lowercase, numbers, and special characters",
        "priority": 1,
        "user_story_id": "US-001",
        "type_id": "123e4567-e89b-12d3-a456-426614174010",
        "creator_id": "123e4567-e89b-12d3-a456-426614174000"
      }
    }
  }'
```

### 8. Update Requirement

```bash
# Update an existing requirement
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 8,
    "method": "tools/call",
    "params": {
      "name": "update_requirement",
      "arguments": {
        "requirement_id": "REQ-001",
        "title": "Enhanced password validation with complexity requirements",
        "description": "The system shall enforce a comprehensive password policy requiring minimum 12 characters, including uppercase, lowercase, numbers, special characters, and no common dictionary words",
        "priority": 1
      }
    }
  }'
```

### 9. Create Requirement Relationship

```bash
# Create a relationship between two requirements
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 9,
    "method": "tools/call",
    "params": {
      "name": "create_relationship",
      "arguments": {
        "source_requirement_id": "REQ-001",
        "target_requirement_id": "REQ-002",
        "relationship_type_id": "123e4567-e89b-12d3-a456-426614174020",
        "created_by": "123e4567-e89b-12d3-a456-426614174000"
      }
    }
  }'
```

## Search Operations via MCP Tools

### 10. Global Search

```bash
# Search across all entities
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 10,
    "method": "tools/call",
    "params": {
      "name": "search_global",
      "arguments": {
        "query": "authentication",
        "entity_types": ["epic", "user_story", "requirement"],
        "limit": 20,
        "offset": 0
      }
    }
  }'
```

### 11. Search Requirements

```bash
# Search specifically within requirements
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 11,
    "method": "tools/call",
    "params": {
      "name": "search_requirements",
      "arguments": {
        "query": "password validation"
      }
    }
  }'
```

## Resource Operations

### 12. Read Epic Resource

```bash
# Read an epic resource
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 12,
    "method": "resources/read",
    "params": {
      "uri": "epic://EP-001"
    }
  }'
```

### 13. Read User Story Resource

```bash
# Read a user story resource
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 13,
    "method": "resources/read",
    "params": {
      "uri": "user_story://US-001"
    }
  }'
```

### 14. Read Requirement Resource

```bash
# Read a requirement resource
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 14,
    "method": "resources/read",
    "params": {
      "uri": "requirement://REQ-001"
    }
  }'
```

## Error Testing

### 15. Invalid Tool Name

```bash
# Test with invalid tool name
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 15,
    "method": "tools/call",
    "params": {
      "name": "invalid_tool",
      "arguments": {}
    }
  }'
```

### 16. Missing Required Arguments

```bash
# Test with missing required arguments
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 16,
    "method": "tools/call",
    "params": {
      "name": "create_epic",
      "arguments": {
        "title": "Test Epic"
      }
    }
  }'
```

### 17. Invalid JSON-RPC Format

```bash
# Test with invalid JSON-RPC format
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer $PAT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "invalid": "request"
  }'
```

## Batch Testing Script

Here's a bash script to run multiple tests:

```bash
#!/bin/bash

# MCP Testing Script
# Usage: ./test_mcp.sh

# Set your token here
PAT_TOKEN="your_pat_token_here"
BASE_URL="http://localhost:8080/api/v1/mcp"

# Function to make MCP request
mcp_request() {
    local id=$1
    local method=$2
    local params=$3
    
    curl -s -X POST $BASE_URL \
        -H "Authorization: Bearer $PAT_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"jsonrpc\": \"2.0\",
            \"id\": $id,
            \"method\": \"$method\",
            \"params\": $params
        }" | jq '.'
}

echo "=== MCP Testing Suite ==="

echo "1. Testing MCP Initialize..."
mcp_request 1 "initialize" '{
    "protocolVersion": "2025-06-18",
    "capabilities": {"elicitation": {}},
    "clientInfo": {"name": "test-client", "version": "1.0.0"}
}'

echo -e "\n2. Testing Tools List..."
mcp_request 2 "tools/list" '{}'

echo -e "\n3. Testing Create Epic..."
mcp_request 3 "tools/call" '{
    "name": "create_epic",
    "arguments": {
        "title": "Test Epic via MCP",
        "description": "Testing epic creation through MCP",
        "priority": 2,
        "creator_id": "123e4567-e89b-12d3-a456-426614174000"
    }
}'

echo -e "\n4. Testing Global Search..."
mcp_request 4 "tools/call" '{
    "name": "search_global",
    "arguments": {
        "query": "test",
        "limit": 5
    }
}'

echo "=== Testing Complete ==="
```

## Expected Response Format

All MCP responses follow JSON-RPC 2.0 format:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Operation completed successfully"
      },
      {
        "type": "data",
        "data": {
          // Actual response data
        }
      }
    ]
  }
}
```

## Troubleshooting

### Common Issues

1. **401 Unauthorized**: Check your PAT token is valid and not expired
2. **Invalid params**: Ensure all required arguments are provided with correct types
3. **Method not found**: Verify the method name is correct (case-sensitive)
4. **Invalid UUID**: Ensure UUIDs are properly formatted or use reference IDs

### Debug Tips

1. Use `jq` to format JSON responses: `curl ... | jq '.'`
2. Check server logs for detailed error messages
3. Verify entity IDs exist before referencing them
4. Test with simple operations first (initialize, tools/list)

## Integration with MCP Clients

These curl commands can be adapted for use with MCP client libraries:

- **Python**: Use `mcp` package
- **TypeScript**: Use `@modelcontextprotocol/sdk`
- **Go**: Use JSON-RPC client libraries

The JSON-RPC format remains the same across all client implementations.