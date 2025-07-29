#!/bin/bash

set -e

BACKUP_DIR="/www/backup/map-memories"
DATE=$(date +%Y%m%d_%H%M%S)

echo "💾 Creating backup..."

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
echo "📊 Backing up database..."
docker-compose -f docker-compose.prod.yml exec -T postgres \
    pg_dump -U mm_user map_memories > $BACKUP_DIR/db_backup_$DATE.sql

# Backup uploads
echo "📁 Backing up uploads..."
tar -czf $BACKUP_DIR/uploads_backup_$DATE.tar.gz \
    -C /www/wwwroot/map-memories-api uploads/

# Backup configuration
echo "⚙️ Backing up configuration..."
tar -czf $BACKUP_DIR/config_backup_$DATE.tar.gz \
    -C /www/wwwroot/map-memories-api \
    .env.production docker-compose.prod.yml nginx.conf ssl/

# Clean old backups (keep last 7 days)
find $BACKUP_DIR -name "*.sql" -mtime +7 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete

echo "✅ Backup completed: $BACKUP_DIR" 