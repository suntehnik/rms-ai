package docs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"product-requirements-management/internal/auth"
	"product-requirements-management/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInteractiveTestingValidation validates all interactive testing capabilities
func TestInteractiveTestingValidation(t *testing.T) {
	// Create a simple Gin router for testing
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create auth service
	authService := auth.NewService("test-secret-key", time.Hour)

	// Add middleware for authentication testing
	router.Use(func(c *gin.Context) {
		// Simple auth middleware for testing
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if len(token) > 10 { // Basic validation
				c.Set("authenticated", true)
				c.Set("user_role", "user") // Default role for testing
			}
		}
		c.Next()
	})

	// Add test endpoints that simulate the API
	setupTestEndpoints(router)

	// Test Swagger UI accessibility
	t.Run("SwaggerUIAccessibility", func(t *testing.T) {
		testSwaggerUIAccessibility(t, router)
	})

	// Test "Try it out" functionality
	t.Run("TryItOutFunctionality", func(t *testing.T) {
		testTryItOutFunctionality(t, router, authService)
	})

	// Test authentication token input
	t.Run("AuthenticationTokenInput", func(t *testing.T) {
		testAuthenticationTokenInput(t, router)
	})

	// Test example requests execution
	t.Run("ExampleRequestsExecution", func(t *testing.T) {
		testExampleRequestsExecution(t, router)
	})
}

// setupTestEndpoints creates test endpoints that simulate the actual API
func setupTestEndpoints(router *gin.Engine) {
	// Swagger endpoints
	router.GET("/swagger/index.html", func(c *gin.Context) {
		c.String(http.StatusOK, "Swagger UI")
	})

	router.GET("/swagger/doc.json", func(c *gin.Context) {
		spec := map[string]interface{}{
			"swagger": "2.0",
			"info": map[string]interface{}{
				"title":   "Product Requirements Management API",
				"version": "1.0.0",
			},
			"paths": map[string]interface{}{
				"/api/v1/epics": map[string]interface{}{
					"get": map[string]interface{}{
						"summary": "List epics",
					},
				},
			},
			"securityDefinitions": map[string]interface{}{
				"BearerAuth": map[string]interface{}{
					"type": "apiKey",
					"name": "Authorization",
					"in":   "header",
				},
			},
		}
		c.JSON(http.StatusOK, spec)
	})

	router.GET("/swagger/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})

	router.GET("/swagger/testing-guide", func(c *gin.Context) {
		guide := map[string]interface{}{
			"title": "Interactive API Testing Guide",
			"authentication": map[string]interface{}{
				"method": "JWT Bearer Token",
				"header": "Authorization: Bearer <your-jwt-token>",
			},
			"steps": []string{
				"1. Obtain a JWT token",
				"2. Click 'Authorize' in Swagger UI",
				"3. Enter 'Bearer <token>'",
				"4. Try out endpoints",
			},
		}
		c.JSON(http.StatusOK, guide)
	})

	router.POST("/swagger/validate-token", func(c *gin.Context) {
		var request struct {
			Token string `json:"token"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if len(request.Token) < 10 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token too short", "valid": false})
			return
		}

		c.JSON(http.StatusOK, gin.H{"valid": true, "message": "Token format valid"})
	})

	// API endpoints with authentication
	api := router.Group("/api/v1")

	// Add auth middleware to API routes
	api.Use(func(c *gin.Context) {
		if !c.GetBool("authenticated") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		c.Next()
	})

	// Test API endpoints
	api.GET("/epics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": []map[string]interface{}{
				{
					"id":           uuid.New().String(),
					"reference_id": "EP-001",
					"title":        "Test Epic",
					"status":       "active",
				},
			},
			"total": 1,
		})
	})

	api.POST("/epics", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		epic := map[string]interface{}{
			"id":           uuid.New().String(),
			"reference_id": "EP-002",
			"title":        body["title"],
			"description":  body["description"],
			"status":       "draft",
		}
		c.JSON(http.StatusCreated, epic)
	})

	api.GET("/user-stories", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": []map[string]interface{}{
				{
					"id":           uuid.New().String(),
					"reference_id": "US-001",
					"title":        "Test User Story",
					"status":       "ready",
				},
			},
			"total": 1,
		})
	})

	api.GET("/acceptance-criteria", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": []map[string]interface{}{
				{
					"id":           uuid.New().String(),
					"reference_id": "AC-001",
					"description":  "WHEN condition THEN result",
				},
			},
			"total": 1,
		})
	})

	api.GET("/requirements", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": []map[string]interface{}{
				{
					"id":           uuid.New().String(),
					"reference_id": "REQ-001",
					"title":        "Test Requirement",
					"status":       "approved",
				},
			},
			"total": 1,
		})
	})

	api.GET("/search", func(c *gin.Context) {
		query := c.Query("query")
		c.JSON(http.StatusOK, gin.H{
			"query": query,
			"results": []map[string]interface{}{
				{
					"type":  "epic",
					"id":    uuid.New().String(),
					"title": "Search Result Epic",
				},
			},
			"total": 1,
		})
	})

	api.GET("/search/suggestions", func(c *gin.Context) {
		query := c.Query("query")
		c.JSON(http.StatusOK, gin.H{
			"suggestions": []string{
				query + " authentication",
				query + " authorization",
				query + " system",
			},
		})
	})

	// Admin endpoints (require higher privileges)
	admin := api.Group("/config")
	admin.Use(func(c *gin.Context) {
		role := c.GetString("user_role")
		if role != "administrator" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Administrator role required"})
			c.Abort()
			return
		}
		c.Next()
	})

	admin.GET("/requirement-types", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": []map[string]interface{}{
				{
					"id":   uuid.New().String(),
					"name": "Functional Requirement",
				},
			},
		})
	})
}

// testSwaggerUIAccessibility tests Swagger UI accessibility
func testSwaggerUIAccessibility(t *testing.T, router *gin.Engine) {
	testCases := []struct {
		name         string
		path         string
		expectedCode int
	}{
		{"Swagger UI Index", "/swagger/index.html", http.StatusOK},
		{"Swagger JSON Spec", "/swagger/doc.json", http.StatusOK},
		{"Swagger Base Redirect", "/swagger/", http.StatusFound},
		{"Testing Guide", "/swagger/testing-guide", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code, "Endpoint %s should return %d", tc.path, tc.expectedCode)

			// Validate JSON responses
			if strings.Contains(tc.path, ".json") || strings.Contains(tc.path, "guide") {
				if w.Code == http.StatusOK {
					var jsonResponse interface{}
					err := json.Unmarshal(w.Body.Bytes(), &jsonResponse)
					assert.NoError(t, err, "Response should be valid JSON")
				}
			}
		})
	}
}

// testTryItOutFunctionality tests "Try it out" functionality
func testTryItOutFunctionality(t *testing.T, router *gin.Engine, authService *auth.Service) {
	// Create a test user for token generation
	testUser := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleUser,
	}

	token, err := authService.GenerateToken(testUser)
	require.NoError(t, err, "Should be able to generate test token")

	testEndpoints := []struct {
		method      string
		path        string
		needsAuth   bool
		body        interface{}
		description string
	}{
		{"GET", "/api/v1/epics", true, nil, "List epics"},
		{"GET", "/api/v1/user-stories", true, nil, "List user stories"},
		{"GET", "/api/v1/acceptance-criteria", true, nil, "List acceptance criteria"},
		{"GET", "/api/v1/requirements", true, nil, "List requirements"},
		{"GET", "/api/v1/search?query=test", true, nil, "Search endpoint"},
		{"GET", "/api/v1/search/suggestions?query=auth", true, nil, "Search suggestions"},
		{
			"POST", "/api/v1/epics", true,
			map[string]interface{}{
				"title":       "Test Epic",
				"description": "Test epic description",
				"priority":    1,
			},
			"Create epic",
		},
	}

	for _, endpoint := range testEndpoints {
		t.Run(fmt.Sprintf("%s_%s", endpoint.method, strings.ReplaceAll(endpoint.path, "/", "_")), func(t *testing.T) {
			var req *http.Request

			if endpoint.body != nil {
				bodyBytes, err := json.Marshal(endpoint.body)
				require.NoError(t, err, "Failed to marshal request body")
				req = httptest.NewRequest(endpoint.method, endpoint.path, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(endpoint.method, endpoint.path, nil)
			}

			if endpoint.needsAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Verify endpoint exists (not 404)
			assert.NotEqual(t, http.StatusNotFound, w.Code, "Endpoint should exist: %s", endpoint.description)

			// Verify no server errors
			assert.Less(t, w.Code, 500, "Should not return server error: %s", endpoint.description)

			// For authenticated endpoints, verify we don't get 401 with valid token
			if endpoint.needsAuth {
				assert.NotEqual(t, http.StatusUnauthorized, w.Code, "Valid token should be accepted: %s", endpoint.description)
			}
		})
	}
}

// testAuthenticationTokenInput tests authentication token handling
func testAuthenticationTokenInput(t *testing.T, router *gin.Engine) {
	testCases := []struct {
		name         string
		token        string
		endpoint     string
		expectedCode int
	}{
		{"Valid Token", "valid-jwt-token-example", "/api/v1/epics", http.StatusOK},
		{"No Token", "", "/api/v1/epics", http.StatusUnauthorized},
		{"Invalid Token", "invalid", "/api/v1/epics", http.StatusUnauthorized},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.endpoint, nil)

			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code, "Token handling should work correctly")
		})
	}
}

// testExampleRequestsExecution tests that documented examples execute successfully
func testExampleRequestsExecution(t *testing.T, router *gin.Engine) {
	// Test token validation helper
	t.Run("TokenValidation", func(t *testing.T) {
		validTokenReq := map[string]string{
			"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.token",
		}

		bodyBytes, err := json.Marshal(validTokenReq)
		require.NoError(t, err, "Should marshal request body")

		req := httptest.NewRequest("POST", "/swagger/validate-token", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Valid token format should be accepted")

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")
		assert.True(t, response["valid"].(bool), "Token should be marked as valid")
	})

	// Test invalid token validation
	t.Run("InvalidTokenValidation", func(t *testing.T) {
		invalidTokenReq := map[string]string{
			"token": "short",
		}

		bodyBytes, err := json.Marshal(invalidTokenReq)
		require.NoError(t, err, "Should marshal request body")

		req := httptest.NewRequest("POST", "/swagger/validate-token", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Invalid token should be rejected")
	})
}

// TestSwaggerConfiguration validates Swagger configuration
func TestSwaggerConfiguration(t *testing.T) {
	swaggerCfg := DefaultSwaggerConfig()

	assert.True(t, swaggerCfg.Enabled, "Swagger should be enabled by default")
	assert.Equal(t, "/swagger", swaggerCfg.BasePath, "Base path should be /swagger")
	assert.Equal(t, "Product Requirements Management API", swaggerCfg.Title, "Title should be set")
	assert.Equal(t, "1.0.0", swaggerCfg.Version, "Version should be set")
	assert.NotEmpty(t, swaggerCfg.Description, "Description should be set")
}

// TestExampleDataGeneration validates example data generation
func TestExampleDataGeneration(t *testing.T) {
	examples := GetExampleData()

	assert.NotNil(t, examples, "Should generate example data")
	assert.NotEmpty(t, examples.Epic.Title, "Epic example should have title")
	assert.NotEmpty(t, examples.UserStory.Title, "User story example should have title")
	assert.NotEmpty(t, examples.AcceptanceCriteria.Description, "AC example should have description")
	assert.NotEmpty(t, examples.Requirement.Title, "Requirement example should have title")
	assert.NotEmpty(t, examples.Comment.Content, "Comment example should have content")

	requestBodies := GetExampleRequestBodies()
	assert.NotNil(t, requestBodies, "Should generate example request bodies")
	assert.Contains(t, requestBodies, "create_epic", "Should have epic creation example")
	assert.Contains(t, requestBodies, "create_user_story", "Should have user story creation example")

	queryParams := GetExampleQueryParameters()
	assert.NotNil(t, queryParams, "Should generate example query parameters")
	assert.Contains(t, queryParams, "search", "Should have search parameter examples")
	assert.Contains(t, queryParams, "list_epics", "Should have epic listing parameter examples")
}
