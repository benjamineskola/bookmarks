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
	links := GetLinks(database.DB, pageNumber, 0)

	switch urlFormat {
	case "json":
		renderJSON(w, links)
	default:
		w.Write([]byte("not implemented"))
	}
}

func showHandler(w http.ResponseWriter, r *http.Request) {
	urlFormat, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	linkID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	link := GetLinkByID(database.DB, uint(linkID))

	switch urlFormat {
	case "json":
		if link.ID == 0 {
			w.Header().Set("Content-Type", "application/json")
			result := renderJSONError(w, nil, http.StatusNotFound)
			w.Write(result)
		} else {
			renderJSON(w, link)
		}

	default:
		w.Write([]byte("not implemented"))
	}
}

func renderJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")

	result, err := json.Marshal(data)
	if err != nil {
		result = renderJSONError(w, err, 0)
	}

	w.Write(result)
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
