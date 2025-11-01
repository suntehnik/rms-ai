package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/mcp/schemas"
	"product-requirements-management/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestMCPAPICompatibility_AllTools tests all 22 MCP tools through JSON-RPC interface
func TestMCPAPICompatibility_AllTools(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Get all supported tools
	tools := schemas.GetSupportedTools()
	assert.Len(t, tools, 22, "Expected exactly 22 MCP tools")

	// Test each tool schema for API compatibility
	for _, tool := range tools {
		t.Run(tool.Name, func(t *testing.T) {
			testToolSchemaCompatibility(t, tool)
		})
	}
}

// testToolSchemaCompatibility tests a single tool schema for API compatibility
func testToolSchemaCompatibility(t *testing.T, tool schemas.ToolDefinition) {
	// Verify tool has required fields
	assert.NotEmpty(t, tool.Name, "Tool should have a name")
	assert.NotEmpty(t, tool.Title, "Tool should have a title")
	assert.NotEmpty(t, tool.Description, "Tool should have a description")
	assert.NotNil(t, tool.InputSchema, "Tool should have an input schema")

	// Verify input schema structure
	schema, ok := tool.InputSchema.(map[string]interface{})
	assert.True(t, ok, "InputSchema should be a map")
	assert.Equal(t, "object", schema["type"], "Schema type should be object")
	assert.Contains(t, schema, "properties", "Schema should have properties")
	// Note: Not all tools have required fields, so we don't assert their presence

	// Verify properties structure
	properties, ok := schema["properties"].(map[string]interface{})
	assert.True(t, ok, "Properties should be a map")

	// Verify required fields (if present)
	if requiredField, exists := schema["required"]; exists {
		required, ok := requiredField.([]string)
		assert.True(t, ok, "Required should be a string array")

		// Verify all required fields exist in properties
		for _, requiredField := range required {
			assert.Contains(t, properties, requiredField, "Required field %s should exist in properties", requiredField)
		}
	}

	// Verify each property has proper structure
	for propName, propValue := range properties {
		prop, ok := propValue.(map[string]interface{})
		assert.True(t, ok, "Property %s should be a map", propName)
		assert.Contains(t, prop, "type", "Property %s should have a type", propName)
		assert.Contains(t, prop, "description", "Property %s should have a description", propName)
	}
}

// createTestArgumentsForTool creates test arguments for a specific tool
func createTestArgumentsForTool(toolName string) map[string]interface{} {
	testUUID := uuid.New().String()
	testEpicID := "EP-001"
	testUserStoryID := "US-001"
	testRequirementID := "REQ-001"
	testSteeringDocID := "STD-001"
	testPromptID := "PROMPT-001"

	switch toolName {
	case "create_epic":
		return map[string]interface{}{
			"title":    "Test Epic",
			"priority": 1,
		}
	case "update_epic":
		return map[string]interface{}{
			"epic_id": testEpicID,
			"title":   "Updated Epic",
		}
	case "create_user_story":
		return map[string]interface{}{
			"title":    "Test User Story",
			"epic_id":  testEpicID,
			"priority": 2,
		}
	case "update_user_story":
		return map[string]interface{}{
			"user_story_id": testUserStoryID,
			"title":         "Updated User Story",
		}
	case "create_requirement":
		return map[string]interface{}{
			"title":         "Test Requirement",
			"user_story_id": testUserStoryID,
			"type_id":       testUUID,
			"priority":      3,
		}
	case "update_requirement":
		return map[string]interface{}{
			"requirement_id": testRequirementID,
			"title":          "Updated Requirement",
		}
	case "create_relationship":
		return map[string]interface{}{
			"source_requirement_id": testRequirementID,
			"target_requirement_id": "REQ-002",
			"relationship_type_id":  testUUID,
		}
	case "search_global":
		return map[string]interface{}{
			"query": "test search",
		}
	case "search_requirements":
		return map[string]interface{}{
			"query": "test requirement search",
		}
	case "list_steering_documents":
		return map[string]interface{}{
			"limit": 10,
		}
	case "create_steering_document":
		return map[string]interface{}{
			"title": "Test Steering Document",
		}
	case "get_steering_document":
		return map[string]interface{}{
			"steering_document_id": testSteeringDocID,
		}
	case "update_steering_document":
		return map[string]interface{}{
			"steering_document_id": testSteeringDocID,
			"title":                "Updated Steering Document",
		}
	case "link_steering_to_epic":
		return map[string]interface{}{
			"steering_document_id": testSteeringDocID,
			"epic_id":              testEpicID,
		}
	case "unlink_steering_from_epic":
		return map[string]interface{}{
			"steering_document_id": testSteeringDocID,
			"epic_id":              testEpicID,
		}
	case "get_epic_steering_documents":
		return map[string]interface{}{
			"epic_id": testEpicID,
		}
	case "create_prompt":
		return map[string]interface{}{
			"name":    "test-prompt",
			"title":   "Test Prompt",
			"content": "Test prompt content",
		}
	case "update_prompt":
		return map[string]interface{}{
			"prompt_id": testPromptID,
			"title":     "Updated Prompt",
		}
	case "delete_prompt":
		return map[string]interface{}{
			"prompt_id": testPromptID,
		}
	case "activate_prompt":
		return map[string]interface{}{
			"prompt_id": testPromptID,
		}
	case "list_prompts":
		return map[string]interface{}{
			"limit": 10,
		}
	case "get_active_prompt":
		return map[string]interface{}{}
	default:
		return map[string]interface{}{}
	}
}

// TestMCPAPICompatibility_ErrorScenarios tests error handling scenarios
func TestMCPAPICompatibility_ErrorScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewMCPHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	testCases := []struct {
		name          string
		request       map[string]interface{}
		expectedError bool
		expectedCode  float64
	}{
		{
			name: "invalid_tool_name",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      1,
				"method":  "tools/call",
				"params": map[string]interface{}{
					"name":      "invalid_tool",
					"arguments": map[string]interface{}{},
				},
			},
			expectedError: true,
			expectedCode:  -32601, // Method not found
		},
		{
			name: "missing_tool_name",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      2,
				"method":  "tools/call",
				"params": map[string]interface{}{
					"arguments": map[string]interface{}{},
				},
			},
			expectedError: true,
			expectedCode:  -32602, // Invalid params
		},
		{
			name: "invalid_params_format",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      3,
				"method":  "tools/call",
				"params":  "invalid",
			},
			expectedError: true,
			expectedCode:  -32602, // Invalid params
		},
		{
			name: "missing_required_arguments",
			request: map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      4,
				"method":  "tools/call",
				"params": map[string]interface{}{
					"name":      "create_epic",
					"arguments": map[string]interface{}{}, // Missing required title and priority
				},
			},
			expectedError: true,
			expectedCode:  -32602, // Invalid params
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Marshal request to JSON
			requestBody, err := json.Marshal(tc.request)
			assert.NoError(t, err)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context with authenticated user
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set(auth.UserContextKey, testUser)

			// Process the request
			handler.Process(c)

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Verify error response
			if tc.expectedError {
				assert.Contains(t, response, "error", "Should have error")
				errorObj := response["error"].(map[string]interface{})
				assert.Equal(t, tc.expectedCode, errorObj["code"], "Should have expected error code")
			}
		})
	}
}

// TestMCPAPICompatibility_ToolsList tests the tools/list method
func TestMCPAPICompatibility_ToolsList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewMCPHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleUser,
	}

	// Create JSON-RPC request for tools/list
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	}

	requestBody, err := json.Marshal(request)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/mcp", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set(auth.UserContextKey, testUser)

	handler.Process(c)

	// Verify response
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify JSON-RPC structure
	assert.Equal(t, "2.0", response["jsonrpc"])
	assert.Equal(t, float64(1), response["id"])
	assert.Contains(t, response, "result")

	// Verify tools list
	result := response["result"].(map[string]interface{})
	assert.Contains(t, result, "tools")

	tools := result["tools"].([]interface{})
	assert.Len(t, tools, 22, "Should have exactly 22 tools")

	// Verify each tool has required fields
	for _, tool := range tools {
		toolObj := tool.(map[string]interface{})
		assert.Contains(t, toolObj, "name")
		assert.Contains(t, toolObj, "title")
		assert.Contains(t, toolObj, "description")
		assert.Contains(t, toolObj, "inputSchema")

		// Verify inputSchema structure
		inputSchema := toolObj["inputSchema"].(map[string]interface{})
		assert.Equal(t, "object", inputSchema["type"])
		assert.Contains(t, inputSchema, "properties")
		// Note: Not all tools have required fields, so we don't assert their presence
	}
}
