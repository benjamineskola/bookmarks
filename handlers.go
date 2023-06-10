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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	urlFormat, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	pageNumber, _ := strconv.Atoi(chi.URLParam(r, "page"))

	switch urlFormat {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		links := GetLinks(database.DB, pageNumber, 0)
		result, err := json.Marshal(links)
		if err != nil {
			result = []byte(fmt.Sprintf("{\"status\": %d, \"message\": \"%s\"}", http.StatusBadRequest, err))
			w.WriteHeader(http.StatusBadRequest)
		}

		w.Write(result)
	default:
		w.Write([]byte("not implemented"))
	}
}
