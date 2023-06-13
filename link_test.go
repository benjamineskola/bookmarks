package main

import (
	"log"
	"os"
	"testing"

	"github.com/benjamineskola/bookmarks/database"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	os.Setenv("ENVIRONMENT", "test")

	database.DB = database.InitDatabase()

	err := database.RunMigrations()
	if err != nil {
		log.Fatalf("failed to migrate database: %s", err)
	}

	database.DB.Exec("DELETE FROM links")

	result := m.Run()

	database.DB.Exec("DELETE FROM links")

	os.Exit(result)
}

func TestLink(t *testing.T) {
	t.Parallel()

	link := NewLink("http://example.com/", "Example Website", "This is just an example.", false)

	assert.Equal(t, "http://example.com/", link.URL.String())
}
