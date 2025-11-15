# Требования к рефакторингу обработчиков MCP

## Введение

**Эпик EP-040 [BE]:** Рефакторинг обработки инструментов MCP: декомпозиция монолитного `ToolsHandler` на специализированные обработчики по доменным сущностям, явные зависимости, фасад для делегирования вызовов, реорганизация файлов и общих утилит при сохранении внешнего JSON-RPC контракта и тестового покрытия.

**Текущая проблема:** `ToolsHandler` стал чрезмерно большим, нарушает SRP/ISP, зависит от множества сервисов и затрудняет поддержку, тестирование и расширение.

## Функциональные требования

### FR1: Декомпозиция ToolsHandler

**User Story (US-185):** As a developer, I want to split the monolithic ToolsHandler into focused handlers by domain entity (epic, user story, requirement, search, steering document, prompt, acceptance criteria), so that responsibilities are isolated and manageable.

#### Acceptance Criteria

1. [AC-701] WHEN refactoring THEN separate handler structs SHALL exist for each domain scope (epic, user story, requirement, search, steering document, prompt, acceptance criteria) with their own entry points
2. [AC-702] ALL domain-specific logic previously in ToolsHandler SHALL be relocated into the corresponding new handler types
3. [AC-703] EACH new handler SHALL encapsulate only its domain responsibilities without cross-entity operations
4. [AC-704] AFTER decomposition THEN ToolsHandler SHALL no longer contain business logic for individual domains

### FR2: Минимизация зависимостей

**User Story (US-186):** As a developer, I want each MCP handler to depend only on the services it needs, so that coupling is reduced and construction is explicit.

#### Acceptance Criteria

1. [AC-705] EACH handler constructor SHALL accept only the domain services it actually uses
2. [AC-706] Handlers SHALL not hold unused dependencies after refactoring
3. [AC-707] Dependency injection SHALL be explicit and type-safe (no reliance on unused fields or global state)
4. [AC-708] Construction of handlers SHALL fail fast when required dependencies are missing

### FR3: Создание фасада

**User Story (US-187):** As a developer, I want ToolsHandler to act as a facade that composes specialized handlers, so that tool calls are routed without duplicating domain logic.

#### Acceptance Criteria

1. [AC-709] ToolsHandler SHALL compose and store instances of each specialized handler
2. [AC-710] Facade construction SHALL wire dependencies to child handlers without duplicating their logic
3. [AC-711] Initialization of ToolsHandler SHALL fail if any required child handler is missing
4. [AC-712] ToolsHandler SHALL expose a single entry point for tool calls while delegating execution to child handlers

### FR4: Делегирование вызовов

**User Story (US-188):** As a developer, I want HandleToolsCall to delegate by tool name to the correct domain handler, so that routing is centralized and behavior remains consistent.

#### Acceptance Criteria

1. [AC-713] HandleToolsCall SHALL route tool invocations by tool name to the matching domain handler method
2. [AC-714] Unhandled tool names SHALL return the same error behavior as before refactoring
3. [AC-715] Delegate routing SHALL not contain business logic beyond dispatch (validation and execution live in domain handlers)
4. [AC-716] Routing SHALL be covered by tests verifying delegation paths for all supported tools

### FR5: Организация файлов и кода

**User Story (US-189):** As a developer, I want handler code and files organized by domain, so that readability and maintainability improve after refactoring.

#### Acceptance Criteria

1. [AC-717] Each domain handler SHALL live in a dedicated file (e.g., mcp_epic_handler.go, mcp_user_story_handler.go, etc.)
2. [AC-718] Legacy handle* methods SHALL be removed or relocated so that domain logic resides in new handler files
3. [AC-719] Package structure SHALL compile successfully with the new handler files and imports
4. [AC-720] Common helper code SHALL be factored out of handler files to avoid duplication across domains

### FR6: Общие утилиты

**User Story (US-190):** As a developer, I want shared helper logic extracted for MCP handlers, so that common functions are reused and code duplication is avoided.

#### Acceptance Criteria

1. [AC-721] Shared helper functions used by multiple handlers SHALL be moved to a common utility or base component
2. [AC-722] Duplicate implementations of shared helpers SHALL be removed from domain handlers
3. [AC-723] Shared utilities SHALL be reusable across handlers without introducing circular dependencies
4. [AC-724] Shared utilities SHALL be covered by unit tests verifying expected behaviors

## Нефункциональные требования

### NFR1: Сохранение внешнего контракта

**User Story (US-191):** As a product stakeholder, I want the MCP tools JSON-RPC contract unchanged after refactoring, so that clients continue to work without updates.

#### Acceptance Criteria

1. [AC-725] JSON-RPC request/response shapes for tools/call SHALL remain identical to pre-refactor behavior
2. [AC-726] All existing tool names and parameters SHALL keep backward compatibility
3. [AC-727] Error codes and messages returned by tools/call SHALL remain unchanged after refactor
4. [AC-728] Regression tests or contract tests SHALL verify client compatibility post-refactor

### NFR2: Тестовое покрытие

**User Story (US-192):** As a QA engineer, I want tests updated for the refactored handlers, so that coverage is preserved or improved after the changes.

#### Acceptance Criteria

1. [AC-729] Existing MCP handler tests SHALL continue to pass after refactoring
2. [AC-730] New unit tests SHALL cover each specialized handler and the ToolsHandler facade routing
3. [AC-731] Test coverage SHALL not decrease compared to pre-refactor baseline
4. [AC-732] CI pipeline SHALL enforce running the updated MCP handler test suite

## Матрица трассировки

| Уровень | ID | Связанные элементы |
| --- | --- | --- |
| Эпик | EP-040 | US-185–US-192 |
| User Story | US-185 | AC-701–AC-704 |
| User Story | US-186 | AC-705–AC-708 |
| User Story | US-187 | AC-709–AC-712 |
| User Story | US-188 | AC-713–AC-716 |
| User Story | US-189 | AC-717–AC-720 |
| User Story | US-190 | AC-721–AC-724 |
| User Story | US-191 | AC-725–AC-728 |
| User Story | US-192 | AC-729–AC-732 |
