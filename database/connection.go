package database

import (
	"log"
	"time"

	"map-memories-api/config"
	"map-memories-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establishes database connection
func Connect() {
	var err error

	// Configure GORM logger
	var gormLogger logger.Interface
	if config.AppConfig.Environment == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Connect to database
	DB, err = gorm.Open(postgres.Open(config.AppConfig.GetDSN()), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected successfully")
}

// AutoMigrate runs database migrations
func AutoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Location{},
		&models.Memory{},
		&models.Media{},
		&models.UserSession{},
		&models.MemoryLike{},
		&models.ShopItem{},
		&models.UserItem{},
		&models.TransactionLog{},
	)

	if err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	log.Println("Database migrations completed successfully")
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// Close closes the database connection
func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Println("Error getting database instance:", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Println("Error closing database connection:", err)
		return
	}

	log.Println("Database connection closed")
}

// Health checks database connection health
func Health() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}
