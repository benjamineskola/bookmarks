package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/benjamineskola/bookmarks/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func noopHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("hello world\n"))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	urlFormat, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	pageNumber, _ := strconv.Atoi(chi.URLParam(r, "page"))

	switch urlFormat {
	case "json":
		w.Header().Set("Content-Type", "application/json")

		links := GetLinks(database.DB, pageNumber, 0)

		result, err := json.Marshal(links)
		if err != nil {
			result = renderJSONError(w, err, 0)
		}

		w.Write(result)
	default:
		w.Write([]byte("not implemented"))
	}
}

func showHandler(w http.ResponseWriter, r *http.Request) {
	urlFormat, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	linkID, _ := strconv.Atoi(chi.URLParam(r, "id"))

	switch urlFormat {
	case "json":
		w.Header().Set("Content-Type", "application/json")

		link := GetLinkByID(database.DB, uint(linkID))

		var (
			result []byte
			err    error
			status int
		)

		if link.ID == 0 {
			status = http.StatusNotFound
		} else {
			result, err = json.Marshal(link)
		}

		if err != nil || status > 0 {
			result = renderJSONError(w, err, status)
		}

		w.Write(result)
	default:
		w.Write([]byte("not implemented"))
	}
}

func renderJSONError(w http.ResponseWriter, err error, status int) []byte {
	var message string

	if status == 0 {
		status = http.StatusBadRequest
	} else {
		message = http.StatusText(status)
	}

	if err != nil {
		message = fmt.Sprintf("%s", err)
	}

	if message == "" {
		message = "Unknown error"
	}

	w.WriteHeader(status)

	return []byte(fmt.Sprintf("{\"status\": %d, \"message\": \"%s\"}", status, message))
}
