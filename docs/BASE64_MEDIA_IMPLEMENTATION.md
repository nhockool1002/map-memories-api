# Base64 Media Storage Implementation

## Tổng quan

Đã cập nhật API Media để tự động convert file upload thành base64 và lưu trực tiếp vào database thay vì lưu file trên filesystem.

## Những thay đổi chính

### 1. Thêm hàm mới trong `utils/file.go`

- **`SaveUploadedFileAsBase64()`**: Hàm mới để convert file upload thành base64 data URL
- Tự động detect MIME type và validate file type
- Convert file content thành base64 string với format data URL: `data:mime/type;base64,base64string`

### 2. Cập nhật API `UploadMedia` trong `controllers/media.go`

- Thay đổi từ `utils.SaveUploadedFile()` sang `utils.SaveUploadedFileAsBase64()`
- File được convert thành base64 và lưu vào trường `file_path` trong database
- Không cần lưu file trên filesystem nữa
- Cập nhật error message phù hợp

### 3. Cập nhật API `ServeMediaFile`

- Thay đổi logic để decode base64 data từ `file_path`
- Extract base64 string từ data URL format
- Decode và serve file content trực tiếp
- Cập nhật error handling cho base64 format

### 4. Cập nhật API `DeleteMedia`

- Loại bỏ logic xóa file từ filesystem
- Chỉ cần xóa record từ database vì file đã được lưu dưới dạng base64

### 5. Cập nhật Model Response

- `MediaResponse.URL` được cập nhật để trỏ đến endpoint `/file`
- `file_path` giờ chứa base64 data URL thay vì đường dẫn file

## Lợi ích

1. **Đơn giản hóa**: Không cần quản lý filesystem, tất cả dữ liệu trong database
2. **Backup dễ dàng**: Chỉ cần backup database, không cần backup files
3. **Deployment đơn giản**: Không cần cấu hình storage path
4. **Consistency**: Tất cả dữ liệu media được lưu trữ thống nhất

## Nhược điểm

1. **Database size**: Base64 encoding làm tăng kích thước dữ liệu khoảng 33%
2. **Performance**: Decode base64 có thể chậm hơn đọc file trực tiếp
3. **Memory usage**: Cần load toàn bộ file vào memory khi decode

## API Endpoints

### Upload Media
```
POST /api/v1/media/upload
Content-Type: multipart/form-data
Authorization: Bearer <token>

Form data:
- memory_id: int (required)
- display_order: int (optional, default: 0)
- file: file (required)
```

**Response:**
```json
{
  "success": true,
  "message": "Media uploaded successfully",
  "data": {
    "id": 1,
    "uuid": "123e4567-e89b-12d3-a456-426614174000",
    "filename": "20231201_143022_abc123.jpg",
    "original_filename": "photo.jpg",
    "file_path": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ...",
    "file_size": 1024,
    "mime_type": "image/jpeg",
    "media_type": "image",
    "display_order": 0,
    "url": "/api/v1/media/123e4567-e89b-12d3-a456-426614174000/file",
    "created_at": "2023-12-01T14:30:22Z"
  }
}
```

### Serve Media File
```
GET /api/v1/media/{uuid}/file
```

**Response:** File content với appropriate headers

## Testing

Sử dụng file `test_base64_upload.html` để test API:

1. Mở file trong browser
2. Nhập JWT token
3. Chọn memory ID
4. Upload file
5. Kiểm tra response và preview (nếu là image)

## Migration Notes

- API hiện tại vẫn tương thích với frontend cũ
- Không cần thay đổi database schema
- Có thể migrate dữ liệu cũ bằng cách convert file cũ thành base64

## Future Considerations

1. **Compression**: Có thể thêm compression cho base64 data
2. **Caching**: Implement caching cho decoded base64 data
3. **CDN**: Có thể serve base64 data qua CDN
4. **Thumbnails**: Tạo thumbnails cho images và lưu dưới dạng base64 