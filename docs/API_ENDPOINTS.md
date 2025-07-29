# Map Memories API - Danh sách đầy đủ API Endpoints

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/auth/register` | Đăng ký tài khoản mới | ❌ |
| `POST` | `/auth/login` | Đăng nhập | ❌ |
| `GET` | `/auth/profile` | Xem profile người dùng | ✅ |
| `PUT` | `/auth/profile` | Cập nhật profile | ✅ |
| `POST` | `/auth/logout` | Đăng xuất | ✅ |

## Location Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/locations` | Danh sách địa điểm (có pagination) | ❌ |
| `POST` | `/locations` | Tạo địa điểm mới | ✅ |
| `GET` | `/locations/{uuid}` | Chi tiết địa điểm | ❌ |
| `PUT` | `/locations/{uuid}` | Cập nhật địa điểm | ✅ |
| `GET` | `/locations/nearby` | Tìm địa điểm gần tọa độ | ❌ |
| `GET` | `/locations/{uuid}/memories` | Kỷ niệm tại địa điểm | ❌ |

## Memory Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/memories` | Danh sách kỷ niệm (có filters) | ❌ |
| `POST` | `/memories` | Tạo kỷ niệm mới | ✅ |
| `GET` | `/memories/{uuid}` | Chi tiết kỷ niệm | ❌ |
| `PUT` | `/memories/{uuid}` | Cập nhật kỷ niệm (owner only) | ✅ |
| `DELETE` | `/memories/{uuid}` | Xóa kỷ niệm (owner only) | ✅ |
| `GET` | `/memories/{memory_uuid}/media` | Media của kỷ niệm | ❌ |

## Media Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/media/upload` | Upload hình ảnh/video | ✅ |
| `GET` | `/media` | Danh sách media (có filters) | ❌ |
| `GET` | `/media/{uuid}` | Thông tin media | ❌ |
| `GET` | `/media/{uuid}/file` | Tải file media | ❌ |
| `PUT` | `/media/{uuid}` | Cập nhật media (owner only) | ✅ |
| `DELETE` | `/media/{uuid}` | Xóa media (owner only) | ✅ |

## Admin Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `DELETE` | `/admin/locations/{uuid}` | Xóa địa điểm | ✅ Admin |
| `GET` | `/admin/memories` | Tất cả kỷ niệm | ✅ Admin |
| `GET` | `/admin/media` | Tất cả media | ✅ Admin |

## Health Check

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/health` | Kiểm tra tình trạng API | ❌ |

## Documentation

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/` | Redirect to Swagger UI | ❌ |
| `GET` | `/swagger/*` | Swagger UI Documentation | ❌ |

---

## Query Parameters Summary

### Pagination (áp dụng cho endpoints trả về danh sách)
- `page` (int): Số trang (default: 1)
- `limit` (int): Số items/trang (default: 20, max: 100)

### Location Filters
- `search` (string): Tìm kiếm trong tên và mô tả
- `country` (string): Lọc theo quốc gia
- `city` (string): Lọc theo thành phố

### Memory Filters
- `user_id` (int): Lọc theo người dùng
- `location_id` (int): Lọc theo địa điểm
- `is_public` (bool): Lọc public/private
- `search` (string): Tìm kiếm trong title và content
- `tags` (string): Lọc theo tags (comma-separated)
- `sort_by` (string): Sắp xếp theo (created_at, visit_date, title)
- `sort_order` (string): Thứ tự (asc, desc)

### Media Filters
- `memory_id` (int): Lọc theo kỷ niệm
- `media_type` (string): Lọc theo loại (image, video)

### Geospatial Search (/locations/nearby)
- `latitude` (float, required): Vĩ độ (-90 to 90)
- `longitude` (float, required): Kinh độ (-180 to 180)
- `radius` (float): Bán kính tìm kiếm (km, default: 10, max: 100)
- `limit` (int): Số kết quả tối đa (default: 20, max: 100)

---

## HTTP Status Codes

### Success Codes
- `200 OK`: Request thành công
- `201 Created`: Tạo resource thành công
- `204 No Content`: Xóa thành công

### Client Error Codes
- `400 Bad Request`: Request không hợp lệ
- `401 Unauthorized`: Chưa xác thực hoặc token không hợp lệ
- `403 Forbidden`: Không có quyền truy cập
- `404 Not Found`: Resource không tồn tại
- `409 Conflict`: Conflict với trạng thái hiện tại
- `413 Payload Too Large`: File upload quá lớn
- `422 Unprocessable Entity`: Validation errors

### Server Error Codes
- `500 Internal Server Error`: Lỗi server
- `503 Service Unavailable`: Service không khả dụng

---

## Authentication

### JWT Token Headers
```
Authorization: Bearer <your-jwt-token>
```

### Admin Access
- Email admin: `admin@mapmemories.com` hoặc `administrator@mapmemories.com`
- Chỉ admin mới có thể xóa locations và truy cập admin endpoints

---

## File Upload Specifications

### Supported File Types
- **Images**: `image/jpeg`, `image/png`, `image/gif`
- **Videos**: `video/mp4`, `video/avi`, `video/mov`

### File Size Limits
- Maximum file size: **50MB**
- Validation based on file signature và extension

### Upload Process
1. POST to `/media/upload` với multipart/form-data
2. Include `memory_id` và optional `display_order`
3. File sẽ được lưu với unique filename
4. Response chứa URL để access file

---

## Rate Limiting

- **Limit**: 60 requests per minute per IP
- **Burst**: 10 requests
- **Response**: 429 Too Many Requests khi vượt giới hạn

---

## Error Response Format

```json
{
  "success": false,
  "message": "Error description",
  "error": {
    "code": "ERROR_CODE",
    "message": "Detailed error message",
    "details": {} // Optional additional details
  }
}
```

### Common Error Codes
- `VALIDATION_ERROR`: Lỗi validation input
- `UNAUTHORIZED`: Lỗi xác thực
- `FORBIDDEN`: Không có quyền truy cập
- `NOT_FOUND`: Resource không tồn tại
- `CONFLICT`: Conflict với dữ liệu hiện tại
- `INTERNAL_ERROR`: Lỗi server
- `RATE_LIMIT_EXCEEDED`: Vượt giới hạn rate limit

---

## Sample API Calls

### Quick Start
```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Get public memories
curl http://localhost:8080/api/v1/memories?is_public=true&limit=5

# Search nearby locations
curl "http://localhost:8080/api/v1/locations/nearby?latitude=21.0285&longitude=105.8542&radius=5"
```

### With Authentication
```bash
export TOKEN="your-jwt-token-here"

# Create location
curl -X POST http://localhost:8080/api/v1/locations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Test Location","latitude":21.0285,"longitude":105.8542}'

# Create memory
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"location_id":1,"title":"Test Memory","content":"Test content","is_public":true}'

# Upload media
curl -X POST http://localhost:8080/api/v1/media/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "memory_id=1" \
  -F "file=@image.jpg"
```

---

Để biết thêm chi tiết về từng endpoint, tham khảo [API Documentation](./API_DOCUMENTATION.md) hoặc truy cập [Swagger UI](http://localhost:8080/swagger/index.html) khi service đang chạy.