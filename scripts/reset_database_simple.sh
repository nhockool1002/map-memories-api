#!/bin/bash

echo "Resetting Database - Xóa hết data hiện tại"
echo "============================================="

echo ""
echo "1. Dừng ứng dụng nếu đang chạy..."
pkill -f "map-memories-api" || true

echo ""
echo "2. Xóa tất cả data và tạo lại schema..."

# Tạo file reset database
cat > reset_db.go << 'EOF'
package main

import (
	"log"
	"map-memories-api/config"
	"map-memories-api/database"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Connect to database
	database.Connect()
	defer database.Close()

	log.Println("Dropping all tables...")
	
	// Drop all tables in correct order (due to foreign keys)
	tables := []string{
		"mm_memories",
		"mm_locations", 
		"mm_media",
		"mm_memory_likes",
		"mm_user_items",
		"mm_shop_items",
		"mm_user_sessions",
		"mm_users",
		"mm_transaction_logs",
	}

	for _, table := range tables {
		log.Printf("Dropping table: %s", table)
		if err := database.DB.Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error; err != nil {
			log.Printf("Error dropping table %s: %v", table, err)
		}
	}

	log.Println("Running migrations...")
	database.AutoMigrate()

	// Run custom migrations
	runCustomMigrations()

	log.Println("Running seed data...")
	database.SeedData()

	log.Println("✅ Database reset completed successfully!")
}

// runCustomMigrations runs custom database migrations
func runCustomMigrations() {
	log.Println("Running custom migrations...")
	
	// Check if is_deleted column exists in mm_locations table
	var columnExists bool
	err := database.DB.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'mm_locations' 
			AND column_name = 'is_deleted'
		)
	`).Scan(&columnExists).Error
	
	if err != nil {
		log.Printf("Error checking is_deleted column: %v", err)
		return
	}
	
	if !columnExists {
		log.Println("Adding is_deleted column to mm_locations table...")
		
		// Add is_deleted column
		if err := database.DB.Exec(`
			ALTER TABLE mm_locations ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE
		`).Error; err != nil {
			log.Printf("Error adding is_deleted column: %v", err)
			return
		}
		
		// Create index
		if err := database.DB.Exec(`
			CREATE INDEX idx_mm_locations_is_deleted ON mm_locations(is_deleted)
		`).Error; err != nil {
			log.Printf("Error creating index: %v", err)
			return
		}
		
		// Update existing records
		if err := database.DB.Exec(`
			UPDATE mm_locations SET is_deleted = FALSE WHERE is_deleted IS NULL
		`).Error; err != nil {
			log.Printf("Error updating existing records: %v", err)
			return
		}
		
		log.Println("Successfully added is_deleted column to mm_locations table")
	} else {
		log.Println("is_deleted column already exists in mm_locations table")
	}
	
	// Check if location_id column in mm_memories table allows NULL
	var locationIdNullable bool
	err = database.DB.Raw(`
		SELECT is_nullable = 'YES' 
		FROM information_schema.columns 
		WHERE table_name = 'mm_memories' 
		AND column_name = 'location_id'
	`).Scan(&locationIdNullable).Error
	
	if err != nil {
		log.Printf("Error checking location_id column: %v", err)
		return
	}
	
	if !locationIdNullable {
		log.Println("Updating location_id column in mm_memories table to allow NULL...")
		
		// Drop NOT NULL constraint
		if err := database.DB.Exec(`
			ALTER TABLE mm_memories ALTER COLUMN location_id DROP NOT NULL
		`).Error; err != nil {
			log.Printf("Error dropping NOT NULL constraint: %v", err)
			return
		}
		
		// Add index
		if err := database.DB.Exec(`
			CREATE INDEX IF NOT EXISTS idx_mm_memories_location_id ON mm_memories(location_id)
		`).Error; err != nil {
			log.Printf("Error creating index: %v", err)
			return
		}
		
		log.Println("Successfully updated location_id column in mm_memories table")
	} else {
		log.Println("location_id column already allows NULL in mm_memories table")
	}
}
EOF

# Chạy script reset
go run reset_db.go

# Xóa file tạm
rm reset_db.go

echo ""
echo "✅ Database đã được reset thành công!"
echo ""
echo "Bây giờ bạn có thể:"
echo "1. Chạy ứng dụng: go run main.go"
echo "2. Test API với dữ liệu mới"
echo "3. Đăng nhập với tài khoản admin: admin/admin"
echo "4. Test soft delete functionality" 