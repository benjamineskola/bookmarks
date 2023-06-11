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

	database.DB = database.InitDatabase()

	err := database.RunMigrations()
	if err != nil {
		log.Fatalf("failed to migrate database: %s", err)
	}

	router.Route("/links", func(router chi.Router) {
		router.Use(middleware.URLFormat)

		router.Get("/", indexHandler)
		router.Get("/page/{page}", indexHandler)
		router.Post("/", noopHandler)

		router.Route("/{id}/", func(router chi.Router) {
			router.Get("/", showHandler)
			router.Put("/", noopHandler)
			router.Delete("/", noopHandler)
		})
	})

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}
