package main

import (
	"net/url"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Link struct {
	gorm.Model

	URL         *datatypes.URL
	Title       string
	Description string
	SavedAt     time.Time
	ReadAt      time.Time
}

func NewLink(urlString string, title string, description string) *Link {
	parsedURL, _ := url.Parse(urlString)
	gormURL := datatypes.URL(*parsedURL)
	link := Link{ //nolint:exhaustruct
		URL:         &gormURL,
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

func (l Link) IsRead() bool {
	return !l.ReadAt.IsZero()
}

func (l Link) HasReadDate() bool {
	return l.ReadAt.Unix() > 0
}
