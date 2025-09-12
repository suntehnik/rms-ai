package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// BenchmarkClient provides HTTP client utilities for API endpoint testing
type BenchmarkClient struct {
	Client  *http.Client
	BaseURL string
	Token   string
	mu      sync.RWMutex
}

// NewBenchmarkClient creates a new HTTP client for benchmark testing with enhanced configuration
func NewBenchmarkClient(baseURL string) *BenchmarkClient {
	return &BenchmarkClient{
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false, // Enable keep-alives for better performance
			},
		},
		BaseURL: baseURL,
	}
}

// NewBenchmarkClientWithTimeout creates a client with custom timeout
func NewBenchmarkClientWithTimeout(baseURL string, timeout time.Duration) *BenchmarkClient {
	return &BenchmarkClient{
		Client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		},
		BaseURL: baseURL,
	}
}

// SetAuthToken sets the JWT token for authenticated requests
func (bc *BenchmarkClient) SetAuthToken(token string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.Token = token
}

// GET performs a GET request to the specified path with enhanced error handling
func (bc *BenchmarkClient) GET(path string) (*http.Response, error) {
	return bc.executeWithRetry("GET", path, nil)
}

// executeWithRetry executes HTTP requests with retry logic for transient failures
func (bc *BenchmarkClient) executeWithRetry(method, path string, body interface{}) (*http.Response, error) {
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := bc.executeRequest(method, path, body)

		// If successful or non-retryable error, return immediately
		if err == nil || !bc.isRetryableError(err) {
			return resp, err
		}

		// If this was the last attempt, return the error
		if attempt == maxRetries {
			return resp, fmt.Errorf("request failed after %d attempts: %w", maxRetries+1, err)
		}

		// Wait before retrying with exponential backoff
		delay := time.Duration(attempt+1) * baseDelay
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("unexpected retry loop exit")
}

// executeRequest performs the actual HTTP request
func (bc *BenchmarkClient) executeRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, bc.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s request: %w", method, err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	bc.addAuthHeader(req)

	resp, err := bc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP %s request failed: %w", method, err)
	}

	return resp, nil
}

// isRetryableError determines if an error should trigger a retry
func (bc *BenchmarkClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	retryableErrors := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no such host",
		"connection reset by peer",
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryableErr) {
			return true
		}
	}

	return false
}

// POST performs a POST request with JSON body and enhanced error handling
func (bc *BenchmarkClient) POST(path string, body interface{}) (*http.Response, error) {
	return bc.executeWithRetry("POST", path, body)
}

// PUT performs a PUT request with JSON body and enhanced error handling
func (bc *BenchmarkClient) PUT(path string, body interface{}) (*http.Response, error) {
	return bc.executeWithRetry("PUT", path, body)
}

// DELETE performs a DELETE request with enhanced error handling
func (bc *BenchmarkClient) DELETE(path string) (*http.Response, error) {
	return bc.executeWithRetry("DELETE", path, nil)
}

// PATCH performs a PATCH request with JSON body and enhanced error handling
func (bc *BenchmarkClient) PATCH(path string, body interface{}) (*http.Response, error) {
	return bc.executeWithRetry("PATCH", path, body)
}

// addAuthHeader adds the JWT token to the request if available
func (bc *BenchmarkClient) addAuthHeader(req *http.Request) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if bc.Token != "" {
		req.Header.Set("Authorization", "Bearer "+bc.Token)
	}
}

// Request represents a single HTTP request for parallel execution
type Request struct {
	Method string
	Path   string
	Body   interface{}
}

// Response represents the result of an HTTP request
type Response struct {
	StatusCode int
	Body       []byte
	Duration   time.Duration
	Error      error
}

// RunParallelRequests executes multiple HTTP requests concurrently
func (bc *BenchmarkClient) RunParallelRequests(requests []Request, concurrency int) ([]Response, error) {
	if concurrency <= 0 {
		concurrency = len(requests)
	}

	responses := make([]Response, len(requests))
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request Request) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			start := time.Now()
			resp, err := bc.executeRequestFromStruct(request)
			duration := time.Since(start)

			response := Response{
				Duration: duration,
				Error:    err,
			}

			if err == nil && resp != nil {
				response.StatusCode = resp.StatusCode
				if resp.Body != nil {
					body, readErr := io.ReadAll(resp.Body)
					resp.Body.Close()
					if readErr == nil {
						response.Body = body
					} else {
						response.Error = readErr
					}
				}
			}

			responses[index] = response
		}(i, req)
	}

	wg.Wait()
	return responses, nil
}

// executeRequestFromStruct executes a single HTTP request from a Request struct
func (bc *BenchmarkClient) executeRequestFromStruct(req Request) (*http.Response, error) {
	switch req.Method {
	case "GET":
		return bc.executeRequest("GET", req.Path, nil)
	case "POST":
		return bc.executeRequest("POST", req.Path, req.Body)
	case "PUT":
		return bc.executeRequest("PUT", req.Path, req.Body)
	case "PATCH":
		return bc.executeRequest("PATCH", req.Path, req.Body)
	case "DELETE":
		return bc.executeRequest("DELETE", req.Path, nil)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", req.Method)
	}
}

// ParseJSONResponse parses a JSON response into the provided interface
func ParseJSONResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return nil
}
