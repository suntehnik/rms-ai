# Requirements Document: MCP Epic Hierarchy Viewer

## Introduction

This document specifies requirements for implementing an MCP tool `tools/epic_hierarchy` that displays the structure of an epic (epic → [steering documents, user stories] → [requirements, acceptance criteria]) in a compact text format. Steering documents and user stories are displayed at the same hierarchical level under the epic. Requirements and acceptance criteria are displayed at the same hierarchical level under each user story. This enables coding agents to quickly assess task coverage and status without manually opening each entity.

## Glossary

- **MCP Tool**: A callable function exposed through the Model Context Protocol that performs specific operations
- **Epic**: A high-level feature or initiative (identified as EP-XXX)
- **Steering Document**: A guiding document linked to an epic that provides context and direction (identified as STD-XXX)
- **User Story**: A user-focused requirement within an epic (identified as US-XXX)
- **Requirement**: A specific technical or functional requirement (identified as REQ-XXX)
- **Acceptance Criteria**: Testable conditions that must be met (identified as AC-XXX)
- **Reference ID**: Human-readable identifier format (EP-XXX, STD-XXX, US-XXX, REQ-XXX, AC-XXX)
- **ASCII Tree**: Text-based hierarchical visualization using characters like ├──, └──, │
- **Status**: Current state of an entity (Backlog, Draft, In Progress, Done, Cancelled, Active, Obsolete)
- **Priority**: Urgency level (P1=Critical, P2=High, P3=Medium, P4=Low)

## Requirements

### Requirement 1: Display Epic Hierarchy Structure

**User Story:** As a coding agent, I want to view an epic hierarchy in compact text form, so that I can understand the full context of the task without opening each entity manually.

#### Acceptance Criteria

1. WHEN the agent invokes `epic_hierarchy` tool with parameter `--epic EP-XXX`, THE Tool SHALL output a tree structure showing "epic → [steering documents, user stories] → [requirements, acceptance criteria]" where steering documents and user stories are at the same hierarchical level under the epic, and requirements and acceptance criteria are at the same hierarchical level under each user story, with status and priority indicators as specified, including proper indentation and branch characters.

2. WHEN the Tool formats the output, THE Tool SHALL use ASCII tree characters (├──, └──, │) to represent hierarchical relationships between entities.

3. WHEN displaying each entity, THE Tool SHALL include the reference ID, status in brackets (except for steering documents which have no status), priority in brackets (for applicable entities), and title on a single line.

4. WHEN the Tool encounters an epic with nested entities, THE Tool SHALL maintain consistent indentation levels (2 spaces per level) throughout the tree structure.

5. WHEN the Tool processes the hierarchy, THE Tool SHALL preserve the natural ordering of entities as stored in the system.

6. WHEN displaying entities under an epic, THE Tool SHALL display steering documents and user stories at the same indentation level, with steering documents appearing first.

7. WHEN displaying entities under a user story, THE Tool SHALL first display all requirements, then display all acceptance criteria.

### Requirement 2: Handle Empty and Missing Data

**User Story:** As a coding agent, I want clear feedback when an epic has no content or doesn't exist, so that I can understand why the output is empty or an error occurred.

#### Acceptance Criteria

1. WHEN an epic has no steering documents and no user stories attached, THE Tool SHALL display "No steering documents or user stories attached" indented under the epic node.

2. WHEN an epic has steering documents but no user stories, THE Tool SHALL display steering documents normally and omit the "No user stories" message.

3. WHEN an epic has user stories but no steering documents, THE Tool SHALL display user stories normally and omit the "No steering documents" message.

4. WHEN a user story has no requirements, THE Tool SHALL display "No requirements" indented under the user story node.

5. WHEN a user story has no acceptance criteria, THE Tool SHALL display "No acceptance criteria" indented under the user story node.

6. WHEN the provided epic reference does not exist in the system, THE Tool SHALL return an error "Epic EP-XXX not found".

### Requirement 3: Display Steering Documents

**User Story:** As a coding agent, I want to see steering documents linked to an epic, so that I can understand the guiding documentation and context alongside user stories.

#### Acceptance Criteria

1. WHEN displaying steering documents under an epic, THE Tool SHALL output each steering document with its reference ID (STD-XXX) and title at the same indentation level as user stories.

2. WHEN displaying steering documents, THE Tool SHALL NOT include status or priority indicators since steering documents do not have these attributes.

3. WHEN displaying steering document description, THE Tool SHALL truncate the text to 80 characters maximum and append "..." if truncation occurs.

4. WHEN multiple steering documents exist for an epic, THE Tool SHALL display each document on its own line with proper tree indentation at the same level as user stories.

5. WHEN a steering document description contains multiple sentences, THE Tool SHALL extract only the first sentence for display.

6. WHEN both steering documents and user stories are present, THE Tool SHALL display all steering documents first, followed by all user stories, maintaining the same indentation level for both.

### Requirement 4: Display Acceptance Criteria Details

**User Story:** As a coding agent, I want to see acceptance criteria for each user story, so that I can understand the validation conditions.

#### Acceptance Criteria

1. WHEN displaying acceptance criteria under a user story, THE Tool SHALL output each AC with its reference ID and the first sentence of its description.

2. WHEN displaying acceptance criteria description, THE Tool SHALL truncate the text to 80 characters maximum and append "..." if truncation occurs.

3. WHEN formatting acceptance criteria, THE Tool SHALL prefix each criterion with its reference ID (AC-XXX).

4. WHEN multiple acceptance criteria exist for a user story, THE Tool SHALL display each criterion on its own line with proper tree indentation at the same level as requirements.

5. WHEN an acceptance criterion description contains multiple sentences, THE Tool SHALL extract only the first sentence for display.

### Requirement 5: Provide Error Handling

**User Story:** As a developer, I want clear error messages, so that I can effectively use and troubleshoot the tool.

#### Acceptance Criteria

5. WHEN the Tool encounters any error condition, THE Tool SHALL return a non-zero exit code to indicate failure to the calling process.

6. WHEN the Tool encounters an error, THE Tool SHALL return an error with human-readable message explaining the failure reason.

## Example Output Format

```
EP-021 [Backlog] [P2] MCP epic hierarchy viewer
│
├── STD-005 Technical Architecture Guidelines
├── STD-006 API Design Standards
│
├─┬ US-064 [Backlog] [P1] Просмотр иерархии эпика
│ │
│ ├── REQ-089 [Draft] [P1] Инструмент выводит структуру эпика
│ ├── REQ-090 [Draft] [P2] Обработка ошибок и документация
│ ├── REQ-091 [Draft] [P2] Поддержка критериев приемки
│ │
│ └── AC-021 — формат дерева, пустые ветки, ошибки, README
│
└─┬ US-065 [Draft] [P2] Another user story
  │
  └── No requirements
```

## Technical Constraints

1. The tool MUST be implemented as an MCP tool callable through the Model Context Protocol
2. The tool MUST use the existing Spexus MCP API for data retrieval
3. The tool MUST support only reference ID formats for epic identification
4. The tool MUST handle Unicode characters in titles and descriptions
6. The tool MUST follow Go coding standards and project structure conventions
