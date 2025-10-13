# Cross-Platform Docker Build Guide

Руководство по кросс-платформенной сборке Docker образов для различных архитектур.

## Обзор

Обновленный скрипт `scripts/build-and-push.sh` поддерживает сборку образов для множественных архитектур с использованием Docker Buildx.

### Поддерживаемые архитектуры

- **linux/amd64** - Intel/AMD x86_64 (стандартные серверы, ПК)
- **linux/arm64** - ARM64/AArch64 (Apple Silicon, AWS Graviton, Raspberry Pi 4+)

## Быстрый старт

### 1. Мультиплатформенная сборка (по умолчанию)

```bash
# Собрать для amd64 + arm64
make build-push
```

### 2. Сборка только для x86_64

```bash
# Собрать только для Intel/AMD
make build-push-x86
```

### 3. Сборка только для ARM64

```bash
# Собрать только для ARM
make build-push-arm
```

### 4. Одноплатформенная сборка

```bash
# Собрать только для текущей платформы
make build-push-single
```

## Детальная конфигурация

### Переменные окружения

```bash
# Целевые платформы (по умолчанию: linux/amd64,linux/arm64)
export PLATFORMS="linux/amd64,linux/arm64"

# Включить/выключить buildx (по умолчанию: true)
export USE_BUILDX=true

# Имя buildx builder (по умолчанию: multiarch-builder)
export BUILDER_NAME="multiarch-builder"
```

### Примеры использования

```bash
# Только x86_64
PLATFORMS=linux/amd64 ./scripts/build-and-push.sh

# Только ARM64
PLATFORMS=linux/arm64 ./scripts/build-and-push.sh

# Множественные платформы
PLATFORMS=linux/amd64,linux/arm64,linux/arm/v7 ./scripts/build-and-push.sh

# Отключить buildx (одна платформа)
USE_BUILDX=false ./scripts/build-and-push.sh
```

## Команды Make

### Основные команды сборки

```bash
make build-push              # Мультиплатформенная сборка (amd64+arm64)
make build-push-x86          # Только x86_64/amd64
make build-push-arm          # Только ARM64
make build-push-single       # Текущая платформа
```

### Управление buildx

```bash
make build-push-setup        # Настроить buildx builder
make build-push-cleanup      # Очистить артефакты сборки
make build-push-cleanup-buildx # Удалить buildx builder
```

### Информация и аутентификация

```bash
make build-push-info         # Показать информацию о сборке
make build-push-auth         # Аутентификация с GitHub Registry
```

## Docker Buildx

### Что такое Buildx?

Docker Buildx - это расширенный инструмент сборки Docker, который поддерживает:
- Мультиплатформенные сборки
- Расширенные возможности кэширования
- Экспорт в различные форматы
- Параллельную сборку

### Автоматическая настройка

Скрипт автоматически:
1. Проверяет доступность buildx
2. Создает builder с именем `multiarch-builder`
3. Настраивает эмуляцию для кросс-компиляции
4. Показывает доступные платформы

### Ручная настройка buildx

```bash
# Создать builder
docker buildx create --name multiarch-builder --driver docker-container --bootstrap

# Использовать builder
docker buildx use multiarch-builder

# Проверить доступные платформы
docker buildx inspect --bootstrap
```

## Особенности кросс-компиляции

### Go приложения

Go отлично подходит для кросс-компиляции:
- Статическая компиляция
- Встроенная поддержка множественных архитектур
- Нет зависимостей от системных библиотек

### Dockerfile оптимизации

Dockerfile уже оптимизирован для кросс-компиляции:
```dockerfile
# Автоматическое определение архитектуры
ARG TARGETPLATFORM
ARG BUILDPLATFORM

# Статическая компиляция
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ...
```

## Деплой мультиплатформенных образов

### Автоматический выбор архитектуры

Docker автоматически выбирает подходящую архитектуру:

```bash
# На x86_64 сервере
docker pull ghcr.io/username/product-requirements-management:latest
# Автоматически выберет linux/amd64

# На ARM64 сервере (Apple Silicon, AWS Graviton)
docker pull ghcr.io/username/product-requirements-management:latest
# Автоматически выберет linux/arm64
```

### Принудительный выбор архитектуры

```bash
# Принудительно x86_64
docker pull --platform linux/amd64 ghcr.io/username/product-requirements-management:latest

# Принудительно ARM64
docker pull --platform linux/arm64 ghcr.io/username/product-requirements-management:latest
```

### Обновление скрипта деплоя

Скрипт `deploy-from-registry.sh` автоматически поддерживает мультиплатформенные образы.

## Производительность

### Время сборки

| Тип сборки | Время | Описание |
|------------|-------|----------|
| Одна платформа | ~2-3 мин | Быстрая сборка |
| Две платформы | ~4-6 мин | Параллельная сборка |
| Эмуляция | +50-100% | При кросс-компиляции |

### Размер образов

Мультиплатформенные образы не увеличивают размер - каждая платформа хранится отдельно.

### Кэширование

Buildx поддерживает продвинутое кэширование:
```bash
# Кэш между сборками
docker buildx build --cache-from type=local,src=/tmp/.buildx-cache
```

## Мониторинг и отладка

### Проверка поддерживаемых платформ

```bash
# Проверить buildx
docker buildx ls

# Проверить доступные платформы
docker buildx inspect multiarch-builder
```

### Проверка образов в registry

```bash
# Показать все архитектуры образа
docker manifest inspect ghcr.io/username/product-requirements-management:latest
```

### Локальное тестирование

```bash
# Запустить x86_64 образ на ARM (с эмуляцией)
docker run --platform linux/amd64 ghcr.io/username/product-requirements-management:latest

# Запустить ARM64 образ на x86_64 (с эмуляцией)
docker run --platform linux/arm64 ghcr.io/username/product-requirements-management:latest
```

## Устранение неполадок

### Buildx не доступен

```bash
# Обновить Docker до последней версии
# Или установить buildx плагин отдельно

# Проверить версию
docker version
docker buildx version
```

### Ошибки эмуляции

```bash
# Установить qemu для эмуляции
docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

# Проверить доступные эмуляторы
ls /proc/sys/fs/binfmt_misc/
```

### Медленная сборка

```bash
# Использовать нативную сборку на соответствующих машинах
# Или настроить удаленные builders

# Создать remote builder
docker buildx create --name remote-builder --driver docker-container --driver-opt network=host ssh://user@arm-server
```

## CI/CD интеграция

### GitHub Actions

```yaml
name: Multi-platform Build

on:
  push:
    branches: [main]
    tags: ['v*']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
        
      - name: Build and Push
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          PLATFORMS: linux/amd64,linux/arm64
        run: ./scripts/build-and-push.sh
```

### Оптимизация для CI

```bash
# Использовать GitHub Actions cache
export BUILDX_CACHE_FROM="type=gha"
export BUILDX_CACHE_TO="type=gha,mode=max"
```

## Рекомендации

### Для разработки

1. **Используйте одноплатформенную сборку** для быстрой итерации
2. **Тестируйте на целевых архитектурах** перед релизом
3. **Настройте локальный buildx** для консистентности

### Для production

1. **Всегда собирайте мультиплатформенные образы** для гибкости
2. **Тестируйте на всех целевых архитектурах**
3. **Используйте конкретные теги версий** вместо `latest`

### Для CI/CD

1. **Кэшируйте слои** для ускорения сборки
2. **Параллелизируйте сборку** разных архитектур
3. **Используйте матричные сборки** в GitHub Actions

## Заключение

Кросс-платформенная сборка обеспечивает:
- **Гибкость деплоя** на различных архитектурах
- **Оптимизацию производительности** (нативное выполнение)
- **Экономию ресурсов** (ARM серверы дешевле)
- **Будущую совместимость** с новыми архитектурами

Используйте `make build-push` для стандартной мультиплатформенной сборки или специализированные команды для конкретных архитектур.