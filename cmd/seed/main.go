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

	// Run database seeding
	database.SeedData()

	log.Println("Seed data completed!")
} 