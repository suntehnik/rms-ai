# Deployment with GitHub Container Registry

Этот документ описывает процесс деплоя с использованием GitHub Container Registry, разделенный на два этапа: сборка и пуш образов, затем деплой из registry.

## Обзор процесса

1. **Сборка и пуш** (`scripts/build-and-push.sh`) - собирает Docker образы и пушит их в GitHub Container Registry
2. **Деплой из registry** (`scripts/deploy-from-registry.sh`) - стягивает образы из registry и разворачивает их

## Настройка GitHub Container Registry

### 1. Создание Personal Access Token

1. Перейдите в GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Создайте новый token с правами:
   - `write:packages` - для пуша образов
   - `read:packages` - для скачивания образов
   - `delete:packages` - для удаления образов (опционально)

### 2. Настройка переменных окружения

```bash
export GITHUB_TOKEN=ghp_your_token_here
export GITHUB_USERNAME=suntehnik
export REPOSITORY_NAME=rms-ai
```

## Сборка и пуш образов

### Команды Make

```bash
# Сборка и пуш в registry
make build-push

# Показать информацию о сборке
make build-push-info

# Аутентификация с GitHub Registry
make build-push-auth
```

### Прямое использование скрипта

```bash
# Основная команда - сборка и пуш
./scripts/build-and-push.sh

# Показать информацию об образе
./scripts/build-and-push.sh info

# Только аутентификация
./scripts/build-and-push.sh auth

# Очистка локальных образов
./scripts/build-and-push.sh cleanup
```

### Переменные окружения для сборки

```bash
# Обязательные
GITHUB_TOKEN=ghp_xxx                    # GitHub Personal Access Token
GITHUB_USERNAME=suntehnik              # GitHub username (автоопределяется из git)

# Опциональные
REPOSITORY_NAME=rms-ai                          # Имя репозитория
VERSION=v1.0.0                         # Версия (автоопределяется из git)
```

### Автоматическое определение версии

Скрипт автоматически определяет версию на основе git:
- Если есть git tag - использует его как версию
- Если нет тега - использует `branch-commit_hash`
- Всегда создает тег `latest`

## Деплой из Registry

### Команды Make

```bash
# Полный деплой из registry
make deploy-prod

# Управление сервисами
make deploy-prod-backup     # Создать бэкап
make deploy-prod-init       # Запустить инициализацию БД
make deploy-prod-migrate    # Запустить миграции
make deploy-prod-update     # Обновить сервисы новым образом
make deploy-prod-restart    # Перезапустить сервисы
make deploy-prod-status     # Показать статус
make deploy-prod-stop       # Остановить сервисы
make deploy-prod-pull       # Скачать образ из registry
```

### Прямое использование скрипта

```bash
# Полный деплой
./scripts/deploy-from-registry.sh deploy

# Управление данными
./scripts/deploy-from-registry.sh backup    # Создать бэкап
./scripts/deploy-from-registry.sh init      # Инициализация БД (./bin/init)
./scripts/deploy-from-registry.sh migrate   # Миграции БД

# Управление сервисами
./scripts/deploy-from-registry.sh update    # Обновить с новым образом
./scripts/deploy-from-registry.sh restart   # Перезапустить сервисы
./scripts/deploy-from-registry.sh status    # Показать статус
./scripts/deploy-from-registry.sh logs      # Показать логи
./scripts/deploy-from-registry.sh stop      # Остановить сервисы

# Работа с образами
./scripts/deploy-from-registry.sh pull      # Скачать образ
```

### Переменные окружения для деплоя

```bash
# Для доступа к registry
GITHUB_TOKEN=ghp_xxx                    # GitHub Personal Access Token
GITHUB_USERNAME=your_username           # GitHub username

# Для выбора версии
VERSION=v1.0.0                         # Версия образа (по умолчанию: latest)
REPOSITORY_NAME=product-requirements-management  # Имя репозитория
```

## Типичные сценарии использования

### 1. Первоначальный деплой

```bash
# 1. Настроить переменные окружения
export GITHUB_TOKEN=ghp_your_token_here
export GITHUB_USERNAME=your_username

# 2. Собрать и запушить образ
make build-push

# 3. Настроить production окружение
cp .env.prod.template .env.prod
# Отредактировать .env.prod

# 4. Развернуть из registry
make deploy-prod
```

### 2. Обновление приложения

```bash
# 1. Собрать новую версию
VERSION=v1.1.0 make build-push

# 2. Обновить деплой
VERSION=v1.1.0 make deploy-prod-update
```

### 3. Откат к предыдущей версии

```bash
# Развернуть конкретную версию
VERSION=v1.0.0 make deploy-prod-update
```

### 4. Обслуживание БД

```bash
# Создать бэкап
make deploy-prod-backup

# Запустить миграции
make deploy-prod-migrate

# Инициализация БД (если нужно)
make deploy-prod-init
```

## Файл deployment-info.json

После успешной сборки создается файл `deployment-info.json` с информацией о сборке:

```json
{
  "image": "ghcr.io/username/product-requirements-management",
  "version": "v1.0.0",
  "latest_tag": "ghcr.io/username/product-requirements-management:latest",
  "version_tag": "ghcr.io/username/product-requirements-management:v1.0.0",
  "git_commit": "abc123...",
  "git_branch": "main",
  "git_tag": "v1.0.0",
  "build_date": "2024-01-15T10:30:00Z",
  "registry": "ghcr.io",
  "repository": "username/product-requirements-management"
}
```

Этот файл используется скриптом деплоя для автоматического определения версии.

## Безопасность

### Рекомендации по токенам

1. **Никогда не коммитьте токены** в git
2. **Используйте переменные окружения** или секреты CI/CD
3. **Регулярно обновляйте токены**
4. **Используйте минимальные права** для токенов

### Настройка CI/CD

Для автоматизации в GitHub Actions:

```yaml
name: Build and Deploy

on:
  push:
    branches: [main]
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build and Push
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GITHUB_USERNAME: ${{ github.actor }}
        run: ./scripts/build-and-push.sh
        
  deploy:
    needs: build
    runs-on: self-hosted  # На production сервере
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GITHUB_USERNAME: ${{ github.actor }}
        run: ./scripts/deploy-from-registry.sh deploy
```

## Мониторинг и логи

### Проверка статуса

```bash
# Статус всех сервисов
make deploy-prod-status

# Логи в реальном времени
make docker-prod-logs

# Проверка здоровья приложения
curl http://localhost:8080/ready
```

### Информация об образах

```bash
# Показать текущие образы
docker images | grep requirements-management

# Показать информацию о запущенных контейнерах
docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}"
```

## Устранение неполадок

### Проблемы с аутентификацией

```bash
# Проверить аутентификацию
docker login ghcr.io

# Переаутентификация
make build-push-auth
```

### Проблемы с образами

```bash
# Принудительно скачать образ
docker pull ghcr.io/username/product-requirements-management:latest

# Очистить локальные образы
docker image prune -f
```

### Проблемы с сервисами

```bash
# Проверить логи
docker logs requirements-app

# Перезапустить сервисы
make deploy-prod-restart

# Полная переустановка
make deploy-prod-stop
make deploy-prod
```

## Сравнение с локальной сборкой

| Аспект | Локальная сборка | Registry сборка |
|--------|------------------|-----------------|
| Скорость деплоя | Медленно (сборка каждый раз) | Быстро (готовый образ) |
| Размер | Требует исходный код | Только образ |
| Версионирование | Ограниченное | Полное с тегами |
| Откат | Сложно | Легко |
| CI/CD | Сложная настройка | Простая интеграция |
| Безопасность | Исходный код на сервере | Только образ |

## Заключение

Использование GitHub Container Registry обеспечивает:
- **Быстрый деплой** готовых образов
- **Версионирование** и легкий откат
- **Безопасность** через отсутствие исходного кода на production
- **Масштабируемость** для множественных серверов
- **Интеграцию с CI/CD** пайплайнами

Этот подход рекомендуется для production окружений.