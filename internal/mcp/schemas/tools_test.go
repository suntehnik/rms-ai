package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSupportedTools_StatusParameters(t *testing.T) {
	tools := GetSupportedTools()

	// Find update_user_story tool
	var updateUserStoryTool *ToolDefinition
	for _, tool := range tools {
		if tool.Name == "update_user_story" {
			updateUserStoryTool = &tool
			break
		}
	}

	assert.NotNil(t, updateUserStoryTool, "update_user_story tool should exist")

	// Check that status parameter exists in schema
	schema, ok := updateUserStoryTool.InputSchema.(map[string]interface{})
	assert.True(t, ok, "InputSchema should be a map")

	properties, ok := schema["properties"].(map[string]interface{})
	assert.True(t, ok, "properties should exist in schema")

	statusParam, ok := properties["status"].(map[string]interface{})
	assert.True(t, ok, "status parameter should exist")

	// Verify status parameter has correct enum values
	enum, ok := statusParam["enum"].([]string)
	assert.True(t, ok, "status parameter should have enum values")
	expectedStatuses := []string{"Backlog", "Draft", "In Progress", "Done", "Cancelled"}
	assert.Equal(t, expectedStatuses, enum, "status enum should match expected values")

	// Find update_requirement tool
	var updateRequirementTool *ToolDefinition
	for _, tool := range tools {
		if tool.Name == "update_requirement" {
			updateRequirementTool = &tool
			break
		}
	}

	assert.NotNil(t, updateRequirementTool, "update_requirement tool should exist")

	// Check that status parameter exists in requirement schema
	reqSchema, ok := updateRequirementTool.InputSchema.(map[string]interface{})
	assert.True(t, ok, "InputSchema should be a map")

	reqProperties, ok := reqSchema["properties"].(map[string]interface{})
	assert.True(t, ok, "properties should exist in schema")

	reqStatusParam, ok := reqProperties["status"].(map[string]interface{})
	assert.True(t, ok, "status parameter should exist")

	// Verify requirement status parameter has correct enum values
	reqEnum, ok := reqStatusParam["enum"].([]string)
	assert.True(t, ok, "status parameter should have enum values")
	expectedReqStatuses := []string{"Draft", "Active", "Obsolete"}
	assert.Equal(t, expectedReqStatuses, reqEnum, "requirement status enum should match expected values")

	// Verify update_epic tool already has status parameter (should exist from before)
	var updateEpicTool *ToolDefinition
	for _, tool := range tools {
		if tool.Name == "update_epic" {
			updateEpicTool = &tool
			break
		}
	}

	assert.NotNil(t, updateEpicTool, "update_epic tool should exist")

	epicSchema, ok := updateEpicTool.InputSchema.(map[string]interface{})
	assert.True(t, ok, "InputSchema should be a map")

	epicProperties, ok := epicSchema["properties"].(map[string]interface{})
	assert.True(t, ok, "properties should exist in schema")

	epicStatusParam, ok := epicProperties["status"].(map[string]interface{})
	assert.True(t, ok, "status parameter should exist")

	// Verify epic status parameter has correct enum values
	epicEnum, ok := epicStatusParam["enum"].([]string)
	assert.True(t, ok, "status parameter should have enum values")
	expectedEpicStatuses := []string{"Backlog", "Draft", "In Progress", "Done", "Cancelled"}
	assert.Equal(t, expectedEpicStatuses, epicEnum, "epic status enum should match expected values")
}

func TestGetToolByName_StatusTools(t *testing.T) {
	// Test that we can retrieve tools by name
	updateUserStoryTool := GetToolByName("update_user_story")
	assert.NotNil(t, updateUserStoryTool, "should be able to get update_user_story tool by name")

	updateRequirementTool := GetToolByName("update_requirement")
	assert.NotNil(t, updateRequirementTool, "should be able to get update_requirement tool by name")

	updateEpicTool := GetToolByName("update_epic")
	assert.NotNil(t, updateEpicTool, "should be able to get update_epic tool by name")

	// Test non-existent tool
	nonExistentTool := GetToolByName("non_existent_tool")
	assert.Nil(t, nonExistentTool, "should return nil for non-existent tool")
}
