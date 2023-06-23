package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	port := "8000"
	if len(os.Args) > 1 && os.Args[1] != "" {
		port = os.Args[1]
	}
	fmt.Printf("Starting asset server server at \"http://localhost:%s\"\n", port)
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	fs := http.FileServer(http.Dir("./static/"))
	router.Handle("/api/chrome-service/v1/static/*", http.StripPrefix("/api/chrome-service/v1/static", fs))
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		log.Fatalf("Chrome-service-api has stopped due to %v", err)
	}
}
