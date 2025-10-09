package jsonrpc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Validator provides JSON-RPC message validation
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequest validates a JSON-RPC request structure
func (v *Validator) ValidateRequest(data []byte) (*JSONRPCRequest, *JSONRPCError) {
	var request JSONRPCRequest

	// Parse JSON
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, NewParseError(fmt.Sprintf("Invalid JSON: %v", err))
	}

	// Validate JSON-RPC version
	if request.JSONRPC != JSONRPCVersion {
		return nil, NewInvalidRequestError(fmt.Sprintf("Invalid jsonrpc version: expected '%s', got '%s'", JSONRPCVersion, request.JSONRPC))
	}

	// Validate method field
	if request.Method == "" {
		return nil, NewInvalidRequestError("Missing required field: method")
	}

	// Validate method name format (should not start with "rpc.")
	if strings.HasPrefix(request.Method, "rpc.") {
		return nil, NewInvalidRequestError("Method names should not start with 'rpc.'")
	}

	return &request, nil
}

// ValidateResponse validates a JSON-RPC response structure
func (v *Validator) ValidateResponse(response *JSONRPCResponse) *JSONRPCError {
	if response == nil {
		return NewInternalError("Response is nil")
	}

	// Validate JSON-RPC version
	if response.JSONRPC != JSONRPCVersion {
		return NewInternalError(fmt.Sprintf("Invalid jsonrpc version in response: expected '%s', got '%s'", JSONRPCVersion, response.JSONRPC))
	}

	// Response must have either result or error, but not both
	hasResult := response.Result != nil
	hasError := response.Error != nil

	if hasResult && hasError {
		return NewInternalError("Response cannot have both result and error")
	}

	if !hasResult && !hasError {
		return NewInternalError("Response must have either result or error")
	}

	// Validate error structure if present
	if hasError {
		if err := v.ValidateError(response.Error); err != nil {
			return err
		}
	}

	return nil
}

// ValidateError validates a JSON-RPC error structure
func (v *Validator) ValidateError(jsonrpcErr *JSONRPCError) *JSONRPCError {
	if jsonrpcErr == nil {
		return NewInternalError("Error object is nil")
	}

	// Validate error code
	if jsonrpcErr.Code == 0 {
		return NewInternalError("Error code cannot be zero")
	}

	// Validate error message
	if jsonrpcErr.Message == "" {
		return NewInternalError("Error message cannot be empty")
	}

	// Validate standard error codes
	if !v.isValidErrorCode(jsonrpcErr.Code) {
		return NewInternalError(fmt.Sprintf("Invalid error code: %d", jsonrpcErr.Code))
	}

	return nil
}

// ValidateNotification validates a JSON-RPC notification structure
func (v *Validator) ValidateNotification(data []byte) (*JSONRPCNotification, *JSONRPCError) {
	var notification JSONRPCNotification

	// Parse JSON
	if err := json.Unmarshal(data, &notification); err != nil {
		return nil, NewParseError(fmt.Sprintf("Invalid JSON: %v", err))
	}

	// Validate JSON-RPC version
	if notification.JSONRPC != JSONRPCVersion {
		return nil, NewInvalidRequestError(fmt.Sprintf("Invalid jsonrpc version: expected '%s', got '%s'", JSONRPCVersion, notification.JSONRPC))
	}

	// Validate method field
	if notification.Method == "" {
		return nil, NewInvalidRequestError("Missing required field: method")
	}

	// Validate method name format
	if strings.HasPrefix(notification.Method, "rpc.") {
		return nil, NewInvalidRequestError("Method names should not start with 'rpc.'")
	}

	return &notification, nil
}

// ValidateMethodName validates a JSON-RPC method name
func (v *Validator) ValidateMethodName(method string) *JSONRPCError {
	if method == "" {
		return NewInvalidRequestError("Method name cannot be empty")
	}

	// Method names starting with "rpc." are reserved
	if strings.HasPrefix(method, "rpc.") {
		return NewInvalidRequestError("Method names starting with 'rpc.' are reserved")
	}

	// Check for valid characters (alphanumeric, underscore, slash, dot)
	for _, char := range method {
		if !isValidMethodChar(char) {
			return NewInvalidRequestError(fmt.Sprintf("Invalid character in method name: %c", char))
		}
	}

	return nil
}

// ValidateParams validates request parameters
func (v *Validator) ValidateParams(params interface{}, expectedType reflect.Type) *JSONRPCError {
	if params == nil {
		return nil // Params are optional
	}

	// If expected type is provided, validate against it
	if expectedType != nil {
		paramType := reflect.TypeOf(params)
		if !paramType.AssignableTo(expectedType) {
			return NewInvalidParamsError(fmt.Sprintf("Invalid parameter type: expected %s, got %s", expectedType, paramType))
		}
	}

	return nil
}

// ValidateID validates a JSON-RPC request/response ID
func (v *Validator) ValidateID(id interface{}) *JSONRPCError {
	if id == nil {
		return nil // ID can be null for notifications
	}

	// ID should be string, number, or null
	switch id.(type) {
	case string, int, int32, int64, float32, float64:
		return nil
	default:
		return NewInvalidRequestError("ID must be a string, number, or null")
	}
}

// isValidErrorCode checks if an error code is valid according to JSON-RPC spec
func (v *Validator) isValidErrorCode(code int) bool {
	// Standard JSON-RPC error codes
	standardCodes := []int{ParseError, InvalidRequest, MethodNotFound, InvalidParams, InternalError}
	for _, stdCode := range standardCodes {
		if code == stdCode {
			return true
		}
	}

	// Server error codes (-32000 to -32099)
	if code >= -32099 && code <= -32000 {
		return true
	}

	// Application defined error codes (any other integer)
	return true
}

// isValidMethodChar checks if a character is valid in a method name
func isValidMethodChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '_' || char == '/' || char == '.'
}

// BatchValidator validates batch requests
type BatchValidator struct {
	validator *Validator
}

// NewBatchValidator creates a new batch validator
func NewBatchValidator() *BatchValidator {
	return &BatchValidator{
		validator: NewValidator(),
	}
}

// ValidateBatch validates a batch of JSON-RPC requests
func (bv *BatchValidator) ValidateBatch(data []byte) ([]JSONRPCRequest, *JSONRPCError) {
	// Try to parse as array
	var rawRequests []json.RawMessage
	if err := json.Unmarshal(data, &rawRequests); err != nil {
		return nil, NewParseError("Invalid batch format")
	}

	// Empty batch is invalid
	if len(rawRequests) == 0 {
		return nil, NewInvalidRequestError("Batch cannot be empty")
	}

	requests := make([]JSONRPCRequest, 0, len(rawRequests))

	// Validate each request in the batch
	for i, rawRequest := range rawRequests {
		request, jsonrpcErr := bv.validator.ValidateRequest(rawRequest)
		if jsonrpcErr != nil {
			return nil, NewInvalidRequestError(fmt.Sprintf("Invalid request at index %d: %s", i, jsonrpcErr.Message))
		}
		requests = append(requests, *request)
	}

	return requests, nil
}

// MessageValidator provides high-level message validation
type MessageValidator struct {
	validator      *Validator
	batchValidator *BatchValidator
}

// NewMessageValidator creates a new message validator
func NewMessageValidator() *MessageValidator {
	return &MessageValidator{
		validator:      NewValidator(),
		batchValidator: NewBatchValidator(),
	}
}

// ValidateMessage validates any JSON-RPC message (single request, batch, or notification)
func (mv *MessageValidator) ValidateMessage(data []byte) (interface{}, *JSONRPCError) {
	// Try to determine message type by parsing structure
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		// Try as array (batch)
		var rawArray []interface{}
		if err := json.Unmarshal(data, &rawArray); err != nil {
			return nil, NewParseError("Invalid JSON format")
		}

		// Validate as batch
		requests, jsonrpcErr := mv.batchValidator.ValidateBatch(data)
		if jsonrpcErr != nil {
			return nil, jsonrpcErr
		}
		return requests, nil
	}

	// Check if it has an ID field to distinguish between request and notification
	if _, hasID := raw["id"]; hasID {
		// Validate as request
		request, jsonrpcErr := mv.validator.ValidateRequest(data)
		if jsonrpcErr != nil {
			return nil, jsonrpcErr
		}
		return request, nil
	} else {
		// Validate as notification
		notification, jsonrpcErr := mv.validator.ValidateNotification(data)
		if jsonrpcErr != nil {
			return nil, jsonrpcErr
		}
		return notification, nil
	}
}
