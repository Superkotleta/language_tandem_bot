# Backup & Restore Guide

Руководство по резервному копированию и восстановлению Language Exchange Bot.

## Обзор стратегии бэкапов

### Типы данных для бэкапа

1. **База данных PostgreSQL** - основная бизнес-логика
2. **Redis кэш** - производительность (опционально)
3. **Конфигурация** - настройки приложения
4. **Логи** - для анализа проблем
5. **Код приложения** - для быстрого восстановления

### Стратегия бэкапов

- **Полные бэкапы**: Ежедневно в 2:00 AM
- **Инкрементальные**: Каждые 6 часов
- **Конфигурация**: При каждом изменении
- **Логи**: Еженедельно (архивирование)

## Бэкап базы данных

### 1. Автоматический бэкап

Создайте скрипт `/opt/language-exchange-bot/backup-database.sh`:

```bash
#!/bin/bash

# Конфигурация
DB_NAME="language_exchange_bot"
DB_USER="bot_user"
DB_HOST="localhost"
DB_PORT="5432"
BACKUP_DIR="/opt/backups/database"
RETENTION_DAYS=30
DATE=$(date +%Y%m%d_%H%M%S)

# Создайте директорию для бэкапов
mkdir -p $BACKUP_DIR

# Полный бэкап базы данных
echo "Starting database backup at $(date)"
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
    --verbose \
    --clean \
    --if-exists \
    --create \
    --format=custom \
    --file="$BACKUP_DIR/full_backup_$DATE.dump"

# Проверьте успешность бэкапа
if [ $? -eq 0 ]; then
    echo "Database backup completed successfully"
    
    # Сожмите бэкап
    gzip "$BACKUP_DIR/full_backup_$DATE.dump"
    
    # Удалите старые бэкапы
    find $BACKUP_DIR -name "full_backup_*.dump.gz" -mtime +$RETENTION_DAYS -delete
    
    # Логирование
    echo "$(date): Database backup completed - full_backup_$DATE.dump.gz" >> /var/log/backup.log
else
    echo "Database backup failed"
    echo "$(date): Database backup failed" >> /var/log/backup.log
    exit 1
fi
```

### 2. Инкрементальный бэкап

Создайте скрипт `/opt/language-exchange-bot/backup-incremental.sh`:

```bash
#!/bin/bash

# Конфигурация
DB_NAME="language_exchange_bot"
DB_USER="bot_user"
DB_HOST="localhost"
DB_PORT="5432"
BACKUP_DIR="/opt/backups/database/incremental"
DATE=$(date +%Y%m%d_%H%M%S)

# Создайте директорию
mkdir -p $BACKUP_DIR

# Инкрементальный бэкап (только изменения за последние 6 часов)
pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME \
    --verbose \
    --format=custom \
    --file="$BACKUP_DIR/incremental_backup_$DATE.dump" \
    --where="updated_at > NOW() - INTERVAL '6 hours'"

# Сожмите бэкап
gzip "$BACKUP_DIR/incremental_backup_$DATE.dump"

echo "$(date): Incremental backup completed - incremental_backup_$DATE.dump.gz" >> /var/log/backup.log
```

### 3. Настройка cron для автоматических бэкапов

```bash
# Добавьте в crontab
sudo crontab -e

# Полные бэкапы каждый день в 2:00 AM
0 2 * * * /opt/language-exchange-bot/backup-database.sh

# Инкрементальные бэкапы каждые 6 часов
0 */6 * * * /opt/language-exchange-bot/backup-incremental.sh

# Очистка старых логов каждую неделю
0 3 * * 0 find /var/log -name "*.log" -mtime +30 -delete
```

## Бэкап Redis

### 1. Автоматический бэкап Redis

Создайте скрипт `/opt/language-exchange-bot/backup-redis.sh`:

```bash
#!/bin/bash

# Конфигурация
REDIS_HOST="localhost"
REDIS_PORT="6379"
BACKUP_DIR="/opt/backups/redis"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=7

# Создайте директорию
mkdir -p $BACKUP_DIR

# Бэкап Redis
echo "Starting Redis backup at $(date)"
redis-cli -h $REDIS_HOST -p $REDIS_PORT --rdb "$BACKUP_DIR/redis_backup_$DATE.rdb"

# Проверьте успешность
if [ $? -eq 0 ]; then
    echo "Redis backup completed successfully"
    
    # Сожмите бэкап
    gzip "$BACKUP_DIR/redis_backup_$DATE.rdb"
    
    # Удалите старые бэкапы
    find $BACKUP_DIR -name "redis_backup_*.rdb.gz" -mtime +$RETENTION_DAYS -delete
    
    echo "$(date): Redis backup completed - redis_backup_$DATE.rdb.gz" >> /var/log/backup.log
else
    echo "Redis backup failed"
    echo "$(date): Redis backup failed" >> /var/log/backup.log
    exit 1
fi
```

### 2. Настройка Redis для бэкапов

```bash
# В /etc/redis/redis.conf
save 900 1      # Сохранять если изменился хотя бы 1 ключ за 900 секунд
save 300 10     # Сохранять если изменилось 10 ключей за 300 секунд
save 60 10000   # Сохранять если изменилось 10000 ключей за 60 секунд

# Включите AOF (Append Only File)
appendonly yes
appendfsync everysec
```

## Бэкап конфигурации

### 1. Скрипт бэкапа конфигурации

Создайте `/opt/language-exchange-bot/backup-config.sh`:

```bash
#!/bin/bash

# Конфигурация
CONFIG_DIR="/opt/language-exchange-bot"
BACKUP_DIR="/opt/backups/config"
DATE=$(date +%Y%m%d_%H%M%S)

# Создайте директорию
mkdir -p $BACKUP_DIR

# Бэкап конфигурации
echo "Starting configuration backup at $(date)"
tar -czf "$BACKUP_DIR/config_backup_$DATE.tar.gz" \
    -C $CONFIG_DIR \
    .env \
    migrations/ \
    docker-compose.yml \
    Dockerfile

# Бэкап systemd сервисов
sudo cp /etc/systemd/system/language-exchange-bot.service "$BACKUP_DIR/"

# Бэкап nginx конфигурации (если используется)
if [ -f /etc/nginx/sites-available/language-exchange-bot ]; then
    sudo cp /etc/nginx/sites-available/language-exchange-bot "$BACKUP_DIR/"
fi

echo "$(date): Configuration backup completed - config_backup_$DATE.tar.gz" >> /var/log/backup.log
```

## Восстановление

### 1. Восстановление базы данных

#### Полное восстановление

```bash
#!/bin/bash
# restore-database.sh

BACKUP_FILE="$1"
DB_NAME="language_exchange_bot"
DB_USER="bot_user"
DB_HOST="localhost"
DB_PORT="5432"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

# Остановите приложение
sudo systemctl stop language-exchange-bot

# Удалите существующую базу данных
sudo -u postgres dropdb $DB_NAME

# Создайте новую базу данных
sudo -u postgres createdb $DB_NAME

# Восстановите из бэкапа
if [[ $BACKUP_FILE == *.gz ]]; then
    gunzip -c $BACKUP_FILE | pg_restore -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME --verbose
else
    pg_restore -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME --verbose $BACKUP_FILE
fi

# Проверьте восстановление
if [ $? -eq 0 ]; then
    echo "Database restored successfully"
    
    # Запустите приложение
    sudo systemctl start language-exchange-bot
    
    echo "$(date): Database restored from $BACKUP_FILE" >> /var/log/restore.log
else
    echo "Database restore failed"
    exit 1
fi
```

#### Восстановление из SQL дампа

```bash
#!/bin/bash
# restore-from-sql.sh

SQL_FILE="$1"
DB_NAME="language_exchange_bot"

if [ -z "$SQL_FILE" ]; then
    echo "Usage: $0 <sql_file>"
    exit 1
fi

# Остановите приложение
sudo systemctl stop language-exchange-bot

# Восстановите из SQL
if [[ $SQL_FILE == *.gz ]]; then
    gunzip -c $SQL_FILE | sudo -u postgres psql $DB_NAME
else
    sudo -u postgres psql $DB_NAME < $SQL_FILE
fi

# Запустите приложение
sudo systemctl start language-exchange-bot
```

### 2. Восстановление Redis

```bash
#!/bin/bash
# restore-redis.sh

REDIS_BACKUP="$1"
REDIS_HOST="localhost"
REDIS_PORT="6379"

if [ -z "$REDIS_BACKUP" ]; then
    echo "Usage: $0 <redis_backup_file>"
    exit 1
fi

# Остановите Redis
sudo systemctl stop redis

# Восстановите данные
if [[ $REDIS_BACKUP == *.gz ]]; then
    gunzip -c $REDIS_BACKUP > /var/lib/redis/dump.rdb
else
    cp $REDIS_BACKUP /var/lib/redis/dump.rdb
fi

# Установите правильные права
sudo chown redis:redis /var/lib/redis/dump.rdb
sudo chmod 660 /var/lib/redis/dump.rdb

# Запустите Redis
sudo systemctl start redis

echo "$(date): Redis restored from $REDIS_BACKUP" >> /var/log/restore.log
```

### 3. Восстановление конфигурации

```bash
#!/bin/bash
# restore-config.sh

CONFIG_BACKUP="$1"
CONFIG_DIR="/opt/language-exchange-bot"

if [ -z "$CONFIG_BACKUP" ]; then
    echo "Usage: $0 <config_backup_file>"
    exit 1
fi

# Остановите приложение
sudo systemctl stop language-exchange-bot

# Восстановите конфигурацию
tar -xzf $CONFIG_BACKUP -C $CONFIG_DIR

# Восстановите systemd сервис
sudo cp $CONFIG_DIR/language-exchange-bot.service /etc/systemd/system/
sudo systemctl daemon-reload

# Восстановите nginx конфигурацию (если есть)
if [ -f "$CONFIG_DIR/language-exchange-bot" ]; then
    sudo cp $CONFIG_DIR/language-exchange-bot /etc/nginx/sites-available/
    sudo nginx -t && sudo systemctl reload nginx
fi

# Запустите приложение
sudo systemctl start language-exchange-bot

echo "$(date): Configuration restored from $CONFIG_BACKUP" >> /var/log/restore.log
```

## Тестирование бэкапов

### 1. Скрипт проверки бэкапов

Создайте `/opt/language-exchange-bot/verify-backup.sh`:

```bash
#!/bin/bash

BACKUP_FILE="$1"
BACKUP_TYPE="$2"

if [ -z "$BACKUP_FILE" ] || [ -z "$BACKUP_TYPE" ]; then
    echo "Usage: $0 <backup_file> <backup_type>"
    echo "Backup types: database, redis, config"
    exit 1
fi

case $BACKUP_TYPE in
    "database")
        echo "Verifying database backup..."
        if [[ $BACKUP_FILE == *.gz ]]; then
            gunzip -t $BACKUP_FILE
        else
            pg_restore --list $BACKUP_FILE > /dev/null
        fi
        ;;
    "redis")
        echo "Verifying Redis backup..."
        if [[ $BACKUP_FILE == *.gz ]]; then
            gunzip -t $BACKUP_FILE
        else
            # Проверьте, что файл является валидным RDB файлом
            file $BACKUP_FILE | grep -q "Redis"
        fi
        ;;
    "config")
        echo "Verifying configuration backup..."
        tar -tzf $BACKUP_FILE > /dev/null
        ;;
    *)
        echo "Unknown backup type: $BACKUP_TYPE"
        exit 1
        ;;
esac

if [ $? -eq 0 ]; then
    echo "Backup verification successful"
    exit 0
else
    echo "Backup verification failed"
    exit 1
fi
```

### 2. Автоматическая проверка бэкапов

```bash
# Добавьте в crontab для проверки бэкапов
0 3 * * * /opt/language-exchange-bot/verify-backup.sh /opt/backups/database/full_backup_$(date +\%Y\%m\%d)_020000.dump.gz database
```

## Мониторинг бэкапов

### 1. Скрипт мониторинга

Создайте `/opt/language-exchange-bot/monitor-backups.sh`:

```bash
#!/bin/bash

BACKUP_DIR="/opt/backups"
LOG_FILE="/var/log/backup-monitor.log"
ALERT_EMAIL="admin@yourdomain.com"

# Проверьте последние бэкапы
LATEST_DB_BACKUP=$(find $BACKUP_DIR/database -name "full_backup_*.dump.gz" -type f -mtime -1 | sort | tail -1)
LATEST_REDIS_BACKUP=$(find $BACKUP_DIR/redis -name "redis_backup_*.rdb.gz" -type f -mtime -1 | sort | tail -1)
LATEST_CONFIG_BACKUP=$(find $BACKUP_DIR/config -name "config_backup_*.tar.gz" -type f -mtime -1 | sort | tail -1)

# Проверьте размеры бэкапов
check_backup_size() {
    local backup_file="$1"
    local min_size="$2"  # Минимальный размер в MB
    
    if [ -f "$backup_file" ]; then
        local size_mb=$(du -m "$backup_file" | cut -f1)
        if [ $size_mb -lt $min_size ]; then
            echo "$(date): WARNING - Backup $backup_file is too small ($size_mb MB)" >> $LOG_FILE
            return 1
        fi
    else
        echo "$(date): ERROR - Backup $backup_file not found" >> $LOG_FILE
        return 1
    fi
    return 0
}

# Проверьте бэкапы
check_backup_size "$LATEST_DB_BACKUP" 10
check_backup_size "$LATEST_REDIS_BACKUP" 1
check_backup_size "$LATEST_CONFIG_BACKUP" 1

# Отправьте алерт если есть проблемы
if [ $? -ne 0 ]; then
    echo "Backup monitoring detected issues" | mail -s "Backup Alert" $ALERT_EMAIL
fi
```

### 2. Настройка мониторинга

```bash
# Добавьте в crontab
0 4 * * * /opt/language-exchange-bot/monitor-backups.sh
```

## Аварийное восстановление

### 1. Полное восстановление системы

```bash
#!/bin/bash
# disaster-recovery.sh

# Восстановите систему из бэкапа
# 1. Восстановите ОС и зависимости
# 2. Восстановите конфигурацию
# 3. Восстановите базу данных
# 4. Восстановите Redis
# 5. Запустите приложение

echo "Starting disaster recovery..."

# Восстановите конфигурацию
/opt/language-exchange-bot/restore-config.sh /opt/backups/config/latest_config.tar.gz

# Восстановите базу данных
/opt/language-exchange-bot/restore-database.sh /opt/backups/database/latest_full.dump.gz

# Восстановите Redis
/opt/language-exchange-bot/restore-redis.sh /opt/backups/redis/latest_redis.rdb.gz

# Запустите все сервисы
sudo systemctl start postgresql
sudo systemctl start redis
sudo systemctl start language-exchange-bot

echo "Disaster recovery completed"
```

### 2. Проверка после восстановления

```bash
#!/bin/bash
# post-recovery-check.sh

echo "Running post-recovery checks..."

# Проверьте статус сервисов
sudo systemctl status language-exchange-bot
sudo systemctl status postgresql
sudo systemctl status redis

# Проверьте подключение к базе данных
pg_isready -h localhost -p 5432

# Проверьте Redis
redis-cli ping

# Проверьте health endpoint
curl -f http://localhost:8080/health

# Проверьте API
curl -H "Authorization: Bearer $API_TOKEN" http://localhost:8080/api/v1/stats

echo "Post-recovery checks completed"
```

## Лучшие практики

### 1. Стратегия бэкапов

- **3-2-1 правило**: 3 копии, 2 разных носителя, 1 оффлайн
- **Тестирование**: Регулярно тестируйте восстановление
- **Мониторинг**: Настройте алерты на проблемы с бэкапами
- **Документация**: Ведите документацию по процедурам

### 2. Безопасность

```bash
# Зашифруйте бэкапы
gpg --symmetric --cipher-algo AES256 backup.dump

# Ограничьте доступ к бэкапам
chmod 600 /opt/backups/database/*.dump.gz
chown backup:backup /opt/backups/database/*.dump.gz
```

### 3. Оптимизация

```bash
# Используйте параллельные бэкапы
pg_dump -j 4 -Fd -f backup_dir database_name

# Сжимайте бэкапы
pg_dump | gzip > backup.sql.gz

# Используйте инкрементальные бэкапы
pg_dump --schema-only > schema.sql
pg_dump --data-only > data.sql
```
