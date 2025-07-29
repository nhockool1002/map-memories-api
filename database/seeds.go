package database

import (
	"log"

	"map-memories-api/models"
	"map-memories-api/utils"
)

// SeedData runs database seeding
func SeedData() {
	log.Println("Starting database seeding...")
	
	// Seed admin user
	seedAdminUser()
	
	log.Println("Database seeding completed successfully")
}

// seedAdminUser creates the admin user
func seedAdminUser() {
	// Check if admin user already exists
	var existingUser models.User
	result := DB.Where("username = ? OR email = ?", "admin", "admin@map-memories.com").First(&existingUser)
	
	if result.Error == nil {
		log.Println("Admin user already exists, skipping...")
		return
	}
	
	// Hash the password
	hashedPassword, err := utils.HashPassword("admin")
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return
	}
	
	// Create admin user
	adminUser := models.User{
		Username:     "admin",
		Email:        "admin@map-memories.com",
		PasswordHash: hashedPassword,
		FullName:     "Administrator",
	}
	
	// Save to database
	if err := DB.Create(&adminUser).Error; err != nil {
		log.Printf("Error creating admin user: %v", err)
		return
	}
	
	log.Printf("Admin user created successfully with ID: %d", adminUser.ID)
} 