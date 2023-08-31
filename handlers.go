package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/benjamineskola/bookmarks/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"gorm.io/datatypes"
)

type TemplateContext struct {
	Authenticated   bool
	CSRFTemplateTag template.HTML
	CurrentPage     int
	LastPage        int
	PrevPage        int
	NextPage        int
}

type SingleTemplateContext struct {
	TemplateContext
	Link *Link
}

type MultiTemplateContext struct {
	TemplateContext
	Links *[]Link
}

var (
	indexTmpl *template.Template //nolint:gochecknoglobals
	showTmpl  *template.Template //nolint:gochecknoglobals
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	urlFormat, _ := r.Context().Value(middleware.URLFormatCtxKey).(string)
	onlyPublic, _ := r.Context().Value("onlyPublic").(bool)
	onlyRead, _ := r.Context().Value("onlyRead").(bool)
	pageNumber, _ := strconv.Atoi(chi.URLParam(r, "page"))

	if pageNumber == 0 {
		pageNumber = 1
	}

	links := GetLinks(database.DB, pageNumber, 0, onlyPublic, onlyRead)

	authenticated := isAuthenticated(r)

	switch urlFormat {
	case "json":
		renderJSON(w, links)
	default:
		if indexTmpl == nil {
			indexTmpl = template.Must(template.ParseFiles("templates/index.html", "templates/base.html"))
		}

		ctx := MultiTemplateContext{Links: links} //nolint:exhaustruct
		ctx.Authenticated = authenticated
		ctx.CurrentPage = pageNumber
		ctx.NextPage = pageNumber + 1
		ctx.PrevPage = pageNumber - 1

		var totalLinks int64

		database.DB.Model(&Link{}).Count(&totalLinks)
		ctx.LastPage = int(math.Ceil(float64(totalLinks) / 50))

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

			_, err := w.Write(result)
			if err != nil {
				log.Panicf("could not write output: %s", err)
			}
		} else {
			renderJSON(w, link)
		}

	default:
		if showTmpl == nil {
			showTmpl = template.Must(template.ParseFiles("templates/show.html", "templates/base.html"))
		}

		ctx := SingleTemplateContext{Link: link} //nolint:exhaustruct
		ctx.Authenticated = true

		err := showTmpl.ExecuteTemplate(w, "base.html", ctx)
		if err != nil {
			log.Printf("error rendering template: %s", err)
		}
	}
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	linkID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	formTmpl := template.Must(template.ParseFiles("templates/form.html", "templates/base.html"))

	link := &Link{}
	if linkID != 0 {
		link = GetLinkByID(database.DB, uint(linkID))
	}

	ctx := SingleTemplateContext{Link: link} //nolint:exhaustruct
	ctx.Authenticated = true
	ctx.CSRFTemplateTag = csrf.TemplateField(r)

	err := formTmpl.ExecuteTemplate(w, "base.html", ctx)
	if err != nil {
		log.Printf("error rendering template: %s", err)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	linkID, _ := strconv.Atoi(chi.URLParam(r, "id"))

	link := &Link{} //nolint:exhaustruct
	if linkID != 0 {
		link = GetLinkByID(database.DB, uint(linkID))
	}

	r.ParseForm()

	parsedURL, _ := url.Parse(r.FormValue("Link.URL"))
	gormURL := datatypes.URL(*parsedURL)
	link.URL = &gormURL
	link.Title = r.FormValue("Link.Title")
	link.Description = r.FormValue("Link.Description")
	link.Public = r.FormValue("Link.Public") == "on"

	if link.IsRead() {
		if r.FormValue("mark_unread") == "on" {
			link.ReadAt = time.Time{}
		}
	} else {
		if r.FormValue("mark_read") == "now" {
			link.ReadAt = time.Now()
		} else if r.FormValue("mark_read") == "sometime" {
			link.ReadAt = time.Unix(0, 0)
		}
	}

	if link.SavedAt.IsZero() {
		link.SavedAt = time.Now()
	}

	_, err := link.Save(database.DB)
	if err != nil {
		log.Panicf("could not save record: %s", err)
	}

	http.Redirect(w, r, "/links/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	linkID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	link := GetLinkByID(database.DB, uint(linkID))

	if link.ID == 0 {
		w.Header().Set("Content-Type", "application/json")
		result := renderJSONError(w, nil, http.StatusNotFound)

		_, err := w.Write(result)
		if err != nil {
			log.Panicf("could not write output: %s", err)
		}
	} else {
		database.DB.Delete(&Link{}, link.ID)
		result := map[string]string{}
		result["result"] = "success"
		renderJSON(w, result)
	}
}

func renderJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")

	result, err := json.Marshal(data)
	if err != nil {
		result = renderJSONError(w, err, 0)
	}

	_, err = w.Write(result)
	if err != nil {
		log.Panicf("could not write output: %s", err)
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

func cookiejar() *sessions.CookieStore {
	SecretKey := []byte(os.Getenv("SECRET_KEY"))

	return sessions.NewCookieStore(SecretKey)
}

func loginFormHandler(w http.ResponseWriter, r *http.Request) {
	formTmpl := template.Must(template.ParseFiles("templates/login_form.html", "templates/base.html"))

	ctx := map[string]interface{}{
		"Authenticated":   isAuthenticated(r),
		"CSRFTemplateTag": csrf.TemplateField(r),
	}

	err := formTmpl.ExecuteTemplate(w, "base.html", ctx)
	if err != nil {
		log.Printf("error rendering template: %s", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := cookiejar().Get(r, "authenticated")
	if err != nil {
		log.Printf("WARN: could not read cookies: %s", err)
	}

	r.ParseForm()
	user, _ := GetValidatedUser(r.FormValue("email"), r.FormValue("password"))
	session.Values["authenticated"] = (user != nil)
	err = session.Save(r, w)

	if err != nil {
		log.Printf("WARN: could not save session: %s", err)

		return
	}

	http.Redirect(w, r, "/links/", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := cookiejar().Get(r, "authenticated")
	if err != nil {
		return
	}

	// nullify the user's session from the cookie Store
	session.Values["authenticated"] = nil
	err = session.Save(r, w)

	c := &http.Cookie{
		Name:     "authenticated",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(w, c)

	if err != nil {
		return
	}
}

func isAuthenticated(r *http.Request) bool {
	session, err := cookiejar().Get(r, "authenticated")
	if err != nil {
		session.Values["authenticated"] = nil

		return false
	}

	return session.Values["authenticated"] == true
}
