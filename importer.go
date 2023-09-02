package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/benjamineskola/bookmarks/database"
)

var errUnhandledDate = errors.New("unhandled date type")

func parseJSONDate(input interface{}) (*time.Time, error) {
	if dateInt, ok := input.(float64); ok {
		date := time.Unix(int64(dateInt), 0)

		return &date, nil
	} else if dateStr, ok := input.(string); ok {
		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return nil, fmt.Errorf("could not parse date from json: %w", err)
		}

		return &date, nil
	}

	return nil, errUnhandledDate
}

func importer(url string, data map[string]interface{}) {
	link := GetLinkByURL(database.DB, url)

	changed := false

	if link.ID == 0 {
		link.URL = parseURL(url)
		changed = true

		tl := make(TagList)
		link.Tags = &tl
	}

	if title, ok := data["Title"].(string); ok {
		if title != "" && link.Title != title {
			changed = true
			link.Title = title
		}
	}

	if description, ok := data["Description"].(string); ok {
		if description != "" && link.Description != description {
			changed = true
			link.Description = description
		}
	}

	if readAt, err := parseJSONDate(data["ReadAt"]); err == nil {
		if !link.ReadAt.Equal(*readAt) {
			changed = true
			link.ReadAt = *readAt
		}
	}

	if savedAt, err := parseJSONDate(data["SavedAt"]); err == nil {
		if !link.SavedAt.Equal(*savedAt) {
			changed = true
			link.SavedAt = *savedAt
		}
	}

	if tagsStr, ok := data["Tags"].(string); ok {
		tl := TagList{}
		tl.Scan(tagsStr)

		for tag := range tl {
			(*link.Tags)[tag] = struct{}{}
		}
	}

	if changed {
		link.Save(database.DB)
	}
}
