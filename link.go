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

func GetLinks(db *gorm.DB, page int, count int) *[]Link {
	var links []Link

	if page < 1 {
		page = 1
	}

	if count < 1 {
		count = 50
	}

	offset := (page - 1) * count

	db.Order("saved_at").Limit(count).Offset(offset).Find(&links)

	return &links
}

func GetLinkByID(db *gorm.DB, id uint) *Link {
	var link Link

	db.First(&link, id)

	return &link
}
