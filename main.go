package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/benjamineskola/bookmarks/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

func serve() {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.GetHead)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	csrfKey := []byte(os.Getenv("CSRF_KEY"))

	if host == "" {
		host = "0.0.0.0"
	}

	if port == "" {
		port = "8080"
	}

	if len(csrfKey) == 0 {
		log.Fatalf("No CSRF_KEY set")
	} else if len(csrfKey) != 32 {
		log.Fatalf("CSRF_KEY must be 32 bytes")
	}

	csrfMiddleware := csrf.Protect(csrfKey,
		csrf.Secure(host != "127.0.0.1"),
		csrf.Path("/"),
	)

	router.Use(csrfMiddleware)

	database.DB = database.InitDatabase()

	err := database.RunMigrations()
	if err != nil {
		log.Fatalf("failed to migrate database: %s", err)
	}

	router.Get("/auth/login/", loginFormHandler)
	router.Post("/auth/login/", loginHandler)
	router.Post("/auth/logout/", logoutHandler)

	router.Route("/links", func(router chi.Router) {
		router.Use(middleware.URLFormat)

		router.Use(middleware.Maybe(middleware.WithValue("onlyPublic", true), func(r *http.Request) bool {
			return !isAuthenticated(r)
		},
		))

		router.Get("/", indexHandler)
		router.With(middleware.WithValue("onlyPublic", true)).Route("/public", func(router chi.Router) {
			router.Get("/", indexHandler)
			router.Get("/page/{page}", indexHandler)
		})
		router.With(middleware.WithValue("onlyRead", true)).Route("/read", func(router chi.Router) {
			router.Get("/", indexHandler)
			router.Get("/page/{page}", indexHandler)
		})

		router.Get("/page/{page}", indexHandler)

		router.Get("/new", formHandler)
		router.Post("/", saveHandler)

		router.Route("/{id}", func(router chi.Router) {
			router.Use(rejectUnauthenticated)

			router.Get("/", showHandler)
			router.Put("/", saveHandler)
			router.Post("/", saveHandler)
			router.Delete("/", deleteHandler)
			router.Get("/edit", formHandler)
		})
	})

	fs := http.FileServer(http.Dir("static"))
	router.Handle("/static/*", http.StripPrefix("/static/", fs))

	log.Printf("listening on %s:%s", host, port)

	err = http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), router)
	if err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}

func add() {
	database.DB = database.InitDatabase()
	err := database.RunMigrations()
	if err != nil {
		log.Fatalf("failed to migrate database: %s", err)
	}

	dec := json.NewDecoder(os.Stdin)
	var data []map[string]interface{}
	err = dec.Decode(&data)

	for _, item := range data {
		if url, ok := item["URL"].(string); ok {
			importer(url, item)
		}
	}
}

func main() {
	flag.Parse()
	args := flag.Args()

	cmd := "serve"
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "serve":
		serve()
	case "add":
		add()
	default:
		log.Fatalf("unknown command %q", cmd)
	}
}
