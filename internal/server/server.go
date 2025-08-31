package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/database"
	"product-requirements-management/internal/logger"
	"product-requirements-management/internal/server/middleware"
	"product-requirements-management/internal/server/routes"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	router *gin.Engine
	db     *database.DB
}

// New creates a new server instance
func New(cfg *config.Config) (*Server, error) {
	// Initialize logger
	logger.Init(&cfg.Log)

	// Initialize database connections
	db, err := database.Initialize(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Set Gin mode based on log level
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Setup routes with database connection
	routes.Setup(router, cfg, db)

	return &Server{
		config: cfg,
		router: router,
		db:     db,
	}, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port)

	srv := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
		// Still try to close database connections
		if s.db != nil {
			s.db.Close()
		}
		return err
	}

	// Close database connections
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			logger.Errorf("Failed to close database connections: %v", err)
		}
	}

	logger.Info("Server exited")
	return nil
}
