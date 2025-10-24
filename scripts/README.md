# PostgreSQL Backup and Restore Scripts

Эти скрипты позволяют создавать бэкапы и восстанавливать PostgreSQL базы данных из Docker контейнеров с использованием переменных окружения из .env файлов.

## Файлы

- `backup_postgres.sh` - Скрипт для создания бэкапов
- `restore_postgres.sh` - Скрипт для восстановления из бэкапов
- `../env.backup.example` - Пример конфигурационного файла

## Быстрый старт

### 1. Настройка конфигурации

Скопируйте пример конфигурации:
```bash
cp .env.backup.example .env.backup
```

Отредактируйте `.env.backup` под ваши настройки:
```bash
DB_HOST=postgres
DB_USER=postgres
DB_NAME=product_requirements
DB_PORT=5432
CONTAINER_NAME=postgres
BACKUP_DIR=./backups
BACKUP_FORMAT=sql
COMPRESS_BACKUP=true
BACKUP_RETENTION_DAYS=7
```

### 2. Создание бэкапа

```bash
# Использовать .env файл по умолчанию
./scripts/backup_postgres.sh

# Использовать конкретный .env файл
./scripts/backup_postgres.sh .env.backup

# Посмотреть справку
./scripts/backup_postgres.sh --help
```

### 3. Восстановление из бэкапа

```bash
# Восстановить из SQL файла
./scripts/restore_postgres.sh backups/backup_postgres_mydatabase_20231023_143022.sql

# Восстановить из сжатого файла
./scripts/restore_postgres.sh backups/backup_postgres_mydatabase_20231023_143022.sql.gz

# Восстановить из custom формата
./scripts/restore_postgres.sh backups/backup_postgres_mydatabase_20231023_143022.dump

# Использовать конкретный .env файл
./scripts/restore_postgres.sh backup.sql .env.backup

# Посмотреть справку
./scripts/restore_postgres.sh --help
```

## Конфигурация

### Обязательные переменные

| Переменная | Описание |
|------------|----------|
| `DB_HOST` | Хост базы данных (имя контейнера или localhost) |
| `DB_USER` | Имя пользователя базы данных |
| `DB_NAME` | Имя базы данных |
| `DB_PASSWORD` | Пароль базы данных (если требуется аутентификация) |

### Опциональные переменные

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `DB_PORT` | `5432` | Порт базы данных |
| `DB_PASSWORD` | - | Пароль базы данных (если требуется) |
| `CONTAINER_NAME` | автоопределение | Конкретное имя контейнера |
| `BACKUP_DIR` | `./backups` | Директория для бэкапов |
| `BACKUP_FORMAT` | `sql` | Формат бэкапа (sql, custom, directory) |
| `COMPRESS_BACKUP` | `true` | Сжимать SQL бэкапы |
| `BACKUP_RETENTION_DAYS` | `7` | Дни хранения старых бэкапов (0 = отключить) |
| `FORCE_RESTORE` | `false` | Пропустить подтверждение при восстановлении |

## Форматы бэкапов

### SQL формат (`BACKUP_FORMAT=sql`)
- Обычный SQL дамп
- Может быть сжат с помощью gzip
- Легко читается и редактируется
- Подходит для небольших и средних баз данных

### Custom формат (`BACKUP_FORMAT=custom`)
- Сжатый бинарный формат PostgreSQL
- Поддерживает параллельное восстановление
- Более эффективен для больших баз данных
- Позволяет выборочное восстановление

### Directory формат (`BACKUP_FORMAT=directory`)
- Создает директорию с файлами
- Поддерживает параллельное создание и восстановление
- Лучший выбор для очень больших баз данных

## Автоматизация

### Добавить в crontab для автоматических бэкапов

```bash
# Редактировать crontab
crontab -e

# Добавить строку для ежедневного бэкапа в 2:00 AM
0 2 * * * /path/to/your/project/scripts/backup_postgres.sh /path/to/your/project/.env.backup >> /var/log/postgres_backup.log 2>&1
```

### Использование в CI/CD

```yaml
# GitHub Actions пример
- name: Backup Database
  run: |
    ./scripts/backup_postgres.sh .env.production
    
- name: Upload Backup
  uses: actions/upload-artifact@v3
  with:
    name: database-backup
    path: backups/
```

## Безопасность

### Пароли

Если требуется пароль для подключения к базе данных:

1. **DB_PASSWORD переменная** (рекомендуется):
   ```bash
   # В .env файле
   DB_PASSWORD=your_secure_password
   ```

2. **Файл .pgpass** (альтернативный способ):
   ```bash
   # Создать файл ~/.pgpass
   echo "localhost:5432:mydatabase:myuser:mypassword" >> ~/.pgpass
   chmod 600 ~/.pgpass
   ```

3. **PGPASSWORD переменная** (устаревший способ):
   ```bash
   PGPASSWORD=your_password
   ```

4. **Docker secrets** (для Docker Swarm):
   ```bash
   echo "mypassword" | docker secret create db_password -
   ```

### Права доступа

```bash
# Установить правильные права на скрипты
chmod 700 scripts/backup_postgres.sh scripts/restore_postgres.sh

# Установить права на .env файлы
chmod 600 .env.backup
```

## Устранение неполадок

### Контейнер не найден
```bash
# Проверить запущенные контейнеры
docker ps

# Проверить все контейнеры
docker ps -a

# Проверить логи контейнера
docker logs your_postgres_container
```

### Ошибки подключения
```bash
# Проверить переменные окружения
docker exec your_postgres_container env | grep PG

# Проверить доступность базы данных
docker exec your_postgres_container psql -U postgres -l
```

### Проблемы с правами
```bash
# Проверить пользователя в контейнере
docker exec your_postgres_container whoami

# Проверить права на файлы
ls -la backups/
```

## Примеры использования

### Миграция между окружениями

```bash
# Создать бэкап на продакшене
./scripts/backup_postgres.sh .env.production

# Восстановить на тестовом окружении
./scripts/restore_postgres.sh backups/backup_prod_mydb_20231023_143022.sql.gz .env.staging
```

### Бэкап перед обновлением

```bash
# Создать бэкап перед миграцией
./scripts/backup_postgres.sh

# Выполнить миграцию
make migrate-up

# В случае проблем - восстановить
./scripts/restore_postgres.sh backups/latest_backup.sql
```

### Регулярные бэкапы с ротацией

```bash
# В .env.backup установить
BACKUP_RETENTION_DAYS=30

# Скрипт будет автоматически удалять бэкапы старше 30 дней
./scripts/backup_postgres.sh .env.backup
```