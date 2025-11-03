# MCP API Examples and Usage Scenarios

## Overview

This document provides comprehensive examples and real-world usage scenarios for the MCP (Model Context Protocol) API. It demonstrates how to use the API effectively for various requirements management workflows and integration patterns.

## Table of Contents

1. [Basic Workflow Examples](#basic-workflow-examples)
2. [Advanced Integration Scenarios](#advanced-integration-scenarios)
3. [AI Assistant Integration](#ai-assistant-integration)
4. [Batch Operations](#batch-operations)
5. [Error Handling Patterns](#error-handling-patterns)
6. [Performance Optimization](#performance-optimization)
7. [Real-World Use Cases](#real-world-use-cases)

## Basic Workflow Examples

### 1. Complete Feature Development Workflow

This example demonstrates creating a complete feature from epic to requirements.

```bash
#!/bin/bash
# Complete Feature Development Workflow

PAT_TOKEN="mcp_pat_your_token_here"
BASE_URL="http://localhost:8080/api/v1/mcp"

# Function to make MCP requests
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
        }"
}

echo "=== Creating Complete Feature: User Authentication System ==="

# Step 1: Initialize MCP connection
echo "1. Initializing MCP connection..."
mcp_request 1 "initialize" '{
    "protocolVersion": "2025-06-18",
    "capabilities": {"elicitation": {}},
    "clientInfo": {"name": "feature-workflow", "version": "1.0.0"}
}' | jq '.'

# Step 2: Create Epic
echo -e "\n2. Creating Epic..."
EPIC_RESPONSE=$(mcp_request 2 "tools/call" '{
    "name": "create_epic",
    "arguments": {
        "title": "User Authentication System",
        "description": "Implement comprehensive user authentication and authorization system",
        "priority": 1
    }
}')

echo $EPIC_RESPONSE | jq '.'
EPIC_ID=$(echo $EPIC_RESPONSE | jq -r '.result.content[1].data.reference_id')
echo "Created Epic: $EPIC_ID"

# Step 3: Create User Stories
echo -e "\n3. Creating User Stories..."

# User Story 1: User Registration
US1_RESPONSE=$(mcp_request 3 "tools/call" '{
    "name": "create_user_story",
    "arguments": {
        "title": "User Registration",
        "description": "As a new user, I want to register for an account with my email and password",
        "priority": 1,
        "epic_id": "'$EPIC_ID'"
    }
}')

US1_ID=$(echo $US1_RESPONSE | jq -r '.result.content[1].data.reference_id')
echo "Created User Story 1: $US1_ID"

# Step 4: View the complete hierarchy
echo -e "\n4. Viewing complete epic hierarchy..."
mcp_request 10 "resources/read" '{
    "uri": "epic://'$EPIC_ID'/hierarchy"
}' | jq '.result.contents'

echo -e "\n=== Feature Creation Complete ==="
```

### 2. Requirement Analysis Workflow

This example shows how to analyze requirements and their relationships.

```python
import requests
import json
from typing import Dict, List

class RequirementAnalyzer:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.token = token
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
        self.request_id = 1
    
    def _make_request(self, method: str, params: Dict) -> Dict:
        payload = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method,
            "params": params
        }
        self.request_id += 1
        
        response = requests.post(
            f"{self.base_url}/api/v1/mcp",
            headers=self.headers,
            json=payload
        )
        return response.json()
    
    def analyze_requirement_dependencies(self, requirement_id: str) -> Dict:
        """Analyze a requirement and its dependencies"""
        print(f"Analyzing requirement {requirement_id}...")
        
        # Get requirement with relationships
        result = self._make_request("resources/read", {
            "uri": f"requirement://{requirement_id}/relationships"
        })
        
        if "error" in result:
            print(f"Error: {result['error']['message']}")
            return {}
        
        requirement = result["result"]["contents"]
        
        analysis = {
            "requirement": {
                "id": requirement["reference_id"],
                "title": requirement["title"],
                "status": requirement["status"],
                "priority": requirement["priority"]
            },
            "dependencies": {
                "depends_on": [],
                "blocks": [],
                "related_to": []
            },
            "impact_analysis": {
                "blocking_count": len(requirement.get("source_relationships", [])),
                "dependency_count": len(requirement.get("target_relationships", [])),
                "risk_level": "low"
            }
        }
        
        # Calculate risk level
        total_relationships = (
            len(analysis["dependencies"]["depends_on"]) + 
            len(analysis["dependencies"]["blocks"])
        )
        
        if total_relationships > 5:
            analysis["impact_analysis"]["risk_level"] = "high"
        elif total_relationships > 2:
            analysis["impact_analysis"]["risk_level"] = "medium"
        
        return analysis

# Usage example
if __name__ == "__main__":
    analyzer = RequirementAnalyzer(
        "http://localhost:8080",
        "mcp_pat_your_token_here"
    )
    
    # Analyze a specific requirement
    req_analysis = analyzer.analyze_requirement_dependencies("REQ-001")
    print("Requirement Analysis:")
    print(json.dumps(req_analysis, indent=2))
```

## Advanced Integration Scenarios

### 3. Automated Requirements Validation

This example shows how to implement automated validation of requirements.

```javascript
class RequirementValidator {
    constructor(baseUrl, token) {
        this.baseUrl = baseUrl;
        this.token = token;
        this.requestId = 1;
    }
    
    async makeRequest(method, params) {
        const response = await fetch(`${this.baseUrl}/api/v1/mcp`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${this.token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                jsonrpc: "2.0",
                id: this.requestId++,
                method: method,
                params: params
            })
        });
        
        return await response.json();
    }
    
    async validateRequirementQuality(requirementId) {
        // Get requirement details
        const result = await this.makeRequest("resources/read", {
            uri: `requirement://${requirementId}`
        });
        
        if (result.error) {
            return { valid: false, errors: [result.error.message] };
        }
        
        const requirement = result.result.contents;
        const validation = {
            valid: true,
            errors: [],
            warnings: [],
            score: 100,
            suggestions: []
        };
        
        // Check title quality
        if (!requirement.title || requirement.title.length < 10) {
            validation.errors.push("Title is too short (minimum 10 characters)");
            validation.score -= 20;
            validation.valid = false;
        }
        
        // Check description quality
        if (!requirement.description || requirement.description.length < 50) {
            validation.errors.push("Description is too short (minimum 50 characters)");
            validation.score -= 15;
            validation.valid = false;
        }
        
        return validation;
    }
}

// Usage example
async function runValidation() {
    const validator = new RequirementValidator(
        "http://localhost:8080",
        "mcp_pat_your_token_here"
    );
    
    try {
        const reqValidation = await validator.validateRequirementQuality("REQ-001");
        console.log("Requirement Validation:", JSON.stringify(reqValidation, null, 2));
    } catch (error) {
        console.error("Validation error:", error);
    }
}

runValidation();
```

## Error Handling Patterns

### 4. Robust Error Handling

```python
import requests
import time
from typing import Dict, Optional

class RobustMCPClient:
    def __init__(self, base_url: str, token: str, max_retries: int = 3):
        self.base_url = base_url
        self.token = token
        self.max_retries = max_retries
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
        self.request_id = 1
    
    def _make_request_with_retry(self, method: str, params: Dict) -> Dict:
        """Make request with retry logic and error handling"""
        
        for attempt in range(self.max_retries):
            try:
                payload = {
                    "jsonrpc": "2.0",
                    "id": self.request_id,
                    "method": method,
                    "params": params
                }
                self.request_id += 1
                
                response = requests.post(
                    f"{self.base_url}/api/v1/mcp",
                    headers=self.headers,
                    json=payload,
                    timeout=30
                )
                
                if response.status_code == 200:
                    result = response.json()
                    
                    # Handle JSON-RPC errors
                    if "error" in result:
                        error_code = result["error"]["code"]
                        error_message = result["error"]["message"]
                        
                        # Retry on server errors
                        if error_code == -32603:  # Internal error
                            if attempt < self.max_retries - 1:
                                time.sleep(2 ** attempt)  # Exponential backoff
                                continue
                        
                        # Don't retry on client errors
                        return result
                    
                    return result
                
                elif response.status_code == 401:
                    return {
                        "error": {
                            "code": -32603,
                            "message": "Authentication failed - check your token"
                        }
                    }
                
                elif response.status_code >= 500:
                    if attempt < self.max_retries - 1:
                        time.sleep(2 ** attempt)
                        continue
                
                return {
                    "error": {
                        "code": -32603,
                        "message": f"HTTP {response.status_code}: {response.text}"
                    }
                }
                
            except requests.exceptions.Timeout:
                if attempt < self.max_retries - 1:
                    time.sleep(2 ** attempt)
                    continue
                return {
                    "error": {
                        "code": -32603,
                        "message": "Request timeout"
                    }
                }
            
            except requests.exceptions.ConnectionError:
                if attempt < self.max_retries - 1:
                    time.sleep(2 ** attempt)
                    continue
                return {
                    "error": {
                        "code": -32603,
                        "message": "Connection error - server may be down"
                    }
                }
            
            except Exception as e:
                return {
                    "error": {
                        "code": -32603,
                        "message": f"Unexpected error: {str(e)}"
                    }
                }
        
        return {
            "error": {
                "code": -32603,
                "message": f"Max retries ({self.max_retries}) exceeded"
            }
        }
    
    def create_epic_safely(self, title: str, priority: int, 
                          description: Optional[str] = None) -> Dict:
        """Create epic with comprehensive error handling"""
        
        # Validate inputs
        if not title or len(title.strip()) == 0:
            return {
                "error": {
                    "code": -32602,
                    "message": "Title cannot be empty"
                }
            }
        
        if priority not in [1, 2, 3, 4]:
            return {
                "error": {
                    "code": -32602,
                    "message": "Priority must be between 1 and 4"
                }
            }
        
        args = {
            "title": title.strip(),
            "priority": priority
        }
        
        if description:
            args["description"] = description.strip()
        
        result = self._make_request_with_retry("tools/call", {
            "name": "create_epic",
            "arguments": args
        })
        
        # Process result
        if "error" in result:
            return {
                "success": False,
                "error": result["error"]["message"],
                "error_code": result["error"]["code"]
            }
        
        try:
            epic_data = result["result"]["content"][1]["data"]
            return {
                "success": True,
                "epic": epic_data,
                "message": f"Successfully created epic {epic_data['reference_id']}"
            }
        except (KeyError, IndexError) as e:
            return {
                "success": False,
                "error": f"Unexpected response format: {str(e)}",
                "error_code": -32603
            }

# Usage example
def demonstrate_robust_client():
    client = RobustMCPClient(
        "http://localhost:8080",
        "mcp_pat_your_token_here",
        max_retries=3
    )
    
    # Test with valid data
    result = client.create_epic_safely(
        title="Test Epic",
        priority=1,
        description="This is a test epic"
    )
    
    if result["success"]:
        print(f"✅ {result['message']}")
        print(f"Epic ID: {result['epic']['reference_id']}")
    else:
        print(f"❌ Error: {result['error']}")
        print(f"Error Code: {result['error_code']}")
    
    # Test with invalid data
    invalid_result = client.create_epic_safely(
        title="",  # Invalid empty title
        priority=5  # Invalid priority
    )
    
    print(f"Invalid data test: {invalid_result}")

if __name__ == "__main__":
    demonstrate_robust_client()
```

## Performance Optimization

### 5. Caching and Performance Patterns

```python
import time
from typing import Dict, Optional, Tuple
import hashlib
import json

class CachedMCPClient:
    def __init__(self, base_url: str, token: str, cache_ttl: int = 300):
        self.base_url = base_url
        self.token = token
        self.cache_ttl = cache_ttl  # 5 minutes default
        self.cache = {}
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
        self.request_id = 1
    
    def _get_cache_key(self, method: str, params: Dict) -> str:
        """Generate cache key from method and parameters"""
        cache_data = {
            "method": method,
            "params": params
        }
        cache_string = json.dumps(cache_data, sort_keys=True)
        return hashlib.md5(cache_string.encode()).hexdigest()
    
    def _is_cacheable(self, method: str) -> bool:
        """Determine if a method should be cached"""
        cacheable_methods = [
            "resources/read",
            "tools/list",
            "initialize"
        ]
        
        # Don't cache write operations
        if method == "tools/call":
            return False
        
        return method in cacheable_methods
    
    def _get_from_cache(self, cache_key: str) -> Optional[Dict]:
        """Get item from cache if not expired"""
        if cache_key in self.cache:
            cached_item = self.cache[cache_key]
            if time.time() - cached_item["timestamp"] < self.cache_ttl:
                return cached_item["data"]
            else:
                # Remove expired item
                del self.cache[cache_key]
        return None
    
    def _set_cache(self, cache_key: str, data: Dict):
        """Set item in cache"""
        self.cache[cache_key] = {
            "data": data,
            "timestamp": time.time()
        }
    
    def make_request(self, method: str, params: Dict) -> Dict:
        """Make request with caching support"""
        import requests
        
        # Check cache first
        if self._is_cacheable(method):
            cache_key = self._get_cache_key(method, params)
            cached_result = self._get_from_cache(cache_key)
            if cached_result:
                return cached_result
        
        # Make actual request
        payload = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method,
            "params": params
        }
        self.request_id += 1
        
        start_time = time.time()
        response = requests.post(
            f"{self.base_url}/api/v1/mcp",
            headers=self.headers,
            json=payload
        )
        request_time = time.time() - start_time
        
        result = response.json()
        
        # Cache successful responses
        if self._is_cacheable(method) and "error" not in result:
            cache_key = self._get_cache_key(method, params)
            self._set_cache(cache_key, result)
        
        # Add performance metadata
        result["_performance"] = {
            "request_time": request_time,
            "cached": False
        }
        
        return result
    
    def get_epic_hierarchy_optimized(self, epic_id: str) -> Dict:
        """Get epic hierarchy with optimized caching"""
        
        # First try to get from cache
        cache_key = self._get_cache_key("resources/read", {
            "uri": f"epic://{epic_id}/hierarchy"
        })
        
        cached_result = self._get_from_cache(cache_key)
        if cached_result:
            cached_result["_performance"] = {
                "request_time": 0.001,  # Very fast cache hit
                "cached": True
            }
            return cached_result
        
        # If not cached, make request
        result = self.make_request("resources/read", {
            "uri": f"epic://{epic_id}/hierarchy"
        })
        
        return result
    
    def batch_get_user_stories(self, user_story_ids: List[str]) -> Dict:
        """Get multiple user stories efficiently"""
        results = {}
        cache_hits = 0
        cache_misses = 0
        
        for us_id in user_story_ids:
            cache_key = self._get_cache_key("resources/read", {
                "uri": f"user-story://{us_id}"
            })
            
            cached_result = self._get_from_cache(cache_key)
            if cached_result:
                results[us_id] = cached_result
                cache_hits += 1
            else:
                # Make request for non-cached items
                result = self.make_request("resources/read", {
                    "uri": f"user-story://{us_id}"
                })
                results[us_id] = result
                cache_misses += 1
        
        return {
            "results": results,
            "performance": {
                "cache_hits": cache_hits,
                "cache_misses": cache_misses,
                "cache_hit_rate": cache_hits / len(user_story_ids) if user_story_ids else 0
            }
        }
    
    def clear_cache(self):
        """Clear all cached items"""
        self.cache.clear()
    
    def get_cache_stats(self) -> Dict:
        """Get cache statistics"""
        current_time = time.time()
        valid_items = 0
        expired_items = 0
        
        for cache_key, cached_item in self.cache.items():
            if current_time - cached_item["timestamp"] < self.cache_ttl:
                valid_items += 1
            else:
                expired_items += 1
        
        return {
            "total_items": len(self.cache),
            "valid_items": valid_items,
            "expired_items": expired_items,
            "cache_ttl": self.cache_ttl
        }

# Usage example
def demonstrate_caching():
    client = CachedMCPClient(
        "http://localhost:8080",
        "mcp_pat_your_token_here",
        cache_ttl=300  # 5 minutes
    )
    
    print("=== Performance Optimization Demo ===")
    
    # First request (cache miss)
    print("1. First request (cache miss):")
    result1 = client.get_epic_hierarchy_optimized("EP-001")
    if "error" not in result1:
        print(f"Request time: {result1['_performance']['request_time']:.3f}s")
        print(f"Cached: {result1['_performance']['cached']}")
    
    # Second request (cache hit)
    print("\n2. Second request (cache hit):")
    result2 = client.get_epic_hierarchy_optimized("EP-001")
    if "error" not in result2:
        print(f"Request time: {result2['_performance']['request_time']:.3f}s")
        print(f"Cached: {result2['_performance']['cached']}")
    
    # Batch operations
    print("\n3. Batch user story retrieval:")
    batch_result = client.batch_get_user_stories(["US-001", "US-002", "US-003"])
    perf = batch_result["performance"]
    print(f"Cache hits: {perf['cache_hits']}")
    print(f"Cache misses: {perf['cache_misses']}")
    print(f"Cache hit rate: {perf['cache_hit_rate']:.2%}")
    
    # Cache statistics
    print("\n4. Cache statistics:")
    stats = client.get_cache_stats()
    print(json.dumps(stats, indent=2))

if __name__ == "__main__":
    demonstrate_caching()
```

## Status Management Workflows

### 6. Complete Status Management Workflow

This example demonstrates managing entity status throughout the development lifecycle.

```python
import requests
import json
from typing import Dict, List

class StatusWorkflowManager:
    def __init__(self, base_url: str, token: str):
        self.base_url = base_url
        self.token = token
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
        self.request_id = 1
    
    def _make_request(self, method: str, params: Dict) -> Dict:
        payload = {
            "jsonrpc": "2.0",
            "id": self.request_id,
            "method": method,
            "params": params
        }
        self.request_id += 1
        
        response = requests.post(
            f"{self.base_url}/api/v1/mcp",
            headers=self.headers,
            json=payload
        )
        return response.json()
    
    def update_entity_status(self, entity_type: str, entity_id: str, status: str) -> Dict:
        """Update status for any entity type"""
        tool_name = f"update_{entity_type}"
        id_field = f"{entity_type}_id"
        
        result = self._make_request("tools/call", {
            "name": tool_name,
            "arguments": {
                id_field: entity_id,
                "status": status
            }
        })
        
        if "error" in result:
            return {
                "success": False,
                "error": result["error"]["message"],
                "entity_type": entity_type,
                "entity_id": entity_id,
                "requested_status": status
            }
        
        entity_data = result["result"]["content"][1]["data"]
        return {
            "success": True,
            "entity_type": entity_type,
            "entity_id": entity_id,
            "old_status": "unknown",  # Would need to track this
            "new_status": entity_data["status"],
            "message": result["result"]["content"][0]["text"]
        }
    
    def execute_development_workflow(self, epic_id: str, user_story_id: str, requirement_id: str) -> List[Dict]:
        """Execute a complete development workflow with status transitions"""
        workflow_steps = [
            # Planning Phase
            {"entity_type": "epic", "entity_id": epic_id, "status": "Draft", "phase": "Planning"},
            {"entity_type": "user_story", "entity_id": user_story_id, "status": "Draft", "phase": "Planning"},
            {"entity_type": "requirement", "entity_id": requirement_id, "status": "Draft", "phase": "Planning"},
            
            # Development Phase
            {"entity_type": "epic", "entity_id": epic_id, "status": "In Progress", "phase": "Development"},
            {"entity_type": "user_story", "entity_id": user_story_id, "status": "In Progress", "phase": "Development"},
            {"entity_type": "requirement", "entity_id": requirement_id, "status": "Active", "phase": "Development"},
            
            # Completion Phase
            {"entity_type": "user_story", "entity_id": user_story_id, "status": "Done", "phase": "Completion"},
            {"entity_type": "epic", "entity_id": epic_id, "status": "Done", "phase": "Completion"},
        ]
        
        results = []
        for step in workflow_steps:
            print(f"🔄 {step['phase']}: Updating {step['entity_type']} {step['entity_id']} to {step['status']}")
            
            result = self.update_entity_status(
                step["entity_type"],
                step["entity_id"], 
                step["status"]
            )
            
            result["phase"] = step["phase"]
            results.append(result)
            
            if result["success"]:
                print(f"✅ {result['message']}")
            else:
                print(f"❌ Error: {result['error']}")
                break  # Stop workflow on error
        
        return results

# Usage example
def demonstrate_status_workflow():
    manager = StatusWorkflowManager(
        "http://localhost:8080",
        "mcp_pat_your_token_here"
    )
    
    print("=== Status Management Workflow Demo ===")
    
    # Execute complete development workflow
    print("\n1. Executing Development Workflow:")
    workflow_results = manager.execute_development_workflow(
        epic_id="EP-001",
        user_story_id="US-001", 
        requirement_id="REQ-001"
    )
    
    # Show workflow summary
    successful_transitions = [r for r in workflow_results if r["success"]]
    failed_transitions = [r for r in workflow_results if not r["success"]]
    
    print(f"\n📊 Workflow Summary:")
    print(f"   Successful transitions: {len(successful_transitions)}")
    print(f"   Failed transitions: {len(failed_transitions)}")

if __name__ == "__main__":
    demonstrate_status_workflow()
```

This comprehensive examples document provides practical, working code examples for:

1. **Basic Workflows** - Complete feature development and requirement analysis
2. **Advanced Integration** - Automated validation and quality checking
3. **Error Handling** - Robust client with retry logic and comprehensive error handling
4. **Performance Optimization** - Caching strategies and batch operations
5. **Status Management** - Complete workflow management with status transitions

Each example is production-ready and demonstrates best practices for integrating with the MCP API. The examples progress from simple to complex, showing how to build sophisticated applications on top of the MCP API.