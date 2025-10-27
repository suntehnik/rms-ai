# MCP Server для Product Requirements Management

MCP (Model Context Protocol) сервер для системы управления требованиями к продукту.

## Установка и сборка

```bash
# Сборка сервера
make build

# Или напрямую через Go
go build -o bin/mcp-server ./cmd/mcp-server
```

## Быстрый старт

### Интерактивная настройка (рекомендуется)

Самый простой способ настроить MCP сервер - использовать интерактивный режим инициализации:

```bash
# Запуск интерактивной настройки
./bin/mcp-server -i

# Или полная форма
./bin/mcp-server --init
```

Интерактивный режим проведет вас через:
1. 🌐 Ввод URL сервера API
2. 🔗 Проверку подключения к серверу
3. 🔑 Ввод учетных данных
4. 🎟️ Автоматическую генерацию Personal Access Token
5. 📝 Создание конфигурационного файла
6. ✅ Проверку готовности к работе

### Ручная настройка

Если вы предпочитаете настроить конфигурацию вручную, следуйте инструкциям в разделе [Конфигурация](#конфигурация).

## Конфигурация

### Аргументы командной строки

```bash
# Использование конфигурации по умолчанию
./bin/mcp-server

# Указание пути к конфигурационному файлу
./bin/mcp-server -config /path/to/config.json

# Интерактивная настройка (создание конфигурации)
./bin/mcp-server -i
./bin/mcp-server --init

# Интерактивная настройка с пользовательским путем
./bin/mcp-server -i -config /path/to/my-config.json

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

#### Автоматическое создание (рекомендуется)

Используйте интерактивный режим для автоматического создания конфигурации:

```bash
./bin/mcp-server -i
```

#### Ручное создание

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

4. Создайте Personal Access Token через веб-интерфейс системы управления требованиями:
   - Войдите в систему
   - Перейдите в настройки профиля
   - Создайте новый PAT с именем "MCP Server"
   - Скопируйте токен в поле `pat_token` конфигурации

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

#### Сценарии развертывания

```bash
# Разработка - конфигурация в текущей директории
./bin/mcp-server -config ./dev-config.json

# Продакшн - конфигурация в системной директории
./bin/mcp-server -config /etc/requirements-mcp/config.json

# Тестирование - временная конфигурация
./bin/mcp-server -config /tmp/test-config.json

# Пользовательская конфигурация
./bin/mcp-server -config ~/my-projects/project-a/mcp-config.json
```

#### Первоначальная настройка

```bash
# Шаг 1: Интерактивная настройка
./bin/mcp-server -i

# Пример вывода:
# 🚀 Welcome to MCP Server Setup!
# 
# This wizard will help you configure the MCP server for your
# Product Requirements Management system.
# 
# 🌐 Please enter the Backend API URL (e.g., https://api.example.com): 
# https://requirements.mycompany.com
# 
# 🔗 Testing server connectivity...
# ✅ Server is reachable and ready
# 
# 🔑 Please enter your username: john.doe
# 🔒 Please enter your password: [hidden]
# 
# 🎟️ Generating Personal Access Token...
# ✅ Token name: MCP Server - hostname - 2024-01-15
# ✅ Expires: 2025-01-15
# 
# 📝 Writing configuration file...
# ✅ Configuration saved to: /home/john/.requirements-mcp/config.json
# 
# 🔍 Validating PAT token...
# ✅ Configuration is valid and ready to use
# 
# 🎉 Setup completed successfully!

# Шаг 2: Проверка конфигурации
./bin/mcp-server 2>&1 | head -5

# Пример вывода:
# {"level":"info","msg":"Starting MCP Server","time":"2024-01-15T10:30:00Z"}
# {"backend_url":"https://requirements.mycompany.com","level":"info","msg":"MCP Server configuration loaded","time":"2024-01-15T10:30:00Z","timeout":"30s"}
```

#### Настройка для разных сред

```bash
# Настройка для разработки
./bin/mcp-server -i -config ./configs/dev-config.json
# URL: http://localhost:8080
# Пользователь: dev-user

# Настройка для тестирования  
./bin/mcp-server -i -config ./configs/test-config.json
# URL: https://test-api.mycompany.com
# Пользователь: test-user

# Настройка для продакшена
sudo ./bin/mcp-server -i -config /etc/requirements-mcp/config.json
# URL: https://api.mycompany.com
# Пользователь: prod-service-account
```

#### Управление несколькими конфигурациями

```bash
# Создание конфигураций для разных проектов
./bin/mcp-server -i -config ~/.requirements-mcp/project-a.json
./bin/mcp-server -i -config ~/.requirements-mcp/project-b.json

# Запуск с конкретной конфигурацией
./bin/mcp-server -config ~/.requirements-mcp/project-a.json
./bin/mcp-server -config ~/.requirements-mcp/project-b.json
```

#### Обновление конфигурации

```bash
# Обновление существующей конфигурации
./bin/mcp-server -i -config ~/.requirements-mcp/config.json

# Пример вывода:
# 🚀 Welcome to MCP Server Setup!
# 
# ⚠️  Existing configuration detected: /home/john/.requirements-mcp/config.json
# 
# Current configuration:
#   Server URL: https://requirements.mycompany.com
#   Created: 2024-01-10 15:30:00
# 
# Do you want to overwrite the existing configuration? (y/N): y
# 
# 📋 Creating backup: /home/john/.requirements-mcp/config.json.backup.2024-01-15-103000
# 
# [продолжается обычный процесс настройки...]
```

## Интеграция с MCP клиентами

### Claude Desktop

#### Базовая интеграция

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

#### Интеграция с автоматическим поиском конфигурации

Если вы используете стандартное расположение конфигурации:

```json
{
  "mcpServers": {
    "requirements-mcp": {
      "command": "/path/to/bin/mcp-server"
    }
  }
}
```

#### Интеграция для разных проектов

```json
{
  "mcpServers": {
    "requirements-project-a": {
      "command": "/path/to/bin/mcp-server",
      "args": ["-config", "/home/user/.requirements-mcp/project-a.json"]
    },
    "requirements-project-b": {
      "command": "/path/to/bin/mcp-server", 
      "args": ["-config", "/home/user/.requirements-mcp/project-b.json"]
    }
  }
}
```

#### Полная конфигурация Claude Desktop

```json
{
  "mcpServers": {
    "requirements-mcp": {
      "command": "/usr/local/bin/mcp-server",
      "args": ["-config", "/home/user/.requirements-mcp/config.json"],
      "env": {
        "LOG_LEVEL": "info"
      }
    }
  },
  "globalShortcut": "CommandOrControl+Shift+C"
}
```

### Другие MCP клиенты

Сервер использует STDIO транспорт и совместим с любыми MCP клиентами, поддерживающими этот протокол.

#### Kiro IDE

Добавьте в конфигурацию Kiro (`~/.kiro/settings/mcp.json`):

```json
{
  "mcpServers": {
    "requirements-mcp": {
      "command": "/path/to/bin/mcp-server",
      "args": ["-config", "/path/to/config.json"],
      "disabled": false,
      "autoApprove": ["search_global", "search_requirements"]
    }
  }
}
```

#### Пользовательские MCP клиенты

Для интеграции с собственными клиентами используйте следующие параметры:

- **Транспорт**: STDIO
- **Протокол**: JSON-RPC 2.0
- **Версия MCP**: 2025-06-18
- **Команда запуска**: `/path/to/bin/mcp-server`
- **Аргументы**: `["-config", "/path/to/config.json"]` (опционально)

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

#### Ошибки конфигурации

1. **Файл конфигурации не найден**
   ```
   Failed to load configuration from /path/to/config.json: failed to read config file
   ```
   **Решение:** 
   - Проверьте путь к файлу и права доступа
   - Используйте `./bin/mcp-server -i` для создания конфигурации
   - Убедитесь, что директория `~/.requirements-mcp` существует

2. **Неверный формат JSON**
   ```
   Failed to load configuration: failed to parse config file
   ```
   **Решение:** 
   - Проверьте синтаксис JSON в конфигурационном файле
   - Используйте `jq . ~/.requirements-mcp/config.json` для валидации
   - Пересоздайте конфигурацию через `./bin/mcp-server -i`

3. **Отсутствуют обязательные параметры**
   ```
   Failed to load configuration: invalid configuration: backend_api_url is required
   ```
   **Решение:** 
   - Добавьте все обязательные параметры в конфигурацию
   - Используйте интерактивный режим для автоматического создания

#### Ошибки подключения

4. **Недоступен бэкенд API**
   ```
   Server error: failed to connect to backend API
   ```
   **Решение:** 
   - Проверьте, что бэкенд API запущен и доступен по указанному URL
   - Убедитесь, что URL включает протокол (http:// или https://)
   - Проверьте сетевое подключение и настройки firewall
   - Попробуйте открыть URL в браузере для проверки доступности

5. **Ошибка аутентификации**
   ```
   Server error: authentication failed: invalid PAT token
   ```
   **Решение:**
   - Проверьте, что PAT токен действителен и не истек
   - Убедитесь, что токен имеет необходимые права доступа
   - Создайте новый PAT через веб-интерфейс
   - Используйте `./bin/mcp-server -i` для автоматической генерации нового токена

#### Ошибки инициализации

6. **Ошибка при интерактивной настройке**
   ```
   Initialization failed: Network Error: Failed to connect to server
   ```
   **Решение:**
   - Проверьте правильность введенного URL сервера
   - Убедитесь, что сервер запущен и доступен
   - Проверьте, что эндпоинт `/ready` отвечает
   - Попробуйте другой URL или обратитесь к администратору

7. **Ошибка создания PAT токена**
   ```
   Initialization failed: Authentication Error: Failed to generate PAT token
   ```
   **Решение:**
   - Проверьте правильность учетных данных
   - Убедитесь, что у пользователя есть права на создание PAT токенов
   - Проверьте, что не превышен лимит токенов для пользователя
   - Обратитесь к администратору системы

#### Ошибки файловой системы

8. **Ошибка записи конфигурации**
   ```
   Initialization failed: File System Error: Failed to write configuration file
   ```
   **Решение:**
   - Проверьте права доступа к директории `~/.requirements-mcp`
   - Убедитесь, что достаточно места на диске
   - Попробуйте запустить с правами администратора (если необходимо)
   - Используйте альтернативный путь: `./bin/mcp-server -i -config /tmp/config.json`

## Расширенная конфигурация

### Переменные окружения

MCP сервер поддерживает переопределение настроек через переменные окружения:

```bash
# Переопределение уровня логирования
export MCP_LOG_LEVEL=debug
./bin/mcp-server

# Переопределение таймаута запросов
export MCP_REQUEST_TIMEOUT=60s
./bin/mcp-server

# Переопределение пути к конфигурации
export MCP_CONFIG_PATH=/custom/path/config.json
./bin/mcp-server
```

### Конфигурация логирования

```json
{
  "backend_api_url": "https://api.example.com",
  "pat_token": "your_token_here",
  "request_timeout": "30s",
  "log_level": "debug"
}
```

Доступные уровни логирования:
- `debug` - Подробная отладочная информация (включая HTTP запросы)
- `info` - Общая информация о работе (по умолчанию)
- `warn` - Предупреждения и нестандартные ситуации
- `error` - Только ошибки

### Мониторинг и диагностика

#### Проверка состояния

```bash
# Проверка загрузки конфигурации
./bin/mcp-server 2>&1 | grep "configuration loaded"

# Проверка подключения к API
./bin/mcp-server 2>&1 | grep -E "(Starting|configuration loaded|error)"

# Мониторинг в реальном времени
./bin/mcp-server 2>&1 | jq -r '.time + " " + .level + " " + .msg'
```

#### Логирование в файл

```bash
# Логирование в файл с ротацией
./bin/mcp-server 2>&1 | tee -a /var/log/mcp-server.log

# Логирование только ошибок
./bin/mcp-server 2>&1 | grep '"level":"error"' >> /var/log/mcp-errors.log
```

### Производительность и оптимизация

#### Настройка таймаутов

```json
{
  "backend_api_url": "https://api.example.com",
  "pat_token": "your_token_here",
  "request_timeout": "60s"
}
```

Рекомендуемые значения:
- **Локальная сеть**: `10s-30s`
- **Интернет**: `30s-60s`
- **Медленное соединение**: `60s-120s`

#### Мониторинг производительности

```bash
# Мониторинг времени ответа
./bin/mcp-server 2>&1 | grep -E "(request|response)" | \
  jq -r '.time + " " + (.duration // "N/A")'

# Статистика по типам запросов
./bin/mcp-server 2>&1 | grep '"method":' | \
  jq -r '.method' | sort | uniq -c
```

## Безопасность

### Управление токенами

#### Ротация PAT токенов

```bash
# 1. Создание нового токена через веб-интерфейс
# 2. Обновление конфигурации
./bin/mcp-server -i -config ~/.requirements-mcp/config.json

# 3. Проверка работы с новым токеном
./bin/mcp-server 2>&1 | head -10
```

#### Безопасное хранение конфигурации

```bash
# Установка правильных прав доступа
chmod 600 ~/.requirements-mcp/config.json
chmod 700 ~/.requirements-mcp/

# Проверка прав доступа
ls -la ~/.requirements-mcp/
# Должно показать: drwx------ для директории
# Должно показать: -rw------- для config.json
```

### Аудит и мониторинг

```bash
# Мониторинг аутентификации
./bin/mcp-server 2>&1 | grep -E "(authentication|token|401|403)"

# Логирование всех операций
./bin/mcp-server 2>&1 | jq -r 'select(.level != "debug") | .time + " " + .msg'
```

## Дополнительная документация

### Подробные руководства

- **[Руководство по устранению неполадок](docs/mcp-troubleshooting-guide.md)** - Комплексное руководство по диагностике и решению проблем
- **[Примеры использования](docs/mcp-usage-examples.md)** - Подробные примеры настройки и использования для различных сценариев
- **[Спецификация MCP сервера](mcp-server-specification.md)** - Техническая спецификация и требования
- **[Дизайн-документ](MCP_DESIGN.md)** - Архитектура и дизайн решения

### Быстрые ссылки

- **Первая настройка**: `./bin/mcp-server -i`
- **Проверка конфигурации**: `./bin/mcp-server 2>&1 | head -5`
- **Отладка**: Установите `"log_level": "debug"` в конфигурации
- **Помощь**: `./bin/mcp-server -h`

## Разработка

### Сборка из исходников

```bash
git clone <repository>
cd product-requirements-management
make build
```

### Запуск тестов

```bash
# Быстрые тесты
make test-fast

# Полные тесты
make test

# Тесты с покрытием
make test-coverage
```

### Структура проекта

```
cmd/mcp-server/                    # Точка входа MCP сервера
├── main.go                       # Основной файл с парсингом аргументов
internal/mcp/                     # Внутренняя логика MCP
├── config.go                     # Загрузка и валидация конфигурации
├── server.go                     # Основная логика MCP сервера
└── client/init/                  # Интерактивная инициализация
    ├── controller.go             # Оркестрация процесса настройки
    ├── input.go                  # Обработка пользовательского ввода
    ├── client.go                 # HTTP клиент для API
    ├── config.go                 # Генерация конфигурации
    └── filesystem.go             # Управление файлами
```

### Отладка

```bash
# Запуск с отладочным логированием
./bin/mcp-server -config <(echo '{"backend_api_url":"http://localhost:8080","pat_token":"test","log_level":"debug"}')

# Тестирование с mock сервером
# (требует запущенный backend API на localhost:8080)
```