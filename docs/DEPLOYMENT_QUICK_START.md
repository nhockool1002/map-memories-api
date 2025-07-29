# Quick Start - Deploy Map Memories API

## 🚀 Deploy nhanh trong 5 phút

### 1. Chuẩn bị server

```bash
# SSH vào server
ssh root@your-server-ip

# Cài đặt Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Cài đặt Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### 2. Clone và cấu hình

```bash
# Tạo thư mục project
mkdir -p /www/wwwroot/map-memories-api
cd /www/wwwroot/map-memories-api

# Clone repository
git clone https://github.com/your-username/map-memories-api.git .

# Tạo file environment
cat > .env.production << EOF
DB_PASSWORD=your_secure_password_here
JWT_SECRET=your_super_secret_jwt_key_change_this_in_production
REDIS_PASSWORD=your_redis_password_here
EOF

# Tạo SSL certificate
mkdir -p ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/nginx.key \
  -out ssl/nginx.crt \
  -subj "/C=VN/ST=HCM/L=Ho Chi Minh City/O=Your Company/CN=your-domain.com"
```

### 3. Deploy

```bash
# Cấp quyền cho scripts
chmod +x scripts/*.sh

# Build và deploy
./scripts/build.sh
./scripts/deploy.sh

# Kiểm tra status
./scripts/monitor.sh
```

### 4. Kiểm tra

```bash
# Health check
curl http://your-domain.com/health

# API documentation
curl http://your-domain.com/swagger/index.html
```

## 📋 Checklist trước khi deploy

- [ ] Thay đổi `your-domain.com` thành domain thực tế
- [ ] Thay đổi passwords trong `.env.production`
- [ ] Cấu hình firewall (ports 80, 443, 8888)
- [ ] Backup dữ liệu hiện tại (nếu có)
- [ ] Kiểm tra disk space (> 10GB free)

## 🔧 Troubleshooting nhanh

### Container không start
```bash
# Kiểm tra logs
docker-compose -f docker-compose.prod.yml logs api

# Restart
docker-compose -f docker-compose.prod.yml restart api
```

### Database connection failed
```bash
# Kiểm tra database
docker-compose -f docker-compose.prod.yml exec postgres psql -U mm_user -d map_memories -c "SELECT 1;"
```

### SSL issues
```bash
# Regenerate certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/nginx.key \
  -out ssl/nginx.crt \
  -subj "/C=VN/ST=HCM/L=Ho Chi Minh City/O=Your Company/CN=your-domain.com"
```

## 📊 Monitoring

```bash
# Xem logs real-time
docker-compose -f docker-compose.prod.yml logs -f api

# Kiểm tra resource usage
docker stats

# Backup dữ liệu
./scripts/backup.sh
```

## 🔄 Update

```bash
# Pull latest code
git pull origin main

# Rebuild và deploy
./scripts/build.sh
./scripts/deploy.sh
```

## 📞 Hỗ trợ

- **Logs**: `docker-compose -f docker-compose.prod.yml logs`
- **Health**: `curl http://your-domain.com/health`
- **Status**: `docker-compose -f docker-compose.prod.yml ps`

---

**Lưu ý**: Đảm bảo thay đổi tất cả placeholder values trước khi deploy! 