# Quick Start - Production Deployment

Быстрый старт для деплоя в production с использованием GitHub Container Registry.

## 🚀 Быстрый старт (5 минут)

### 1. Настройка GitHub Token

```bash
# Создайте Personal Access Token в GitHub с правами write:packages
export GITHUB_TOKEN=ghp_your_token_here
export GITHUB_USERNAME=suntehnik
```

### 2. Сборка и пуш

```bash
# Собрать и запушить в GitHub Registry
make build-push
```

### 3. Настройка production

```bash
# Скопировать и настроить окружение
cp .env.prod.template .env.prod
# Отредактировать .env.prod с вашими значениями
```

### 4. Деплой

```bash
# Развернуть из registry
make deploy-prod
```

## 📋 Основные команды

### Сборка и пуш
```bash
make build-push              # Мультиплатформенная сборка (amd64+arm64)
make build-push-x86          # Только x86_64/amd64
make build-push-arm          # Только ARM64
make build-push-single       # Текущая платформа
make build-push-info         # Показать информацию
```

### Деплой и управление
```bash
make deploy-prod             # Полный деплой
make deploy-prod-update      # Обновить сервисы
make deploy-prod-restart     # Перезапустить
make deploy-prod-status      # Показать статус
make deploy-prod-backup      # Создать бэкап
make deploy-prod-migrate     # Запустить миграции
make deploy-prod-init        # Инициализация БД
```

## 🔧 Управление версиями и архитектурами

```bash
# Деплой конкретной версии
VERSION=v1.0.0 make build-push
VERSION=v1.0.0 make deploy-prod-update

# Сборка для конкретной архитектуры
make build-push-x86          # Только x86_64
make build-push-arm          # Только ARM64

# Откат к предыдущей версии
VERSION=v0.9.0 make deploy-prod-update
```

## 📊 Мониторинг

```bash
# Проверить статус
make deploy-prod-status

# Посмотреть логи
make docker-prod-logs

# Проверить здоровье
curl http://localhost:8080/ready
```

## 🆘 Быстрое решение проблем

### Сервисы не запускаются
```bash
make deploy-prod-stop
make deploy-prod
```

### Проблемы с БД
```bash
make deploy-prod-backup      # Сначала бэкап!
make deploy-prod-migrate     # Затем миграции
```

### Проблемы с образами
```bash
make deploy-prod-pull        # Перескачать образ
make deploy-prod-restart     # Перезапустить
```

## 📁 Структура файлов

```
scripts/
├── build-and-push.sh       # Сборка и пуш в registry
├── deploy-from-registry.sh # Деплой из registry
└── deploy-prod.sh          # Legacy локальная сборка

docker-compose.prod.yml     # Production конфигурация
.env.prod.template         # Шаблон окружения
.env.prod                  # Ваши настройки (создать)
deployment-info.json       # Информация о сборке (автосоздается)
```

## 🔐 Безопасность

1. **Измените все пароли** в `.env.prod`
2. **Сгенерируйте JWT_SECRET** (32+ символа)
3. **Настройте HTTPS** для production
4. **Не коммитьте** `.env.prod` в git

## 📖 Подробная документация

- [CROSS-PLATFORM-BUILD.md](CROSS-PLATFORM-BUILD.md) - Кросс-платформенная сборка
- [DEPLOYMENT-REGISTRY.md](DEPLOYMENT-REGISTRY.md) - Полное руководство по деплою
- [DEPLOYMENT.md](DEPLOYMENT.md) - Общее руководство по деплою
- [README.md](README.md) - Основная документация проекта

---

**Готово!** Ваше приложение должно быть доступно по адресу http://localhost:8080