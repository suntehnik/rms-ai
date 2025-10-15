# Requirements Document

## Introduction

Данная функциональность позволяет пользователям системы управлять steering документами и связывать их с эпиками для предоставления дополнительного контекста модели при реализации. Steering документы содержат важные инструкции, стандарты, нормы команды и полезную информацию о проекте, которые должны учитываться при выполнении задач.

## Requirements

### Requirement 1

**User Story:** Как пользователь системы, я хочу создавать и управлять steering документами как полноценными сущностями в системе, чтобы определить стандарты и инструкции для команды

#### Acceptance Criteria

1. WHEN пользователь создает новый steering документ THEN система SHALL создать запись в таблице steering_documents с уникальным ID и reference_id (STD-001)
2. WHEN пользователь создает steering документ THEN система SHALL сохранить поля: title, description, creator_id, created_at, updated_at
3. WHEN пользователь редактирует steering документ THEN система SHALL обновить запись через GORM и автоматически обновить поле updated_at
4. WHEN пользователь удаляет steering документ THEN система SHALL удалить запись через GORM с CASCADE удалением связанных отношений
5. WHEN пользователь просматривает список steering документов THEN система SHALL получить записи через GORM с пагинацией
6. WHEN система работает с steering документами THEN система SHALL использовать GORM модели и методы для всех операций с БД

### Requirement 2

**User Story:** Как пользователь системы, я хочу связывать steering документы с эпиками через отношения многие-ко-многим, чтобы модель имела релевантный контекст при реализации

#### Acceptance Criteria

1. WHEN пользователь привязывает steering документ к эпику THEN система SHALL создать запись в таблице epic_steering_documents с epic_id и steering_document_id
2. WHEN пользователь просматривает эпик THEN система SHALL отображать список связанных steering документов через JOIN запрос
3. WHEN пользователь отвязывает steering документ от эпика THEN система SHALL удалить соответствующую запись из таблицы epic_steering_documents
4. WHEN модель работает с эпиком THEN система SHALL автоматически включать связанные steering документы в контекст через API
6. WHEN steering документ удаляется THEN система SHALL автоматически удалить все связи с эпиками (CASCADE DELETE)

### Requirement 3

**User Story:** Как пользователь системы, я хочу искать и сортировать steering документы, чтобы быстро находить нужную информацию

#### Acceptance Criteria

1. WHEN пользователь ищет steering документы THEN система SHALL использовать full-text search по полям title и description
2. WHEN пользователь сортирует steering документы THEN система SHALL поддерживать ORDER BY по created_at, title, updated_at через GORM
3. WHEN пользователь фильтрует документы THEN система SHALL поддерживать фильтрацию по creator_id через GORM Where условия
4. WHEN система выполняет поиск THEN система SHALL использовать PostgreSQL full-text search возможности
5. WHEN пользователь получает результаты поиска THEN система SHALL возвращать результаты с пагинацией

### Requirement 4

**User Story:** Как пользователь системы, я хочу связывать steering документы с эпиками, чтобы они автоматически применялись при работе с конкретными эпиками

#### Acceptance Criteria

1. WHEN пользователь связывает steering документ с эпиком THEN система SHALL использовать связь через таблицу epic_steering_documents
2. WHEN система определяет применимые документы для эпика THEN система SHALL выполнять JOIN запрос через GORM
3. WHEN пользователь работает с эпиком THEN система SHALL автоматически включать связанные steering документы в контекст
4. WHEN связь создается или удаляется THEN система SHALL использовать GORM Association методы
5. WHEN эпик удаляется THEN система SHALL автоматически удалить связи через CASCADE DELETE

### Requirement 5

**User Story:** Как пользователь системы, я хочу иметь контролируемый доступ к steering документам в соответствии с существующими ролями, чтобы обеспечить безопасность данных

#### Acceptance Criteria

1. WHEN пользователь с ролью 'Administrator' работает с документами THEN система SHALL разрешить все операции CRUD
2. WHEN пользователь с ролью 'User' читает steering документ THEN система SHALL разрешить чтение всех документов
3. WHEN пользователь с ролью 'User' создает или редактирует документ THEN система SHALL разрешить операции только для документов где он является creator
4. WHEN пользователь с ролью 'Commenter' работает с документами THEN система SHALL разрешить только чтение документов
5. WHEN система проверяет права THEN система SHALL использовать существующий JWT middleware для аутентификации

### Requirement 6

**User Story:** Как разработчик клиентского приложения, я хочу иметь полноценный REST API для управления steering документами, чтобы интегрировать функциональность в веб-интерфейс

#### Acceptance Criteria

1. WHEN клиент отправляет POST /api/v1/steering-documents THEN система SHALL создать новый steering документ и вернуть HTTP 201 с данными
2. WHEN клиент отправляет GET /api/v1/steering-documents THEN система SHALL вернуть список документов с пагинацией и фильтрацией
3. WHEN клиент отправляет GET /api/v1/steering-documents/:id THEN система SHALL вернуть конкретный документ или HTTP 404
4. WHEN клиент отправляет PUT /api/v1/steering-documents/:id THEN система SHALL обновить документ и создать версию
5. WHEN клиент отправляет DELETE /api/v1/steering-documents/:id THEN система SHALL удалить документ и все связи
6. WHEN клиент отправляет GET /api/v1/epics/:id/steering-documents THEN система SHALL вернуть связанные с эпиком документы
7. WHEN клиент отправляет POST /api/v1/epics/:epic_id/steering-documents/:doc_id THEN система SHALL создать связь между эпиком и документом
8. WHEN клиент отправляет DELETE /api/v1/epics/:epic_id/steering-documents/:doc_id THEN система SHALL удалить связь
9. WHEN клиент отправляет GET /api/v1/steering-documents/:id/versions THEN система SHALL вернуть историю версий документа
10. WHEN все API endpoints вызываются THEN система SHALL требовать JWT аутентификацию и проверять права доступа

### Requirement 7

**User Story:** Как пользователь MCP клиента, я хочу управлять steering документами через MCP протокол, чтобы интегрировать их в рабочий процесс с AI моделями

#### Acceptance Criteria

1. WHEN MCP клиент запрашивает список steering документов THEN система SHALL предоставить MCP tool "list_steering_documents" для получения списка документов
2. WHEN MCP клиент создает steering документ THEN система SHALL предоставить MCP tool "create_steering_document" с параметрами title и description
3. WHEN MCP клиент обновляет steering документ THEN система SHALL предоставить MCP tool "update_steering_document" с параметрами steering_document_id (UUID или STD-XXX), title, description
4. WHEN MCP клиент получает steering документ THEN система SHALL предоставить MCP tool "get_steering_document" с параметром steering_document_id (UUID или STD-XXX)
5. WHEN MCP клиент связывает документ с эпиком THEN система SHALL предоставить MCP tool "link_steering_to_epic" с параметрами steering_document_id (UUID или STD-XXX) и epic_id (UUID или EP-XXX)
6. WHEN MCP клиент отвязывает документ от эпика THEN система SHALL предоставить MCP tool "unlink_steering_from_epic" с параметрами steering_document_id (UUID или STD-XXX) и epic_id (UUID или EP-XXX)
7. WHEN MCP клиент получает документы эпика THEN система SHALL предоставить MCP tool "get_epic_steering_documents" с параметром epic_id (UUID или EP-XXX)
8. WHEN MCP tools вызываются THEN система SHALL использовать существующую MCP инфраструктуру в internal/handlers/mcp_tools_handler.go
9. WHEN MCP tools выполняются THEN система SHALL проверять права доступа через getUserFromContext() функцию с использованием существующей PAT аутентификации
10. WHEN MCP возвращает результаты THEN система SHALL использовать существующий ToolResponse формат с ContentItem массивом