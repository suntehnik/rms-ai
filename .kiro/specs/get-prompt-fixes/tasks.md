# Implementation Plan

Convert the MCP prompt response compliance design into a series of prompts for a code-generation LLM that will implement each step with incremental progress. Make sure that each prompt builds on the previous prompts, and ends with wiring things together. There should be no hanging or orphaned code that isn't integrated into a previous step. Focus ONLY on tasks that involve writing, modifying, or testing code.

- [x] 1. Create MCP-compliant data structures and validation
  - Create MCPRole enum type with validation methods in models package
  - Create ContentChunk struct for structured content representation
  - Update PromptMessage struct to use structured content array instead of plain string
  - Add validation methods for MCP compliance checking
  - _Requirements: FR1, FR2_

- [x] 2. Add role field to Prompt model with database migration
  - Add Role field to Prompt model with appropriate GORM tags and default value
  - Create database migration to add role column with check constraint
  - Update Prompt model's BeforeCreate hook to set default role if not specified
  - Update Prompt model's validation tags to include role validation
  - _Requirements: FR4_

- [x] 3. Implement prompt validation service layer
  - Create PromptValidator struct with ValidateForMCP method
  - Add validation logic for role enum values and content requirements
  - Integrate validator into PromptService with structured error logging
  - Add helper method to transform plain string content into ContentChunk array
  - _Requirements: FR3_

- [x] 4. Update MCP prompt definition generation
  - Modify PromptService.GetMCPPromptDefinition to use prompt.Role instead of hardcoded "system"
  - Transform prompt content from string to structured ContentChunk array
  - Add validation call before generating MCP response
  - Update error handling to return MCP-compliant errors for validation failures
  - _Requirements: FR1, FR2, FR3_

- [x] 5. Update prompt creation and management tools
  - Modify MCP prompt tools to accept and validate role parameter
  - Update CreatePromptRequest struct to include role field with validation
  - Update prompt creation logic to use specified role or default to "assistant"
  - Ensure all prompt management operations preserve role information
  - _Requirements: FR4_

- [ ] 6. Write comprehensive unit tests for MCP compliance
  - Test MCPRole validation with valid and invalid values
  - Test ContentChunk transformation from plain string content
  - Test PromptValidator with various valid and invalid prompt configurations
  - Test GetMCPPromptDefinition with compliant and non-compliant stored data
  - Test error handling and logging for validation failures
  - _Requirements: FR1, FR2, FR3_

- [ ]* 7. Write integration tests for end-to-end MCP prompt flow
  - Test complete prompts/get request flow with valid prompts
  - Test prompts/get error responses for invalid stored data
  - Test database migration and backward compatibility
  - Test MCP response format compliance against specification
  - _Requirements: All functional requirements_

- [ ] 8. Update existing prompts and finalize integration
  - Run database migration to add role field to existing prompts
  - Update any existing test data or fixtures to include role information
  - Verify all MCP prompt endpoints return compliant responses
  - Add observability logging for validation metrics and error tracking
  - _Requirements: All functional requirements_