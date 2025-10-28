package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestListResponse(t *testing.T) {
	// Test data
	testData := []string{"item1", "item2", "item3"}
	totalCount := int64(100)
	limit := 10
	offset := 20

	// Create a test response
	response := ListResponse[string]{
		Data:       testData,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}

	// Verify the structure
	assert.Equal(t, testData, response.Data)
	assert.Equal(t, totalCount, response.TotalCount)
	assert.Equal(t, limit, response.Limit)
	assert.Equal(t, offset, response.Offset)

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	// Verify JSON structure
	var jsonResponse map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonResponse)
	assert.NoError(t, err)

	assert.Equal(t, float64(totalCount), jsonResponse["total_count"])
	assert.Equal(t, float64(limit), jsonResponse["limit"])
	assert.Equal(t, float64(offset), jsonResponse["offset"])
	assert.Len(t, jsonResponse["data"], len(testData))
}

func TestSendListResponse(t *testing.T) {
	// Setup Gin in test mode
	gin.SetMode(gin.TestMode)

	// Create test data
	testData := []map[string]string{
		{"id": "1", "name": "Test Item 1"},
		{"id": "2", "name": "Test Item 2"},
	}

	// Create test router and handler
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		SendListResponse(c, testData, 50, 10, 0)
	})

	// Create test request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Equal(t, float64(50), response["total_count"])
	assert.Equal(t, float64(10), response["limit"])
	assert.Equal(t, float64(0), response["offset"])
	assert.Len(t, response["data"], 2)

	// Verify data content
	data := response["data"].([]interface{})
	firstItem := data[0].(map[string]interface{})
	assert.Equal(t, "1", firstItem["id"])
	assert.Equal(t, "Test Item 1", firstItem["name"])
}

func TestPaginationParams(t *testing.T) {
	t.Run("SetDefaults with zero values", func(t *testing.T) {
		params := PaginationParams{}
		params.SetDefaults()

		assert.Equal(t, 50, params.Limit)
		assert.Equal(t, 0, params.Offset)
	})

	t.Run("SetDefaults with negative offset", func(t *testing.T) {
		params := PaginationParams{
			Limit:  25,
			Offset: -10,
		}
		params.SetDefaults()

		assert.Equal(t, 25, params.Limit)
		assert.Equal(t, 0, params.Offset)
	})

	t.Run("SetDefaults preserves valid values", func(t *testing.T) {
		params := PaginationParams{
			Limit:  25,
			Offset: 100,
		}
		params.SetDefaults()

		assert.Equal(t, 25, params.Limit)
		assert.Equal(t, 100, params.Offset)
	})
}
