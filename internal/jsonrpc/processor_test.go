package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestGinContext(baseCtx context.Context) (context.Context, *gin.Context) {
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/", nil)
	ctxWithGin := context.WithValue(baseCtx, "gin_context", c)
	req = req.WithContext(ctxWithGin)
	c.Request = req

	return ctxWithGin, c
}

func TestProcessor_RegisterHandler(t *testing.T) {
	processor := NewProcessor()

	handler := func(c *gin.Context, params interface{}) (interface{}, error) {
		return "test result", nil
	}

	processor.RegisterHandler("test_method", handler)

	assert.True(t, processor.HasMethod("test_method"))
	assert.False(t, processor.HasMethod("nonexistent_method"))

	methods := processor.GetRegisteredMethods()
	assert.Contains(t, methods, "test_method")
}

func TestProcessor_ProcessRequest_Success(t *testing.T) {
	processor := NewProcessor()

	// Register a test handler
	processor.RegisterHandler("test_method", func(c *gin.Context, params interface{}) (interface{}, error) {
		return map[string]interface{}{"status": "ok", "params": params}, nil
	})

	// Test request
	requestData := `{"jsonrpc":"2.0","id":1,"method":"test_method","params":{"key":"value"}}`

	ctx, ginCtx := createTestGinContext(context.Background())
	responseData, err := processor.ProcessRequest(ctx, ginCtx, []byte(requestData))
	require.NoError(t, err)
	require.NotNil(t, responseData)

	var response JSONRPCResponse
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID.GetValue())
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)
}

func TestProcessor_ProcessRequest_MethodNotFound(t *testing.T) {
	processor := NewProcessor()

	requestData := `{"jsonrpc":"2.0","id":1,"method":"nonexistent_method"}`

	ctx, ginCtx := createTestGinContext(context.Background())
	responseData, err := processor.ProcessRequest(ctx, ginCtx, []byte(requestData))
	require.NoError(t, err)
	require.NotNil(t, responseData)

	var response JSONRPCResponse
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID.GetValue())
	assert.Nil(t, response.Result)
	require.NotNil(t, response.Error)
	assert.Equal(t, MethodNotFound, response.Error.Code)
}

func TestProcessor_ProcessRequest_HandlerError(t *testing.T) {
	processor := NewProcessor()

	// Register a handler that returns an error
	processor.RegisterHandler("error_method", func(c *gin.Context, params interface{}) (interface{}, error) {
		return nil, errors.New("handler error")
	})

	requestData := `{"jsonrpc":"2.0","id":1,"method":"error_method"}`

	ctx, ginCtx := createTestGinContext(context.Background())
	responseData, err := processor.ProcessRequest(ctx, ginCtx, []byte(requestData))
	require.NoError(t, err)
	require.NotNil(t, responseData)

	var response JSONRPCResponse
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID.GetValue())
	assert.Nil(t, response.Result)
	require.NotNil(t, response.Error)
	assert.Equal(t, InternalError, response.Error.Code)
}

func TestProcessor_ProcessRequest_ParseError(t *testing.T) {
	processor := NewProcessor()

	// Invalid JSON
	requestData := `{"jsonrpc":"2.0","id":1,"method":"test"`

	ctx, ginCtx := createTestGinContext(context.Background())
	responseData, err := processor.ProcessRequest(ctx, ginCtx, []byte(requestData))
	require.NoError(t, err)
	require.NotNil(t, responseData)

	var response JSONRPCResponse
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Nil(t, response.ID) // Parse error means we can't get the ID
	assert.Nil(t, response.Result)
	require.NotNil(t, response.Error)
	assert.Equal(t, ParseError, response.Error.Code)
}

func TestProcessor_ProcessRequest_Notification(t *testing.T) {
	processor := NewProcessor()

	// Register a test handler
	processor.RegisterHandler("notification_method", func(c *gin.Context, params interface{}) (interface{}, error) {
		return "notification processed", nil
	})

	// Notification (no ID field)
	requestData := `{"jsonrpc":"2.0","method":"notification_method","params":{"key":"value"}}`

	ctx, ginCtx := createTestGinContext(context.Background())
	responseData, err := processor.ProcessRequest(ctx, ginCtx, []byte(requestData))
	require.NoError(t, err)
	assert.Nil(t, responseData) // No response for notifications
}

func TestBatchProcessor_ProcessBatch(t *testing.T) {
	processor := NewProcessor()
	batchProcessor := NewBatchProcessor(processor)

	// Register test handlers
	processor.RegisterHandler("test_method", func(c *gin.Context, params interface{}) (interface{}, error) {
		return map[string]interface{}{"status": "ok"}, nil
	})

	processor.RegisterHandler("notification_method", func(c *gin.Context, params interface{}) (interface{}, error) {
		return "notification processed", nil
	})

	// Batch request with regular request and notification
	batchData := `[
		{"jsonrpc":"2.0","id":1,"method":"test_method","params":{"key":"value"}},
		{"jsonrpc":"2.0","method":"notification_method","params":{"key":"value"}},
		{"jsonrpc":"2.0","id":2,"method":"test_method","params":{"key":"value2"}}
	]`

	ctx, ginCtx := createTestGinContext(context.Background())
	responseData, err := batchProcessor.ProcessBatch(ctx, ginCtx, []byte(batchData))
	require.NoError(t, err)
	require.NotNil(t, responseData)

	var responses []JSONRPCResponse
	err = json.Unmarshal(responseData, &responses)
	require.NoError(t, err)

	// Should have 2 responses (notification doesn't get a response)
	assert.Len(t, responses, 2)

	// Check first response
	assert.Equal(t, "2.0", responses[0].JSONRPC)
	assert.Equal(t, 1, responses[0].ID.GetValue())
	assert.Nil(t, responses[0].Error)

	// Check second response
	assert.Equal(t, "2.0", responses[1].JSONRPC)
	assert.Equal(t, 2, responses[1].ID.GetValue())
	assert.Nil(t, responses[1].Error)
}

func TestBatchProcessor_ProcessBatch_EmptyBatch(t *testing.T) {
	processor := NewProcessor()
	batchProcessor := NewBatchProcessor(processor)

	// Empty batch
	batchData := `[]`

	ctx, ginCtx := createTestGinContext(context.Background())
	responseData, err := batchProcessor.ProcessBatch(ctx, ginCtx, []byte(batchData))
	require.NoError(t, err)
	require.NotNil(t, responseData)

	var response JSONRPCResponse
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Nil(t, response.ID)
	require.NotNil(t, response.Error)
	assert.Equal(t, InvalidRequest, response.Error.Code)
}

func TestBatchProcessor_ProcessBatch_SingleRequest(t *testing.T) {
	processor := NewProcessor()
	batchProcessor := NewBatchProcessor(processor)

	// Register test handler
	processor.RegisterHandler("test_method", func(c *gin.Context, params interface{}) (interface{}, error) {
		return map[string]interface{}{"status": "ok"}, nil
	})

	// Single request (not an array)
	requestData := `{"jsonrpc":"2.0","id":1,"method":"test_method","params":{"key":"value"}}`

	ctx, ginCtx := createTestGinContext(context.Background())
	responseData, err := batchProcessor.ProcessBatch(ctx, ginCtx, []byte(requestData))
	require.NoError(t, err)
	require.NotNil(t, responseData)

	var response JSONRPCResponse
	err = json.Unmarshal(responseData, &response)
	require.NoError(t, err)

	assert.Equal(t, "2.0", response.JSONRPC)
	assert.Equal(t, 1, response.ID.GetValue())
	assert.Nil(t, response.Error)
}

func TestValidationHelper_ExtractParams(t *testing.T) {
	vh := NewValidationHelper()

	t.Run("valid params", func(t *testing.T) {
		params := map[string]interface{}{
			"name": "test",
			"age":  25,
		}

		var target struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		err := vh.ExtractParams(params, &target)
		require.NoError(t, err)
		assert.Equal(t, "test", target.Name)
		assert.Equal(t, 25, target.Age)
	})

	t.Run("nil params", func(t *testing.T) {
		var target struct {
			Name string `json:"name"`
		}

		err := vh.ExtractParams(nil, &target)
		require.Error(t, err)
		jsonrpcErr, ok := err.(*JSONRPCError)
		require.True(t, ok)
		assert.Equal(t, InvalidParams, jsonrpcErr.Code)
	})

	t.Run("invalid params format", func(t *testing.T) {
		// Create a circular reference that can't be marshaled
		params := make(map[string]interface{})
		params["self"] = params

		var target struct {
			Name string `json:"name"`
		}

		err := vh.ExtractParams(params, &target)
		require.Error(t, err)
		jsonrpcErr, ok := err.(*JSONRPCError)
		require.True(t, ok)
		assert.Equal(t, InvalidParams, jsonrpcErr.Code)
	})
}

func TestValidationHelper_ValidateParams(t *testing.T) {
	vh := NewValidationHelper()

	t.Run("nil validator", func(t *testing.T) {
		err := vh.ValidateParams("test", nil)
		assert.NoError(t, err)
	})

	t.Run("successful validation", func(t *testing.T) {
		validator := func(params interface{}) error {
			return nil
		}

		err := vh.ValidateParams("test", validator)
		assert.NoError(t, err)
	})

	t.Run("validation failure", func(t *testing.T) {
		validator := func(params interface{}) error {
			return errors.New("validation failed")
		}

		err := vh.ValidateParams("test", validator)
		require.Error(t, err)
		jsonrpcErr, ok := err.(*JSONRPCError)
		require.True(t, ok)
		assert.Equal(t, ValidationError, jsonrpcErr.Code)
	})
}

// Mock logger for testing
type mockLogger struct {
	infoCalls  []string
	errorCalls []string
	debugCalls []string
}

func (m *mockLogger) Info(msg string, fields ...interface{}) {
	m.infoCalls = append(m.infoCalls, msg)
}

func (m *mockLogger) Error(msg string, fields ...interface{}) {
	m.errorCalls = append(m.errorCalls, msg)
}

func (m *mockLogger) Debug(msg string, fields ...interface{}) {
	m.debugCalls = append(m.debugCalls, msg)
}

func TestProcessor_SetLogger(t *testing.T) {
	processor := NewProcessor()
	mockLog := &mockLogger{}

	processor.SetLogger(mockLog)

	// Register a handler to trigger logging
	processor.RegisterHandler("test_method", func(ctx context.Context, params interface{}) (interface{}, error) {
		return "ok", nil
	})

	// Should have logged the handler registration
	assert.Len(t, mockLog.debugCalls, 1)
	assert.Contains(t, mockLog.debugCalls[0], "Registered JSON-RPC handler")
}
