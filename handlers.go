package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/benjamineskola/bookmarks/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type TemplateContext struct {
	Authenticated bool
}

type SingleTemplateContext struct {
	TemplateContext
	Link *Link
}

type MultiTemplateContext struct {
	TemplateContext
	Links *[]Link
}

type LinkRequest struct {
	*Link
}

func (l *LinkRequest) Bind(r *http.Request) error {
	if l.Link == nil {
		return errors.New("link not defined? what")
	}

	return nil
}

var (
	indexTmpl *template.Template //nolint:gochecknoglobals
	showTmpl  *template.Template //nolint:gochecknoglobals
)

func noopHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("hello world\n"))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	urlFormat, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	onlyPublic, _ := r.Context().Value("onlyPublic").(bool)
	pageNumber, _ := strconv.Atoi(chi.URLParam(r, "page"))

	var links *[]Link

	if onlyPublic {
		links = GetPublicLinks(database.DB, pageNumber, 0)
	} else {
		links = GetLinks(database.DB, pageNumber, 0)
	}

	authenticated := r.Header.Get("Authorization") != ""

	switch urlFormat {
	case "json":
		renderJSON(w, links)
	default:
		if indexTmpl == nil {
			indexTmpl = template.Must(template.ParseFiles("templates/index.html", "templates/base.html"))
		}

		ctx := MultiTemplateContext{Links: links}
		ctx.Authenticated = authenticated

		err := indexTmpl.ExecuteTemplate(w, "base.html", ctx)
		if err != nil {
			log.Printf("error rendering template: %s", err)
		}
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
		if showTmpl == nil {
			showTmpl = template.Must(template.ParseFiles("templates/show.html", "templates/base.html"))
		}

		ctx := SingleTemplateContext{Link: link}
		ctx.Authenticated = true

		err := showTmpl.ExecuteTemplate(w, "base.html", ctx)
		if err != nil {
			log.Printf("error rendering template: %s", err)
		}
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	data := &LinkRequest{}
	if err := render.Bind(r, data); err != nil {
		w.WriteHeader(500)
		return
	}

	link := data.Link
	link.Save(database.DB)

	render.Status(r, http.StatusCreated)
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
