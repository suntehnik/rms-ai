package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// JSONRPCVersion represents the JSON-RPC version
const JSONRPCVersion = "2.0"

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc" validate:"required,eq=2.0"`
	ID      *JSONRPCID  `json:"id"`
	Method  string      `json:"method" validate:"required"`
	Params  interface{} `json:"params"`
}

// JSONRPCID represents a JSON-RPC ID that can be string, number, or null
type JSONRPCID struct {
	Value interface{}
}

// UnmarshalJSON implements custom unmarshaling for JSONRPCID
func (id *JSONRPCID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		id.Value = nil
		return nil
	}

	// Try to unmarshal as int first
	var intVal int
	if err := json.Unmarshal(data, &intVal); err == nil {
		id.Value = intVal
		return nil
	}

	// Try to unmarshal as string
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		id.Value = strVal
		return nil
	}

	// Try to unmarshal as float64 (fallback)
	var floatVal float64
	if err := json.Unmarshal(data, &floatVal); err == nil {
		id.Value = floatVal
		return nil
	}

	return fmt.Errorf("invalid JSON-RPC ID format")
}

// MarshalJSON implements custom marshaling for JSONRPCID
func (id *JSONRPCID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.Value)
}

// IsNull returns true if the ID is null
func (id *JSONRPCID) IsNull() bool {
	return id == nil || id.Value == nil
}

// GetValue returns the underlying value
func (id *JSONRPCID) GetValue() interface{} {
	if id == nil {
		return nil
	}
	return id.Value
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      *JSONRPCID    `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *JSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// JSONRPCNotification represents a JSON-RPC 2.0 notification (no ID field)
type JSONRPCNotification struct {
	JSONRPC string      `json:"jsonrpc" validate:"required,eq=2.0"`
	Method  string      `json:"method" validate:"required"`
	Params  interface{} `json:"params"`
}

// IsNotification checks if the request is a notification (no ID field)
func (r *JSONRPCRequest) IsNotification() bool {
	return r.ID == nil || r.ID.IsNull()
}

// NewSuccessResponse creates a successful JSON-RPC response
func NewSuccessResponse(id interface{}, result interface{}) *JSONRPCResponse {
	var jsonrpcID *JSONRPCID
	if id != nil {
		jsonrpcID = &JSONRPCID{Value: id}
	}
	return &JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      jsonrpcID,
		Result:  result,
	}
}

// NewErrorResponse creates an error JSON-RPC response
func NewErrorResponse(id interface{}, err *JSONRPCError) *JSONRPCResponse {
	var jsonrpcID *JSONRPCID
	if id != nil {
		jsonrpcID = &JSONRPCID{Value: id}
	}
	return &JSONRPCResponse{
		JSONRPC: JSONRPCVersion,
		ID:      jsonrpcID,
		Error:   err,
	}
}

// ParseRequest parses a JSON-RPC request from raw JSON
func ParseRequest(data []byte) (*JSONRPCRequest, error) {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, NewJSONRPCError(ParseError, "Parse error", err.Error())
	}

	// Validate JSON-RPC version
	if req.JSONRPC != JSONRPCVersion {
		return nil, NewJSONRPCError(InvalidRequest, "Invalid Request", "jsonrpc field must be '2.0'")
	}

	// Validate required fields
	if req.Method == "" {
		return nil, NewJSONRPCError(InvalidRequest, "Invalid Request", "method field is required")
	}

	return &req, nil
}

// ParseNotification parses a JSON-RPC notification from raw JSON
func ParseNotification(data []byte) (*JSONRPCNotification, error) {
	var notif JSONRPCNotification
	if err := json.Unmarshal(data, &notif); err != nil {
		return nil, NewJSONRPCError(ParseError, "Parse error", err.Error())
	}

	// Validate JSON-RPC version
	if notif.JSONRPC != JSONRPCVersion {
		return nil, NewJSONRPCError(InvalidRequest, "Invalid Request", "jsonrpc field must be '2.0'")
	}

	// Validate required fields
	if notif.Method == "" {
		return nil, NewJSONRPCError(InvalidRequest, "Invalid Request", "method field is required")
	}

	return &notif, nil
}
