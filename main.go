package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	m "github.com/RedHatInsights/chrome-service-backend/rest/middleware"
	"github.com/RedHatInsights/chrome-service-backend/rest/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	godotenv.Load()
	flag.Parse()
	initDependencies()
	cfg := config.Get()
	router := chi.NewRouter()
	metricsRouter := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/health", HealthProbe)

	router.With(routes.PrometheusMiddleware).Route("/api/chrome-service/v1/", func(subrouter chi.Router) {
		subrouter.Use(m.ParseHeaders)
		subrouter.Use(m.InjectUser)
		subrouter.Get("/hello-world", HelloWorld)
		subrouter.Route("/last-visited", routes.MakeLastVisitedRoutes)
		subrouter.Route("/favorite-pages", routes.MakeFavoritePagesRoutes)
		subrouter.Route("/self-report", routes.MakeSelfReportRoutes)
		subrouter.Route("/user", routes.MakeUserIdentityRoutes)
	})

	router.Route("/wss/chrome-service/v1/", func(subrouter chi.Router) {
		subrouter.Route("/ws", routes.MakeWsRoute)
	})

	metricsRouter.Handle("/metrics", promhttp.Handler())

	go func() {
		metricsStringAddr := fmt.Sprintf(":%s", strconv.Itoa(cfg.MetricsPort))
		if err := http.ListenAndServe(metricsStringAddr, metricsRouter); err != nil {
			log.Fatalf("Metrics server stopped %v", err)
		}
	}()

	serverStringAddr := fmt.Sprintf(":%s", strconv.Itoa(cfg.WebPort))
	if err := http.ListenAndServe(serverStringAddr, router); err != nil {
		log.Fatalf("Chrome-service-api has stopped due to %v", err)
	}
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
