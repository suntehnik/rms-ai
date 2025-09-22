# Product Requirements Management API - Developer Guide

## Overview

This comprehensive developer guide provides everything you need to integrate with the Product Requirements Management API. The API enables management of product requirements through a hierarchical structure of Epics → User Stories → Requirements, with full support for comments, relationships, and deletion workflows.

## 📚 Documentation Resources

### Interactive Documentation
- **[Swagger UI](swagger-ui.html)** - Interactive API explorer with request/response examples
- **[Complete HTML Documentation](api-documentation.html)** - Comprehensive HTML reference

### Reference Documentation  
- **[Markdown Documentation](api-documentation.md)** - Complete API reference in Markdown format
- **[TypeScript Interfaces](api-types.ts)** - Generated TypeScript definitions for all API types
- **[JSON Schema](api-documentation.json)** - Machine-readable API documentation
- **[OpenAPI Specification](../openapi-v3.yaml)** - Complete OpenAPI 3.0.3 specification

## 🚀 Quick Start

### 1. Authentication

All API endpoints (except login and health checks) require JWT authentication:

```bash
# Login to get JWT token
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "your_username", "password": "your_password"}'

# Use token in subsequent requests
curl -X GET http://localhost:8080/api/v1/epics \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 2. Basic Operations

#### Create an Epic
```bash
curl -X POST http://localhost:8080/api/v1/epics \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "User Authentication System",
    "description": "Implement comprehensive user authentication",
    "priority": 1,
    "creator_id": "user-uuid"
  }'
```

#### List Epics with Filtering
```bash
curl -X GET "http://localhost:8080/api/v1/epics?status=In Progress&limit=10&include=creator,assignee" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Search Across All Entities
```bash
curl -X GET "http://localhost:8080/api/v1/search?q=authentication&entity_types=epic,user_story" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🏗️ API Architecture

### Base URLs
- **Development**: `http://localhost:8080`
- **Production**: `https://api.requirements.example.com`

### Entity Hierarchy
```
Epic (EP-001)
├── User Story (US-001)
│   ├── Acceptance Criteria (AC-001)
│   └── Requirements (REQ-001)
└── User Story (US-002)
    └── Requirements (REQ-002)
```

### Status Workflows

#### Epic & User Story Status Flow
```
Backlog → Draft → In Progress → Done
    ↓
Cancelled (from any state)
```

#### Requirement Status Flow
```
Draft → Active → Obsolete
```

### Priority Levels
- **1**: Critical (highest urgency)
- **2**: High (important)  
- **3**: Medium (normal)
- **4**: Low (can be deferred)

## 🔧 Client Implementation

### TypeScript/JavaScript Client

```typescript
import { ApiClient, Epic, CreateEpicRequest } from './api-types';

class RequirementsApiClient implements ApiClient {
  private baseUrl: string;
  private token?: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await fetch(`${this.baseUrl}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credentials)
    });
    
    const result = await response.json();
    this.token = result.token;
    return result;
  }

  async createEpic(epic: CreateEpicRequest): Promise<Epic> {
    const response = await fetch(`${this.baseUrl}/api/v1/epics`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`
      },
      body: JSON.stringify(epic)
    });
    
    return response.json();
  }

  async getEpics(params?: {
    status?: EpicStatus;
    priority?: Priority;
    limit?: number;
    offset?: number;
  }): Promise<EpicListResponse> {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          queryParams.append(key, value.toString());
        }
      });
    }

    const response = await fetch(
      `${this.baseUrl}/api/v1/epics?${queryParams}`,
      {
        headers: { 'Authorization': `Bearer ${this.token}` }
      }
    );
    
    return response.json();
  }
}
```

### Python Client Example

```python
import requests
from typing import Optional, Dict, Any

class RequirementsApiClient:
    def __init__(self, base_url: str):
        self.base_url = base_url
        self.token = None
        
    def login(self, username: str, password: str) -> Dict[str, Any]:
        response = requests.post(
            f"{self.base_url}/auth/login",
            json={"username": username, "password": password}
        )
        response.raise_for_status()
        result = response.json()
        self.token = result["token"]
        return result
        
    def _headers(self) -> Dict[str, str]:
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        return headers
        
    def create_epic(self, epic_data: Dict[str, Any]) -> Dict[str, Any]:
        response = requests.post(
            f"{self.base_url}/api/v1/epics",
            json=epic_data,
            headers=self._headers()
        )
        response.raise_for_status()
        return response.json()
        
    def get_epics(self, **params) -> Dict[str, Any]:
        response = requests.get(
            f"{self.base_url}/api/v1/epics",
            params=params,
            headers=self._headers()
        )
        response.raise_for_status()
        return response.json()
```

## 💬 Comment System Integration

### General Comments
```typescript
// Create a general comment on an epic
const comment = await client.createComment('epic', epicId, {
  content: "This epic needs more detailed requirements"
});

// Reply to a comment (threading)
const reply = await client.createComment('epic', epicId, {
  content: "I agree, let's schedule a requirements workshop",
  parent_comment_id: comment.id
});
```

### Inline Comments
```typescript
// Create inline comment linked to specific text
const inlineComment = await client.createInlineComment('epic', epicId, {
  content: "This section needs clarification",
  linked_text: "user authentication flow",
  text_position_start: 45,
  text_position_end: 68
});

// Validate inline comments after content changes
const validation = await client.validateInlineComments('epic', epicId, {
  comments: [{
    comment_id: inlineComment.id,
    text_position_start: 45,
    text_position_end: 68
  }]
});
```

### Comment Resolution
```typescript
// Resolve a comment
await client.resolveComment(commentId);

// Get comments by status
const unresolvedComments = await client.getCommentsByStatus('unresolved');
```

## 🗑️ Safe Deletion Workflows

### Deletion Process
```typescript
// 1. Validate deletion to check dependencies
const validation = await client.validateEpicDeletion(epicId);

if (!validation.can_delete) {
  // Show dependencies to user
  console.log('Cannot delete due to dependencies:', validation.dependencies);
  console.log('Warnings:', validation.warnings);
  return;
}

// 2. Perform comprehensive deletion
const result = await client.deleteEpicComprehensive(epicId);

if (result.success) {
  console.log('Deleted entities:', result.deleted_entities);
  // Update UI to reflect deletions
} else {
  console.error('Deletion failed:', result.message);
}
```

### Dependency Types
- **child**: Direct child entities (e.g., User Stories under Epic)
- **reference**: Entities that reference this entity
- **relationship**: Requirement relationships that would be broken

## 🔍 Search Implementation

### Global Search
```typescript
// Search across all entity types
const results = await client.search({
  q: "authentication",
  entity_types: "epic,user_story,requirement",
  limit: 20
});

// Process search results
results.results.forEach(result => {
  console.log(`${result.entity_type}: ${result.reference_id} - ${result.title}`);
  if (result.highlight) {
    console.log(`Highlight: ${result.highlight}`);
  }
});
```

### Search Suggestions (Autocomplete)
```typescript
// Get suggestions for autocomplete
const suggestions = await client.getSearchSuggestions({
  query: "auth",
  limit: 10
});

// Use suggestions in UI
suggestions.titles.forEach(title => {
  // Add to autocomplete dropdown
});
```

## 📊 Pagination & Filtering

### Standard Pagination
```typescript
// All list endpoints support pagination
const epics = await client.getEpics({
  limit: 50,        // Items per page (max 100)
  offset: 0,        // Items to skip
  order_by: "created_at",
  status: "In Progress",
  priority: 1
});

console.log(`Showing ${epics.data.length} of ${epics.total_count} epics`);
```

### Including Related Data
```typescript
// Include related entities in response
const epics = await client.getEpics({
  include: "creator,assignee,user_stories,comments"
});

// Access populated fields
epics.data.forEach(epic => {
  console.log(`Epic: ${epic.title}`);
  console.log(`Creator: ${epic.creator?.username}`);
  console.log(`User Stories: ${epic.user_stories?.length || 0}`);
});
```

## ⚠️ Error Handling

### Standard Error Response
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed"
  }
}
```

### Common Error Codes
- **VALIDATION_ERROR**: Request validation failed
- **AUTHENTICATION_REQUIRED**: JWT token required
- **INSUFFICIENT_PERMISSIONS**: User lacks required permissions
- **ENTITY_NOT_FOUND**: Requested entity doesn't exist
- **DELETION_CONFLICT**: Entity has dependencies preventing deletion
- **INTERNAL_ERROR**: Server-side error

### Error Handling Example
```typescript
try {
  const epic = await client.createEpic(epicData);
} catch (error) {
  if (error.response?.status === 400) {
    const errorData = error.response.data;
    if (errorData.error.code === 'VALIDATION_ERROR') {
      // Handle validation errors
      console.error('Validation failed:', errorData.error.message);
    }
  } else if (error.response?.status === 401) {
    // Handle authentication errors
    console.error('Authentication required');
    // Redirect to login
  } else if (error.response?.status === 403) {
    // Handle authorization errors
    console.error('Insufficient permissions');
  }
}
```

## 🔐 Security Best Practices

### JWT Token Management
```typescript
class TokenManager {
  private token?: string;
  private refreshToken?: string;
  private expiresAt?: Date;

  setTokens(loginResponse: LoginResponse) {
    this.token = loginResponse.token;
    this.expiresAt = new Date(loginResponse.expires_at);
  }

  getValidToken(): string | null {
    if (!this.token || !this.expiresAt) {
      return null;
    }

    // Check if token expires within 5 minutes
    const fiveMinutesFromNow = new Date(Date.now() + 5 * 60 * 1000);
    if (this.expiresAt <= fiveMinutesFromNow) {
      // Token is expired or about to expire
      return null;
    }

    return this.token;
  }

  clearTokens() {
    this.token = undefined;
    this.refreshToken = undefined;
    this.expiresAt = undefined;
  }
}
```

### Request Interceptors
```typescript
// Add token to all requests
axios.interceptors.request.use(
  (config) => {
    const token = tokenManager.getValidToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Handle authentication errors
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      tokenManager.clearTokens();
      // Redirect to login
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

## 🧪 Testing Your Integration

### Unit Tests
```typescript
describe('RequirementsApiClient', () => {
  let client: RequirementsApiClient;
  let mockFetch: jest.MockedFunction<typeof fetch>;

  beforeEach(() => {
    mockFetch = jest.fn();
    global.fetch = mockFetch;
    client = new RequirementsApiClient('http://localhost:8080');
  });

  it('should create epic successfully', async () => {
    const mockEpic = { id: '123', title: 'Test Epic', reference_id: 'EP-001' };
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockEpic
    } as Response);

    const result = await client.createEpic({
      title: 'Test Epic',
      priority: 1,
      creator_id: 'user-123'
    });

    expect(result).toEqual(mockEpic);
    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/epics',
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({
          'Content-Type': 'application/json'
        })
      })
    );
  });
});
```

### Integration Tests
```typescript
describe('API Integration', () => {
  let client: RequirementsApiClient;
  let authToken: string;

  beforeAll(async () => {
    client = new RequirementsApiClient('http://localhost:8080');
    
    // Login with test credentials
    const loginResponse = await client.login({
      username: 'test_user',
      password: 'test_password'
    });
    authToken = loginResponse.token;
  });

  it('should create and retrieve epic', async () => {
    // Create epic
    const epic = await client.createEpic({
      title: 'Integration Test Epic',
      priority: 2,
      creator_id: 'test-user-id'
    });

    expect(epic.title).toBe('Integration Test Epic');
    expect(epic.reference_id).toMatch(/^EP-\d+$/);

    // Retrieve epic
    const retrievedEpic = await client.getEpic(epic.id);
    expect(retrievedEpic.id).toBe(epic.id);
  });
});
```

## 📈 Performance Optimization

### Caching Strategy
```typescript
class CachedApiClient {
  private cache = new Map<string, { data: any; expires: number }>();
  private client: RequirementsApiClient;

  constructor(client: RequirementsApiClient) {
    this.client = client;
  }

  async getEpic(id: string): Promise<Epic> {
    const cacheKey = `epic:${id}`;
    const cached = this.cache.get(cacheKey);
    
    if (cached && cached.expires > Date.now()) {
      return cached.data;
    }

    const epic = await this.client.getEpic(id);
    
    // Cache for 5 minutes
    this.cache.set(cacheKey, {
      data: epic,
      expires: Date.now() + 5 * 60 * 1000
    });

    return epic;
  }

  invalidateCache(pattern?: string) {
    if (pattern) {
      for (const key of this.cache.keys()) {
        if (key.includes(pattern)) {
          this.cache.delete(key);
        }
      }
    } else {
      this.cache.clear();
    }
  }
}
```

### Batch Operations
```typescript
// Instead of multiple individual requests
const epics = await Promise.all([
  client.getEpic('epic1'),
  client.getEpic('epic2'),
  client.getEpic('epic3')
]);

// Use list endpoint with filtering
const epics = await client.getEpics({
  // Filter to get specific epics if possible
  include: 'creator,assignee'
});
```

## 🔄 Real-time Updates

### WebSocket Integration (Future Enhancement)
```typescript
class RealtimeApiClient extends RequirementsApiClient {
  private ws?: WebSocket;
  private eventHandlers = new Map<string, Function[]>();

  connect() {
    this.ws = new WebSocket('ws://localhost:8080/ws');
    
    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.handleEvent(data.type, data.payload);
    };
  }

  on(eventType: string, handler: Function) {
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, []);
    }
    this.eventHandlers.get(eventType)!.push(handler);
  }

  private handleEvent(type: string, payload: any) {
    const handlers = this.eventHandlers.get(type) || [];
    handlers.forEach(handler => handler(payload));
  }
}

// Usage
const realtimeClient = new RealtimeApiClient('http://localhost:8080');
realtimeClient.connect();

realtimeClient.on('epic.updated', (epic) => {
  console.log('Epic updated:', epic);
  // Update UI
});

realtimeClient.on('comment.created', (comment) => {
  console.log('New comment:', comment);
  // Show notification
});
```

## 📋 Configuration Management

### Admin Operations
```typescript
// Manage requirement types (Admin only)
const requirementTypes = await client.getRequirementTypes();

const newType = await client.createRequirementType({
  name: 'Security Requirement',
  description: 'Requirements related to security and compliance'
});

// Manage relationship types
const relationshipTypes = await client.getRelationshipTypes();

const newRelationType = await client.createRelationshipType({
  name: 'implements',
  description: 'Indicates that one requirement implements another'
});
```

## 🚀 Deployment Considerations

### Environment Configuration
```typescript
interface ApiConfig {
  baseUrl: string;
  timeout: number;
  retryAttempts: number;
  retryDelay: number;
}

const configs: Record<string, ApiConfig> = {
  development: {
    baseUrl: 'http://localhost:8080',
    timeout: 30000,
    retryAttempts: 3,
    retryDelay: 1000
  },
  staging: {
    baseUrl: 'https://api-staging.requirements.example.com',
    timeout: 15000,
    retryAttempts: 2,
    retryDelay: 2000
  },
  production: {
    baseUrl: 'https://api.requirements.example.com',
    timeout: 10000,
    retryAttempts: 1,
    retryDelay: 3000
  }
};
```

### Rate Limiting Handling
```typescript
class RateLimitedClient {
  private requestQueue: Array<() => Promise<any>> = [];
  private processing = false;
  private lastRequestTime = 0;
  private minInterval = 100; // Minimum time between requests (ms)

  async makeRequest<T>(requestFn: () => Promise<T>): Promise<T> {
    return new Promise((resolve, reject) => {
      this.requestQueue.push(async () => {
        try {
          const result = await requestFn();
          resolve(result);
        } catch (error) {
          reject(error);
        }
      });

      this.processQueue();
    });
  }

  private async processQueue() {
    if (this.processing || this.requestQueue.length === 0) {
      return;
    }

    this.processing = true;

    while (this.requestQueue.length > 0) {
      const now = Date.now();
      const timeSinceLastRequest = now - this.lastRequestTime;

      if (timeSinceLastRequest < this.minInterval) {
        await new Promise(resolve => 
          setTimeout(resolve, this.minInterval - timeSinceLastRequest)
        );
      }

      const request = this.requestQueue.shift()!;
      this.lastRequestTime = Date.now();

      try {
        await request();
      } catch (error) {
        if (error.response?.status === 429) {
          // Rate limited, wait and retry
          const retryAfter = error.response.headers['retry-after'] || 60;
          await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
          this.requestQueue.unshift(request); // Put request back at front
        }
      }
    }

    this.processing = false;
  }
}
```

## 📞 Support & Resources

### Getting Help
- **API Documentation**: Complete reference in multiple formats
- **Interactive Testing**: Use Swagger UI for live API exploration
- **TypeScript Support**: Full type definitions for type-safe development
- **Error Codes**: Comprehensive error handling documentation

### Best Practices Summary
1. **Always authenticate** requests with valid JWT tokens
2. **Handle errors gracefully** with proper error codes
3. **Use pagination** for large data sets
4. **Cache frequently accessed data** to improve performance
5. **Validate deletions** before performing destructive operations
6. **Include related data** when needed to reduce API calls
7. **Implement retry logic** for transient failures
8. **Monitor rate limits** and implement backoff strategies

### Community & Contributions
- Report issues and request features through the project repository
- Contribute to documentation improvements
- Share integration examples and best practices

---

*This developer guide is generated from OpenAPI specification version 1.0.0*
*Last updated: Auto-generated*