package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"product-requirements-management/internal/jsonrpc"
	"product-requirements-management/internal/mcp"

	"github.com/sirupsen/logrus"
)

// InitializeHandler handles MCP initialize method requests
type InitializeHandler struct {
	capabilitiesManager  *mcp.CapabilitiesManager
	systemPromptProvider *mcp.SystemPromptProvider
	logger               *logrus.Logger
}

// NewInitializeHandler creates a new initialize handler instance
func NewInitializeHandler(
	toolProvider mcp.ToolProvider,
	promptProvider mcp.PromptProvider,
	promptService mcp.PromptServiceInterface,
	logger *logrus.Logger,
) *InitializeHandler {
	capabilitiesManager := mcp.NewCapabilitiesManager(toolProvider, promptProvider)
	systemPromptProvider := mcp.NewSystemPromptProvider(promptService)

	return &InitializeHandler{
		capabilitiesManager:  capabilitiesManager,
		systemPromptProvider: systemPromptProvider,
		logger:               logger,
	}
}

// InitializeRequest represents the MCP initialize request structure
type InitializeRequest struct {
	JSONRPC string           `json:"jsonrpc" validate:"required,eq=2.0"`
	ID      interface{}      `json:"id"`
	Method  string           `json:"method" validate:"required,eq=initialize"`
	Params  InitializeParams `json:"params"`
}

// InitializeParams represents the parameters for the initialize request
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion" validate:"required"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

// ClientCapabilities represents the capabilities declared by the client
type ClientCapabilities struct {
	Elicitation *ElicitationCapability `json:"elicitation,omitempty"`
	Sampling    *SamplingCapability    `json:"sampling,omitempty"`
	Roots       *RootsCapability       `json:"roots,omitempty"`
}

// ElicitationCapability represents client's elicitation capability
type ElicitationCapability struct{}

// SamplingCapability represents client's sampling capability
type SamplingCapability struct{}

// RootsCapability represents client's roots capability
type RootsCapability struct{}

// ClientInfo represents information about the client
type ClientInfo struct {
	Name    string `json:"name" validate:"required"`
	Version string `json:"version" validate:"required"`
}

// InitializeResponse represents the MCP initialize response structure
type InitializeResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      interface{}      `json:"id"`
	Result  InitializeResult `json:"result"`
}

// InitializeResult represents the result of the initialize request
type InitializeResult struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    mcp.ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo             `json:"serverInfo"`
	Instructions    string                 `json:"instructions"`
}

// ServerInfo represents information about the server
type ServerInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
}

// Server information constants as per specification
const (
	ServerName      = "spexus mcp"
	ServerTitle     = "MCP server for requirements management system"
	ServerVersion   = "1.0.0"
	ProtocolVersion = "2025-03-26"
)

var SupportedProtocolVersions = []string{ProtocolVersion, "2025-06-18"}

// Initialize error codes
const (
	InvalidProtocolVersion = "INVALID_PROTOCOL_VERSION"
	MalformedRequest       = "MALFORMED_REQUEST"
	InternalError          = "INTERNAL_ERROR"
)

// InitializeError represents an initialize-specific error
type InitializeError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error implements the error interface
func (e *InitializeError) Error() string {
	return fmt.Sprintf("Initialize error %s: %s", e.Code, e.Message)
}

// NewInitializeError creates a new initialize error
func NewInitializeError(code, message string, data interface{}) *InitializeError {
	return &InitializeError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// HandleInitialize processes the MCP initialize request and returns a complete JSON-RPC 2.0 response
func (h *InitializeHandler) HandleInitialize(ctx context.Context, request *InitializeRequest) (*jsonrpc.JSONRPCResponse, error) {
	// Validate and parse parameters
	initParams, err := h.validateInitializeParams(request.Params)
	if err != nil {
		h.logger.WithError(err).Error("Initialize request validation failed")
		return CreateInitializeErrorResponse(request.ID, err), nil
	}

	// Validate protocol version
	if err := h.validateProtocolVersion(initParams.ProtocolVersion); err != nil {
		h.logger.WithFields(logrus.Fields{
			"received_version":  initParams.ProtocolVersion,
			"supported_version": ProtocolVersion,
		}).Error("Protocol version validation failed")
		return CreateInitializeErrorResponse(request.ID, err), nil
	}

	// Generate server capabilities
	capabilities, err := h.capabilitiesManager.GenerateCapabilities(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate server capabilities")
		return CreateInitializeErrorResponse(request.ID, NewInitializeError(InternalError, "Failed to generate capabilities", nil)), nil
	}

	// Get system instructions
	instructions, err := h.systemPromptProvider.GetInstructions(ctx)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get system instructions, using empty string")
		instructions = ""
	}

	// Create initialize result
	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities:    *capabilities,
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Title:   ServerTitle,
			Version: ServerVersion,
		},
		Instructions: instructions,
	}

	h.logger.WithFields(logrus.Fields{
		"protocol_version": result.ProtocolVersion,
		"client_name":      initParams.ClientInfo.Name,
		"client_version":   initParams.ClientInfo.Version,
	}).Info("Initialize request processed successfully")

	// Create JSON-RPC 2.0 success response with matching ID
	return jsonrpc.NewSuccessResponse(request.ID, result), nil
}

// HandleInitializeFromParams processes the MCP initialize request from raw parameters (legacy interface)
func (h *InitializeHandler) HandleInitializeFromParams(ctx context.Context, params interface{}) (interface{}, error) {
	// Validate and parse parameters
	initParams, err := h.validateInitializeParamsFromInterface(params)
	if err != nil {
		h.logger.WithError(err).Error("Initialize request validation failed")
		return nil, err
	}

	// Validate protocol version
	if err := h.validateProtocolVersion(initParams.ProtocolVersion); err != nil {
		h.logger.WithFields(logrus.Fields{
			"received_version":  initParams.ProtocolVersion,
			"supported_version": ProtocolVersion,
		}).Error("Protocol version validation failed")
		return nil, err
	}

	// Generate server capabilities
	capabilities, err := h.capabilitiesManager.GenerateCapabilities(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate server capabilities")
		return nil, NewInitializeError(InternalError, "Failed to generate capabilities", nil)
	}

	// Get system instructions
	instructions, err := h.systemPromptProvider.GetInstructions(ctx)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get system instructions, using empty string")
		instructions = ""
	}

	// Create initialize result
	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities:    *capabilities,
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Title:   ServerTitle,
			Version: ServerVersion,
		},
		Instructions: instructions,
	}

	h.logger.WithFields(logrus.Fields{
		"protocol_version": result.ProtocolVersion,
		"client_name":      initParams.ClientInfo.Name,
		"client_version":   initParams.ClientInfo.Version,
	}).Info("Initialize request processed successfully")

	return result, nil
}

// validateInitializeParams validates and parses initialize parameters
func (h *InitializeHandler) validateInitializeParams(params InitializeParams) (*InitializeParams, error) {
	// Validate protocol version
	if params.ProtocolVersion == "" {
		return nil, NewInitializeError(MalformedRequest, "protocolVersion is required", nil)
	}

	// Validate client info
	if params.ClientInfo.Name == "" {
		return nil, NewInitializeError(MalformedRequest, "clientInfo.name is required", nil)
	}

	if params.ClientInfo.Version == "" {
		return nil, NewInitializeError(MalformedRequest, "clientInfo.version is required", nil)
	}

	return &params, nil
}

// validateInitializeParamsFromInterface validates and parses initialize parameters from interface{} (legacy)
func (h *InitializeHandler) validateInitializeParamsFromInterface(params interface{}) (*InitializeParams, error) {
	if params == nil {
		return nil, NewInitializeError(MalformedRequest, "Initialize parameters are required", nil)
	}

	// Convert params to map for easier handling
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return nil, NewInitializeError(MalformedRequest, "Invalid parameter format", nil)
	}

	// Extract protocol version
	protocolVersion, ok := paramsMap["protocolVersion"].(string)
	if !ok || protocolVersion == "" {
		return nil, NewInitializeError(MalformedRequest, "protocolVersion is required", nil)
	}

	// Extract client info
	clientInfoRaw, ok := paramsMap["clientInfo"]
	if !ok {
		return nil, NewInitializeError(MalformedRequest, "clientInfo is required", nil)
	}

	clientInfoMap, ok := clientInfoRaw.(map[string]interface{})
	if !ok {
		return nil, NewInitializeError(MalformedRequest, "Invalid clientInfo format", nil)
	}

	clientName, ok := clientInfoMap["name"].(string)
	if !ok || clientName == "" {
		return nil, NewInitializeError(MalformedRequest, "clientInfo.name is required", nil)
	}

	clientVersion, ok := clientInfoMap["version"].(string)
	if !ok || clientVersion == "" {
		return nil, NewInitializeError(MalformedRequest, "clientInfo.version is required", nil)
	}

	// Parse client capabilities (optional)
	var capabilities ClientCapabilities
	if capabilitiesRaw, exists := paramsMap["capabilities"]; exists {
		if capabilitiesMap, ok := capabilitiesRaw.(map[string]interface{}); ok {
			// Parse elicitation capability
			if _, hasElicitation := capabilitiesMap["elicitation"]; hasElicitation {
				capabilities.Elicitation = &ElicitationCapability{}
			}

			// Parse sampling capability
			if _, hasSampling := capabilitiesMap["sampling"]; hasSampling {
				capabilities.Sampling = &SamplingCapability{}
			}

			// Parse roots capability
			if _, hasRoots := capabilitiesMap["roots"]; hasRoots {
				capabilities.Roots = &RootsCapability{}
			}
		}
	}

	return &InitializeParams{
		ProtocolVersion: protocolVersion,
		Capabilities:    capabilities,
		ClientInfo: ClientInfo{
			Name:    clientName,
			Version: clientVersion,
		},
	}, nil
}

// validateProtocolVersion validates the protocol version with comprehensive checks
func (h *InitializeHandler) validateProtocolVersion(version string) error {
	// Check for empty version
	if version == "" {
		return NewInitializeError(
			InvalidProtocolVersion,
			"Protocol version cannot be empty",
			map[string]interface{}{
				"supported_versions": SupportedProtocolVersions,
				"received_version":   version,
			},
		)
	}

	// Check if version follows expected format (YYYY-MM-DD)
	if !isValidProtocolVersionFormat(version) {
		return NewInitializeError(
			InvalidProtocolVersion,
			fmt.Sprintf("Invalid protocol version format. Expected format: YYYY-MM-DD, got: %s", version),
			map[string]interface{}{
				"supported_versions": SupportedProtocolVersions,
				"received_version":   version,
				"expected_format":    "YYYY-MM-DD",
			},
		)
	}

	// Check if version is in supported versions
	if !slices.Contains(SupportedProtocolVersions, version) {
		return NewInitializeError(
			InvalidProtocolVersion,
			fmt.Sprintf("Unsupported protocol version. Expected one of: %v", SupportedProtocolVersions),
			map[string]interface{}{
				"supported_versions": SupportedProtocolVersions,
				"received_version":   version,
			},
		)
	}
	return nil
}

// isValidProtocolVersionFormat checks if the protocol version follows YYYY-MM-DD format
func isValidProtocolVersionFormat(version string) bool {
	// Basic length check
	if len(version) != 10 {
		return false
	}

	// Check format: YYYY-MM-DD
	if version[4] != '-' || version[7] != '-' {
		return false
	}

	// Check if year, month, day are numeric
	year := version[0:4]
	month := version[5:7]
	day := version[8:10]

	for _, char := range year + month + day {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// CreateInitializeErrorResponse creates a JSON-RPC error response for initialize errors
func CreateInitializeErrorResponse(id interface{}, err error) *jsonrpc.JSONRPCResponse {
	var jsonrpcError *jsonrpc.JSONRPCError

	if initErr, ok := err.(*InitializeError); ok {
		// Map initialize error to JSON-RPC error with proper codes
		var code int
		switch initErr.Code {
		case InvalidProtocolVersion:
			// Protocol version errors are considered invalid requests
			code = jsonrpc.InvalidRequest
		case MalformedRequest:
			// Malformed requests map to invalid request
			code = jsonrpc.InvalidRequest
		case InternalError:
			// Internal errors map to internal error
			code = jsonrpc.InternalError
		default:
			// Unknown initialize error codes default to internal error
			code = jsonrpc.InternalError
		}

		jsonrpcError = &jsonrpc.JSONRPCError{
			Code:    code,
			Message: initErr.Message,
			Data:    initErr.Data,
		}
	} else if jsonrpcErr, ok := err.(*jsonrpc.JSONRPCError); ok {
		// Already a JSON-RPC error, use as-is
		jsonrpcError = jsonrpcErr
	} else {
		// Generic error - map to internal error
		jsonrpcError = &jsonrpc.JSONRPCError{
			Code:    jsonrpc.InternalError,
			Message: "Internal server error",
			Data:    err.Error(),
		}
	}

	return jsonrpc.NewErrorResponse(id, jsonrpcError)
}

// ValidateJSONRPCRequest performs basic JSON-RPC 2.0 validation
func ValidateJSONRPCRequest(data []byte) error {
	if len(data) == 0 {
		return NewInitializeError(MalformedRequest, "Empty request body", nil)
	}

	// Try to parse as basic JSON first
	var basicJSON map[string]interface{}
	if err := json.Unmarshal(data, &basicJSON); err != nil {
		return NewInitializeError(MalformedRequest, "Invalid JSON format", err.Error())
	}

	// Check required JSON-RPC fields
	jsonrpcField, exists := basicJSON["jsonrpc"]
	if !exists {
		return NewInitializeError(MalformedRequest, "Missing required field: jsonrpc", nil)
	}

	if jsonrpcStr, ok := jsonrpcField.(string); !ok || jsonrpcStr != "2.0" {
		return NewInitializeError(MalformedRequest, "Invalid jsonrpc field. Must be '2.0'", nil)
	}

	methodField, exists := basicJSON["method"]
	if !exists {
		return NewInitializeError(MalformedRequest, "Missing required field: method", nil)
	}

	if _, ok := methodField.(string); !ok {
		return NewInitializeError(MalformedRequest, "Method field must be a string", nil)
	}

	return nil
}

// ProcessInitializeRequest processes a complete JSON-RPC initialize request with comprehensive validation
func (h *InitializeHandler) ProcessInitializeRequest(ctx context.Context, data []byte) (*jsonrpc.JSONRPCResponse, error) {
	// Perform basic JSON-RPC validation first
	if err := ValidateJSONRPCRequest(data); err != nil {
		h.logger.WithError(err).Error("Basic JSON-RPC validation failed")
		// For basic validation errors, we don't have a valid ID, so use null
		return CreateInitializeErrorResponse(nil, err), nil
	}

	// Validate and parse the complete initialize request
	request, err := ValidateInitializeRequest(data)
	if err != nil {
		h.logger.WithError(err).Error("Initialize request parsing failed")
		// For parsing errors, we don't have a valid ID, so use null
		return CreateInitializeErrorResponse(nil, err), nil
	}

	// Process the initialize request
	response, processErr := h.HandleInitialize(ctx, request)
	if processErr != nil {
		h.logger.WithError(processErr).Error("Initialize request processing failed")
		return CreateInitializeErrorResponse(request.ID, processErr), nil
	}

	return response, nil
}

// ValidateInitializeRequest validates the complete initialize request structure
func ValidateInitializeRequest(data []byte) (*InitializeRequest, error) {
	// Validate JSON format first
	if len(data) == 0 {
		return nil, NewInitializeError(MalformedRequest, "Empty request body", nil)
	}

	// Parse as generic JSON-RPC request first
	genericReq, err := jsonrpc.ParseRequest(data)
	if err != nil {
		// Map JSON-RPC parsing errors to initialize errors
		if jsonrpcErr, ok := err.(*jsonrpc.JSONRPCError); ok {
			switch jsonrpcErr.Code {
			case jsonrpc.ParseError:
				return nil, NewInitializeError(MalformedRequest, "Invalid JSON format", jsonrpcErr.Data)
			case jsonrpc.InvalidRequest:
				return nil, NewInitializeError(MalformedRequest, jsonrpcErr.Message, jsonrpcErr.Data)
			default:
				return nil, NewInitializeError(MalformedRequest, "Request parsing failed", jsonrpcErr.Data)
			}
		}
		return nil, NewInitializeError(MalformedRequest, "Request parsing failed", err.Error())
	}

	// Validate JSON-RPC version
	if genericReq.JSONRPC != jsonrpc.JSONRPCVersion {
		return nil, NewInitializeError(MalformedRequest,
			fmt.Sprintf("Invalid JSON-RPC version. Expected '%s', got '%s'", jsonrpc.JSONRPCVersion, genericReq.JSONRPC),
			map[string]interface{}{
				"expected": jsonrpc.JSONRPCVersion,
				"received": genericReq.JSONRPC,
			})
	}

	// Validate method
	if genericReq.Method != "initialize" {
		return nil, NewInitializeError(MalformedRequest,
			fmt.Sprintf("Invalid method. Expected 'initialize', got '%s'", genericReq.Method),
			map[string]interface{}{
				"expected": "initialize",
				"received": genericReq.Method,
			})
	}

	// Create initialize request
	initReq := &InitializeRequest{
		JSONRPC: genericReq.JSONRPC,
		ID:      genericReq.ID.GetValue(),
		Method:  genericReq.Method,
	}

	// Validate and parse parameters
	if genericReq.Params == nil {
		return nil, NewInitializeError(MalformedRequest, "Initialize parameters are required", nil)
	}

	paramsMap, ok := genericReq.Params.(map[string]interface{})
	if !ok {
		return nil, NewInitializeError(MalformedRequest, "Parameters must be an object", nil)
	}

	// Extract and validate protocol version
	protocolVersionRaw, exists := paramsMap["protocolVersion"]
	if !exists {
		return nil, NewInitializeError(MalformedRequest, "protocolVersion is required", nil)
	}

	protocolVersion, ok := protocolVersionRaw.(string)
	if !ok {
		return nil, NewInitializeError(MalformedRequest, "protocolVersion must be a string", nil)
	}

	if protocolVersion == "" {
		return nil, NewInitializeError(MalformedRequest, "protocolVersion cannot be empty", nil)
	}

	initReq.Params.ProtocolVersion = protocolVersion

	// Extract and validate client info
	clientInfoRaw, exists := paramsMap["clientInfo"]
	if !exists {
		return nil, NewInitializeError(MalformedRequest, "clientInfo is required", nil)
	}

	clientInfoMap, ok := clientInfoRaw.(map[string]interface{})
	if !ok {
		return nil, NewInitializeError(MalformedRequest, "clientInfo must be an object", nil)
	}

	// Validate client name
	clientNameRaw, exists := clientInfoMap["name"]
	if !exists {
		return nil, NewInitializeError(MalformedRequest, "clientInfo.name is required", nil)
	}

	clientName, ok := clientNameRaw.(string)
	if !ok {
		return nil, NewInitializeError(MalformedRequest, "clientInfo.name must be a string", nil)
	}

	if clientName == "" {
		return nil, NewInitializeError(MalformedRequest, "clientInfo.name cannot be empty", nil)
	}

	// Validate client version
	clientVersionRaw, exists := clientInfoMap["version"]
	if !exists {
		return nil, NewInitializeError(MalformedRequest, "clientInfo.version is required", nil)
	}

	clientVersion, ok := clientVersionRaw.(string)
	if !ok {
		return nil, NewInitializeError(MalformedRequest, "clientInfo.version must be a string", nil)
	}

	if clientVersion == "" {
		return nil, NewInitializeError(MalformedRequest, "clientInfo.version cannot be empty", nil)
	}

	initReq.Params.ClientInfo = ClientInfo{
		Name:    clientName,
		Version: clientVersion,
	}

	// Extract capabilities (optional)
	if capabilitiesRaw, exists := paramsMap["capabilities"]; exists {
		if capabilitiesMap, ok := capabilitiesRaw.(map[string]interface{}); ok {
			// Parse elicitation capability
			if _, hasElicitation := capabilitiesMap["elicitation"]; hasElicitation {
				initReq.Params.Capabilities.Elicitation = &ElicitationCapability{}
			}

			// Parse sampling capability
			if _, hasSampling := capabilitiesMap["sampling"]; hasSampling {
				initReq.Params.Capabilities.Sampling = &SamplingCapability{}
			}

			// Parse roots capability
			if _, hasRoots := capabilitiesMap["roots"]; hasRoots {
				initReq.Params.Capabilities.Roots = &RootsCapability{}
			}
		}
		// If capabilities is not an object, we ignore it (it's optional)
	}

	return initReq, nil
}
