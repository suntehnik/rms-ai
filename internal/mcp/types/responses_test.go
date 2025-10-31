package types

import (
	"testing"
)

func TestCreateToolResponse(t *testing.T) {
	// Test with message only
	response := CreateToolResponse("Test message", nil)
	if len(response.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(response.Content))
	}
	if response.Content[0].Type != "text" {
		t.Errorf("Expected type 'text', got %s", response.Content[0].Type)
	}
	if response.Content[0].Text != "Test message" {
		t.Errorf("Expected 'Test message', got %s", response.Content[0].Text)
	}

	// Test with message and data
	data := map[string]string{"key": "value"}
	response = CreateToolResponse("Test message", data)
	if len(response.Content) != 2 {
		t.Errorf("Expected 2 content items, got %d", len(response.Content))
	}
}

func TestCreateSuccessResponse(t *testing.T) {
	response := CreateSuccessResponse("Success message")
	if len(response.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(response.Content))
	}
	if response.Content[0].Text != "Success message" {
		t.Errorf("Expected 'Success message', got %s", response.Content[0].Text)
	}
}
