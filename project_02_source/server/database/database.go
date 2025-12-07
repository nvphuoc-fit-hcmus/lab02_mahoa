package database

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the SQLite database and runs migrations
func InitDB(models ...interface{}) error {
	var err error

	// Open SQLite database
	DB, err = gorm.Open(sqlite.Open("storage/app.db"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the database schema
	if len(models) > 0 {
		err = DB.AutoMigrate(models...)
		if err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
	}

	log.Println("✅ Database initialized successfully")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// InitTestDB initializes an in-memory SQLite database for testing
func InitTestDB(models ...interface{}) error {
	var err error

	// Use in-memory database for tests
	DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Auto-migrate the database schema
	if len(models) > 0 {
		err = DB.AutoMigrate(models...)
		if err != nil {
			return fmt.Errorf("failed to migrate test database: %w", err)
		}
	}

	log.Println("✅ Test database initialized successfully")
	return nil
}
