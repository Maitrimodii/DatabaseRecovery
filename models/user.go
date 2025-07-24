package models

import (
	"context"
	"databaseRecovery/database"
	"gorm.io/gorm"
	"log"
	"time"
)

// User represents a sample model
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"type:varchar(100)"`
	Email     string `gorm:"type:varchar(100);unique"`
	CreatedAt time.Time
}

// Database operations

// CreateUser creates a new user
func CreateUser(ctx context.Context, name, email string) error {
	db, err := database.GetDB() // Get the database instance with recovery

	if err != nil {
		return err
	}
	user := User{Name: name, Email: email}
	return db.WithContext(ctx).Create(&user).Error
}

func GetUsers(ctx context.Context) ([]User, error) {
	db, err := database.GetDB() // Get the database instance with recovery

	if err != nil {
		return nil, err
	}
	var user User
	db.Find(&user)

	return []User{user}, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(ctx context.Context, id uint) (*User, error) {
	db, err := database.GetDB() // Get the database instance with recovery

	if err != nil {
		return nil, err
	}
	var user User
	err = db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserEmail updates a user's email by ID
func UpdateUserEmail(ctx context.Context, id uint, newEmail string) error {
	db, err := database.GetDB() // Get the database instance with recovery

	if err != nil {
		return err
	}
	return db.WithContext(ctx).Model(&User{}).Where("id = ?", id).Update("email", newEmail).Error
}

func DeleteUsers(ctx context.Context) error {
	db, err := database.GetDB() // Get the database instance with recovery

	if err != nil {
		return err
	}
	return db.WithContext(ctx).Delete(&User{}).Error
}

// DeleteUser deletes a user by ID
func DeleteUser(ctx context.Context, id uint) error {
	db, err := database.GetDB() // Get the database instance with recovery

	if err != nil {
		return err
	}
	return db.WithContext(ctx).Delete(&User{}, id).Error
}

// Retry executes a database operation with retries
func Retry(ctx context.Context, operation func(*gorm.DB) error, maxRetries int, retryInterval time.Duration) error {
	var lastErr error
	db, err := database.GetDB() // Get the database instance with recovery

	if err != nil {
		return err
	}
	for i := 0; i < maxRetries; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := operation(db)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Printf("Operation failed (attempt %d/%d): %v", i+1, maxRetries, err)

		// Check if error is retryable (customize based on your database)
		if !isRetryableError(err) {
			return err
		}

		// Attempt global reconnection before next retry
		if err := database.DbInstance.Reconnect(); err != nil {
			log.Printf("Reconnection failed: %v", err)
		}
		db, err = database.GetDB() // Get updated DB instance

		if err != nil {
			return err
		}
		time.Sleep(retryInterval * time.Duration(i+1)) // Exponential backoff
	}

	return lastErr
}

// isRetryableError determines if the error is retryable
func isRetryableError(err error) bool {
	// Customize for your database (e.g., PostgreSQL)
	return err.Error() == "dial tcp: i/o timeout" || err.Error() == "connection refused"
}
