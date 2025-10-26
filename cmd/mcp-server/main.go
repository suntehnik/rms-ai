package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"product-requirements-management/internal/mcp"
)

// main is the entry point for the MCP Server console application.
// It initializes the server, sets up signal handling for graceful shutdown,
// and starts the STDIO message processing loop.
//
// Command line usage:
//
//	mcp-server                           # Uses default config: ~/.requirements-mcp/config.json
//	mcp-server -config /path/to/config   # Uses specified config file
//	mcp-server -h                        # Shows help
func main() {
	// Parse command line arguments
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file (default: ~/.requirements-mcp/config.json)")
	flag.Parse()

	// Set default config path if not provided
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get user home directory: %v\n", err)
			os.Exit(1)
		}
		configPath = filepath.Join(homeDir, ".requirements-mcp", "config.json")
	}

	// Load configuration from specified path
	config, err := mcp.LoadConfigFromPath(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration from %s: %v\n", configPath, err)
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
