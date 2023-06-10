package database

import (
	"errors"
	"fmt"
	"os"

	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3" // doesn't need to be referenced
	_ "github.com/golang-migrate/migrate/v4/source/file"      // doesn't need to be referenced
	_ "github.com/mattn/go-sqlite3"                           // doesn't need to be referenced
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB //nolint:gochecknoglobals

func getDBPath() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	return fmt.Sprintf("data/%s.sqlite3", env)
}

func InitDatabase() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(getDBPath()))

	return db
}

func RunMigrations() error {
	migrateDir := "file://migrations/"

	m, err := migrate.New(migrateDir, "sqlite3://"+getDBPath())
	if err != nil {
		return fmt.Errorf("failed to initialise migrations: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
