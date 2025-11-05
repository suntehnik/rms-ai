# Requirements Document

## Introduction

This document specifies the requirements for implementing an MCP (Model Context Protocol) tool that allows MCP clients to retrieve all requirements linked to a specific user story. The tool `get_user_story_requirements` will provide structured access to requirement information through the MCP interface, enabling AI applications to understand and work with requirement hierarchies.

## Glossary

- **MCP Client**: An application or service that connects to MCP servers to access tools, resources, and prompts
- **User Story**: A high-level requirement written from the user's perspective, identified by reference ID format "US-XXX"
- **Requirement**: A detailed specification linked to a user story, identified by reference ID format "REQ-XXX"
- **Reference ID**: Human-readable identifier for entities (e.g., "US-047", "REQ-048")
- **MCP Tool**: A function exposed by MCP server that clients can invoke to perform actions
- **Requirements Management System**: The backend system that stores and manages requirements data

## Requirements

### Requirement 1

**User Story:** As an MCP client, I want to get a list of all requirements linked to a specific user story, so that I can see their structure and details via the MCP interface.

#### Acceptance Criteria

1. WHEN MCP клиент вызывает инструмент `get_user_story_requirements` с валидным user_story, THE система SHALL возвращать список всех требований, связанных с указанной пользовательской историей.

2. WHEN MCP клиент передает user_story в инструмент `get_user_story_requirements`, THE система SHALL проверять существование пользовательской истории и возвращать ошибку "User story not found" если история не существует.

3. WHEN система успешно находит требования для пользовательской истории, THE ответ SHALL содержать массив объектов требований с полями: reference_id (REQ-XXX), title, description, status (Draft/Active/Obsolete), priority (1-4), type_name (название типа требования), creator_username, assignee_username (если назначен), created_at, updated_at.

4. WHEN пользовательская история существует но не имеет связанных требований, THE система SHALL возвращать пустой массив requirements с кодом успеха.

5. WHEN MCP клиент передает user_story, THE система SHALL принимать только формат "US-XXX" где XXX - числовой идентификатор.

6. WHEN система возвращает список требований, THE требования SHALL быть отсортированы по приоритету (1-4) и дате создания (created_at DESC).

7. THE MCP инструмент `get_user_story_requirements` SHALL иметь следующую схему: name: "get_user_story_requirements", description: "Get all requirements linked to a specific user story", inputSchema с обязательным параметром user_story типа string с описанием "User story reference ID (e.g., 'US-047')" и паттерном валидации "^US-\\d+$".

8. THE успешный ответ SHALL иметь структуру MCP content с типом "text" и отформатированным текстом, содержащим найденные требования согласно критерию 3. Схема ответа: {"content": [{"type": "text", "text": "Found X requirements for user story US-047:\n\nREQ-XXX: [title] (Priority: [1-4], Status: [Draft/Active/Obsolete], Type: [type_name], Creator: [creator_username], Assignee: [assignee_username])\n[description]\nCreated: [created_at]\n\n[повторить для каждого требования]"}]}

9. WHEN система возвращает информацию о типах требований, THE она SHALL использовать данные из ресурса requirements://requirements-types (согласно REQ-038) и возвращать имена типов (type_name) вместо внутренних идентификаторов (type_id) для обеспечения читаемости ответа.

### Requirement 2

**User Story:** As a system administrator, I want the MCP tool to validate input parameters correctly, so that invalid requests are handled gracefully with clear error messages.

#### Acceptance Criteria

1. GIVEN пользовательская история с reference_id "US-999" НЕ существует в системе AND формат "US-999" корректен (соответствует паттерну "^US-\\d+$") WHEN MCP клиент вызывает инструмент get_user_story_requirements с параметром user_story = "US-999" THEN система возвращает ошибку с сообщением "User story not found" AND ошибка содержит reference_id "US-999" для идентификации AND HTTP статус соответствует стандартам MCP для ошибок "не найдено"

2. GIVEN система настроена на валидацию формата reference_id WHEN MCP клиент вызывает инструмент get_user_story_requirements с некорректными параметрами: user_story = "USER-047" (неправильный префикс), "US-" (отсутствие номера), "US-ABC" (нечисловой суффикс), "047" (только число без префикса) THEN возвращается ошибка валидации "Invalid user story reference ID format" WHEN параметр user_story = "US-047" (корректный формат) THEN валидация проходит успешно (если история существует)

### Requirement 3

**User Story:** As an MCP client, I want to receive properly formatted and sorted requirement data, so that I can present it in a consistent and useful manner.

#### Acceptance Criteria

1. GIVEN пользовательская история US-047 существует в системе AND US-047 имеет связанные требования AND каждое требование имеет: reference_id (REQ-XXX), title (не пустой), description (может быть пустым), status (Draft/Active/Obsolete), priority (1-4), type_name, creator_username, assignee_username (может быть null), created_at, updated_at (валидные даты) WHEN MCP клиент вызывает инструмент get_user_story_requirements с параметром user_story = "US-047" THEN система возвращает успешный ответ со структурой MCP content AND требования отсортированы по приоритету (1-4) затем по дате создания (DESC) AND все поля заполнены AND type_name отображается вместо type_id

2. GIVEN пользовательская история US-047 существует AND US-047 НЕ имеет связанных требований (пустая связь) AND US-047 имеет валидные поля WHEN MCP клиент вызывает инструмент get_user_story_requirements с параметром user_story = "US-047" THEN система возвращает успешный ответ со структурой MCP content с сообщением "Found 0 requirements for user story US-047. No requirements are currently linked to this user story." AND статус ответа указывает на успех (не ошибку) AND сообщение ясно объясняет отсутствие требований

3. GIVEN пользовательская история US-047 существует AND US-047 имеет требования с разными приоритетами и датами создания WHEN MCP клиент вызывает get_user_story_requirements для US-047 THEN требования возвращаются отсортированными: сначала по приоритету (1-4), затем по дате создания (DESC)

### Requirement 4

**User Story:** As an MCP client, I want the tool to integrate properly with existing MCP resources, so that type information is displayed in a human-readable format.

#### Acceptance Criteria

1. GIVEN система имеет типы требований в ресурсе requirements://requirements-types: "Functional", "Interface", "Non-Functional" AND пользовательская история US-047 имеет требования с разными типами WHEN MCP клиент вызывает get_user_story_requirements для US-047 THEN в ответе отображаются имена типов, а не ID: "REQ-048: ... (Type: Functional, ...)", "REQ-054: ... (Type: Interface, ...)" AND НЕ отображаются внутренние идентификаторы типов AND имена типов соответствуют данным из ресурса requirements://requirements-types

2. GIVEN MCP сервер предоставляет инструмент get_user_story_requirements WHEN MCP клиент запрашивает список доступных инструментов (tools/list) THEN в списке присутствует инструмент со следующими характеристиками: name: "get_user_story_requirements", description: "Get all requirements linked to a specific user story", inputSchema с обязательным параметром user_story типа string с описанием "User story reference ID (e.g., 'US-047')" и паттерном валидации "^US-\\d+$" AND схема точно соответствует спецификации AND валидация паттерна работает корректно