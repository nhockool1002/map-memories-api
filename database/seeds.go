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
	
	// Seed shop items
	seedShopItems()
	
	// Seed user items for admin
	seedUserItems()
	
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

// seedShopItems creates sample shop items with the 3 markers
func seedShopItems() {
	log.Println("Seeding shop items...")
	
	// Check if shop items already exist
	var count int64
	DB.Model(&models.ShopItem{}).Count(&count)
	if count > 0 {
		log.Println("Shop items already exist, skipping...")
		return
	}
	
	// Base64 strings for the 3 markers
	marker1Base64 := "data:image/png;base64,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"
	
	marker2Base64 := "data:image/png;base64,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"
	
	marker3Base64 := "data:image/png;base64,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"

	shopItems := []models.ShopItem{
		{
			Name:        "Red Star Marker",
			Description: "A beautiful red star marker for your locations",
			ImageBase64: marker1Base64,
			Price:       1000, // 10 coins
			Stock:       50,
			ItemType:    "marker",
			IsActive:    true,
		},
		{
			Name:        "Blue Diamond Marker",
			Description: "A stunning blue diamond marker for your special places",
			ImageBase64: marker2Base64,
			Price:       2000, // 20 coins
			Stock:       30,
			ItemType:    "marker",
			IsActive:    true,
		},
		{
			Name:        "Green Heart Marker",
			Description: "A lovely green heart marker for your favorite memories",
			ImageBase64: marker3Base64,
			Price:       1500, // 15 coins
			Stock:       40,
			ItemType:    "marker",
			IsActive:    true,
		},
	}
	
	for _, item := range shopItems {
		if err := DB.Create(&item).Error; err != nil {
			log.Printf("Error creating shop item %s: %v", item.Name, err)
			continue
		}
		log.Printf("Created shop item: %s (ID: %d)", item.Name, item.ID)
	}
}

// seedUserItems creates sample user items for admin user
func seedUserItems() {
	log.Println("Seeding user items for admin...")
	
	// Get admin user
	var adminUser models.User
	if err := DB.Where("email = ?", "admin@map-memories.com").First(&adminUser).Error; err != nil {
		log.Printf("Error finding admin user: %v", err)
		return
	}
	
	// Check if user items already exist for admin
	var count int64
	DB.Model(&models.UserItem{}).Where("user_id = ?", adminUser.ID).Count(&count)
	if count > 0 {
		log.Println("User items for admin already exist, skipping...")
		return
	}
	
	// Get all shop items
	var shopItems []models.ShopItem
	if err := DB.Find(&shopItems).Error; err != nil {
		log.Printf("Error fetching shop items: %v", err)
		return
	}
	
	// Create user items for admin (give admin 1 of each marker - one-time purchase)
	userItems := []models.UserItem{
		{
			UserID:     adminUser.ID,
			ShopItemID: shopItems[0].ID, // Red Star Marker
			Quantity:   1,
		},
		{
			UserID:     adminUser.ID,
			ShopItemID: shopItems[1].ID, // Blue Diamond Marker
			Quantity:   1,
		},
		{
			UserID:     adminUser.ID,
			ShopItemID: shopItems[2].ID, // Green Heart Marker
			Quantity:   1,
		},
	}
	
	for _, userItem := range userItems {
		if err := DB.Create(&userItem).Error; err != nil {
			log.Printf("Error creating user item: %v", err)
			continue
		}
		log.Printf("Created user item for admin: %d x ShopItem ID %d (one-time purchase)", userItem.Quantity, userItem.ShopItemID)
	}
} 