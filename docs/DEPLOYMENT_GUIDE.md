# Hướng dẫn Deploy Map Memories API lên aaPanel

## Tổng quan

Hướng dẫn này sẽ giúp bạn đóng gói ứng dụng Map Memories API thành Docker images và deploy lên aaPanel một cách an toàn và hiệu quả.

## Mục lục

1. [Chuẩn bị môi trường](#chuẩn-bị-môi-trường)
2. [Đóng gói Docker Images](#đóng-gói-docker-images)
3. [Push Images lên Registry](#push-images-lên-registry)
4. [Cấu hình aaPanel](#cấu-hình-aapanel)
5. [Deploy ứng dụng](#deploy-ứng-dụng)
6. [Monitoring và Logs](#monitoring-và-logs)
7. [Backup và Recovery](#backup-và-recovery)
8. [Troubleshooting](#troubleshooting)

---

## 1. Chuẩn bị môi trường

### 1.1 Yêu cầu hệ thống

```bash
# Kiểm tra Docker
docker --version
docker-compose --version

# Kiểm tra Git
git --version

# Kiểm tra disk space
df -h
```

### 1.2 Cấu trúc thư mục

```
map-memories-api/
├── Dockerfile
├── docker-compose.yml
├── docker-compose.prod.yml
├── .env.production
├── nginx.conf
├── ssl/
├── docs/
└── scripts/
    ├── build.sh
    ├── deploy.sh
    └── backup.sh
```

### 1.3 Tạo file môi trường production

```bash
# Tạo file .env.production
cat > .env.production << EOF
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=mm_user
DB_PASSWORD=your_secure_password_here
DB_NAME=map_memories

# JWT Configuration
JWT_SECRET=your_super_secret_jwt_key_change_this_in_production
JWT_EXPIRY=86400

# Server Configuration
PORT=8080
ENV=production

# File Upload
UPLOAD_PATH=/app/uploads
MAX_FILE_SIZE=50MB

# Redis (optional)
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=

# External Services
EXTERNAL_URL=https://your-domain.com
EOF
```

---

## 2. Đóng gói Docker Images

### 2.1 Tạo Dockerfile tối ưu

```dockerfile
# Multi-stage build cho production
FROM golang:1.23-alpine AS builder

# Cài đặt dependencies
RUN apk add --no-cache git curl

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install swag for documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
RUN swag init -g main.go

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

# Install ca-certificates và curl
RUN apk --no-cache add ca-certificates curl

# Create app user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary từ builder stage
COPY --from=builder /app/main .

# Copy swagger docs
COPY --from=builder /app/docs ./docs

# Create uploads directory
RUN mkdir -p /app/uploads && chown appuser:appuser /app/uploads

# Switch to app user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run application
CMD ["./main"]
```

### 2.2 Tạo docker-compose.prod.yml

```yaml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: mm_postgres_prod
    restart: unless-stopped
    environment:
      POSTGRES_DB: map_memories
      POSTGRES_USER: mm_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      PGDATA: /var/lib/postgresql/data/pgdata
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./database/migrations:/docker-entrypoint-initdb.d
    networks:
      - mm_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U mm_user -d map_memories"]
      interval: 30s
      timeout: 10s
      retries: 5

  # Go API Application
  api:
    build:
      context: .
      dockerfile: Dockerfile
    image: map-memories-api:latest
    container_name: mm_api_prod
    restart: unless-stopped
    environment:
      - ENV=production
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=mm_user
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=map_memories
      - JWT_SECRET=${JWT_SECRET}
      - JWT_EXPIRY=86400
      - PORT=8080
      - UPLOAD_PATH=/app/uploads
      - MAX_FILE_SIZE=50MB
    volumes:
      - uploads_data:/app/uploads
    networks:
      - mm_network
    depends_on:
      postgres:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Redis for caching (optional)
  redis:
    image: redis:7-alpine
    container_name: mm_redis_prod
    restart: unless-stopped
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    networks:
      - mm_network

  # Nginx reverse proxy
  nginx:
    image: nginx:alpine
    container_name: mm_nginx_prod
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - uploads_data:/var/www/uploads:ro
    networks:
      - mm_network
    depends_on:
      - api

volumes:
  postgres_data:
    driver: local
  uploads_data:
    driver: local
  redis_data:
    driver: local

networks:
  mm_network:
    driver: bridge
```

### 2.3 Tạo script build

```bash
#!/bin/bash
# scripts/build.sh

set -e

echo "🚀 Building Map Memories API Docker images..."

# Load environment variables
if [ -f .env.production ]; then
    export $(cat .env.production | grep -v '^#' | xargs)
fi

# Build API image
echo "📦 Building API image..."
docker build -t map-memories-api:latest .

# Build with specific tag
TAG=${1:-latest}
docker tag map-memories-api:latest map-memories-api:$TAG

echo "✅ Build completed successfully!"
echo "📋 Images created:"
docker images | grep map-memories-api
```

### 2.4 Tạo script deploy

```bash
#!/bin/bash
# scripts/deploy.sh

set -e

echo "🚀 Deploying Map Memories API..."

# Load environment variables
if [ -f .env.production ]; then
    export $(cat .env.production | grep -v '^#' | xargs)
fi

# Stop existing containers
echo "🛑 Stopping existing containers..."
docker-compose -f docker-compose.prod.yml down

# Pull latest images (if using registry)
# docker pull your-registry/map-memories-api:latest

# Start services
echo "▶️ Starting services..."
docker-compose -f docker-compose.prod.yml up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 30

# Check health
echo "🔍 Checking service health..."
docker-compose -f docker-compose.prod.yml ps

echo "✅ Deployment completed successfully!"
echo "🌐 API URL: http://your-domain.com"
echo "📊 Health check: http://your-domain.com/health"
```

---

## 3. Push Images lên Registry

### 3.1 Sử dụng Docker Hub

```bash
# Login to Docker Hub
docker login

# Tag image
docker tag map-memories-api:latest your-username/map-memories-api:latest

# Push to Docker Hub
docker push your-username/map-memories-api:latest
```

### 3.2 Sử dụng Private Registry

```bash
# Login to private registry
docker login your-registry.com

# Tag image
docker tag map-memories-api:latest your-registry.com/map-memories-api:latest

# Push to registry
docker push your-registry.com/map-memories-api:latest
```

### 3.3 Tạo script push

```bash
#!/bin/bash
# scripts/push.sh

set -e

REGISTRY=${1:-"your-registry.com"}
TAG=${2:-"latest"}

echo "📤 Pushing images to $REGISTRY..."

# Tag images
docker tag map-memories-api:latest $REGISTRY/map-memories-api:$TAG

# Push images
docker push $REGISTRY/map-memories-api:$TAG

echo "✅ Images pushed successfully!"
```

---

## 4. Cấu hình aaPanel

### 4.1 Cài đặt aaPanel

```bash
# Download và cài đặt aaPanel
wget -O install.sh http://www.aapanel.com/script/install_6.0_en.sh && sudo bash install.sh

# Truy cập panel
# http://your-server-ip:8888
```

### 4.2 Cài đặt Docker trong aaPanel

```bash
# SSH vào server
ssh root@your-server-ip

# Cài đặt Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Cài đặt Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Thêm user vào docker group
sudo usermod -aG docker $USER
```

### 4.3 Cấu hình Firewall

```bash
# Mở ports cần thiết
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 8888/tcp
sudo ufw enable
```

---

## 5. Deploy ứng dụng

### 5.1 Chuẩn bị server

```bash
# Tạo thư mục project
mkdir -p /www/wwwroot/map-memories-api
cd /www/wwwroot/map-memories-api

# Clone repository
git clone https://github.com/your-username/map-memories-api.git .

# Copy environment file
cp .env.example .env.production

# Edit environment variables
nano .env.production
```

### 5.2 Cấu hình SSL Certificate

```bash
# Tạo thư mục SSL
mkdir -p ssl

# Copy certificates từ aaPanel
cp /www/server/panel/vhost/cert/your-domain.com/* ssl/

# Hoặc tạo self-signed certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/nginx.key \
  -out ssl/nginx.crt \
  -subj "/C=VN/ST=HCM/L=Ho Chi Minh City/O=Your Company/CN=your-domain.com"
```

### 5.3 Cấu hình Nginx

```nginx
# nginx.conf
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

    # Upstream API
    upstream api_backend {
        server api:8080;
    }

    # HTTP to HTTPS redirect
    server {
        listen 80;
        server_name your-domain.com;
        return 301 https://$server_name$request_uri;
    }

    # HTTPS server
    server {
        listen 443 ssl http2;
        server_name your-domain.com;

        # SSL configuration
        ssl_certificate /etc/nginx/ssl/nginx.crt;
        ssl_certificate_key /etc/nginx/ssl/nginx.key;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

        # API endpoints
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            
            proxy_pass http://api_backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # Timeouts
            proxy_connect_timeout 30s;
            proxy_send_timeout 30s;
            proxy_read_timeout 30s;
        }

        # Health check
        location /health {
            proxy_pass http://api_backend/health;
            proxy_set_header Host $host;
        }

        # Swagger documentation
        location /swagger/ {
            proxy_pass http://api_backend/swagger/;
            proxy_set_header Host $host;
        }

        # Static files (uploads)
        location /uploads/ {
            alias /var/www/uploads/;
            expires 1y;
            add_header Cache-Control "public, immutable";
        }

        # Root redirect to swagger
        location = / {
            return 301 /swagger/index.html;
        }
    }
}
```

### 5.4 Deploy ứng dụng

```bash
# Build images
./scripts/build.sh

# Deploy
./scripts/deploy.sh

# Kiểm tra status
docker-compose -f docker-compose.prod.yml ps
```

---

## 6. Monitoring và Logs

### 6.1 Cấu hình logging

```bash
# Tạo thư mục logs
mkdir -p logs

# Cấu hình log rotation
cat > /etc/logrotate.d/map-memories << EOF
/www/wwwroot/map-memories-api/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 root root
    postrotate
        docker-compose -f /www/wwwroot/map-memories-api/docker-compose.prod.yml restart api
    endscript
}
EOF
```

### 6.2 Monitoring script

```bash
#!/bin/bash
# scripts/monitor.sh

echo "🔍 Monitoring Map Memories API..."

# Check container status
echo "📊 Container Status:"
docker-compose -f docker-compose.prod.yml ps

# Check resource usage
echo "💾 Resource Usage:"
docker stats --no-stream

# Check logs
echo "📝 Recent Logs:"
docker-compose -f docker-compose.prod.yml logs --tail=20 api

# Health check
echo "🏥 Health Check:"
curl -f http://localhost/health || echo "❌ Health check failed"
```

### 6.3 Cấu hình alerts

```bash
# Tạo script alert
cat > scripts/alert.sh << 'EOF'
#!/bin/bash

# Alert configuration
WEBHOOK_URL="your-webhook-url"
SERVICE_NAME="Map Memories API"

# Check service health
if ! curl -f http://localhost/health > /dev/null 2>&1; then
    curl -X POST $WEBHOOK_URL \
        -H "Content-Type: application/json" \
        -d "{\"text\":\"🚨 $SERVICE_NAME is down!\"}"
fi
EOF

chmod +x scripts/alert.sh

# Add to crontab
echo "*/5 * * * * /www/wwwroot/map-memories-api/scripts/alert.sh" | crontab -
```

---

## 7. Backup và Recovery

### 7.1 Backup script

```bash
#!/bin/bash
# scripts/backup.sh

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
```

### 7.2 Recovery script

```bash
#!/bin/bash
# scripts/restore.sh

set -e

BACKUP_FILE=$1
BACKUP_DIR="/www/backup/map-memories"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

echo "🔄 Restoring from backup: $BACKUP_FILE"

# Stop services
docker-compose -f docker-compose.prod.yml down

# Restore database
echo "📊 Restoring database..."
docker-compose -f docker-compose.prod.yml up -d postgres
sleep 10
docker-compose -f docker-compose.prod.yml exec -T postgres \
    psql -U mm_user -d map_memories < $BACKUP_DIR/$BACKUP_FILE

# Restore uploads
echo "📁 Restoring uploads..."
tar -xzf $BACKUP_DIR/uploads_backup_*.tar.gz -C /www/wwwroot/map-memories-api/

# Start services
echo "▶️ Starting services..."
docker-compose -f docker-compose.prod.yml up -d

echo "✅ Restore completed!"
```

---

## 8. Troubleshooting

### 8.1 Common Issues

#### Container không start
```bash
# Kiểm tra logs
docker-compose -f docker-compose.prod.yml logs api

# Kiểm tra port conflicts
netstat -tulpn | grep :8080

# Restart container
docker-compose -f docker-compose.prod.yml restart api
```

#### Database connection failed
```bash
# Kiểm tra database
docker-compose -f docker-compose.prod.yml exec postgres psql -U mm_user -d map_memories -c "SELECT 1;"

# Restart database
docker-compose -f docker-compose.prod.yml restart postgres
```

#### SSL certificate issues
```bash
# Kiểm tra certificate
openssl x509 -in ssl/nginx.crt -text -noout

# Regenerate certificate
./scripts/generate-ssl.sh
```

### 8.2 Performance tuning

```bash
# Tăng memory limit cho containers
docker-compose -f docker-compose.prod.yml down
docker system prune -f
docker-compose -f docker-compose.prod.yml up -d

# Optimize database
docker-compose -f docker-compose.prod.yml exec postgres \
    psql -U mm_user -d map_memories -c "VACUUM ANALYZE;"
```

### 8.3 Security checklist

- [ ] Change default passwords
- [ ] Enable firewall
- [ ] Configure SSL certificates
- [ ] Set up monitoring
- [ ] Regular backups
- [ ] Update dependencies
- [ ] Configure rate limiting
- [ ] Enable security headers

---

## 9. Maintenance

### 9.1 Update application

```bash
#!/bin/bash
# scripts/update.sh

set -e

echo "🔄 Updating Map Memories API..."

# Backup before update
./scripts/backup.sh

# Pull latest code
git pull origin main

# Rebuild images
./scripts/build.sh

# Deploy
./scripts/deploy.sh

echo "✅ Update completed!"
```

### 9.2 Cleanup script

```bash
#!/bin/bash
# scripts/cleanup.sh

echo "🧹 Cleaning up..."

# Remove unused images
docker image prune -f

# Remove unused volumes
docker volume prune -f

# Remove unused networks
docker network prune -f

# Clean logs
find /www/wwwroot/map-memories-api/logs -name "*.log" -mtime +30 -delete

echo "✅ Cleanup completed!"
```

---

## 10. Kết luận

Với hướng dẫn này, bạn đã có thể:

1. ✅ Đóng gói ứng dụng thành Docker images
2. ✅ Deploy lên aaPanel an toàn
3. ✅ Cấu hình SSL và security
4. ✅ Monitoring và logging
5. ✅ Backup và recovery
6. ✅ Maintenance và updates

### Liên hệ hỗ trợ

Nếu gặp vấn đề, hãy kiểm tra:
- Logs: `docker-compose -f docker-compose.prod.yml logs`
- Health check: `curl http://your-domain.com/health`
- Container status: `docker-compose -f docker-compose.prod.yml ps`

---

**Lưu ý**: Đảm bảo thay đổi tất cả placeholder values (your-domain.com, your-registry.com, etc.) thành giá trị thực tế của bạn trước khi deploy. 