package main

import (
	"time"

	"gorm.io/gorm"
)

type Link struct {
	gorm.Model

	URL         string
	Title       string
	Description string
	SavedAt     time.Time
	ReadAt      time.Time
}

func NewLink(url string, title string, description string) *Link {
	link := Link{ //nolint:exhaustruct
		URL:         url,
		Title:       title,
		Description: description,
	}

	return &link
}
