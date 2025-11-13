package middleware

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                     string
		allowedOrigins           string
		requestOrigin            string
		expectedAllowOrigin      string
		expectedAllowCredentials string
		method                   string
		expectedStatus           int
	}{
		{
			name:                     "allowed origin with credentials",
			allowedOrigins:           "http://localhost:5173,http://localhost:3000",
			requestOrigin:            "http://localhost:5173",
			expectedAllowOrigin:      "http://localhost:5173",
			expectedAllowCredentials: "true",
			method:                   "GET",
			expectedStatus:           200,
		},
		{
			name:                     "allowed origin OPTIONS request",
			allowedOrigins:           "http://localhost:5173",
			requestOrigin:            "http://localhost:5173",
			expectedAllowOrigin:      "http://localhost:5173",
			expectedAllowCredentials: "true",
			method:                   "OPTIONS",
			expectedStatus:           204,
		},
		{
			name:                     "disallowed origin",
			allowedOrigins:           "http://localhost:5173",
			requestOrigin:            "http://evil.com",
			expectedAllowOrigin:      "http://localhost:5173",
			expectedAllowCredentials: "",
			method:                   "GET",
			expectedStatus:           200,
		},
		{
			name:                     "no origin header",
			allowedOrigins:           "http://localhost:5173",
			requestOrigin:            "",
			expectedAllowOrigin:      "http://localhost:5173",
			expectedAllowCredentials: "", // No credentials when no origin
			method:                   "GET",
			expectedStatus:           200,
		},
		{
			name:                     "wildcard allowed origins",
			allowedOrigins:           "*",
			requestOrigin:            "http://any-origin.com",
			expectedAllowOrigin:      "http://any-origin.com",
			expectedAllowCredentials: "true",
			method:                   "GET",
			expectedStatus:           200,
		},
		{
			name:                     "default origins when env not set",
			allowedOrigins:           "",
			requestOrigin:            "http://localhost:5173",
			expectedAllowOrigin:      "http://localhost:5173",
			expectedAllowCredentials: "true",
			method:                   "GET",
			expectedStatus:           200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.allowedOrigins != "" {
				os.Setenv("CORS_ALLOWED_ORIGINS", tt.allowedOrigins)
			} else {
				os.Unsetenv("CORS_ALLOWED_ORIGINS")
			}
			defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

			// Create test router
			router := gin.New()
			router.Use(CORS())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "ok"})
			})

			// Create request
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert status
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Assert CORS headers
			assert.Equal(t, tt.expectedAllowOrigin, w.Header().Get("Access-Control-Allow-Origin"))

			if tt.expectedAllowCredentials != "" {
				assert.Equal(t, tt.expectedAllowCredentials, w.Header().Get("Access-Control-Allow-Credentials"))
			}

			// Assert other CORS headers are present
			assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
			assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))

			// Verify X-Auth-Skip is in allowed headers
			allowedHeaders := w.Header().Get("Access-Control-Allow-Headers")
			assert.Contains(t, allowedHeaders, "X-Auth-Skip")
		})
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	router := gin.New()
	router.Use(CORS())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	// Create preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert preflight response
	assert.Equal(t, 204, w.Code)
	assert.Equal(t, "http://localhost:5173", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORS_MultipleOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)

	os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, http://localhost:5173, http://localhost:8080")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	origins := []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://localhost:8080",
	}

	for _, origin := range origins {
		t.Run("origin_"+origin, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", origin)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			assert.Equal(t, origin, w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		})
	}
}
