package main

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TagList []string

func (tl *TagList) Scan(src any) error {
	var source string
	switch s := src.(type) {
	case string:
		source = s
	default:
		return errors.New("Incompatible type")
	}

	trimmed := strings.Trim(source, "{}")
	split := strings.Split(trimmed, ",")

	*tl = split

	return nil
}

func (tl TagList) Value() (driver.Value, error) {
	return fmt.Sprintf("{%s}", strings.Join(tl, ",")), nil
}

type Link struct {
	gorm.Model

	URL         *datatypes.URL
	Title       string
	Description string
	SavedAt     time.Time
	ReadAt      time.Time
	Public      bool
	Tags        *TagList
}

func NewLink(urlString string, title string, description string, public bool) *Link {
	parsedURL, _ := url.Parse(urlString)
	gormURL := datatypes.URL(*parsedURL)
	link := Link{ //nolint:exhaustruct
		URL:         &gormURL,
		Title:       title,
		Description: description,
		Public:      public,
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

	db.Order("saved_at desc").Limit(count).Offset(offset).Find(&links)

	return &links
}

func GetPublicLinks(db *gorm.DB, page int, count int) *[]Link {
	var links []Link

	if page < 1 {
		page = 1
	}

	if count < 1 {
		count = 50
	}

	offset := (page - 1) * count

	db.Where("public = ?", true).Order("saved_at desc").Limit(count).Offset(offset).Find(&links)

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

func (l Link) Save(db *gorm.DB) (uint, error) {
	result := db.Save(&l)

	return l.ID, result.Error
}
