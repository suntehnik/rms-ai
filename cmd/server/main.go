package main

import (
	"log"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/server"
)

//	@title			Product Requirements Management API
//	@version		1.0.0
//	@description	API for managing product requirements through hierarchical structure of Epics, User Stories, and Requirements. This API uses JWT-based authentication with role-based access control. Three user roles are supported: Administrator (full access), User (entity management), and Commenter (view and comment only).
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT token authentication. Include 'Bearer ' followed by your JWT token. Example: 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' Tokens expire after 1 hour and must be refreshed through re-authentication.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and start server
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
