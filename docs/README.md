# Map Memories API Documentation

## Tổng quan

Dự án Map Memories API cung cấp 3 file documentation khác nhau để phục vụ các nhu cầu khác nhau của frontend developers.

## Các file documentation

### 1. `API_ENDPOINTS.txt` - Documentation chi tiết
- **Mục đích**: Documentation đầy đủ và chi tiết nhất
- **Đối tượng**: Developers cần hiểu sâu về API
- **Nội dung**:
  - Tất cả endpoints với mô tả chi tiết
  - Request/Response formats đầy đủ
  - Error codes và messages
  - Authentication flow
  - File upload guidelines
  - Development tips

### 2. `API_QUICK_REFERENCE.txt` - Tóm tắt nhanh
- **Mục đích**: Tham khảo nhanh cho development
- **Đối tượng**: Frontend developers cần tham khảo nhanh
- **Nội dung**:
  - Danh sách endpoints cơ bản
  - Authentication flow
  - Common request formats
  - Error codes
  - Development tips

### 3. `API_ENDPOINTS.json` - Machine-readable format
- **Mục đích**: Để frontend có thể parse và sử dụng programmatically
- **Đối tượng**: Frontend frameworks, code generators
- **Nội dung**:
  - Structured JSON data
  - Endpoints với metadata
  - Response formats
  - Error responses
  - Configuration data

## Cách sử dụng

### Cho Frontend Development

1. **Bắt đầu**: Đọc `API_QUICK_REFERENCE.txt` để hiểu tổng quan
2. **Chi tiết**: Tham khảo `API_ENDPOINTS.txt` khi cần thông tin chi tiết
3. **Tự động hóa**: Sử dụng `API_ENDPOINTS.json` để tạo API clients

### Cho API Integration

```javascript
// Ví dụ sử dụng JSON data
const apiConfig = require('./docs/API_ENDPOINTS.json');

// Lấy base URL
const baseUrl = apiConfig.api_info.base_url;

// Lấy endpoint info
const loginEndpoint = apiConfig.endpoints.authentication.login;
const loginUrl = baseUrl + loginEndpoint.url;

// Tạo request
fetch(loginUrl, {
  method: loginEndpoint.method,
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password123'
  })
});
```

### Cho Testing

1. **Swagger UI**: Truy cập `http://localhost:8222/swagger/index.html`
2. **Postman**: Import endpoints từ JSON file
3. **Curl**: Sử dụng examples từ documentation

## Authentication Flow

### 1. Register User
```bash
curl -X POST http://localhost:8222/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user123",
    "email": "user@example.com",
    "password": "password123",
    "full_name": "John Doe"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8222/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### 3. Use Token
```bash
curl -X GET http://localhost:8222/api/v1/auth/profile \
  -H "Authorization: Bearer <your_jwt_token>"
```

## File Upload

### Upload Media
```bash
curl -X POST http://localhost:8222/api/v1/media/upload \
  -H "Authorization: Bearer <your_jwt_token>" \
  -F "memory_id=1" \
  -F "file=@/path/to/image.jpg" \
  -F "display_order=1"
```

## Error Handling

### Common Error Responses
```json
{
  "success": false,
  "message": "Error message",
  "error": {
    "code": "ERROR_CODE",
    "message": "Detailed error message"
  }
}
```

### HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request (Validation errors)
- `401` - Unauthorized (Invalid/missing token)
- `403` - Forbidden (Insufficient permissions)
- `404` - Not Found
- `409` - Conflict (Resource already exists)
- `413` - Request Entity Too Large (File too big)
- `500` - Internal Server Error

## Development Tips

### 1. Always handle authentication
```javascript
const headers = {
  'Content-Type': 'application/json',
  'Authorization': `Bearer ${token}`
};
```

### 2. Handle pagination
```javascript
const response = await fetch('/api/v1/locations?page=1&limit=20');
const data = await response.json();
const { data: locations, pagination } = data;
```

### 3. Handle file uploads
```javascript
const formData = new FormData();
formData.append('memory_id', memoryId);
formData.append('file', file);
formData.append('display_order', 1);

const response = await fetch('/api/v1/media/upload', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});
```

### 4. Error handling
```javascript
try {
  const response = await fetch(url, options);
  const data = await response.json();
  
  if (!data.success) {
    throw new Error(data.message);
  }
  
  return data.data;
} catch (error) {
  console.error('API Error:', error);
  // Handle error appropriately
}
```

## Testing

### Default Admin User
- **Username**: `admin`
- **Password**: `admin`
- **Email**: `admin@map-memories.com`

### Health Check
```bash
curl http://localhost:8222/health
```

### Swagger Documentation
Truy cập: `http://localhost:8222/swagger/index.html`

## Support

Nếu có vấn đề hoặc cần hỗ trợ:
1. Kiểm tra logs của API server
2. Sử dụng Swagger UI để test endpoints
3. Tham khảo error codes trong documentation
4. Kiểm tra authentication token

## Version History

- **v1.0.0**: Initial API documentation
- **v1.0.1**: Added JSON format for programmatic access
- **v1.0.2**: Added quick reference guide
- **v1.0.3**: Updated with authentication fixes 