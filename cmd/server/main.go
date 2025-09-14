package main

import (
	"log"
	"product-requirements-management/internal/config"
	"product-requirements-management/internal/server"

	_ "product-requirements-management/docs" // Import generated docs
)

//	@title			Product Requirements Management API
//	@version		1.0.0
//	@description	Comprehensive API for managing product requirements through hierarchical structure of Epics → User Stories → Requirements. Features include full-text search, comment system, relationship mapping, and configurable workflows.
//
//	## API Overview
//	This API provides complete lifecycle management for product requirements with the following key capabilities:
//
//	### Core Entities
//	- **Epics**: High-level features or initiatives (EP-001, EP-002...)
//	- **User Stories**: Feature requirements within epics (US-001, US-002...)
//	- **Acceptance Criteria**: Testable conditions for user stories (AC-001, AC-002...)
//	- **Requirements**: Detailed technical specifications (REQ-001, REQ-002...)
//	- **Comments**: Discussion system with threading and inline comments
//
//	### Key Features
//	- **Hierarchical Structure**: Organized Epic → User Story → Requirement relationships
//	- **Full-Text Search**: PostgreSQL-powered search across all entities with filtering
//	- **Comment System**: General and inline comments with resolution tracking
//	- **Relationship Mapping**: Link requirements with depends_on, blocks, relates_to relationships
//	- **Status Management**: Configurable workflows for each entity type
//	- **Reference IDs**: Human-readable identifiers for easy tracking
//
//	### Authentication & Authorization
//	JWT-based authentication with role-based access control:
//	- **Administrator**: Full system access including configuration management
//	- **User**: Create, read, update entities and manage assignments
//	- **Commenter**: Read access and comment creation only
//
//	### Common Patterns
//
//	#### Pagination
//	Most list endpoints support pagination with `limit` and `offset` parameters:
//	- `limit`: Maximum results (1-100, default 25)
//	- `offset`: Number of results to skip (default 0)
//	- Response includes total count for pagination calculation
//
//	#### Filtering & Sorting
//	List endpoints support filtering by common fields:
//	- `creator_id`, `assignee_id`: Filter by user relationships
//	- `status`, `priority`: Filter by entity state
//	- `order_by`: Sort by field with ASC/DESC (e.g., "created_at DESC")
//
//	#### Error Handling
//	Consistent error responses across all endpoints:
//	- **400 Bad Request**: Validation errors, malformed input
//	- **401 Unauthorized**: Missing or invalid authentication
//	- **403 Forbidden**: Insufficient permissions
//	- **404 Not Found**: Resource not found
//	- **409 Conflict**: Business logic conflicts (dependencies exist)
//	- **500 Internal Server Error**: System errors
//
//	#### Search Capabilities
//	Advanced search with multiple filter options:
//	- Full-text search across titles, descriptions, and content
//	- Entity type filtering (epics, user_stories, requirements, etc.)
//	- Status, priority, and date range filtering
//	- Relevance-based ranking with configurable sorting
//
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT token authentication. Include 'Bearer ' followed by your JWT token. Example: 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...' Tokens expire after 1 hour and must be refreshed through re-authentication.

//	@tag.name			epics
//	@tag.description	Epic management endpoints for high-level features and initiatives. Epics serve as containers for user stories and provide project-level organization.

//	@tag.name			user-stories
//	@tag.description	User story management within epics. User stories represent feature requirements from the user perspective and contain acceptance criteria and detailed requirements.

//	@tag.name			acceptance-criteria
//	@tag.description	Acceptance criteria management for user stories. Define testable conditions that must be met for user story completion using EARS format (Easy Approach to Requirements Syntax).

//	@tag.name			requirements
//	@tag.description	Detailed requirement management with relationship mapping. Requirements provide technical specifications and can be linked with various relationship types (depends_on, blocks, relates_to, conflicts_with, derives_from).

//	@tag.name			comments
//	@tag.description	Comment system for collaboration and feedback. Supports both general comments and inline comments with threading, resolution tracking, and entity associations.

//	@tag.name			search
//	@tag.description	Full-text search capabilities across all entities. Provides advanced filtering, sorting, and suggestion features for efficient content discovery.

//	@tag.name			navigation
//	@tag.description	Hierarchical navigation and entity relationship endpoints. Retrieve entity hierarchies, paths, and relationship structures for navigation interfaces.

//	@tag.name			configuration
//	@tag.description	System configuration management for requirement types, relationship types, and status models. Administrative endpoints for customizing system behavior and workflows.

//	@tag.name			deletion
//	@tag.description	Comprehensive deletion management with dependency validation. Provides safe deletion with cascade options and dependency impact analysis.

//	@tag.name			health
//	@tag.description	System health and monitoring endpoints for service status, database connectivity, and operational metrics.

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
