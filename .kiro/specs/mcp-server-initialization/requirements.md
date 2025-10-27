# Requirements Document

## Introduction

This document outlines the requirements for implementing an interactive initialization mode for the MCP Server console application. The initialization mode will allow users to configure authentication interactively by prompting for server details, credentials, and automatically generating a Personal Access Token (PAT) with a 1-year expiration. This feature will simplify the initial setup process and eliminate the need for manual PAT generation and configuration file creation.

## Glossary

- **MCP Server**: The console Go application that implements the Model Context Protocol server
- **PAT**: Personal Access Token used for authentication with the backend API
- **Backend API**: The Product Requirements Management System's REST API
- **Configuration File**: JSON file stored at ~/.requirements-mcp/config.json containing server settings
- **Initialization Mode**: Interactive setup mode activated by command-line flags

## Requirements

### Requirement 1: Command-Line Flag Support

**User Story:** As a user, I want to activate initialization mode using command-line flags, so that I can easily set up the MCP server without manual configuration.

#### Acceptance Criteria

1. WHEN the user runs the command with `-i` flag THEN the MCP Server SHALL enter initialization mode
2. WHEN the user runs the command with `--init` flag THEN the MCP Server SHALL enter initialization mode
3. WHEN initialization mode is active THEN the MCP Server SHALL not start normal server operations
4. WHEN initialization mode is active THEN the MCP Server SHALL display a welcome message explaining the setup process
5. WHEN both `-i` and `-config` flags are provided THEN the system SHALL use the specified config path for output

### Requirement 2: Interactive Server Configuration

**User Story:** As a user, I want to provide server connection details interactively, so that the system can connect to my backend API.

#### Acceptance Criteria

1. WHEN initialization starts THEN the system SHALL prompt for "Backend API URL" with example format
2. WHEN user enters URL THEN the system SHALL validate URL format using Go's url.Parse
3. WHEN URL is invalid THEN the system SHALL display error message and re-prompt
4. WHEN URL is valid THEN the system SHALL test connectivity by making GET request to /ready endpoint
5. WHEN connectivity test fails THEN the system SHALL display error and allow retry or different URL
6. WHEN connectivity succeeds THEN the system SHALL proceed to credential collection

### Requirement 3: Interactive Credential Collection

**User Story:** As a user, I want to provide my credentials interactively, so that the system can authenticate and generate a PAT token.

#### Acceptance Criteria

1. WHEN collecting credentials THEN the system SHALL prompt for "Username"
2. WHEN collecting credentials THEN the system SHALL prompt for "Password" with hidden input
3. WHEN password input is active THEN the system SHALL not echo characters to terminal
4. WHEN credentials are collected THEN the system SHALL validate they are not empty
5. WHEN credentials are empty THEN the system SHALL display error and re-prompt

### Requirement 4: Authentication and PAT Generation

**User Story:** As a user, I want the system to automatically authenticate and generate a PAT token, so that I don't need to manually create tokens through the web interface.

#### Acceptance Criteria

1. WHEN credentials are provided THEN the system SHALL make POST request to /auth/login endpoint
2. WHEN login request includes THEN the system SHALL send JSON body with username and password
3. WHEN authentication succeeds THEN the system SHALL extract JWT token from response
4. WHEN authentication fails THEN the system SHALL display error message and allow retry
5. WHEN JWT token is obtained THEN the system SHALL make POST request to create PAT with 1-year expiration
6. WHEN creating PAT THEN the system SHALL use Authorization header with Bearer JWT token
7. WHEN PAT creation succeeds THEN the system SHALL extract PAT token from response
8. WHEN PAT creation fails THEN the system SHALL display error and exit gracefully

### Requirement 5: Configuration File Generation

**User Story:** As a user, I want the system to automatically generate a configuration file, so that the MCP server can use the obtained credentials for future operations.

#### Acceptance Criteria

1. WHEN PAT token is obtained THEN the system SHALL create configuration directory ~/.requirements-mcp if it doesn't exist
2. WHEN creating directory THEN the system SHALL set appropriate permissions (0755)
3. WHEN generating config THEN the system SHALL create JSON structure with backend_api_url and pat_token
4. WHEN generating config THEN the system SHALL include default values for request_timeout ("30s") and log_level ("info")
5. WHEN writing config file THEN the system SHALL use proper JSON formatting with indentation
6. WHEN config file exists THEN the system SHALL prompt user for confirmation before overwriting
7. WHEN user confirms overwrite THEN the system SHALL backup existing config with timestamp suffix
8. WHEN config is written THEN the system SHALL set file permissions to 0600 for security

### Requirement 6: PAT Token Configuration

**User Story:** As a system, I want to create PAT tokens with appropriate settings, so that they provide secure long-term access for MCP operations.

#### Acceptance Criteria

1. WHEN creating PAT THEN the system SHALL set expiration to 1 year from creation date
2. WHEN creating PAT THEN the system SHALL set name to "MCP Server - {hostname} - {timestamp}"
3. WHEN creating PAT THEN the system SHALL include description explaining purpose
4. WHEN PAT request is made THEN the system SHALL use POST /api/v1/pats endpoint
5. WHEN PAT is created THEN the system SHALL store only the token value in configuration
6. WHEN PAT creation includes metadata THEN the system SHALL not store sensitive metadata in config file

### Requirement 7: Error Handling and User Feedback

**User Story:** As a user, I want clear error messages and guidance when setup fails, so that I can resolve issues and complete configuration.

#### Acceptance Criteria

1. WHEN network errors occur THEN the system SHALL display user-friendly error messages
2. WHEN authentication fails THEN the system SHALL indicate whether it's username, password, or server issue
3. WHEN URL is unreachable THEN the system SHALL suggest checking server status and network connectivity
4. WHEN PAT creation fails THEN the system SHALL display specific error from API response
5. WHEN file system errors occur THEN the system SHALL display permission and path information
6. WHEN any step fails THEN the system SHALL provide option to retry or exit
7. WHEN initialization completes successfully THEN the system SHALL display success message with next steps

### Requirement 8: Security Considerations

**User Story:** As a security-conscious user, I want the initialization process to handle credentials securely, so that sensitive information is protected.

#### Acceptance Criteria

1. WHEN collecting password THEN the system SHALL use terminal package to hide input
2. WHEN making HTTP requests THEN the system SHALL use HTTPS and validate certificates
3. WHEN storing credentials THEN the system SHALL never store username or password in config file
4. WHEN storing PAT THEN the system SHALL set restrictive file permissions (0600)
5. WHEN logging operations THEN the system SHALL never log passwords or tokens
6. WHEN errors occur THEN the system SHALL not include sensitive data in error messages
7. WHEN process is interrupted THEN the system SHALL clean up any temporary credentials from memory

### Requirement 9: Configuration Validation

**User Story:** As a user, I want the system to validate the generated configuration, so that I can be confident the setup will work correctly.

#### Acceptance Criteria

1. WHEN configuration is generated THEN the system SHALL validate JSON structure
2. WHEN configuration is complete THEN the system SHALL test PAT token by making authenticated request
3. WHEN testing PAT THEN the system SHALL make GET request to /auth/profile endpoint
4. WHEN PAT test succeeds THEN the system SHALL display user information from response
5. WHEN PAT test fails THEN the system SHALL display error and offer to retry setup
6. WHEN validation passes THEN the system SHALL confirm configuration is ready for use

### Requirement 10: User Experience and Guidance

**User Story:** As a user, I want clear guidance throughout the setup process, so that I understand what information is needed and what the system is doing.

#### Acceptance Criteria

1. WHEN initialization starts THEN the system SHALL display welcome message explaining the process
2. WHEN prompting for input THEN the system SHALL provide examples and format requirements
3. WHEN processing steps THEN the system SHALL display progress indicators
4. WHEN making network requests THEN the system SHALL show "Connecting..." or similar status
5. WHEN setup completes THEN the system SHALL display summary of what was configured
6. WHEN setup completes THEN the system SHALL provide instructions for next steps
7. WHEN user needs help THEN the system SHALL provide option to display help information

### Requirement 11: Backup and Recovery

**User Story:** As a user, I want protection against losing existing configuration, so that I can safely run initialization on systems with existing setups.

#### Acceptance Criteria

1. WHEN config file exists THEN the system SHALL detect existing configuration
2. WHEN existing config detected THEN the system SHALL prompt user with options: overwrite, backup, or cancel
3. WHEN user chooses backup THEN the system SHALL create backup with timestamp suffix
4. WHEN backup is created THEN the system SHALL confirm backup location to user
5. WHEN user chooses cancel THEN the system SHALL exit without making changes
6. WHEN overwrite is chosen THEN the system SHALL proceed with new configuration

### Requirement 12: Integration with Existing Configuration System

**User Story:** As a developer, I want initialization mode to integrate seamlessly with existing configuration loading, so that the generated config works with normal server operations.

#### Acceptance Criteria

1. WHEN generating config THEN the system SHALL use same JSON structure as existing Config struct
2. WHEN setting default values THEN the system SHALL match defaults in existing validation logic
3. WHEN writing config file THEN the system SHALL use same file path resolution as LoadConfigFromPath
4. WHEN config is generated THEN it SHALL pass validation in existing Config.Validate method
5. WHEN normal server starts THEN it SHALL successfully load and use generated configuration
6. WHEN testing integration THEN the system SHALL verify config works with existing MCP server startup