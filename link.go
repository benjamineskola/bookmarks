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

type TagList map[string]struct{}

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

	*tl = make(map[string]struct{}, len(split))
	for _, tag := range split {
		(*tl)[tag] = struct{}{}
	}

	return nil
}

func (tl TagList) Value() (driver.Value, error) {
	tags := make([]string, len(tl))
	for tag := range tl {
		tags = append(tags, tag)
	}

	return fmt.Sprintf("{%s}", strings.Join(tags, ",")), nil
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

func parseURL(urlString string) *datatypes.URL {
	parsedURL, _ := url.Parse(urlString)
	gormURL := datatypes.URL(*parsedURL)
	return &gormURL
}

func NewLink(urlString string, title string, description string, public bool) *Link {
	link := Link{ //nolint:exhaustruct
		URL:         parseURL(urlString),
		Title:       title,
		Description: description,
		Public:      public,
	}

	return &link
}

func GetLinks(db *gorm.DB, page int, count int, onlyPublic bool, onlyRead bool) *[]Link {
	var links []Link

	if page < 1 {
		page = 1
	}

	if count < 1 {
		count = 50
	}

	offset := (page - 1) * count

	query := db

	if onlyPublic {
		query = query.Where("public = ?", true)
	}

	if onlyRead {
		query = query.Where("read_at >= ?", 0).Order("read_at desc")
	} else {
		query = query.Order("saved_at desc")
	}

	query = query.Limit(count).Offset(offset)

	query.Find(&links)

	return &links
}

func GetLinkByID(db *gorm.DB, id uint) *Link {
	var link Link

	db.First(&link, id)

	return &link
}

func GetLinkByURL(db *gorm.DB, url string) *Link {
	var link Link

	db.Where("url = ?", url).First(&link)

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
