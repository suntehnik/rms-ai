package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// Handler represents a JSON-RPC method handler
type Handler func(c *gin.Context, params interface{}) (interface{}, error)

// Processor handles JSON-RPC 2.0 request processing
type Processor struct {
	handlers    map[string]Handler
	errorMapper *ErrorMapper
	logger      Logger
}

// Logger interface for logging JSON-RPC operations
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// DefaultLogger provides a simple logger implementation
type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string, fields ...interface{}) {
	log.Printf("[INFO] %s %v", msg, fields)
}

func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	log.Printf("[ERROR] %s %v", msg, fields)
}

func (l *DefaultLogger) Debug(msg string, fields ...interface{}) {
	log.Printf("[DEBUG] %s %v", msg, fields)
}

// NewProcessor creates a new JSON-RPC processor
func NewProcessor() *Processor {
	return &Processor{
		handlers:    make(map[string]Handler),
		errorMapper: NewErrorMapper(),
		logger:      &DefaultLogger{},
	}
}

// SetLogger sets a custom logger for the processor
func (p *Processor) SetLogger(logger Logger) {
	p.logger = logger
}

// RegisterHandler registers a handler for a specific method
func (p *Processor) RegisterHandler(method string, handler Handler) {
	p.handlers[method] = handler
	p.logger.Debug("Registered JSON-RPC handler", "method", method)
}

// ProcessRequest processes a JSON-RPC request and returns a response
func (p *Processor) ProcessRequest(ctx context.Context, ginCtx *gin.Context, requestData []byte) ([]byte, error) {
	// Log the incoming request body
	p.logger.Debug("JSON-RPC request body", "request", string(requestData))

	// Parse the request
	request, err := ParseRequest(requestData)
	if err != nil {
		// For parse errors, we can't get the ID, so use nil
		response := NewErrorResponse(nil, err.(*JSONRPCError))
		responseData, _ := json.Marshal(response)
		p.logger.Debug("JSON-RPC response body", "response", string(responseData))
		return responseData, nil
	}

	var idValue interface{}
	if request.ID != nil {
		idValue = request.ID.GetValue()
	}
	p.logger.Info("Processing JSON-RPC request", "method", request.Method, "id", idValue)

	// Check if it's a notification (no response expected)
	if request.IsNotification() {
		err := p.processNotification(ctx, ginCtx, request)
		if err != nil {
			p.logger.Error("Error processing notification", "method", request.Method, "error", err)
		}
		return nil, nil // No response for notifications
	}

	// Process the request and create response
	response := p.processRequestWithResponse(ctx, ginCtx, request)

	// Marshal the response
	responseData, err := json.Marshal(response)
	if err != nil {
		p.logger.Error("Error marshaling response", "error", err)
		// Create a fallback error response
		var idValue interface{}
		if request.ID != nil {
			idValue = request.ID.GetValue()
		}
		fallbackResponse := NewErrorResponse(idValue, NewInternalError("Failed to marshal response"))
		responseData, _ = json.Marshal(fallbackResponse)
	}

	// Log the outgoing response body
	p.logger.Debug("JSON-RPC response body", "response", string(responseData))

	return responseData, nil
}

// processRequestWithResponse processes a request that expects a response
func (p *Processor) processRequestWithResponse(ctx context.Context, ginCtx *gin.Context, request *JSONRPCRequest) *JSONRPCResponse {
	// Find the handler
	handler, exists := p.handlers[request.Method]
	if !exists {
		p.logger.Error("Method not found", "method", request.Method)
		var idValue interface{}
		if request.ID != nil {
			idValue = request.ID.GetValue()
		}
		return NewErrorResponse(idValue, NewMethodNotFoundError(request.Method))
	}

	// Execute the handler
	result, err := handler(ginCtx, request.Params)
	if err != nil {
		p.logger.Error("Handler error", "method", request.Method, "error", err)
		jsonrpcErr := p.errorMapper.MapError(err)
		var idValue interface{}
		if request.ID != nil {
			idValue = request.ID.GetValue()
		}
		return NewErrorResponse(idValue, jsonrpcErr)
	}

	var idValue interface{}
	if request.ID != nil {
		idValue = request.ID.GetValue()
	}
	p.logger.Info("Request processed successfully", "method", request.Method, "id", idValue)
	return NewSuccessResponse(idValue, result)
}

// processNotification processes a notification (no response)
func (p *Processor) processNotification(ctx context.Context, ginCtx *gin.Context, request *JSONRPCRequest) error {
	// Find the handler
	handler, exists := p.handlers[request.Method]
	if !exists {
		return fmt.Errorf("method not found: %s", request.Method)
	}

	// Execute the handler
	_, err := handler(ginCtx, request.Params)
	return err
}

// GetRegisteredMethods returns a list of registered method names
func (p *Processor) GetRegisteredMethods() []string {
	methods := make([]string, 0, len(p.handlers))
	for method := range p.handlers {
		methods = append(methods, method)
	}
	return methods
}

// HasMethod checks if a method is registered
func (p *Processor) HasMethod(method string) bool {
	_, exists := p.handlers[method]
	return exists
}

// BatchProcessor handles batch JSON-RPC requests
type BatchProcessor struct {
	processor *Processor
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(processor *Processor) *BatchProcessor {
	return &BatchProcessor{
		processor: processor,
	}
}

// ProcessBatch processes a batch of JSON-RPC requests
func (bp *BatchProcessor) ProcessBatch(ctx context.Context, ginCtx *gin.Context, batchData []byte) ([]byte, error) {
	// Try to parse as array first
	var requests []json.RawMessage
	if err := json.Unmarshal(batchData, &requests); err != nil {
		// Not a batch, process as single request
		return bp.processor.ProcessRequest(ctx, ginCtx, batchData)
	}

	if len(requests) == 0 {
		// Empty batch is invalid
		response := NewErrorResponse(nil, NewInvalidRequestError("Empty batch"))
		return json.Marshal(response)
	}

	responses := make([]*JSONRPCResponse, 0, len(requests))

	// Process each request in the batch
	for _, requestData := range requests {
		responseData, err := bp.processor.ProcessRequest(ctx, ginCtx, requestData)
		if err != nil {
			// This shouldn't happen as ProcessRequest handles errors internally
			continue
		}

		// Skip notifications (no response)
		if responseData == nil {
			continue
		}

		var response JSONRPCResponse
		if err := json.Unmarshal(responseData, &response); err == nil {
			responses = append(responses, &response)
		}
	}

	// If no responses (all notifications), return empty
	if len(responses) == 0 {
		return nil, nil
	}

	return json.Marshal(responses)
}

// ValidationHelper provides utilities for parameter validation
type ValidationHelper struct{}

// NewValidationHelper creates a new validation helper
func NewValidationHelper() *ValidationHelper {
	return &ValidationHelper{}
}

// ValidateParams validates parameters against expected structure
func (vh *ValidationHelper) ValidateParams(params interface{}, validator func(interface{}) error) error {
	if validator == nil {
		return nil
	}

	if err := validator(params); err != nil {
		return NewValidationError(err.Error())
	}

	return nil
}

// ExtractParams extracts and validates parameters from the request
func (vh *ValidationHelper) ExtractParams(params interface{}, target interface{}) error {
	if params == nil {
		return NewInvalidParamsError("Parameters are required")
	}

	// Convert params to JSON and back to target type for type safety
	data, err := json.Marshal(params)
	if err != nil {
		return NewInvalidParamsError("Invalid parameter format")
	}

	if err := json.Unmarshal(data, target); err != nil {
		return NewInvalidParamsError(fmt.Sprintf("Parameter validation failed: %v", err))
	}

	return nil
}
