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
	initpkg "product-requirements-management/internal/mcp/client/init"
)

// runInitialization handles the interactive initialization mode for setting up
// the MCP server configuration. It guides users through server connection setup,
// credential collection, PAT token generation, and configuration file creation.
func runInitialization(configPath string) error {
	// Create initialization controller
	controller := initpkg.NewInitController()

	// Run the initialization process
	if err := controller.RunInitialization(configPath); err != nil {
		// Handle different types of initialization errors
		if initErr, ok := err.(*initpkg.InitError); ok {
			return handleInitializationError(initErr)
		}
		return fmt.Errorf("initialization failed: %w", err)
	}

	return nil
}

// handleInitializationError provides user-friendly error handling for different error types.
func handleInitializationError(err *initpkg.InitError) error {
	// Display the user-friendly error with guidance
	err.DisplayError()

	// Return a simple error for the exit code
	return fmt.Errorf("initialization failed")
}

// main is the entry point for the MCP Server console application.
// It initializes the server, sets up signal handling for graceful shutdown,
// and starts the STDIO message processing loop.
//
// Command line usage:
//
//	mcp-server                           # Uses default config: ~/.requirements-mcp/config.json
//	mcp-server -config /path/to/config   # Uses specified config file
//	mcp-server -i                        # Run in initialization mode
//	mcp-server --init                    # Run in initialization mode
//	mcp-server -h                        # Shows help
func main() {
	// Parse command line arguments
	var (
		configPath string
		initMode   bool
		initLong   bool
	)
	flag.StringVar(&configPath, "config", "", "Path to configuration file (default: ~/.requirements-mcp/config.json)")
	flag.BoolVar(&initMode, "i", false, "Run in initialization mode")
	flag.BoolVar(&initLong, "init", false, "Run in initialization mode")
	flag.Parse()

	// Check if initialization mode is requested
	if initMode || initLong {
		if err := runInitialization(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Initialization failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

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
