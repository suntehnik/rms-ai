package tools

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateRequiredArgs(t *testing.T) {
	args := map[string]interface{}{
		"title":    "Test Title",
		"priority": 1,
	}

	// Test with all required args present
	err := validateRequiredArgs(args, []string{"title", "priority"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test with missing required arg
	err = validateRequiredArgs(args, []string{"title", "priority", "missing"})
	if err == nil {
		t.Error("Expected error for missing argument, got nil")
	}
}

func TestGetStringArg(t *testing.T) {
	args := map[string]interface{}{
		"title":   "Test Title",
		"number":  123,
		"boolean": true,
	}

	// Test valid string
	val, ok := getStringArg(args, "title")
	if !ok || val != "Test Title" {
		t.Errorf("Expected 'Test Title', got %s (ok: %v)", val, ok)
	}

	// Test non-string value
	_, ok = getStringArg(args, "number")
	if ok {
		t.Error("Expected false for non-string value")
	}

	// Test missing key
	_, ok = getStringArg(args, "missing")
	if ok {
		t.Error("Expected false for missing key")
	}
}

func TestGetIntArg(t *testing.T) {
	args := map[string]interface{}{
		"int_val":    123,
		"float_val":  456.0,
		"string_val": "789",
	}

	// Test valid int
	val, ok := getIntArg(args, "int_val")
	if !ok || val != 123 {
		t.Errorf("Expected 123, got %d (ok: %v)", val, ok)
	}

	// Test valid float (should convert to int)
	val, ok = getIntArg(args, "float_val")
	if !ok || val != 456 {
		t.Errorf("Expected 456, got %d (ok: %v)", val, ok)
	}

	// Test string value (should fail)
	_, ok = getIntArg(args, "string_val")
	if ok {
		t.Error("Expected false for string value")
	}
}

func TestGetUUIDArg(t *testing.T) {
	testUUID := uuid.New()
	args := map[string]interface{}{
		"valid_uuid":   testUUID.String(),
		"invalid_uuid": "not-a-uuid",
		"number":       123,
	}

	// Test valid UUID
	val, ok := getUUIDArg(args, "valid_uuid")
	if !ok || val != testUUID {
		t.Errorf("Expected %s, got %s (ok: %v)", testUUID, val, ok)
	}

	// Test invalid UUID
	_, ok = getUUIDArg(args, "invalid_uuid")
	if ok {
		t.Error("Expected false for invalid UUID")
	}

	// Test non-string value
	_, ok = getUUIDArg(args, "number")
	if ok {
		t.Error("Expected false for non-string value")
	}
}
