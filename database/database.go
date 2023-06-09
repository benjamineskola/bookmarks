package database

import (
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3" // doesn't need to be referenced
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB //nolint:gochecknoglobals

func InitDatabase() *gorm.DB {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	db, _ := gorm.Open(sqlite.Open(fmt.Sprintf("data/%s.sqlite3", env)))

	return db
}
