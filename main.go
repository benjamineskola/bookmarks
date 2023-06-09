package main

import (
	"log"
	"net/http"

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
	router.Use(middleware.URLFormat)

	database.DB = database.InitDatabase()

	router.Get("/links", noopHandler)
	router.Get("/links/page/{page}", noopHandler)
	router.Get("/links/{id}", noopHandler)
	router.Put("/links/{id}", noopHandler)
	router.Delete("/links/{id}", noopHandler)
	router.Post("/links", noopHandler)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}

func noopHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("hello world\n"))
}
