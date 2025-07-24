package main

import (
	"context"
	"databaseRecovery/constants"
	"databaseRecovery/database"
	"databaseRecovery/models"
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

func main() {
	// Database configuration
	config := database.DBConfig{
		DSN:             constants.Dns,
		MaxRetries:      constants.MaxRetries,
		RetryInterval:   time.Second,
		MaxIdleConns:    constants.MaxIdleConns,
		MaxOpenConns:    constants.MaxOpenConns,
		ConnMaxLifetime: time.Hour,
	}

	// Initialize database
	_, err := database.InitDB(config)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a context with timeout for operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//time.Sleep(1 * time.Minute)
	err = models.CreateUser(ctx, "John Doe", "john@example.com")
	if err != nil {
		log.Printf("Failed to create user: %v", err)
	} else {
		log.Println("User created successfully")
	}

	err = models.Retry(ctx, func(db *gorm.DB) error {
		user := models.User{Name: "Jane Doe", Email: "jane@example.com"}
		return db.WithContext(ctx).Create(&user).Error
	}, 3, time.Second)
	if err != nil {
		log.Printf("Failed to create user with retry: %v", err)
	} else {
		log.Println("User created successfully with retry")
	}

	users, err := models.GetUsers(ctx)
	if err != nil {
		log.Printf("Failed to get users: %v", err)
	} else {
		fmt.Println(users)
	}

	//time.Sleep(20 * time.Second)

	user, err := models.GetUserByID(ctx, 2)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
	} else {
		log.Printf("Got user: %+v", user)
	}

	err = models.UpdateUserEmail(ctx, 1, "john.doe@example.com")
	if err != nil {
		log.Printf("Failed to update user: %v", err)
	} else {
		log.Println("User email updated successfully")
	}

	// Example 5: Delete a user
	err = models.DeleteUsers(ctx)

	if err != nil {
		log.Printf("Failed to delete users: %v", err)
	} else {
		log.Println("Users deleted successfully")
	}

}
