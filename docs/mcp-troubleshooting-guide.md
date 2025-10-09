# MCP API Troubleshooting Guide

## Overview

This guide provides comprehensive troubleshooting information for common issues encountered when using the MCP (Model Context Protocol) API. It covers authentication problems, request formatting issues, validation errors, and performance concerns.

## Table of Contents

1. [Authentication Issues](#authentication-issues)
2. [Request Format Problems](#request-format-problems)
3. [Parameter Validation Errors](#parameter-validation-errors)
4. [URI Scheme Issues](#uri-scheme-issues)
5. [Tool Execution Errors](#tool-execution-errors)
6. [Performance Issues](#performance-issues)
7. [Network and Connectivity](#network-and-connectivity)
8. [Debugging Tools and Techniques](#debugging-tools-and-techniques)

## Authentication Issues

### 1. 401 Unauthorized Error

**Symptoms:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "Authentication required"
  }
}
```

**Common Causes:**
- Missing Authorization header
- Invalid PAT token format
- Expired PAT token
- Incorrect token prefix

**Solutions:**

#### Check Authorization Header
```bash
# Correct format
curl -H "Authorization: Bearer mcp_pat_your_token_here"

# Common mistakes
curl -H "Authorization: mcp_pat_your_token_here"  # Missing "Bearer"
curl -H "Authorization: Bearer your_token_here"   # Missing "mcp_pat_" prefix
```

#### Verify Token Format
PAT tokens must follow the format: `mcp_pat_[base64_encoded_secret]`

```bash
# Valid token example
mcp_pat_YWJjZGVmZ2hpams1bG1ub3BxcnN0dXZ3eHl6MTIzNDU2

# Invalid formats
pat_YWJjZGVmZ2hpams1bG1ub3BxcnN0dXZ3eHl6MTIzNDU2    # Wrong prefix
mcp_YWJjZGVmZ2hpams1bG1ub3BxcnN0dXZ3eHl6MTIzNDU2     # Missing "pat_"
```

#### Test Token Validity
```bash
# Test with a simple MCP request
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-06-18",
      "capabilities": {},
      "clientInfo": {"name": "test", "version": "1.0.0"}
    }
  }'
```

### 2. Token Expiration

**Symptoms:**
- Previously working token suddenly returns 401
- Error message about expired token

**Solutions:**
1. Check token expiration date in your PAT management system
2. Generate a new PAT token
3. Update your application configuration with the new token

### 3. Insufficient Permissions

**Symptoms:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "Insufficient permissions"
  }
}
```

**Solutions:**
1. Verify your user account has the necessary permissions
2. Check if your PAT token has the required scopes
3. Contact your system administrator for permission updates

## Request Format Problems

### 1. Invalid JSON-RPC Format

**Symptoms:**
```json
{
  "jsonrpc": "2.0",
  "id": null,
  "error": {
    "code": -32600,
    "message": "Invalid Request"
  }
}
```

**Common Causes:**
- Missing required JSON-RPC fields
- Incorrect JSON-RPC version
- Invalid JSON syntax

**Solutions:**

#### Verify Required Fields
```json
{
  "jsonrpc": "2.0",     // Required: must be "2.0"
  "id": 1,              // Required: unique request identifier
  "method": "tools/call", // Required: method name
  "params": {}          // Required for most methods
}
```

#### Common JSON-RPC Mistakes
```json
// ❌ Wrong version
{
  "jsonrpc": "1.0",
  "id": 1,
  "method": "tools/call"
}

// ❌ Missing ID
{
  "jsonrpc": "2.0",
  "method": "tools/call"
}

// ❌ Invalid JSON syntax
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "create_epic",
    "arguments": {
      "title": "Test Epic",  // ❌ Trailing comma
    }
  }
}
```

### 2. Content-Type Issues

**Symptoms:**
- Requests fail with parsing errors
- Server returns 400 Bad Request

**Solutions:**
```bash
# Always include Content-Type header
curl -H "Content-Type: application/json"

# Common mistake - missing header
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "initialize"}'
```

## Parameter Validation Errors

### 1. Missing Required Parameters

**Symptoms:**
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

**Solutions:**

#### Check Tool Requirements
```bash
# ❌ Missing required 'priority' parameter
{
  "name": "create_epic",
  "arguments": {
    "title": "Test Epic"
    // Missing "priority" field
  }
}

# ✅ Correct with all required parameters
{
  "name": "create_epic",
  "arguments": {
    "title": "Test Epic",
    "priority": 1
  }
}
```

#### Validate Parameter Types
```json
// ❌ Wrong data types
{
  "name": "create_epic",
  "arguments": {
    "title": "Test Epic",
    "priority": "1"        // Should be integer, not string
  }
}

// ✅ Correct data types
{
  "name": "create_epic",
  "arguments": {
    "title": "Test Epic",
    "priority": 1          // Integer value
  }
}
```

### 2. Invalid UUID Format

**Symptoms:**
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

**Solutions:**

#### UUID Validation
```javascript
// JavaScript UUID validation
function isValidUUID(uuid) {
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
  return uuidRegex.test(uuid);
}

// Example usage
const assigneeId = "123e4567-e89b-12d3-a456-426614174000";
if (!isValidUUID(assigneeId)) {
  console.error("Invalid UUID format");
}
```

#### Reference ID Alternative
```json
// Instead of UUID, you can use reference IDs
{
  "name": "update_epic",
  "arguments": {
    "epic_id": "EP-001",    // Reference ID instead of UUID
    "title": "Updated Epic"
  }
}
```

### 3. Parameter Length Limits

**Symptoms:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": "Title exceeds maximum length of 500 characters"
  }
}
```

**Solutions:**

#### Check Length Limits
| Field | Maximum Length |
|-------|----------------|
| `title` | 500 characters |
| `description` | 50,000 characters |

```javascript
// JavaScript length validation
function validateTitle(title) {
  if (title.length > 500) {
    throw new Error(`Title too long: ${title.length}/500 characters`);
  }
}

function validateDescription(description) {
  if (description && description.length > 50000) {
    throw new Error(`Description too long: ${description.length}/50000 characters`);
  }
}
```

## URI Scheme Issues

### 1. Invalid URI Format

**Symptoms:**
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

**Common URI Mistakes:**
```bash
# ❌ Invalid schemes
invalid://EP-001
Epic://EP-001          # Case sensitive
epic//EP-001           # Missing colon

# ✅ Valid schemes
epic://EP-001
user-story://US-001
requirement://REQ-001
acceptance-criteria://AC-001
```

### 2. Reference ID Format Issues

**Symptoms:**
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

**Reference ID Rules:**
- Must be uppercase prefix
- Must use hyphen separator
- Must have numeric suffix

```bash
# ❌ Invalid formats
epic://ep-001          # Lowercase prefix
epic://EP_001          # Underscore separator
epic://EP-            # Missing number
epic://EP-01A          # Non-numeric suffix

# ✅ Valid formats
epic://EP-001
epic://EP-1
epic://EP-1234
```

### 3. Scheme-Prefix Mismatch

**Symptoms:**
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

**Solution:**
Ensure the reference ID prefix matches the URI scheme:

```bash
# ❌ Mismatched scheme and prefix
epic://US-001          # Epic scheme with User Story prefix

# ✅ Correct matching
epic://EP-001           # Epic scheme with Epic prefix
user-story://US-001     # User Story scheme with User Story prefix
```

### 4. Unsupported Sub-paths

**Symptoms:**
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

**Supported Sub-paths:**
```bash
# Epic sub-paths
epic://EP-001/hierarchy      ✅
epic://EP-001/user-stories   ✅
epic://EP-001/requirements   ❌ Not supported

# User Story sub-paths
user-story://US-001/requirements        ✅
user-story://US-001/acceptance-criteria ✅
user-story://US-001/hierarchy           ❌ Not supported

# Requirement sub-paths
requirement://REQ-001/relationships     ✅
requirement://REQ-001/dependencies      ❌ Not supported

# Acceptance Criteria sub-paths
acceptance-criteria://AC-001/            ❌ No sub-paths supported
```

## Tool Execution Errors

### 1. Unknown Tool Name

**Symptoms:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": "Unknown tool: invalid_tool"
  }
}
```

**Solutions:**

#### List Available Tools
```bash
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list"
  }'
```

#### Verify Tool Names
Available tools (case-sensitive):
- `create_epic`
- `update_epic`
- `create_user_story`
- `update_user_story`
- `create_requirement`
- `update_requirement`
- `create_relationship`
- `search_global`
- `search_requirements`

### 2. Entity Not Found

**Symptoms:**
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

**Solutions:**

#### Verify Entity Exists
```bash
# Check if epic exists using resources/read
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "resources/read",
    "params": {
      "uri": "epic://EP-001"
    }
  }'
```

#### Search for Entity
```bash
# Search for the entity
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "search_global",
      "arguments": {
        "query": "EP-001"
      }
    }
  }'
```

### 3. Relationship Creation Errors

**Symptoms:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32603,
    "message": "Internal error",
    "data": "Failed to create relationship: circular dependency detected"
  }
}
```

**Common Issues:**
- Circular dependencies
- Invalid relationship type
- Requirements from different user stories

**Solutions:**

#### Validate Relationship Logic
```bash
# Check existing relationships before creating new ones
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "resources/read",
    "params": {
      "uri": "requirement://REQ-001/relationships"
    }
  }'
```

## Performance Issues

### 1. Slow Response Times

**Symptoms:**
- Requests taking longer than expected
- Timeouts on large operations

**Solutions:**

#### Use Pagination for Search
```json
{
  "name": "search_global",
  "arguments": {
    "query": "authentication",
    "limit": 10,        // Smaller page size
    "offset": 0
  }
}
```

#### Optimize Resource Queries
```bash
# Instead of full hierarchy
epic://EP-001/hierarchy

# Use specific sub-paths when possible
epic://EP-001/user-stories
```

#### Implement Client-Side Caching
```javascript
class MCPClientWithCache {
  constructor(baseUrl, token) {
    this.client = new MCPClient(baseUrl, token);
    this.cache = new Map();
  }
  
  async getResource(uri, ttl = 300000) { // 5 minute TTL
    const cached = this.cache.get(uri);
    if (cached && Date.now() - cached.timestamp < ttl) {
      return cached.data;
    }
    
    const data = await this.client.readResource(uri);
    this.cache.set(uri, { data, timestamp: Date.now() });
    return data;
  }
}
```

### 2. Rate Limiting

**Symptoms:**
- 429 Too Many Requests errors
- Requests being rejected

**Solutions:**

#### Implement Request Throttling
```javascript
class ThrottledMCPClient {
  constructor(baseUrl, token, requestsPerSecond = 10) {
    this.client = new MCPClient(baseUrl, token);
    this.requestQueue = [];
    this.processing = false;
    this.interval = 1000 / requestsPerSecond;
  }
  
  async request(method, params) {
    return new Promise((resolve, reject) => {
      this.requestQueue.push({ method, params, resolve, reject });
      this.processQueue();
    });
  }
  
  async processQueue() {
    if (this.processing || this.requestQueue.length === 0) return;
    
    this.processing = true;
    while (this.requestQueue.length > 0) {
      const { method, params, resolve, reject } = this.requestQueue.shift();
      try {
        const result = await this.client.request(method, params);
        resolve(result);
      } catch (error) {
        reject(error);
      }
      await new Promise(resolve => setTimeout(resolve, this.interval));
    }
    this.processing = false;
  }
}
```

## Network and Connectivity

### 1. Connection Refused

**Symptoms:**
```bash
curl: (7) Failed to connect to localhost port 8080: Connection refused
```

**Solutions:**
1. Verify server is running: `ps aux | grep server`
2. Check port availability: `netstat -tlnp | grep 8080`
3. Verify server configuration
4. Check firewall settings

### 2. SSL/TLS Issues

**Symptoms:**
```bash
curl: (60) SSL certificate problem: self signed certificate
```

**Solutions:**
```bash
# For development/testing only - skip SSL verification
curl -k -X POST https://localhost:8080/api/v1/mcp

# Better solution - add certificate to trust store
# or use proper SSL certificates
```

### 3. DNS Resolution Issues

**Symptoms:**
```bash
curl: (6) Could not resolve host: your-server.com
```

**Solutions:**
1. Check DNS configuration
2. Use IP address instead of hostname
3. Verify network connectivity
4. Check /etc/hosts file for local development

## Debugging Tools and Techniques

### 1. Request/Response Logging

#### Using curl with verbose output
```bash
curl -v -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {}
  }'
```

#### Using jq for JSON formatting
```bash
curl -X POST http://localhost:8080/api/v1/mcp \
  -H "Authorization: Bearer mcp_pat_your_token" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | jq '.'
```

### 2. Server Log Analysis

#### Check server logs
```bash
# If using systemd
journalctl -u your-service-name -f

# If using Docker
docker logs -f container-name

# If using direct execution
tail -f /var/log/your-app/app.log
```

#### Look for specific error patterns
```bash
# Search for authentication errors
grep -i "authentication\|unauthorized" /var/log/your-app/app.log

# Search for validation errors
grep -i "validation\|invalid.*param" /var/log/your-app/app.log

# Search for performance issues
grep -i "slow\|timeout\|performance" /var/log/your-app/app.log
```

### 3. Network Debugging

#### Test connectivity
```bash
# Basic connectivity
ping your-server.com

# Port connectivity
telnet your-server.com 8080
# or
nc -zv your-server.com 8080
```

#### Analyze HTTP traffic
```bash
# Using tcpdump
sudo tcpdump -i any -A -s 0 'port 8080'

# Using Wireshark for GUI analysis
# Capture on port 8080 and analyze HTTP traffic
```

### 4. JSON-RPC Validation Tools

#### Online JSON-RPC validators
- Use online JSON validators to check request format
- Validate JSON-RPC 2.0 compliance

#### Custom validation script
```python
import json
import jsonschema

# JSON-RPC 2.0 schema
jsonrpc_schema = {
    "type": "object",
    "properties": {
        "jsonrpc": {"const": "2.0"},
        "method": {"type": "string"},
        "params": {"type": ["object", "array"]},
        "id": {"type": ["string", "number", "null"]}
    },
    "required": ["jsonrpc", "method"],
    "additionalProperties": False
}

def validate_jsonrpc_request(request_data):
    try:
        jsonschema.validate(request_data, jsonrpc_schema)
        print("✅ Valid JSON-RPC 2.0 request")
        return True
    except jsonschema.ValidationError as e:
        print(f"❌ Invalid JSON-RPC request: {e.message}")
        return False

# Example usage
request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
        "name": "create_epic",
        "arguments": {"title": "Test", "priority": 1}
    }
}

validate_jsonrpc_request(request)
```

## Common Error Patterns and Solutions

### Error Pattern Matrix

| Error Code | Common Cause | Quick Fix |
|------------|--------------|-----------|
| -32700 | Invalid JSON | Validate JSON syntax |
| -32600 | Invalid Request | Check JSON-RPC format |
| -32601 | Method Not Found | Verify method name |
| -32602 | Invalid Params | Check required parameters |
| -32603 | Internal Error | Check server logs |
| -32002 | Not Found | Verify entity exists |

### Quick Diagnostic Checklist

1. **Authentication**
   - [ ] Authorization header present
   - [ ] Token format correct (mcp_pat_...)
   - [ ] Token not expired

2. **Request Format**
   - [ ] Content-Type: application/json
   - [ ] Valid JSON syntax
   - [ ] JSON-RPC 2.0 format
   - [ ] Required fields present

3. **Parameters**
   - [ ] All required parameters provided
   - [ ] Correct data types
   - [ ] Valid UUIDs/reference IDs
   - [ ] Length limits respected

4. **URIs (for resources/read)**
   - [ ] Valid scheme
   - [ ] Correct reference ID format
   - [ ] Scheme-prefix match
   - [ ] Supported sub-path

5. **Network**
   - [ ] Server running
   - [ ] Port accessible
   - [ ] Network connectivity
   - [ ] DNS resolution

## Getting Help

### Information to Collect

When reporting issues, include:

1. **Request Details**
   - Complete curl command or code
   - Request payload
   - Expected vs actual response

2. **Environment Information**
   - Server version
   - Client library version
   - Operating system
   - Network configuration

3. **Error Information**
   - Complete error response
   - Server logs (if accessible)
   - Timestamp of the issue

4. **Reproduction Steps**
   - Minimal example to reproduce
   - Frequency of the issue
   - Workarounds attempted

### Support Channels

1. Check server logs for detailed error information
2. Review this troubleshooting guide
3. Consult the [MCP API Documentation](mcp-api-documentation.md)
4. Test with the [MCP Testing Guide](mcp-testing-guide.md)
5. Contact your system administrator or development team