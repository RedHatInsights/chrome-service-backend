package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	fs := http.FileServer(http.Dir("./static/"))
	router.Handle("/api/chrome-service/v1/static/*", http.StripPrefix("/api/chrome-service/v1/static", fs))
	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatalf("Chrome-service-api has stopped due to %v", err)
	}
}
