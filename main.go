package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/benjamineskola/bookmarks/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.GetHead)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))

	expectedUser := os.Getenv("AUTH_USER")
	expectedPass := os.Getenv("AUTH_PASSWORD")

	if expectedUser == "" || expectedPass == "" {
		log.Fatal("no user and password defined")
	}

	log.Printf("enabling auth for user %s", expectedUser)
	auth := middleware.BasicAuth("bookmarks", map[string]string{expectedUser: expectedPass})

	database.DB = database.InitDatabase()

	err := database.RunMigrations()
	if err != nil {
		log.Fatalf("failed to migrate database: %s", err)
	}

	router.Route("/links", func(router chi.Router) {
		router.Use(middleware.URLFormat)
		router.Use(middleware.Maybe(auth, func(r *http.Request) bool {
			return !strings.HasPrefix(r.URL.Path, "/links/public")
		}))

		router.Get("/", indexHandler)
		router.With(middleware.WithValue("onlyPublic", true)).Route("/public", func(router chi.Router) {
			router.Get("/", indexHandler)
			router.Get("/page/{page}", indexHandler)
		})

		router.Get("/page/{page}", indexHandler)

		router.Get("/new", newFormHandler)
		router.Post("/", createHandler)

		router.Route("/{id}/", func(router chi.Router) {
			router.Get("/", showHandler)
			router.Put("/", noopHandler)
			router.Delete("/", noopHandler)
		})
	})

	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/*", http.StripPrefix("/static/", fs))

	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on %s:%s", host, port)

	err = http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), router)
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}
