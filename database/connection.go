package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"time"
)

// DBConfig holds database connection details
type DBConfig struct {
	DSN             string
	MaxRetries      int
	RetryInterval   time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// Database holds the GORM DB instance
type Database struct {
	*gorm.DB
	config DBConfig
}

// Global database instance
var DbInstance *Database

// InitDB initializes the database with a retry mechanism
func InitDB(config DBConfig) (*Database, error) {
	var db *gorm.DB
	var err error

	// Retry loop for initial connection
	for i := 0; i < config.MaxRetries; i++ {
		db, err = gorm.Open(postgres.Open(config.DSN), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %	v", i+1, config.MaxRetries, err)
		time.Sleep(config.RetryInterval * time.Duration(i+1)) // Exponential backoff
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d retries: %w", config.MaxRetries, err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DbInstance = &Database{DB: db, config: config}
	log.Println("Database connected successfully")
	return DbInstance, nil
}

// Reconnect attempts to reconnect to the database
func (d *Database) Reconnect() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.Close()

	// Reinitialize connection
	newDB, err := InitDB(d.config)
	if err != nil {
		return fmt.Errorf("reconnection failed: %w", err)
	}
	d.DB = newDB.DB
	log.Println("Database reconnected successfully")
	return nil
}

// GetDB returns the global database instance, reconnecting if necessary
func GetDB() (*gorm.DB, error) {
	if DbInstance == nil {
		log.Fatal("Database not initialized")
	}

	// Check connection health
	sqlDB, err := DbInstance.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		log.Println("Database connection lost, attempting to reconnect")
		if err := DbInstance.Reconnect(); err != nil {
			log.Printf("Failed to reconnect: %v", err)
			return nil, err
		}
	}

	return DbInstance.DB, nil
}
