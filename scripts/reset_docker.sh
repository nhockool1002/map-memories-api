#!/bin/bash

echo "Resetting Database với Docker - Xóa hết data hiện tại"
echo "====================================================="

echo ""
echo "1. Dừng tất cả containers..."
docker-compose down

echo ""
echo "2. Xóa tất cả volumes để xóa data..."
docker-compose down -v

echo ""
echo "3. Xóa tất cả images (optional)..."
read -p "Bạn có muốn xóa tất cả images không? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Xóa images..."
    docker-compose down --rmi all
    docker system prune -f
fi

echo ""
echo "4. Xóa tất cả containers và networks..."
docker system prune -f

echo ""
echo "5. Khởi động lại database..."
docker-compose up -d postgres

echo ""
echo "6. Đợi database khởi động..."
sleep 10

echo ""
echo "7. Chạy migration và seed data..."
go run cmd/migrate/main.go
go run cmd/seed/main.go

echo ""
echo "✅ Database đã được reset thành công với Docker!"
echo ""
echo "Bây giờ bạn có thể:"
echo "1. Khởi động ứng dụng: docker-compose up"
echo "2. Hoặc chạy ứng dụng local: go run main.go"
echo "3. Test API với dữ liệu mới"
echo "4. Đăng nhập với tài khoản admin: admin/admin" 