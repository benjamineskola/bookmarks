package main

import (
	"errors"
	"fmt"
	"log"
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

func mergeStringField(orig string, field interface{}, changed bool) (string, bool) {
	if repl, ok := field.(string); ok {
		if repl != "" && orig != repl {
			return repl, true
		}
	}

	return orig, changed
}

func mergeDateField(orig time.Time, field interface{}, changed bool) (time.Time, bool) {
	if repl, err := parseJSONDate(field); err == nil {
		if !orig.Equal(*repl) && (orig.Unix() < 1 || repl.Unix() > 0) {
			return *repl, true
		}
	}

	return orig, changed
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

	link.Title, changed = mergeStringField(link.Title, data["Title"], changed)
	link.Description, changed = mergeStringField(link.Description, data["Description"], changed)

	link.ReadAt, changed = mergeDateField(link.ReadAt, data["ReadAt"], changed)
	link.SavedAt, changed = mergeDateField(link.SavedAt, data["SavedAt"], changed)

	if tagsStr, ok := data["Tags"].(string); ok {
		link.Tags.Merge(NewTagListFromString(tagsStr))
	}

	if changed {
		_, err := link.Save(database.DB)
		if err != nil {
			log.Fatalf("could not save link %q: %s", link.URL, err)
		}
	}
}
