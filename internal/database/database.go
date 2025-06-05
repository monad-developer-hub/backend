package database

import (
	"log"

	"monad-devhub-be/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initialize creates a new database connection
func Initialize(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	log.Println("Database connection established")
	return db, nil
}

// Migrate runs all database migrations
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.Project{},
		&models.TeamMember{},
		&models.Submission{},
		&models.AnalyticsStats{},
		&models.Transaction{},
		&models.Contract{},
		&models.ContractStats{},
		&models.AdminUser{},
	)

	if err != nil {
		return err
	}

	log.Println("Database migrations completed")
	return nil
}
