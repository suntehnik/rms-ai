package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"product-requirements-management/internal/mcp"
)

// main is the entry point for the MCP Server console application.
// It initializes the server, sets up signal handling for graceful shutdown,
// and starts the STDIO message processing loop.
func main() {
	// Load configuration from ~/.requirements-mcp/config.json
	config, err := mcp.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Create MCP server instance
	server, err := mcp.NewServer(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create MCP server: %v\n", err)
		os.Exit(1)
	}

	// Set up graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Run()
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		server.Shutdown()
	case err := <-errChan:
		if err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}
}
