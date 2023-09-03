package main

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/benjamineskola/bookmarks/database"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var errIncompatibleType = errors.New("incompatible type")

type TagList map[string]struct{}

func NewTagListFromString(src string) TagList {
	trimmed := strings.Trim(src, "{}")
	tags := strings.Split(trimmed, ",")

	tl := make(TagList, len(tags))
	for _, tag := range tags {
		tl[tag] = struct{}{}
	}

	return tl
}

func (tl *TagList) Scan(src any) error {
	switch source := src.(type) {
	case string:
		*tl = NewTagListFromString(source)
	default:
		return errIncompatibleType
	}

	return nil
}

func (tl TagList) Value() (driver.Value, error) { //nolint:unparam
	tags := make([]string, 0, len(tl))
	for tag := range tl {
		tags = append(tags, tag)
	}

	return fmt.Sprintf("{%s}", strings.Join(tags, ",")), nil
}

func (tl *TagList) Merge(other TagList) {
	for tag := range other {
		(*tl)[tag] = struct{}{}
	}
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
	link := Link{ //nolint:exhaustruct
		URL:         parseURL(urlString),
		Title:       title,
		Description: description,
		Public:      public,
	}

	return &link
}

func GetLinks(page int, count int, onlyPublic bool, onlyRead bool) (*[]Link, int64) {
	var links []Link

	if page < 1 {
		page = 1
	}

	if count < 1 {
		count = 50
	}

	offset := (page - 1) * count

	query := database.DB

	if onlyPublic {
		query = query.Where("public = ?", true)
	}

	if onlyRead {
		query = query.Where("read_at >= ?", time.Unix(0, 0)).Order("read_at desc")
	} else {
		query = query.Order("saved_at desc")
	}

	var totalCount int64

	query.Model(&Link{}).Count(&totalCount) //nolint:exhaustruct
	query = query.Limit(count).Offset(offset)
	query.Find(&links)

	return &links, totalCount
}

func GetLinkByID(id uint) *Link {
	var link Link

	database.DB.First(&link, id)

	return &link
}

func GetLinkByURL(url string) *Link {
	var link Link

	normalisedURL := normaliseURLString(url)

	database.DB.Where("url = ?", normalisedURL).First(&link)

	return &link
}

func (l Link) IsRead() bool {
	return !l.ReadAt.IsZero()
}

func (l Link) HasReadDate() bool {
	return l.ReadAt.Unix() > 0
}

func (l Link) Save() (uint, error) {
	result := database.DB.Save(&l)

	return l.ID, result.Error
}
