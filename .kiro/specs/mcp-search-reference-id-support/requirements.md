# Requirements Document

## Introduction

This feature enhances the MCP (Model Context Protocol) search functionality to support searching by reference IDs. Currently, the MCP agent cannot find documents when searching by reference IDs like "US-119", "EP-006", etc. This enhancement will enable the search_global tool to properly handle reference ID queries, improving the MCP agent's ability to locate specific entities.

## Glossary

- **MCP Agent**: The Model Context Protocol agent that uses search tools to find documents and entities
- **Reference ID**: Human-readable identifiers for entities (e.g., EP-001, US-119, REQ-045, AC-023)
- **Search Service**: The internal service responsible for performing search operations across all entities
- **Entity Types**: The different types of entities in the system (epic, user_story, acceptance_criteria, requirement, steering_document)
- **Global Search**: The search_global MCP tool that searches across all entity types
- **Query Pattern**: The search query format used by the MCP agent

## Requirements

### Requirement 1

**User Story:** As an MCP agent, I want to search for entities by their reference IDs, so that I can quickly locate specific documents when given reference ID patterns.

#### Acceptance Criteria

1. WHEN the MCP agent searches with parameter "query" set to "US-119", THE Search Service SHALL return the matching user story entity
2. WHEN the MCP agent searches with parameter "query" set to "EP-006", THE Search Service SHALL return the matching epic entity  
3. WHEN the MCP agent searches with parameter "query" set to "REQ-045", THE Search Service SHALL return the matching requirement entity
4. WHEN the MCP agent searches with parameter "query" set to "AC-023", THE Search Service SHALL return the matching acceptance criteria entity
5. WHEN the MCP agent searches with parameter "query" set to "STD-012", THE Search Service SHALL return the matching steering document entity

### Requirement 2

**User Story:** As an MCP agent, I want reference ID searches to work with entity type filtering for both direct and hierarchical lookups, so that I can find specific entities or their related child entities.

#### Acceptance Criteria

1. WHEN the MCP agent searches with parameter "query" matching a reference ID of USER_STORY and entity_types ["user_story"], THE Search Service SHALL return the matching user story entity
2. WHEN the MCP agent searches with parameter "query" matching a reference ID of EPIC and entity_types ["user_story"], THE Search Service SHALL return all user story entities that belong to that epic
3. WHEN the MCP agent searches with parameter "query" matching a reference ID of EPIC and entity_types ["user_story", "requirement"], THE Search Service SHALL return all user story and requirement entities that belong to that epic
4. WHEN the MCP agent searches with parameter "query" matching a reference ID of USER_STORY and entity_types ["requirement"], THE Search Service SHALL return all requirement entities that belong to that user story
5. WHEN the MCP agent searches with parameter "query" matching a reference ID of USER_STORY and entity_types ["acceptance_criteria"], THE Search Service SHALL return all acceptance criteria entities that belong to that user story
6. WHEN the MCP agent searches with parameter "query" matching a reference ID of STEERING_DOCUMENT and entity_types ["epic"], THE Search Service SHALL return all epic entities linked to that steering document

### Requirement 3

**User Story:** As an MCP agent, I want reference ID searches to be case-insensitive and flexible, so that I can find entities regardless of case variations.

#### Acceptance Criteria

1. WHEN the MCP agent searches for "us-119", THE Search Service SHALL return the same results as "US-119"
2. WHEN the MCP agent searches for "ep-006", THE Search Service SHALL return the same results as "EP-006"
3. WHEN the MCP agent searches for "req-045", THE Search Service SHALL return the same results as "REQ-045"
4. WHEN the MCP agent searches for "ac-023", THE Search Service SHALL return the same results as "AC-023"
5. WHEN the MCP agent searches for "std-012", THE Search Service SHALL return the same results as "STD-012"

### Requirement 4

**User Story:** As an MCP agent, I want reference ID searches to have high priority in search results, so that exact reference ID matches appear first in the results.

#### Acceptance Criteria

1. WHEN the MCP agent searches for a reference ID, THE Search Service SHALL prioritize exact reference ID matches over partial text matches
2. WHEN the MCP agent searches for "US-119", THE Search Service SHALL return the exact US-119 entity as the first result
3. WHEN the MCP agent searches for a reference ID that appears in multiple entity descriptions, THE Search Service SHALL rank the entity with the matching reference ID highest
4. WHERE multiple entities match the search query, THE Search Service SHALL sort results with reference ID matches first
5. THE Search Service SHALL maintain existing relevance scoring for non-reference ID searches

### Requirement 5

**User Story:** As an MCP agent, I want reference ID searches to work efficiently, so that search performance remains optimal even with the new functionality.

#### Acceptance Criteria

1. WHEN the MCP agent performs a reference ID search, THE Search Service SHALL complete the search within the same performance bounds as text searches
2. WHEN the Search Service detects a reference ID pattern, THE Search Service SHALL use optimized database queries for reference ID lookups
3. WHERE the search query is clearly a reference ID pattern, THE Search Service SHALL prioritize direct reference ID matching over full-text search
4. THE Search Service SHALL maintain existing caching mechanisms for reference ID searches
5. THE Search Service SHALL not degrade performance of existing text-based searches