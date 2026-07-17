package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
	"github.com/RedHatInsights/chrome-service-backend/rest/database"
	"github.com/RedHatInsights/chrome-service-backend/rest/featureflags"
	"github.com/RedHatInsights/chrome-service-backend/rest/kafka"
	"github.com/RedHatInsights/chrome-service-backend/rest/logger"
	m "github.com/RedHatInsights/chrome-service-backend/rest/middleware"
	"github.com/RedHatInsights/chrome-service-backend/rest/routes"
	"github.com/RedHatInsights/chrome-service-backend/rest/securitylog"
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
	util.CreateChromeConfiguration()
}

func main() {
	cfg := config.Get()
	service.LoadBaseLayout()
	featureflags.Init(cfg)
	setupGlobalLogger(cfg)
	defer logger.FlushCloudWatch()
	router := chi.NewRouter()
	metricsRouter := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.InjectLogger)
	router.Use(middleware.RequestLogger(logger.NewLogger(cfg, logrus.StandardLogger())))
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
		subrouter.Use(logger.EnrichLoggerWithIdentity)
		subrouter.Use(m.InjectUser)
		subrouter.Get("/hello-world", HelloWorld)
		subrouter.Route("/last-visited", routes.MakeLastVisitedRoutes)
		subrouter.Route("/recently-used-workspaces", routes.MakeRecentlyUsedWorkspacesRoutes)
		subrouter.Route("/favorite-pages", routes.MakeFavoritePagesRoutes)
		subrouter.Route("/self-report", routes.MakeSelfReportRoutes)
		subrouter.Route("/user", routes.MakeUserIdentityRoutes)
		subrouter.Route("/emit-message", routes.BroadcastMessage)
		subrouter.Route("/dashboard-templates", routes.MakeDashboardTemplateRoutes)
		subrouter.Route("/api-docs", routes.MakeApiDocsRoutes)
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
			logger.FlushCloudWatch()
			log.Fatalf("Metrics server stopped %v", err)
		}
	}()

	securitylog.LogStartup("chrome-service-backend", cfg.WebPort)

	serverStringAddr := fmt.Sprintf(":%s", strconv.Itoa(cfg.WebPort))
	server := &http.Server{Addr: serverStringAddr, Handler: router}

	// Handle SIGTERM/SIGINT for graceful shutdown with security logging.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		securitylog.LogShutdown("chrome-service-backend", fmt.Sprintf("received %s", sig))
		logger.FlushCloudWatch()
		server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		securitylog.LogShutdown("chrome-service-backend", fmt.Sprintf("server error: %v", err))
		logger.FlushCloudWatch()
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
	logger.SetupCloudWatch(opts)
}
