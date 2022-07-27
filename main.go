package main

import (
	"flag"
	"net/http"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

var routes = flag.Bool("routes", false, "Generate router documentation")

func main() {
	godotenv.Load()
	flag.Parse()
	initDependencies()
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/health", HealthProbe)

	router.Route("/api/chrome/v1/", func(subrouter chi.Router) {
		subrouter.Get("/hello-world", HelloWorld)
	})

	http.ListenAndServe(":8000", router)
}

func HelloWorld(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("que lo que manin"))
}

func HealthProbe(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte("Why yes thank you, I am quite healthy :D"))
}

func initDependencies() {
	config.Init()
	database.Init()
}
