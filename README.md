# Map Memories API

## Giới thiệu

Map Memories là một API RESTful được xây dựng bằng Go và PostgreSQL, cho phép người dùng tạo và quản lý những kỷ niệm gắn liền với các địa điểm trên bản đồ. Ứng dụng hỗ trợ upload hình ảnh/video, tìm kiếm geospatial và chia sẻ kỷ niệm công khai.

## Tính năng chính

- ✅ **Xác thực người dùng**: Đăng ký, đăng nhập với JWT authentication
- ✅ **Quản lý địa điểm**: Tạo, sửa, xóa và tìm kiếm địa điểm với tọa độ GPS
- ✅ **Quản lý kỷ niệm**: Viết bài kỷ niệm, đính kèm media, phân loại với tags
- ✅ **Upload media**: Hỗ trợ upload hình ảnh (JPEG, PNG, GIF) và video (MP4, AVI, MOV)
- ✅ **Tìm kiếm geospatial**: Tìm địa điểm và kỷ niệm trong bán kính từ tọa độ
- ✅ **Phân quyền**: Hệ thống phân quyền user/admin với middleware bảo mật
- ✅ **Swagger documentation**: Tài liệu API tự động và interactive
- ✅ **Docker hóa**: Dễ dàng triển khai với Docker Compose

## Công nghệ sử dụng

- **Backend**: Go 1.21, Gin framework
- **Database**: PostgreSQL 15 + PostGIS extension
- **ORM**: GORM
- **Authentication**: JWT tokens
- **Documentation**: Swagger/OpenAPI 3.0
- **Containerization**: Docker & Docker Compose
- **File Upload**: Multipart form upload với validation

## Cấu trúc Database

### Bảng chính (prefix: `mm_`)

- `mm_users`: Thông tin người dùng
- `mm_locations`: Địa điểm với tọa độ GPS
- `mm_memories`: Bài viết kỷ niệm
- `mm_media`: File hình ảnh/video
- `mm_user_sessions`: Session quản lý JWT
- `mm_memory_likes`: Lượt thích bài viết

### Tính năng PostGIS

- Sử dụng geometry columns cho tìm kiếm hiệu quả
- Hỗ trợ tìm kiếm trong bán kính (ST_DWithin)
- Tính khoảng cách chính xác (ST_Distance)

## Cài đặt và chạy

### 1. Yêu cầu hệ thống

- Docker và Docker Compose
- Git
- Port 8080 và 5432 không bị sử dụng

### 2. Clone repository

```bash
git clone <repository-url>
cd map-memories-api
```

### 3. Cấu hình environment

```bash
cp .env.example .env
```

Chỉnh sửa file `.env` theo nhu cầu:

```env
# Database
DB_HOST=postgres
DB_USER=mm_user
DB_PASSWORD=mm_password
DB_NAME=map_memories

# JWT Secret (đổi trong production)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Server
PORT=8080
HOST=0.0.0.0

# Upload settings
MAX_FILE_SIZE=50MB
UPLOAD_PATH=/app/uploads
```

### 4. Chạy với Docker Compose

```bash
# Chạy tất cả services
docker-compose up -d

# Xem logs
docker-compose logs -f api

# Dừng services
docker-compose down
```

### 5. Kiểm tra kết nối

```bash
# Health check
curl http://localhost:8080/health

# Swagger documentation
open http://localhost:8080/swagger/index.html
```

## Cách sử dụng API

### 1. Đăng ký tài khoản

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User"
  }'
```

### 2. Đăng nhập

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 3. Tạo địa điểm

```bash
curl -X POST http://localhost:8080/api/v1/locations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Hồ Gươm",
    "description": "Hồ Hoàn Kiếm, trung tâm Hà Nội",
    "latitude": 21.0285,
    "longitude": 105.8542,
    "address": "Hoàn Kiếm, Hà Nội",
    "city": "Hà Nội",
    "country": "Việt Nam"
  }'
```

### 4. Tạo kỷ niệm

```bash
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "location_id": 1,
    "title": "Dạo quanh Hồ Gươm",
    "content": "Buổi chiều đẹp trời đi dạo quanh hồ Gươm...",
    "is_public": true,
    "tags": ["hanoi", "hoangkiem", "travel"]
  }'
```

### 5. Upload hình ảnh

```bash
curl -X POST http://localhost:8080/api/v1/media/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "memory_id=1" \
  -F "file=@/path/to/image.jpg"
```

### 6. Tìm kiếm địa điểm gần đó

```bash
curl "http://localhost:8080/api/v1/locations/nearby?latitude=21.0285&longitude=105.8542&radius=5&limit=10"
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Đăng ký
- `POST /api/v1/auth/login` - Đăng nhập
- `GET /api/v1/auth/profile` - Xem profile
- `PUT /api/v1/auth/profile` - Cập nhật profile
- `POST /api/v1/auth/logout` - Đăng xuất

### Locations
- `GET /api/v1/locations` - Danh sách địa điểm
- `POST /api/v1/locations` - Tạo địa điểm mới
- `GET /api/v1/locations/{uuid}` - Chi tiết địa điểm
- `PUT /api/v1/locations/{uuid}` - Cập nhật địa điểm
- `DELETE /api/v1/admin/locations/{uuid}` - Xóa địa điểm (admin)
- `GET /api/v1/locations/nearby` - Tìm địa điểm gần đó
- `GET /api/v1/locations/{uuid}/memories` - Kỷ niệm tại địa điểm

### Memories
- `GET /api/v1/memories` - Danh sách kỷ niệm
- `POST /api/v1/memories` - Tạo kỷ niệm mới
- `GET /api/v1/memories/{uuid}` - Chi tiết kỷ niệm
- `PUT /api/v1/memories/{uuid}` - Cập nhật kỷ niệm
- `DELETE /api/v1/memories/{uuid}` - Xóa kỷ niệm

### Media
- `POST /api/v1/media/upload` - Upload file
- `GET /api/v1/media` - Danh sách media
- `GET /api/v1/media/{uuid}` - Thông tin media
- `GET /api/v1/media/{uuid}/file` - Tải file
- `PUT /api/v1/media/{uuid}` - Cập nhật media
- `DELETE /api/v1/media/{uuid}` - Xóa media
- `GET /api/v1/memories/{uuid}/media` - Media của kỷ niệm

### Admin
- `DELETE /api/v1/admin/locations/{uuid}` - Xóa địa điểm
- `GET /api/v1/admin/memories` - Tất cả kỷ niệm
- `GET /api/v1/admin/media` - Tất cả media

## Development

### Chạy môi trường development

```bash
# Chỉ chạy database
docker-compose up -d postgres

# Cài đặt dependencies
go mod download

# Chạy app
go run main.go
```

### Generate Swagger docs

```bash
# Cài đặt swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g main.go
```

### Database migrations

Database sẽ tự động migrate khi khởi động. Manual migration:

```bash
# Vào container database
docker-compose exec postgres psql -U mm_user -d map_memories

# Chạy custom migrations
\i /docker-entrypoint-initdb.d/001_initial_schema.sql
```

## Testing

### Test với curl

```bash
# Health check
curl http://localhost:8080/health

# Test công khai endpoints
curl http://localhost:8080/api/v1/locations
curl http://localhost:8080/api/v1/memories?is_public=true
```

### Test với Postman

Import Swagger spec từ `http://localhost:8080/swagger/doc.json` vào Postman để test interactive.

## Production Deployment

### 1. Cấu hình bảo mật

```env
ENV=production
JWT_SECRET=<random-256-bit-key>
DB_PASSWORD=<strong-password>
```

### 2. SSL/HTTPS

Cấu hình nginx reverse proxy với SSL:

```nginx
server {
    listen 443 ssl;
    server_name api.mapmemories.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 3. Backup database

```bash
# Backup
docker-compose exec postgres pg_dump -U mm_user map_memories > backup.sql

# Restore
docker-compose exec -T postgres psql -U mm_user map_memories < backup.sql
```

## Troubleshooting

### Lỗi thường gặp

1. **Port đã được sử dụng**
   ```bash
   docker-compose down
   sudo lsof -i :8080
   sudo lsof -i :5432
   ```

2. **Database connection failed**
   ```bash
   docker-compose logs postgres
   docker-compose restart postgres
   ```

3. **Permission denied khi upload**
   ```bash
   chmod 755 uploads/
   chown -R 1000:1000 uploads/
   ```

4. **JWT token expired**
   - Token có thời hạn 24 giờ
   - Đăng nhập lại để lấy token mới

### Logs và monitoring

```bash
# API logs
docker-compose logs -f api

# Database logs
docker-compose logs -f postgres

# All services
docker-compose logs -f
```

## Contributing

1. Fork repository
2. Tạo feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Tạo Pull Request

## License

Dự án được phát hành dưới [MIT License](LICENSE).

## Support

- **GitHub Issues**: [Tạo issue mới](https://github.com/your-repo/issues)
- **Email**: support@mapmemories.com
- **Documentation**: [Swagger UI](http://localhost:8080/swagger/index.html)

## Changelog

### v1.0.0 (2024-01-XX)
- ✅ API cơ bản cho authentication, locations, memories, media
- ✅ PostgreSQL + PostGIS integration
- ✅ Docker containerization
- ✅ Swagger documentation
- ✅ File upload với validation
- ✅ Geospatial search features
