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
	"github.com/RedHatInsights/chrome-service-backend/rest/logger"
	m "github.com/RedHatInsights/chrome-service-backend/rest/middleware"
	"github.com/RedHatInsights/chrome-service-backend/rest/routes"
	"github.com/RedHatInsights/chrome-service-backend/rest/service"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func init() {
	godotenv.Load()
	flag.Parse()
	database.Init()
	util.InitUserIdentitiesCache()
}

func main() {
	cfg := config.Get()
	service.LoadBaseLayout()
	featureflags.Init(cfg)
	setupGlobalLogger(cfg)
	router := chi.NewRouter()
	metricsRouter := chi.NewRouter()

	routerLogger := logrus.New()
	router.Use(middleware.RequestID)
	router.Use(middleware.RequestLogger(logger.NewLogger(cfg, routerLogger)))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/health", HealthProbe)
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/", fs)

	fsApiSpec := http.FileServer(http.Dir("./spec/"))
	http.Handle("/spec/", fsApiSpec)

	// can't be in sub router as we don't enforce identity header
	router.Handle("/api/chrome-service/v1/static/*", http.StripPrefix("/api/chrome-service/v1/static", fs))
	router.Handle("/api/chrome-service/v1/spec/*", http.StripPrefix("/api/chrome-service/v1/spec", fsApiSpec))

	router.Route("/api/chrome-service/v1/", func(subrouter chi.Router) {
		subrouter.Use(m.ParseHeaders)
		subrouter.Use(m.InjectUser)
		subrouter.Get("/hello-world", HelloWorld)
		subrouter.Route("/last-visited", routes.MakeLastVisitedRoutes)
		subrouter.Route("/favorite-pages", routes.MakeFavoritePagesRoutes)
		subrouter.Route("/self-report", routes.MakeSelfReportRoutes)
		subrouter.Route("/user", routes.MakeUserIdentityRoutes)
		subrouter.Route("/emit-message", routes.BroadcastMessage)
		subrouter.Route("/dashboard-templates", routes.MakeDashboardTemplateRoutes)
	})

	// We might want to set up some event listeners at some point, but the pod will
	// have to restart for these to take effect. We can't enable and disable websockets on the fly
	if featureflags.IsEnabled("chrome-service.websockets.enabled") {
		// start the connection hub
		go connectionhub.ConnectionHub.Run()
		logrus.Infoln("Enabling WebSockets")
		kafka.InitializeConsumers()
		router.Route("/wss/chrome-service/v1/", func(subrouter chi.Router) {
			subrouter.Use(cors.Handler(cors.Options{
				AllowedOrigins: []string{
					"wss://stage.foo.redhat.com:1337",
					"wss://prod.foo.redhat.com:1337",
				},
			}))
			subrouter.Route("/ws", routes.MakeWsRoute)
		})
	} else {
		logrus.Infoln("WebSockets are currently disabled")
	}

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

func setupGlobalLogger(opts *config.ChromeServiceConfig) {
	logLevel, err := logrus.ParseLevel(opts.LogLevel)
	if err != nil {
		logLevel = logrus.ErrorLevel
	}
	logrus.SetLevel(logLevel)
}
