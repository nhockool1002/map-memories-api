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

	// Run database migrations
	database.AutoMigrate()

	// Run custom migrations
	runCustomMigrations()

	log.Println("All migrations completed successfully")
}

// runCustomMigrations runs custom database migrations
func runCustomMigrations() {
	log.Println("Running custom migrations...")

	// Check if deleted_at column exists in mm_locations table
	var columnExists bool
	err := database.DB.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'mm_locations' 
			AND column_name = 'deleted_at'
		)
	`).Scan(&columnExists).Error

	if err != nil {
		log.Printf("Error checking deleted_at column: %v", err)
		return
	}

	if !columnExists {
		log.Println("Adding deleted_at column to mm_locations table...")

		// Add deleted_at column
		if err := database.DB.Exec(`
			ALTER TABLE mm_locations ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE
		`).Error; err != nil {
			log.Printf("Error adding deleted_at column: %v", err)
			return
		}

		// Create index
		if err := database.DB.Exec(`
			CREATE INDEX idx_mm_locations_deleted_at ON mm_locations(deleted_at)
		`).Error; err != nil {
			log.Printf("Error creating index: %v", err)
			return
		}

		log.Println("Successfully added deleted_at column to mm_locations table")
	} else {
		log.Println("deleted_at column already exists in mm_locations table")
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
