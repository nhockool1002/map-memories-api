# Seed Data Documentation

## Tổng quan

Dự án Map Memories API có sẵn seed data để tạo user admin mặc định khi khởi động ứng dụng.

## User Admin

### Thông tin đăng nhập
- **Username**: `admin`
- **Password**: `admin`
- **Email**: `admin@map-memories.com`
- **Full Name**: `Administrator`

### Cách hoạt động

1. **Tự động**: Khi khởi động ứng dụng, seed data sẽ tự động chạy và tạo user admin nếu chưa tồn tại.

2. **Thủ công**: Bạn có thể chạy seed data riêng biệt bằng cách:
   ```bash
   go run cmd/seed/main.go
   ```

### Tính năng

- **Kiểm tra trùng lặp**: Seed data sẽ kiểm tra xem user admin đã tồn tại chưa (dựa trên username hoặc email)
- **Hash password**: Password được hash bằng bcrypt trước khi lưu vào database
- **Logging**: Có log để theo dõi quá trình tạo seed data

### Cấu trúc file

```
database/
├── connection.go    # Kết nối database
├── seeds.go         # Seed data logic
└── migrations/      # Database migrations

cmd/
└── seed/
    └── main.go      # Script chạy seed data riêng biệt
```

### Sử dụng

1. **Khởi động ứng dụng bình thường**:
   ```bash
   go run main.go
   ```
   Seed data sẽ tự động chạy.

2. **Chạy seed data riêng biệt**:
   ```bash
   go run cmd/seed/main.go
   ```

3. **Đăng nhập với user admin**:
   ```bash
   curl -X POST http://localhost:8222/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
       "email": "admin@map-memories.com",
       "password": "admin"
     }'
   ```

### Lưu ý

- Seed data chỉ chạy một lần và không tạo lại user nếu đã tồn tại
- Password được hash an toàn bằng bcrypt
- User admin có đầy đủ quyền truy cập vào tất cả API endpoints
- Email và username là duy nhất trong hệ thống 