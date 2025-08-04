#!/bin/bash

echo "Resetting Database - Xóa hết data hiện tại"
echo "============================================="

# Load environment variables if .env exists
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Database connection parameters
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-map_memories}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}

echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "User: $DB_USER"

# Function to execute SQL command
execute_sql() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "$1"
}

# Function to check if database exists
check_database() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -lqt | cut -d \| -f 1 | grep -qw $DB_NAME
}

echo ""
echo "1. Kiểm tra kết nối database..."
if ! check_database; then
    echo "❌ Không thể kết nối đến database '$DB_NAME'"
    echo "Vui lòng kiểm tra thông tin kết nối trong file .env hoặc biến môi trường"
    exit 1
fi
echo "✅ Kết nối database thành công"

echo ""
echo "2. Xóa tất cả data trong các bảng..."

# Drop all data from tables (in correct order due to foreign keys)
echo "   - Xóa memories..."
execute_sql "DELETE FROM mm_memories;"

echo "   - Xóa locations..."
execute_sql "DELETE FROM mm_locations;"

echo "   - Xóa media..."
execute_sql "DELETE FROM mm_media;"

echo "   - Xóa memory likes..."
execute_sql "DELETE FROM mm_memory_likes;"

echo "   - Xóa user items..."
execute_sql "DELETE FROM mm_user_items;"

echo "   - Xóa shop items..."
execute_sql "DELETE FROM mm_shop_items;"

echo "   - Xóa user sessions..."
execute_sql "DELETE FROM mm_user_sessions;"

echo "   - Xóa users..."
execute_sql "DELETE FROM mm_users;"

echo "   - Xóa transaction logs..."
execute_sql "DELETE FROM mm_transaction_logs;"

echo ""
echo "3. Reset sequences..."
execute_sql "ALTER SEQUENCE mm_users_id_seq RESTART WITH 1;"
execute_sql "ALTER SEQUENCE mm_locations_id_seq RESTART WITH 1;"
execute_sql "ALTER SEQUENCE mm_memories_id_seq RESTART WITH 1;"
execute_sql "ALTER SEQUENCE mm_media_id_seq RESTART WITH 1;"
execute_sql "ALTER SEQUENCE mm_shop_items_id_seq RESTART WITH 1;"
execute_sql "ALTER SEQUENCE mm_user_items_id_seq RESTART WITH 1;"
execute_sql "ALTER SEQUENCE mm_memory_likes_id_seq RESTART WITH 1;"
execute_sql "ALTER SEQUENCE mm_transaction_logs_id_seq RESTART WITH 1;"

echo ""
echo "4. Chạy migration để tạo lại schema..."
go run cmd/migrate/main.go

echo ""
echo "5. Chạy seed data..."
go run cmd/seed/main.go

echo ""
echo "✅ Database đã được reset thành công!"
echo ""
echo "Bây giờ bạn có thể:"
echo "1. Chạy ứng dụng: go run main.go"
echo "2. Test API với dữ liệu mới"
echo "3. Đăng nhập với tài khoản admin: admin/admin" 