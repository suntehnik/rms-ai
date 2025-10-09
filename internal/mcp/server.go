package mcp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Server represents the MCP Server console application.
// It handles STDIO communication with AI hosts and forwards JSON-RPC messages
// to the backend API server.
type Server struct {
	config     *Config
	httpClient *http.Client
	logger     *logrus.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewServer creates a new MCP Server instance with the provided configuration.
func NewServer(config *Config) (*Server, error) {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Set up logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.GetRequestTimeout(),
	}

	return &Server{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Run starts the MCP server and begins processing STDIO messages.
// This method blocks until the server is shut down or an error occurs.
func (s *Server) Run() error {
	s.logger.Info("Starting MCP Server")
	s.logger.WithFields(logrus.Fields{
		"backend_url": s.config.BackendAPIURL,
		"timeout":     s.config.GetRequestTimeout(),
	}).Info("MCP Server configuration loaded")

	// Create scanner for STDIN
	scanner := bufio.NewScanner(os.Stdin)

	// Process messages from STDIN
	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			s.logger.Info("Shutdown requested, stopping message processing")
			return nil
		default:
			// Process the message
			message := scanner.Bytes()
			if len(message) == 0 {
				continue
			}

			s.wg.Add(1)
			go func(msg []byte) {
				defer s.wg.Done()
				s.processMessage(msg)
			}(message)
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		s.logger.WithError(err).Error("Error reading from STDIN")
		return fmt.Errorf("STDIN read error: %w", err)
	}

	s.logger.Info("STDIN closed, shutting down")
	return nil
}

// processMessage handles a single JSON-RPC message by forwarding it to the backend API.
func (s *Server) processMessage(message []byte) {
	s.logger.WithField("message_size", len(message)).Debug("Processing message")

	// Forward message to backend API
	response, err := s.forwardToBackend(message)
	if err != nil {
		s.writeError(fmt.Errorf("backend communication failed: %w", err))
		return
	}

	// Write response to STDOUT
	s.writeResponse(response)
}

// forwardToBackend sends a JSON-RPC message to the backend API and returns the response.
func (s *Server) forwardToBackend(message []byte) ([]byte, error) {
	// Create HTTP request
	req, err := http.NewRequestWithContext(s.ctx, "POST", s.config.BackendAPIURL+"/api/v1/mcp", bytes.NewReader(message))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.PATToken)

	// Send request
	s.logger.WithField("url", req.URL.String()).Debug("Sending request to backend")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		s.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
			"response":    string(responseBody),
		}).Error("Backend API returned error")

		if resp.StatusCode == 401 {
			return nil, fmt.Errorf("authentication failed: invalid PAT token")
		}

		return nil, fmt.Errorf("backend API error: %s", resp.Status)
	}

	s.logger.WithField("response_size", len(responseBody)).Debug("Received response from backend")
	return responseBody, nil
}

// writeResponse writes a successful response to STDOUT.
func (s *Server) writeResponse(data []byte) {
	if _, err := os.Stdout.Write(data); err != nil {
		s.logger.WithError(err).Error("Failed to write response to STDOUT")
	}

	// Ensure response ends with newline for proper JSON-RPC framing
	if len(data) > 0 && data[len(data)-1] != '\n' {
		os.Stdout.Write([]byte("\n"))
	}
}

// writeError writes an error message to STDERR.
func (s *Server) writeError(err error) {
	s.logger.WithError(err).Error("MCP Server error")
	fmt.Fprintf(os.Stderr, "MCP Server Error: %v\n", err)
}

// Shutdown gracefully shuts down the MCP server.
func (s *Server) Shutdown() {
	s.logger.Info("Shutting down MCP Server")

	// Cancel context to stop processing
	s.cancel()

	// Wait for ongoing operations to complete with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("All operations completed, shutdown complete")
	case <-time.After(5 * time.Second):
		s.logger.Warn("Shutdown timeout reached, forcing exit")
	}
}
