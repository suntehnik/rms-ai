# Дизайн-документ реализации MCP сервера

## 1. Введение

Этот документ описывает дизайн и архитектуру для интеграции поддержки протокола `Model Context Protocol` (MCP) в существующую систему управления требованиями. Цель — предоставить программируемый интерфейс для взаимодействия с системой через AI агенты (Claude Desktop и другие).

Интеграция будет состоять из двух основных компонентов:
1.  **MCP Server** (консольное Go приложение): Standalone приложение, которое запускается AI хостом, общается с ним через STDIO и делает HTTP запросы к Backend API.
2.  **Backend API Handler**: Новый эндпоинт `/api/v1/mcp` в существующем монолите для обработки MCP запросов.

Дизайн следует принципам, изложенным в `docs/MCP_Integration_Strategy.md`, но с ключевым изменением в механизме аутентификации: вместо OAuth будет использоваться **аутентификация по персональным токенам доступа (Personal Access Tokens - PAT)**.

### 1.1. MCP Архитектура

Model Context Protocol использует трехуровневую архитектуру:

```
┌──────────────────────────────────────┐
│   MCP Host (Claude Desktop)         │  - AI приложение
│   - Управляет взаимодействием       │  - Обрабатывает запросы пользователя
│   - Встроенный MCP Client           │  - Вызывает AI модель
└─────────────┬────────────────────────┘
              │
              │ JSON-RPC 2.0 messages
              │ via Transport Layer (STDIO/HTTP)
              │
┌─────────────▼────────────────────────┐
│   MCP Server (наше приложение)       │  - Провайдер контекста
│   - Предоставляет Resources          │  - Выполняет Tools
│   - Предоставляет Tools              │  - Генерирует Prompts
│   - Предоставляет Prompts            │
└─────────────┬────────────────────────┘
              │
              │ HTTP API calls (с PAT)
              │
┌─────────────▼────────────────────────┐
│   Backend API (Go/Gin)               │  - Бизнес-логика
│   - Обрабатывает /api/v1/mcp        │  - Валидация PAT
└──────────────────────────────────────┘
```

### 1.2. Слои протокола

**Transport Layer** (Транспортный слой):
- **STDIO**: Для локальных MCP серверов, запускаемых как процессы
- **HTTP with SSE**: Для удаленных серверов (будущая поддержка)

**Protocol Layer** (Слой протокола):
- **JSON-RPC 2.0**: Формат всех сообщений
- **Lifecycle**: Инициализация, обмен capabilities, работа
- **Messages**: Requests, Responses, Notifications

### 1.3. Возможности (Capabilities)

**Server Capabilities** (что предоставляет наш MCP Server):
- **Resources**: Контекстные данные (эпики, истории, требования)
- **Tools**: Исполняемые функции (создание, обновление, поиск)
- **Prompts**: Шаблоны взаимодействия (анализ качества, генерация критериев)

**Client Capabilities** (что может делать AI Host):
- **Sampling**: Запрос LLM инференса от сервера (для ИИ-функций)
  - Сервер может запросить у клиента выполнить LLM completion
  - Требует подтверждения пользователя (human-in-the-loop)
  - Используется для сложных ИИ-задач
- **Roots**: Указание корневых контекстов
  - Определяет boundaries для доступа к файловой системе
  - Координирует рабочие директории
  - В нашем случае не используется (работаем с Backend API)

### 1.4. Версионирование протокола

**Формат версии:** `YYYY-MM-DD` (дата последних breaking changes)

**Текущая версия:** `2025-06-18`

**Правила совместимости:**
- Backwards-compatible изменения НЕ меняют версию
- Breaking changes требуют новую версию
- Клиент и сервер ДОЛЖНЫ договориться об одной версии

**Состояния версий:**
- **Draft**: В разработке
- **Current**: Готова к использованию, может получать backwards-compatible обновления
- **Final**: Завершена, не меняется

**Negotiation при инициализации:**
```json
// Client → Server
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-06-18",
    "capabilities": {...},
    "clientInfo": {
      "name": "claude-desktop",
      "version": "1.0.0"
    }
  }
}

// Server → Client
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2025-06-18",
    "capabilities": {...},
    "serverInfo": {
      "name": "requirements-mcp-server",
      "version": "0.1.0"
    }
  }
}
```

Если версии несовместимы, клиент может gracefully завершить соединение.

## 2. Аутентификация: Personal Access Tokens (PAT)

Аутентификация всех MCP запросов производится с помощью PAT. Система PAT уже реализована в проекте и включает:

- Генерацию токенов через веб-интерфейс
- Безопасное хранение (только хеш в БД)
- Валидацию и проверку срока действия
- REST API эндпоинты (`/api/v1/pats`)

MCP Server будет использовать существующую инфраструктуру PAT для аутентификации всех запросов к Backend API через заголовок:
```
Authorization: Bearer <personal_access_token>
```

## 3. Дизайн MCP Server

MCP Server состоит из двух компонентов:

### 3.1. MCP Server (консольное приложение)

Standalone Go приложение, которое:

**Инициализация:**
1. Запускается AI хостом (Claude Desktop)
2. Считывает конфигурацию (URL Backend API, PAT токен)
3. Устанавливает STDIO соединение с AI хостом
4. Получает `initialize` запрос от клиента с версией протокола
5. Проверяет совместимость версий
6. Отвечает с своей версией и capabilities
7. Получает `initialized` notification - готов к работе

**Lifecycle (жизненный цикл):**
1. **Initialize**:
   - Клиент отправляет `initialize` с `protocolVersion` и `capabilities`
   - Сервер отвечает с согласованной `protocolVersion` и своими `capabilities`
   - Если версии несовместимы - возвращает ошибку
2. **Initialized**:
   - Клиент отправляет `initialized` notification
   - Сервер готов к обработке запросов
3. **Active**: Обработка resources/tools/prompts запросов
4. **Shutdown**:
   - Клиент отправляет `shutdown` request
   - Сервер завершает работу и закрывает соединения

**Обработка запросов:**
- Слушает STDIN для JSON-RPC 2.0 сообщений от AI хоста
- Парсит методы: `resources/list`, `resources/read`, `tools/list`, `tools/call`, `prompts/list`, `prompts/get`
- Преобразует MCP запросы в HTTP вызовы к Backend API
- Добавляет PAT в заголовок `Authorization: Bearer <token>`
- Возвращает ответы через STDOUT в формате JSON-RPC 2.0

**Capabilities (объявляемые при инициализации):**
```json
{
  "capabilities": {
    "resources": {
      "subscribe": true,
      "listChanged": true
    },
    "tools": {},
    "prompts": {}
  }
}
```

### 3.2. Backend API Handler (`/api/v1/mcp`)

Новый эндпоинт в существующем Go/Gin монолите:

**Точка входа:**
- **Метод**: `POST`
- **Путь**: `/api/v1/mcp`
- **Content-Type**: `application/json`
- **Authorization**: `Bearer <PAT>`

**Обработка запросов:**
1. **Аутентификация**: Валидация PAT из заголовка Authorization
2. **Парсинг**: Разбор JSON-RPC 2.0 сообщения
3. **Роутинг**: Определение типа запроса (resource/tool/prompt)
4. **Выполнение**: Вызов соответствующих сервисов и репозиториев
5. **Ответ**: Формирование JSON-RPC 2.0 ответа

### 3.3. MCP Primitives (примитивы)

#### 3.3.1. Resources (Ресурсы)

Resources предоставляют контекстные данные AI модели. MCP поддерживает два типа ресурсов:

**1. Direct Resources** - конкретные ресурсы с фиксированным URI
**2. Resource Templates** - шаблоны для динамического поиска ресурсов

**Методы:**
- `resources/list` - Список прямых ресурсов
- `resources/templates/list` - Список шаблонов ресурсов (для поиска/фильтрации)
- `resources/read` - Чтение конкретного ресурса

**Direct Resources URI схема:**
```
epic://{id}                    # Детали эпика
epic://{id}/hierarchy          # Эпик с полной иерархией
user-story://{id}              # Пользовательская история
user-story://{id}/requirements # История с требованиями
requirement://{id}             # Требование
requirement://{id}/relationships # Требование со связями
acceptance-criteria://{id}     # Критерий приемки
```

**Resource Templates** (для динамического поиска):
```
epics://list?status={status}&priority={priority}
user-stories://list?epic_id={epic_id}&status={status}
requirements://search?query={query}&type={type}
```

Templates позволяют AI модели динамически искать ресурсы с параметрами.

**Пример direct resource:**
```json
{
  "uri": "epic://550e8400-e29b-41d4-a716-446655440000",
  "name": "Epic: User Authentication System",
  "description": "Complete authentication system with OAuth 2.0",
  "mimeType": "application/json",
  "text": "{\"id\":\"550e8400...\", \"title\":\"User Authentication\", ...}"
}
```

**Пример resource template:**
```json
{
  "uriTemplate": "epics://list?status={status}&priority={priority}",
  "name": "Search Epics",
  "description": "Find epics by status and priority",
  "mimeType": "application/json"
}
```

#### 3.3.2. Tools (Инструменты)

Tools позволяют AI выполнять действия с подтверждением пользователя.

**Принципы:**
- **Model-controlled**: AI модель решает когда вызывать tool
- **User oversight**: Пользователь видит и подтверждает выполнение
- **JSON Schema validation**: Все параметры валидируются

**Методы:**
- `tools/list` - Список доступных инструментов
- `tools/call` - Вызов инструмента

**Категории tools:**

| Tool Name | Описание | REST Эндпоинт |
| :--- | :--- | :--- |
| `create_epic` | Создать эпик | `POST /api/v1/epics` |
| `update_epic` | Обновить эпик | `PUT /api/v1/epics/{id}` |
| `create_user_story` | Создать пользовательскую историю | `POST /api/v1/user-stories` |
| `update_user_story` | Обновить пользовательскую историю | `PUT /api/v1/user-stories/{id}` |
| `create_requirement` | Создать требование | `POST /api/v1/requirements` |
| `update_requirement` | Обновить требование | `PUT /api/v1/requirements/{id}` |
| `create_relationship` | Создать связь между требованиями | `POST /api/v1/requirement-relationships` |
| `search_global` | Глобальный поиск | `GET /api/v1/search` |
| `search_requirements` | Поиск требований | `GET /api/v1/requirements/search` |

**Пример tool definition с JSON Schema:**
```json
{
  "name": "create_epic",
  "description": "Create a new epic in the system",
  "inputSchema": {
    "type": "object",
    "properties": {
      "title": {
        "type": "string",
        "description": "Epic title",
        "maxLength": 500
      },
      "priority": {
        "type": "integer",
        "description": "Priority (1=Critical, 2=High, 3=Medium, 4=Low)",
        "enum": [1, 2, 3, 4]
      },
      "description": {
        "type": "string",
        "description": "Detailed description",
        "maxLength": 50000
      },
      "assignee_id": {
        "type": "string",
        "format": "uuid",
        "description": "Optional assignee UUID"
      }
    },
    "required": ["title", "priority"]
  }
}
```

#### 3.3.3. Prompts (Промпты)

Prompts предоставляют параметризованные шаблоны для типовых задач.

**Принципы:**
- **User-controlled**: Пользователь выбирает когда использовать prompt
- **Structured templates**: Промпты с четкой структурой и метаданными
- **Context gathering**: Сервер собирает необходимый контекст из системы

**Методы:**
- `prompts/list` - Список доступных промптов
- `prompts/get` - Получить промпт с заполненными аргументами

**Категории prompts:**

| Prompt Name | Описание | Аргументы |
| :--- | :--- | :--- |
| `analyze_requirement_quality` | Анализ качества требования | `requirement_id` |
| `suggest_acceptance_criteria` | Предложить критерии приемки | `requirement_id` |
| `identify_dependencies` | Выявить зависимости | `requirement_id` |
| `generate_user_story` | Сгенерировать пользовательскую историю | `epic_id`, `description` |
| `decompose_epic` | Декомпозировать эпик на истории | `epic_id` |
| `suggest_test_scenarios` | Предложить тест-сценарии | `requirement_id` |

**Пример prompt definition:**
```json
{
  "name": "analyze_requirement_quality",
  "description": "Analyze the quality of a requirement and suggest improvements",
  "arguments": [
    {
      "name": "requirement_id",
      "description": "UUID of the requirement to analyze",
      "required": true
    }
  ]
}
```

### 3.4. Примеры взаимодействия

#### Пример 1: Вызов tool для создания эпика

**1. AI Host → MCP Server (STDIO):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "create_epic",
    "arguments": {
      "title": "User Authentication System",
      "priority": 1,
      "description": "Implement OAuth 2.0 authentication"
    }
  }
}
```

**2. MCP Server → Backend API (HTTP):**
```http
POST /api/v1/mcp HTTP/1.1
Authorization: Bearer <PAT>
Content-Type: application/json

{
  "method": "tools/call",
  "tool": "create_epic",
  "arguments": {
    "title": "User Authentication System",
    "priority": 1,
    "description": "Implement OAuth 2.0 authentication"
  }
}
```

**3. Backend API → MCP Server (HTTP Response):**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "reference_id": "EP-001",
    "title": "User Authentication System",
    "priority": 1,
    "status": "draft"
  }
}
```

**4. MCP Server → AI Host (STDIO):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Epic created successfully with ID: EP-001"
      }
    ]
  }
}
```

#### Пример 2: Чтение resource

**1. AI Host → MCP Server (STDIO):**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "resources/read",
  "params": {
    "uri": "epic://550e8400-e29b-41d4-a716-446655440000/hierarchy"
  }
}
```

**2. MCP Server → Backend API (HTTP):**
```http
POST /api/v1/mcp HTTP/1.1
Authorization: Bearer <PAT>

{
  "method": "resources/read",
  "uri": "epic://550e8400-e29b-41d4-a716-446655440000/hierarchy"
}
```

**3. Backend возвращает полную иерархию эпика с историями и требованиями**

### 3.5. Обработка ИИ-функций через Sampling

Для функций, требующих ИИ-анализа (из `jtbd-mcp-mapping.md`), используется MCP Sampling capability:

**Сценарий:** AI хост запрашивает prompt для анализа качества требования

**1. AI Host запрашивает prompt:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "prompts/get",
  "params": {
    "name": "analyze_requirement_quality",
    "arguments": {
      "requirement_id": "550e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

**2. MCP Server возвращает сформированный промпт с контекстом:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "description": "Analyze requirement quality",
    "messages": [
      {
        "role": "user",
        "content": {
          "type": "text",
          "text": "Analyze this requirement:\n\nTitle: OAuth 2.0 Authentication\nDescription: System must support OAuth 2.0...\n\nProvide quality assessment."
        }
      }
    ]
  }
}
```

**3. AI Host использует собственную LLM для анализа и возвращает результат пользователю**

Таким образом, MCP Server предоставляет контекст через prompts, а AI Host выполняет LLM инференс.

## 4. Маппинг моделей данных

MCP-ресурсы будут напрямую соответствовать моделям данных из `swagger.yaml`.

| MCP Ресурс (`type`) | Модель в `swagger.yaml` | Ключевые поля |
| :--- | :--- | :--- |
| `epic` | `product-requirements-management_internal_models.Epic` | `id`, `reference_id`, `title`, `description`, `status`, `priority` |
| `user_story` | `product-requirements-management_internal_models.UserStory` | `id`, `reference_id`, `title`, `description`, `status`, `priority`, `epic_id` |
| `requirement` | `product-requirements-management_internal_models.Requirement` | `id`, `reference_id`, `title`, `description`, `status`, `priority`, `user_story_id` |
| `comment` | `product-requirements-management_internal_models.Comment` | `id`, `content`, `author_id`, `entity_id`, `entity_type` |

## 5. Обработка ошибок

Ошибки будут возвращаться в стандартном формате JSON-RPC 2.0, как определено в спецификации MCP.

**Пример ошибки (токен недействителен):**
```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32001,
    "message": "Authentication failed: Invalid or expired token",
    "data": {
      "type": "auth_error"
    }
  },
  "id": "request-id-123"
}
```

Сервер будет мапить внутренние ошибки (например, "сущность не найдена", "неверный переход статуса") на соответствующие коды ошибок MCP.

## 6. Концепции реализации

### 6.1. Stateful vs Stateless

**MCP Server (stateful):**
- Поддерживает постоянное соединение с AI хостом через STDIO
- Хранит состояние сессии (capabilities, конфигурацию)
- Управляет lifecycle (initialize → ready → active → shutdown)

**Backend API (stateless):**
- Каждый HTTP запрос независим
- Аутентификация через PAT в каждом запросе
- Не хранит состояние между запросами

### 6.2. Discovery и Metadata

**Resources:**
- Предоставляют четкие метаданные (name, description, mimeType)
- URI схема должна быть понятной и предсказуемой
- Templates позволяют AI модели исследовать доступные данные

**Tools:**
- JSON Schema обеспечивает самодокументированность
- Четкие описания каждого параметра
- Validation на стороне сервера

**Prompts:**
- Структурированные аргументы с описаниями
- Сервер собирает необходимый контекст автоматически
- Результат готов для передачи в LLM

### 6.3. Collaborative Design

MCP сервер должен:
- **Работать совместно с другими серверами**: AI хост может подключать множество MCP серверов одновременно
- **Специализироваться**: Наш сервер фокусируется на управлении требованиями
- **Быть композируемым**: Tools и Resources могут комбинироваться с другими доменами

### 6.4. User Control

**Принципы контроля пользователя:**
- **Tools**: AI предлагает, пользователь подтверждает выполнение
- **Prompts**: Пользователь явно выбирает prompt для использования
- **Resources**: Автоматически загружаются по запросу AI (read-only)

### 6.5. Context Management

**Эффективное предоставление контекста:**
- Direct Resources для известных сущностей (epic://ID)
- Templates для поиска и фильтрации (epics://list?status=active)
- Hierarchical Resources для связанных данных (epic://ID/hierarchy)
- Metadata Resources для справочной информации (requirement-types, statuses)

### 6.6. Роль клиента (AI Host)

MCP Client встроен в AI хост (Claude Desktop) и управляет:

**Connection Management:**
- Запуск MCP Server процесса
- Установка и поддержание STDIO соединения
- Обработка переподключений при сбоях
- Graceful shutdown при завершении

**Protocol Coordination:**
- Negotiation версии протокола
- Обмен capabilities
- Валидация запросов/ответов
- Обработка ошибок

**User Control & Security:**
- **Human-in-the-loop** для tool execution
  - Пользователь видит какой tool будет вызван
  - Пользователь видит параметры
  - Пользователь подтверждает или отклоняет
- **Transparent interactions**
  - Показывает что делает сервер
  - Логирует все операции
- **Permission management**
  - Контролирует доступ к capabilities
  - Может ограничить функциональность сервера

**Sampling Requests:**
Если наш MCP Server захочет использовать LLM (например, для анализа):
1. Сервер отправляет `sampling/createMessage` request клиенту
2. Клиент показывает запрос пользователю
3. Пользователь подтверждает
4. Клиент вызывает LLM (Claude)
5. Результат возвращается серверу

В нашей реализации мы будем использовать это через Prompts (клиент сам вызывает LLM с промптом от сервера).
