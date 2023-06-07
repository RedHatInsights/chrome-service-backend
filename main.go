package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/featureflags"
	"github.com/RedHatInsights/chrome-service-backend/rest/kafka"
	m "github.com/RedHatInsights/chrome-service-backend/rest/middleware"
	"github.com/RedHatInsights/chrome-service-backend/rest/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func init() {
	godotenv.Load()
	flag.Parse()
	database.Init()
}

func main() {
	cfg := config.Get()
	featureflags.Init(cfg)
	setupLogger(cfg)
	router := chi.NewRouter()
	metricsRouter := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/health", HealthProbe)
	fs := http.FileServer(http.Dir("./static/"))

	// can't be in sub router as we don't enforce identity header
	router.Handle("/api/chrome-service/v1/static/*", http.StripPrefix("/api/chrome-service/v1/static", fs))

	router.Route("/api/chrome-service/v1/", func(subrouter chi.Router) {
		subrouter.Use(m.ParseHeaders)
		subrouter.Use(m.InjectUser)
		subrouter.Get("/hello-world", HelloWorld)
		subrouter.Route("/last-visited", routes.MakeLastVisitedRoutes)
		subrouter.Route("/favorite-pages", routes.MakeFavoritePagesRoutes)
		subrouter.Route("/self-report", routes.MakeSelfReportRoutes)
		subrouter.Route("/user", routes.MakeUserIdentityRoutes)
		subrouter.Route("/emit-message", routes.BroadcastMessage)
	})

	// We might want to setup some event listeners at some point, but the pod will
	// have to restart for these to take effect. We can't enable and disable websockets on the fly
	if featureflags.IsEnabled("chrome-service.websockets.enabled") {
		// start the connection hub
		go connectionhub.ConnectionHub.Run()
		logrus.Infoln("Enabling WebSockets")
		kafka.InitializeConsumers()
		router.Route("/wss/chrome-service/v1/", func(subrouter chi.Router) {
			subrouter.Route("/ws", routes.MakeWsRoute)
		})
	} else {
		logrus.Infoln("WebSockets are currently disabled")
	}

	metricsRouter.Handle("/metrics", promhttp.Handler())

	go func() {
		metricsStringAddr := fmt.Sprintf("localhost:%s", strconv.Itoa(cfg.MetricsPort))
		if err := http.ListenAndServe(metricsStringAddr, metricsRouter); err != nil {
			log.Fatalf("Metrics server stopped %v", err)
		}
	}()

	serverStringAddr := fmt.Sprintf("localhost:%s", strconv.Itoa(cfg.WebPort))
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

func setupLogger(opts *config.ChromeServiceConfig) {
	logLevel, err := logrus.ParseLevel(opts.LogLevel)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)
}
