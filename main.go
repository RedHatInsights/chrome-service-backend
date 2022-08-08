package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	m "github.com/RedHatInsights/chrome-service-backend/rest/middleware"
	"github.com/RedHatInsights/chrome-service-backend/rest/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	godotenv.Load()
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	flag.Parse()
	initDependencies()
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/health", HealthProbe)

	router.Route("/api/chrome/v1/", func(subrouter chi.Router) {
		subrouter.Use(m.ParseHeaders)
		subrouter.Get("/hello-world", HelloWorld)
		subrouter.Route("/last-visited", routes.MakeLastVisitedRoutes)
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
