# Примеры использования скриптов бэкапа

## Настройка

### 1. Создание конфигурационного файла

```bash
# Скопировать пример
cp .env.backup.example .env.backup

# Отредактировать под ваши настройки
nano .env.backup
```

### 2. Пример .env.backup для локальной разработки

```bash
# Локальная разработка с Docker Compose
DB_HOST=postgres
DB_USER=postgres
DB_NAME=product_requirements
DB_PORT=5432
DB_PASSWORD=dev_password
CONTAINER_NAME=product-requirements-postgres-1
BACKUP_DIR=./backups
BACKUP_FORMAT=sql
COMPRESS_BACKUP=true
BACKUP_RETENTION_DAYS=7
```

### 3. Пример .env.production

```bash
# Продакшен окружение
DB_HOST=prod-postgres
DB_USER=prod_user
DB_NAME=prod_requirements
DB_PORT=5432
DB_PASSWORD=secure_production_password
CONTAINER_NAME=prod-postgres-container
BACKUP_DIR=/var/backups/postgres
BACKUP_FORMAT=custom
COMPRESS_BACKUP=false
BACKUP_RETENTION_DAYS=30
```

## Использование через Make команды

### Создание бэкапов

```bash
# Создать бэкап с настройками по умолчанию (.env)
make backup-db

# Создать бэкап продакшена
make backup-db-prod

# Создать бэкап staging окружения
make backup-db-staging

# Посмотреть доступные бэкапы
make list-backups

# Очистить старые бэкапы
make clean-old-backups
```

### Восстановление

```bash
# Восстановить из бэкапа
make restore-db BACKUP_FILE=backups/backup_postgres_mydb_20231023_143022.sql

# Восстановить продакшен
make restore-db-prod BACKUP_FILE=backups/backup_prod_20231023_143022.dump
```

### Безопасные миграции

```bash
# Создать бэкап перед миграцией
make backup-before-migrate

# Создать бэкап и выполнить миграцию
make migrate-safe
```

## Прямое использование скриптов

### Создание бэкапов

```bash
# Использовать .env по умолчанию
./scripts/backup_postgres.sh

# Использовать конкретный .env файл
./scripts/backup_postgres.sh .env.production

# Посмотреть справку
./scripts/backup_postgres.sh --help
```

### Восстановление

```bash
# Восстановить из SQL файла
./scripts/restore_postgres.sh backups/backup.sql

# Восстановить с конкретным .env
./scripts/restore_postgres.sh backups/backup.sql .env.production

# Принудительное восстановление без подтверждения
FORCE_RESTORE=true ./scripts/restore_postgres.sh backups/backup.sql
```

## Автоматизация

### Crontab для регулярных бэкапов

```bash
# Редактировать crontab
crontab -e

# Добавить ежедневный бэкап в 2:00 AM
0 2 * * * cd /path/to/project && make backup-db-prod >> /var/log/postgres_backup.log 2>&1

# Добавить еженедельную очистку старых бэкапов в воскресенье в 3:00 AM
0 3 * * 0 cd /path/to/project && make clean-old-backups >> /var/log/postgres_cleanup.log 2>&1
```

### Systemd timer (альтернатива cron)

Создать `/etc/systemd/system/postgres-backup.service`:
```ini
[Unit]
Description=PostgreSQL Database Backup
After=docker.service

[Service]
Type=oneshot
User=your_user
WorkingDirectory=/path/to/project
ExecStart=/usr/bin/make backup-db-prod
StandardOutput=journal
StandardError=journal
```

Создать `/etc/systemd/system/postgres-backup.timer`:
```ini
[Unit]
Description=Run PostgreSQL backup daily
Requires=postgres-backup.service

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

Активировать:
```bash
sudo systemctl enable postgres-backup.timer
sudo systemctl start postgres-backup.timer
```

## Сценарии использования

### 1. Миграция между окружениями

```bash
# На продакшене
make backup-db-prod

# Скопировать бэкап на staging
scp backups/backup_prod_*.sql.gz staging-server:/tmp/

# На staging сервере
make restore-db BACKUP_FILE=/tmp/backup_prod_*.sql.gz
```

### 2. Тестирование миграций

```bash
# Создать бэкап перед тестированием
make backup-db

# Выполнить миграцию
make migrate-up

# Если что-то пошло не так - откатиться
make migrate-down
make restore-db BACKUP_FILE=backups/latest_backup.sql
```

### 3. Регулярное обслуживание

```bash
#!/bin/bash
# maintenance.sh

# Создать бэкап
make backup-db-prod

# Очистить старые бэкапы
make clean-old-backups

# Проверить состояние базы данных
docker exec postgres psql -U postgres -d product_requirements -c "SELECT version();"

# Обновить статистику
docker exec postgres psql -U postgres -d product_requirements -c "ANALYZE;"
```

## Мониторинг и логирование

### Логирование бэкапов

```bash
# Создать директорию для логов
mkdir -p logs

# Запустить бэкап с логированием
make backup-db-prod 2>&1 | tee logs/backup_$(date +%Y%m%d_%H%M%S).log
```

### Проверка успешности бэкапа

```bash
#!/bin/bash
# check_backup.sh

BACKUP_DIR="./backups"
LATEST_BACKUP=$(ls -t $BACKUP_DIR/backup_*.sql* 2>/dev/null | head -n1)

if [[ -n "$LATEST_BACKUP" ]]; then
    SIZE=$(du -sh "$LATEST_BACKUP" | cut -f1)
    echo "✅ Latest backup: $LATEST_BACKUP ($SIZE)"
    
    # Проверить, что бэкап не пустой
    if [[ "$SIZE" == "0B" ]] || [[ "$SIZE" == "0" ]]; then
        echo "❌ Warning: Backup file is empty!"
        exit 1
    fi
else
    echo "❌ No backups found in $BACKUP_DIR"
    exit 1
fi
```

## Устранение неполадок

### Проблема: Контейнер не найден

```bash
# Проверить запущенные контейнеры
docker ps

# Найти PostgreSQL контейнеры
docker ps | grep postgres

# Проверить имя контейнера в docker-compose
docker-compose ps
```

### Проблема: Ошибка доступа

```bash
# Проверить переменные окружения в контейнере
docker exec your_postgres_container env | grep PG

# Проверить пользователей базы данных
docker exec your_postgres_container psql -U postgres -c "\du"

# Проверить базы данных
docker exec your_postgres_container psql -U postgres -l
```

### Проблема: Недостаточно места

```bash
# Проверить место на диске
df -h

# Проверить размер базы данных
docker exec your_postgres_container psql -U postgres -d your_db -c "SELECT pg_size_pretty(pg_database_size('your_db'));"

# Использовать сжатие
echo "COMPRESS_BACKUP=true" >> .env.backup
```