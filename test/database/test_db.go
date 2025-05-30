package database

import (
	"fmt"
	"gorm.io/gorm"
	"linked-clone/internal/config"
	"linked-clone/internal/domain/entities"
	"linked-clone/internal/infrastructure/database"
)

type TestDB struct {
	DB *gorm.DB
}

func NewTestDB(cfg *config.Config) (*TestDB, error) {
	db, err := database.NewPostgreSQLConnection(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	err = db.AutoMigrate(
		&entities.User{},
		&entities.Post{},
		&entities.Like{},
		&entities.Comment{},
		&entities.Job{},
		&entities.Application{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %w", err)
	}

	return &TestDB{DB: db}, nil
}

func (tdb *TestDB) Clean() error {

	tables := []string{
		"likes", "comments", "applications", "posts", "jobs", "users",
	}

	for _, table := range tables {
		if err := tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)).Error; err != nil {
			return fmt.Errorf("failed to clean table %s: %w", table, err)
		}
	}

	return nil
}

func (tdb *TestDB) Close() error {
	sqlDB, err := tdb.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
