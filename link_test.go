package main

import (
	"log"
	"os"
	"testing"

	"github.com/benjamineskola/bookmarks/config"
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

	config.Config = config.MakeConfig()
	config.Config.URLNormalisations.AddWWW = []string{"theguardian.com"}

	result := m.Run()

	database.DB.Exec("DELETE FROM links")

	os.Exit(result)
}

func TestLink(t *testing.T) {
	t.Parallel()

	link := NewLink("http://example.com/", "Example Website", "TestLink example", false)

	assert.Equal(t, "http://example.com/", link.URL.String())
}

func TestLinkTags(t *testing.T) {
	t.Parallel()

	link := NewLink("http://example.com/", "Example Website", "TestLinkTags example", false)
	tags := map[string]struct{}{"foo": {}, "bar": {}}
	tl := TagList(tags)
	link.Tags = &tl

	id, _ := link.Save()

	actual := GetLinkByID(id)
	expected := TagList(map[string]struct{}{"foo": {}, "bar": {}})
	assert.Equal(t, &expected, actual.Tags)
}

func TestGetLinkByURLNormalises(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Input    string
		Expected string
	}{
		{Input: "https://theguardian.com", Expected: "https://www.theguardian.com"},
		{Input: "https://www.theguardian.com", Expected: "https://www.theguardian.com"},
		{Input: "https://nottheguardian.com", Expected: ""},
	}

	link := NewLink("https://www.theguardian.com", "Example Website", "TestGetLinkByURLNormalises example", false)
	_, err := link.Save()
	assert.Nil(t, err)

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Input, func(t *testing.T) {
			t.Parallel()

			actual := GetLinkByURL(testCase.Input)

			if testCase.Expected == "" {
				assert.Nil(t, actual.URL)
			} else {
				assert.NotNil(t, actual.URL)
				assert.Equal(t, testCase.Expected, actual.URL.String())
			}
		})
	}
}

func TestGetLinkByURLNormalisesSlashes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Input    string
		Expected string
	}{
		{Input: "https://normaliseslashwith.com/", Expected: "https://normaliseslashwith.com/"},
		{Input: "https://normaliseslashwith.com", Expected: "https://normaliseslashwith.com/"},
		{Input: "https://normaliseslashwithout.com", Expected: "https://normaliseslashwithout.com"},
		{Input: "https://normaliseslashwithout.com/", Expected: "https://normaliseslashwithout.com"},
	}

	link := NewLink("https://normaliseslashwith.com/",
		"Example Website", "TestGetLinkByURLNormalisesSlashes example", false)
	_, err := link.Save()
	assert.Nil(t, err)

	link = NewLink("https://normaliseslashwithout.com",
		"Example Website", "TestGetLinkByURLNormalisesSlashes example", false)
	_, err = link.Save()
	assert.Nil(t, err)

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Input, func(t *testing.T) {
			t.Parallel()

			actual := GetLinkByURL(testCase.Input)

			assert.NotNil(t, actual.URL)
			assert.Equal(t, testCase.Expected, actual.URL.String())
		})
	}
}
