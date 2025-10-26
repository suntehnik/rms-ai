# MCP Server для Product Requirements Management

MCP (Model Context Protocol) сервер для системы управления требованиями к продукту.

## Установка и сборка

```bash
# Сборка сервера
make build

# Или напрямую через Go
go build -o bin/mcp-server ./cmd/mcp-server
```

## Конфигурация

### Аргументы командной строки

```bash
# Использование конфигурации по умолчанию
./bin/mcp-server

# Указание пути к конфигурационному файлу
./bin/mcp-server -config /path/to/config.json

# Показать справку
./bin/mcp-server -h
```

### Конфигурационный файл

По умолчанию сервер ищет конфигурацию в `~/.requirements-mcp/config.json`.

Пример конфигурации (`config.example.json`):

```json
{
  "backend_api_url": "http://localhost:8080",
  "pat_token": "your_personal_access_token_here",
  "request_timeout": "30s",
  "log_level": "info"
}
```

#### Параметры конфигурации

- `backend_api_url` (обязательный) - URL бэкенд API сервера
- `pat_token` (обязательный) - Personal Access Token для аутентификации
- `request_timeout` (опциональный) - Таймаут HTTP запросов (по умолчанию: "30s")
- `log_level` (опциональный) - Уровень логирования: debug, info, warn, error (по умолчанию: "info")

### Создание конфигурации

1. Создайте директорию для конфигурации:
   ```bash
   mkdir -p ~/.requirements-mcp
   ```

2. Скопируйте пример конфигурации:
   ```bash
   cp config.example.json ~/.requirements-mcp/config.json
   ```

3. Отредактируйте конфигурацию, указав правильные значения:
   ```bash
   nano ~/.requirements-mcp/config.json
   ```

## Использование

### Запуск с конфигурацией по умолчанию

```bash
./bin/mcp-server
```

Сервер будет искать конфигурацию в `~/.requirements-mcp/config.json`.

### Запуск с пользовательской конфигурацией

```bash
./bin/mcp-server -config /path/to/my-config.json
```

### Примеры использования

```bash
# Разработка - конфигурация в текущей директории
./bin/mcp-server -config ./dev-config.json

# Продакшн - конфигурация в системной директории
./bin/mcp-server -config /etc/requirements-mcp/config.json

# Тестирование - временная конфигурация
./bin/mcp-server -config /tmp/test-config.json
```

## Интеграция с MCP клиентами

### Claude Desktop

Добавьте в конфигурацию Claude Desktop (`~/.claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "requirements-mcp": {
      "command": "/path/to/bin/mcp-server",
      "args": ["-config", "/path/to/config.json"]
    }
  }
}
```

### Другие MCP клиенты

Сервер использует STDIO транспорт и совместим с любыми MCP клиентами, поддерживающими этот протокол.

## Возможности

MCP сервер предоставляет следующие возможности:

### Tools (Инструменты)
- Создание и управление эпиками
- Создание и управление пользовательскими историями
- Создание и управление требованиями
- Поиск по всем сущностям
- Управление связями между требованиями

### Resources (Ресурсы)
- Доступ к иерархии проекта
- Чтение содержимого эпиков, историй и требований
- Доступ к метаданным и статусам

### Prompts (Шаблоны)
- Шаблоны для создания требований
- Шаблоны для анализа проекта
- Шаблоны для планирования разработки

## Отладка

### Уровни логирования

- `debug` - Подробная отладочная информация
- `info` - Общая информация о работе (по умолчанию)
- `warn` - Предупреждения
- `error` - Только ошибки

### Проверка конфигурации

```bash
# Проверить, что конфигурация загружается корректно
./bin/mcp-server -config /path/to/config.json 2>&1 | head -10
```

### Типичные ошибки

1. **Файл конфигурации не найден**
   ```
   Failed to load configuration from /path/to/config.json: failed to read config file
   ```
   Решение: Проверьте путь к файлу и права доступа.

2. **Неверный формат JSON**
   ```
   Failed to load configuration: failed to parse config file
   ```
   Решение: Проверьте синтаксис JSON в конфигурационном файле.

3. **Отсутствуют обязательные параметры**
   ```
   Failed to load configuration: invalid configuration: backend_api_url is required
   ```
   Решение: Добавьте все обязательные параметры в конфигурацию.

4. **Недоступен бэкенд API**
   ```
   Server error: failed to connect to backend API
   ```
   Решение: Проверьте, что бэкенд API запущен и доступен по указанному URL.

## Разработка

### Сборка из исходников

```bash
git clone <repository>
cd product-requirements-management
make build
```

### Запуск тестов

```bash
make test-fast
```

### Структура проекта

```
cmd/mcp-server/          # Точка входа MCP сервера
├── main.go             # Основной файл с парсингом аргументов
internal/mcp/           # Внутренняя логика MCP
├── config.go          # Загрузка и валидация конфигурации
├── server.go          # Основная логика MCP сервера
└── ...
```