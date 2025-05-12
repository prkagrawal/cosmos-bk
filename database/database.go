package database

import (
	"fmt"
	"os"

	"github.com/prkagrawal/cosmos-bk2/graph/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() error {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	return nil
}

func Migrate() error {
	// First migrate all other tables
	err := DB.AutoMigrate(
		&model.User{},
		&model.Skill{},
		&model.Cause{},
		&model.Nonprofit{},
		&model.Project{},
		&model.Application{},
		&model.Engagement{},
		&model.HoursLogged{},
	)
	if err != nil {
		return err
	}

	// Manually handle the days_available column conversion
	return DB.Transaction(func(tx *gorm.DB) error {
		// Add new temporary column
		if err := tx.Exec(`
					ALTER TABLE users 
					ADD COLUMN days_available_temp jsonb
			`).Error; err != nil {
			return err
		}

		// Convert data from old column to new
		if err := tx.Exec(`
					UPDATE users 
					SET days_available_temp = to_jsonb(days_available::text::jsonb)
			`).Error; err != nil {
			return err
		}

		// Drop old column
		if err := tx.Exec(`
					ALTER TABLE users 
					DROP COLUMN days_available
			`).Error; err != nil {
			return err
		}

		// Rename temporary column
		return tx.Exec(`
					ALTER TABLE users 
					RENAME COLUMN days_available_temp TO days_available
			`).Error
	})
}
