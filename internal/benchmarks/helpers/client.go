package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// NewBenchmarkClient creates a new HTTP client for benchmark testing
func NewBenchmarkClient(baseURL string) *BenchmarkClient {
	return &BenchmarkClient{
		Client: &http.Client{
			Timeout: 30 * time.Second,
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

// GET performs a GET request to the specified path
func (bc *BenchmarkClient) GET(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", bc.BaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	bc.addAuthHeader(req)
	return bc.Client.Do(req)
}

// POST performs a POST request with JSON body
func (bc *BenchmarkClient) POST(path string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", bc.BaseURL+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	bc.addAuthHeader(req)
	return bc.Client.Do(req)
}

// PUT performs a PUT request with JSON body
func (bc *BenchmarkClient) PUT(path string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("PUT", bc.BaseURL+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create PUT request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	bc.addAuthHeader(req)
	return bc.Client.Do(req)
}

// DELETE performs a DELETE request
func (bc *BenchmarkClient) DELETE(path string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", bc.BaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DELETE request: %w", err)
	}

	bc.addAuthHeader(req)
	return bc.Client.Do(req)
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
			resp, err := bc.executeRequest(request)
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

// executeRequest executes a single HTTP request
func (bc *BenchmarkClient) executeRequest(req Request) (*http.Response, error) {
	switch req.Method {
	case "GET":
		return bc.GET(req.Path)
	case "POST":
		return bc.POST(req.Path, req.Body)
	case "PUT":
		return bc.PUT(req.Path, req.Body)
	case "DELETE":
		return bc.DELETE(req.Path)
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