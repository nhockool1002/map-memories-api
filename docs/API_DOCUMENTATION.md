# Map Memories API - Tài liệu tích hợp API hoàn chỉnh

## Tổng quan

Map Memories API là một RESTful API được thiết kế để quản lý kỷ niệm và địa điểm trên bản đồ. API hỗ trợ đầy đủ các tính năng CRUD, xác thực người dùng, upload media và tìm kiếm geospatial.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

API sử dụng JWT (JSON Web Token) để xác thực. Token có thời hạn 24 giờ.

### Headers yêu cầu

```
Authorization: Bearer <your-jwt-token>
Content-Type: application/json
```

## Error Responses

Tất cả lỗi tuân theo format chuẩn:

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

### HTTP Status Codes

- `200` - OK
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `413` - Payload Too Large
- `422` - Unprocessable Entity
- `500` - Internal Server Error

## Pagination

Các endpoint trả về danh sách hỗ trợ pagination:

### Request Parameters
- `page` (int): Số trang (default: 1)
- `limit` (int): Số items/trang (default: 20, max: 100)

### Response Format
```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": [...],
  "pagination": {
    "current_page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  }
}
```

---

# 1. Authentication API

## 1.1 Đăng ký người dùng

**Endpoint:** `POST /auth/register`

### Request Body
```json
{
  "username": "string (required, 3-50 chars)",
  "email": "string (required, valid email)",
  "password": "string (required, min 6 chars)",
  "full_name": "string (optional)"
}
```

### Response (201)
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": 1,
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "username": "testuser",
      "email": "test@example.com",
      "full_name": "Test User",
      "avatar_url": "",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

### Validation Errors (400)
```json
{
  "success": false,
  "message": "Validation failed",
  "error": {
    "code": "VALIDATION_ERROR",
    "details": [
      {
        "field": "email",
        "message": "email must be a valid email address"
      }
    ]
  }
}
```

## 1.2 Đăng nhập

**Endpoint:** `POST /auth/login`

### Request Body
```json
{
  "email": "test@example.com",
  "password": "password123"
}
```

### Response (200)
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": { /* User object */ },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

## 1.3 Xem profile

**Endpoint:** `GET /auth/profile`
**Auth:** Required

### Response (200)
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "id": 1,
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "username": "testuser",
    "email": "test@example.com",
    "full_name": "Test User",
    "avatar_url": "",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

## 1.4 Cập nhật profile

**Endpoint:** `PUT /auth/profile`
**Auth:** Required

### Request Body
```json
{
  "full_name": "New Full Name",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

### Response (200)
```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": { /* Updated user object */ }
}
```

---

# 2. Locations API

## 2.1 Danh sách địa điểm

**Endpoint:** `GET /locations`

### Query Parameters
- `page` (int): Pagination
- `limit` (int): Items per page
- `search` (string): Tìm kiếm trong tên và mô tả
- `country` (string): Filter theo quốc gia
- `city` (string): Filter theo thành phố

### Response (200)
```json
{
  "success": true,
  "message": "Locations retrieved successfully",
  "data": [
    {
      "id": 1,
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Hồ Gươm",
      "description": "Hồ Hoàn Kiếm, trung tâm Hà Nội",
      "latitude": 21.0285,
      "longitude": 105.8542,
      "address": "Hoàn Kiếm, Hà Nội",
      "country": "Việt Nam",
      "city": "Hà Nội",
      "memory_count": 15,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": { /* Pagination object */ }
}
```

## 2.2 Tạo địa điểm mới

**Endpoint:** `POST /locations`
**Auth:** Required

### Request Body
```json
{
  "name": "Hồ Gươm",
  "description": "Hồ Hoàn Kiếm, trung tâm Hà Nội",
  "latitude": 21.0285,
  "longitude": 105.8542,
  "address": "Hoàn Kiếm, Hà Nội",
  "country": "Việt Nam",
  "city": "Hà Nội"
}
```

### Response (201)
```json
{
  "success": true,
  "message": "Location created successfully",
  "data": { /* Location object */ }
}
```

## 2.3 Chi tiết địa điểm

**Endpoint:** `GET /locations/{uuid}`

### Response (200)
```json
{
  "success": true,
  "message": "Location retrieved successfully",
  "data": { /* Location object with memory count */ }
}
```

## 2.4 Tìm kiếm địa điểm gần đó

**Endpoint:** `GET /locations/nearby`

### Query Parameters
- `latitude` (float, required): Vĩ độ (-90 to 90)
- `longitude` (float, required): Kinh độ (-180 to 180)
- `radius` (float): Bán kính tìm kiếm (km, default: 10, max: 100)
- `limit` (int): Số kết quả tối đa (default: 20, max: 100)

### Response (200)
```json
{
  "success": true,
  "message": "Found 5 locations within 10.0 km",
  "data": [
    {
      /* Location objects sorted by distance */
    }
  ]
}
```

## 2.5 Kỷ niệm tại địa điểm

**Endpoint:** `GET /locations/{uuid}/memories`

### Query Parameters
- `page`, `limit`: Pagination
- `is_public` (bool): Filter public/private

### Response (200)
```json
{
  "success": true,
  "message": "Location memories retrieved successfully",
  "data": [ /* Array of memory objects */ ],
  "pagination": { /* Pagination object */ }
}
```

---

# 3. Memories API

## 3.1 Danh sách kỷ niệm

**Endpoint:** `GET /memories`

### Query Parameters
- `page`, `limit`: Pagination
- `user_id` (int): Filter theo user
- `location_id` (int): Filter theo địa điểm
- `is_public` (bool): Filter public/private
- `search` (string): Tìm kiếm trong title và content
- `tags` (string): Filter theo tags (comma-separated)
- `sort_by` (string): Sắp xếp theo (created_at, visit_date, title)
- `sort_order` (string): Thứ tự (asc, desc)

### Response (200)
```json
{
  "success": true,
  "message": "Memories retrieved successfully",
  "data": [
    {
      "id": 1,
      "uuid": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Dạo quanh Hồ Gươm",
      "content": "Buổi chiều đẹp trời đi dạo...",
      "visit_date": "2024-01-15",
      "is_public": true,
      "tags": ["hanoi", "hoangkiem", "travel"],
      "like_count": 10,
      "media_count": 3,
      "is_liked": false,
      "user": { /* User object */ },
      "location": { /* Location object */ },
      "media": [ /* Array of media objects */ ],
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": { /* Pagination object */ }
}
```

## 3.2 Tạo kỷ niệm mới

**Endpoint:** `POST /memories`
**Auth:** Required

### Request Body
```json
{
  "location_id": 1,
  "title": "Dạo quanh Hồ Gươm",
  "content": "Buổi chiều đẹp trời đi dạo quanh hồ Gươm...",
  "visit_date": "2024-01-15",
  "is_public": true,
  "tags": ["hanoi", "hoangkiem", "travel"]
}
```

### Response (201)
```json
{
  "success": true,
  "message": "Memory created successfully",
  "data": { /* Memory object with user and location */ }
}
```

## 3.3 Chi tiết kỷ niệm

**Endpoint:** `GET /memories/{uuid}`

### Response (200)
```json
{
  "success": true,
  "message": "Memory retrieved successfully",
  "data": { /* Complete memory object with relationships */ }
}
```

## 3.4 Cập nhật kỷ niệm

**Endpoint:** `PUT /memories/{uuid}`
**Auth:** Required (Owner only)

### Request Body
```json
{
  "title": "New title",
  "content": "Updated content",
  "visit_date": "2024-01-16",
  "is_public": false,
  "tags": ["updated", "tags"]
}
```

### Response (200)
```json
{
  "success": true,
  "message": "Memory updated successfully",
  "data": { /* Updated memory object */ }
}
```

## 3.5 Xóa kỷ niệm

**Endpoint:** `DELETE /memories/{uuid}`
**Auth:** Required (Owner only)

### Response (200)
```json
{
  "success": true,
  "message": "Memory deleted successfully",
  "data": null
}
```

---

# 4. Media API

## 4.1 Upload media

**Endpoint:** `POST /media/upload`
**Auth:** Required
**Content-Type:** `multipart/form-data`

### Form Data
- `memory_id` (int, required): ID của memory
- `display_order` (int, optional): Thứ tự hiển thị (default: 0)
- `file` (file, required): File upload

### Supported File Types
- **Images**: JPEG, PNG, GIF
- **Videos**: MP4, AVI, MOV
- **Max size**: 50MB

### Response (201)
```json
{
  "success": true,
  "message": "Media uploaded successfully",
  "data": {
    "id": 1,
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "filename": "20240115_123456_abc12345.jpg",
    "original_filename": "vacation_photo.jpg",
    "file_path": "/app/uploads/20240115_123456_abc12345.jpg",
    "file_size": 2048576,
    "mime_type": "image/jpeg",
    "media_type": "image",
    "display_order": 0,
    "url": "/api/v1/media/550e8400-e29b-41d4-a716-446655440000/file",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

## 4.2 Danh sách media

**Endpoint:** `GET /media`

### Query Parameters
- `page`, `limit`: Pagination
- `memory_id` (int): Filter theo memory
- `media_type` (string): Filter theo loại (image, video)

### Response (200)
```json
{
  "success": true,
  "message": "Media retrieved successfully",
  "data": [ /* Array of media objects */ ],
  "pagination": { /* Pagination object */ }
}
```

## 4.3 Thông tin media

**Endpoint:** `GET /media/{uuid}`

### Response (200)
```json
{
  "success": true,
  "message": "Media retrieved successfully",
  "data": { /* Media object */ }
}
```

## 4.4 Tải file media

**Endpoint:** `GET /media/{uuid}/file`

### Response (200)
- **Content-Type**: Original file MIME type
- **Content-Disposition**: `inline; filename="original_filename.jpg"`
- **Cache-Control**: `public, max-age=31536000`
- **Body**: File binary data

## 4.5 Media của kỷ niệm

**Endpoint:** `GET /memories/{memory_uuid}/media`

### Response (200)
```json
{
  "success": true,
  "message": "Memory media retrieved successfully",
  "data": [
    { /* Media objects sorted by display_order */ }
  ]
}
```

## 4.6 Cập nhật media

**Endpoint:** `PUT /media/{uuid}`
**Auth:** Required (Owner only)

### Request Body
```json
{
  "display_order": 1
}
```

### Response (200)
```json
{
  "success": true,
  "message": "Media updated successfully",
  "data": { /* Updated media object */ }
}
```

## 4.7 Xóa media

**Endpoint:** `DELETE /media/{uuid}`
**Auth:** Required (Owner only)

### Response (200)
```json
{
  "success": true,
  "message": "Media deleted successfully",
  "data": null
}
```

---

# 5. Admin API

## 5.1 Xóa địa điểm (Admin only)

**Endpoint:** `DELETE /admin/locations/{uuid}`
**Auth:** Required (Admin only)

### Response (200)
```json
{
  "success": true,
  "message": "Location deleted successfully",
  "data": null
}
```

### Conflict Error (409)
```json
{
  "success": false,
  "message": "Cannot delete location with associated memories",
  "error": {
    "code": "LOCATION_HAS_MEMORIES",
    "details": {
      "memory_count": 5
    }
  }
}
```

---

# 6. Health Check

## 6.1 Kiểm tra tình trạng service

**Endpoint:** `GET /health`

### Response (200)
```json
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "status": "healthy",
    "database": "connected"
  }
}
```

### Unhealthy Response (503)
```json
{
  "success": false,
  "message": "Service unhealthy",
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "details": {
      "database": "unhealthy"
    }
  }
}
```

---

# 7. Integration Examples

## 7.1 Complete User Journey

### Step 1: Đăng ký và đăng nhập
```bash
# Đăng ký
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "traveler",
    "email": "traveler@example.com",
    "password": "securepass123",
    "full_name": "Travel Enthusiast"
  }'

# Lưu token từ response
export JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Step 2: Tạo địa điểm
```bash
curl -X POST http://localhost:8080/api/v1/locations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "name": "Hồ Gươm",
    "description": "Hồ Hoàn Kiếm, trung tâm Hà Nội",
    "latitude": 21.0285,
    "longitude": 105.8542,
    "address": "Hoàn Kiếm, Hà Nội",
    "city": "Hà Nội",
    "country": "Việt Nam"
  }'

# Lưu location_id từ response
export LOCATION_ID=1
```

### Step 3: Tạo kỷ niệm
```bash
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "location_id": '$LOCATION_ID',
    "title": "Buổi chiều thư giãn tại Hồ Gươm",
    "content": "Hôm nay thời tiết đẹp, tôi quyết định đi dạo quanh Hồ Gươm. Không khí trong lành, nhiều người tập thể dục...",
    "visit_date": "2024-01-15",
    "is_public": true,
    "tags": ["hanoi", "relaxing", "weekend"]
  }'

# Lưu memory_id từ response
export MEMORY_ID=1
```

### Step 4: Upload hình ảnh
```bash
curl -X POST http://localhost:8080/api/v1/media/upload \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -F "memory_id=$MEMORY_ID" \
  -F "display_order=1" \
  -F "file=@./sunset_at_hoan_kiem.jpg"
```

### Step 5: Tìm kiếm địa điểm gần đó
```bash
curl "http://localhost:8080/api/v1/locations/nearby?latitude=21.0285&longitude=105.8542&radius=2&limit=10"
```

## 7.2 Error Handling Examples

### Invalid Authentication
```bash
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid_token" \
  -d '{...}'

# Response: 401 Unauthorized
{
  "success": false,
  "message": "Invalid or expired token",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired token",
    "details": "token signature invalid"
  }
}
```

### Validation Error
```bash
curl -X POST http://localhost:8080/api/v1/locations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "name": "",
    "latitude": 100,
    "longitude": -200
  }'

# Response: 400 Bad Request
{
  "success": false,
  "message": "Validation failed",
  "error": {
    "code": "VALIDATION_ERROR",
    "details": [
      {
        "field": "name",
        "message": "name is required"
      },
      {
        "field": "latitude",
        "message": "latitude must be between -90 and 90"
      },
      {
        "field": "longitude",
        "message": "longitude must be between -180 and 180"
      }
    ]
  }
}
```

## 7.3 Rate Limiting

API có giới hạn 60 requests/phút cho mỗi IP. Khi vượt giới hạn:

```json
{
  "success": false,
  "message": "Rate limit exceeded",
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests, please try again later",
    "details": {
      "retry_after": 60
    }
  }
}
```

---

# 8. Best Practices

## 8.1 Authentication
- Luôn kiểm tra token expiry
- Implement refresh token logic
- Store token securely (không trong localStorage cho production)

## 8.2 File Upload
- Validate file type trước khi upload
- Implement progress tracking cho large files
- Handle upload errors gracefully

## 8.3 Geospatial Queries
- Sử dụng reasonable radius values (tránh queries quá rộng)
- Cache kết quả location search khi có thể
- Validate coordinates trước khi gửi request

## 8.4 Error Handling
- Luôn check `success` field trong response
- Implement retry logic cho network errors
- Show user-friendly error messages

## 8.5 Performance
- Sử dụng pagination cho large datasets
- Cache static data (như locations)
- Implement lazy loading cho media

---

# 9. SDK và Integration Libraries

## 9.1 JavaScript/TypeScript Example

```typescript
class MapMemoriesAPI {
  private baseURL = 'http://localhost:8080/api/v1';
  private token: string | null = null;

  async login(email: string, password: string) {
    const response = await fetch(`${this.baseURL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    
    const data = await response.json();
    if (data.success) {
      this.token = data.data.access_token;
    }
    return data;
  }

  async createMemory(memoryData: any) {
    const response = await fetch(`${this.baseURL}/memories`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.token}`
      },
      body: JSON.stringify(memoryData)
    });
    
    return response.json();
  }

  async uploadMedia(memoryId: number, file: File) {
    const formData = new FormData();
    formData.append('memory_id', memoryId.toString());
    formData.append('file', file);

    const response = await fetch(`${this.baseURL}/media/upload`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.token}`
      },
      body: formData
    });
    
    return response.json();
  }
}
```

## 9.2 Python Example

```python
import requests
from typing import Optional, Dict, Any

class MapMemoriesClient:
    def __init__(self, base_url: str = "http://localhost:8080/api/v1"):
        self.base_url = base_url
        self.token: Optional[str] = None
    
    def login(self, email: str, password: str) -> Dict[str, Any]:
        response = requests.post(
            f"{self.base_url}/auth/login",
            json={"email": email, "password": password}
        )
        data = response.json()
        if data["success"]:
            self.token = data["data"]["access_token"]
        return data
    
    def _headers(self) -> Dict[str, str]:
        headers = {"Content-Type": "application/json"}
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
        return headers
    
    def create_location(self, location_data: Dict[str, Any]) -> Dict[str, Any]:
        response = requests.post(
            f"{self.base_url}/locations",
            json=location_data,
            headers=self._headers()
        )
        return response.json()
    
    def search_nearby_locations(self, lat: float, lng: float, radius: float = 10) -> Dict[str, Any]:
        params = {
            "latitude": lat,
            "longitude": lng,
            "radius": radius
        }
        response = requests.get(
            f"{self.base_url}/locations/nearby",
            params=params
        )
        return response.json()
```

---

# 10. Changelog và Versioning

## API Versioning
- Current version: v1
- Version trong URL: `/api/v1/`
- Backward compatibility được maintain trong minor updates

## v1.0.0 Features
- ✅ Complete CRUD cho tất cả resources
- ✅ JWT Authentication
- ✅ File upload với validation
- ✅ PostGIS geospatial queries
- ✅ Pagination và filtering
- ✅ Swagger documentation
- ✅ Error handling chuẩn
- ✅ Docker containerization

## Future Versions (Roadmap)
- v1.1.0: Push notifications, WebSocket support
- v1.2.0: Advanced search, full-text search
- v1.3.0: Social features (follow users, comments)
- v2.0.0: GraphQL support, advanced analytics

---

Đây là tài liệu API hoàn chỉnh cho Map Memories. Để biết thêm chi tiết và test interactive, truy cập [Swagger UI](http://localhost:8080/swagger/index.html) khi service đang chạy.