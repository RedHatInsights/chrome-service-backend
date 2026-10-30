package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

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
	"runtime/debug"
)

func main() {
	cfg := config.Get()
	setupGlobalLogger(cfg)
	if err := run(cfg); err != nil {
		fields := logrus.Fields{"error_type": rootErrorType(err)}
		var failure *lifecycleError
		if errors.As(err, &failure) {
			fields["component"] = failure.component
			if failure.panicType != "" {
				fields["panic_type"] = failure.panicType
				fields["stack"] = failure.stack
			}
		}
		var databaseFailure *database.InitError
		if errors.As(err, &databaseFailure) {
			fields["stage"] = databaseFailure.Stage
		}
		logrus.WithFields(fields).Error("service stopped unexpectedly")
		logger.FlushCloudWatch()
		os.Exit(1)
	}
	logger.FlushCloudWatch()
}

type lifecycleError struct {
	component string
	err       error
	panicType string
	stack     string
}

func (e *lifecycleError) Error() string { return e.component + " failed" }
func (e *lifecycleError) Unwrap() error { return e.err }

func rootErrorType(err error) string {
	for errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}
	if err == nil {
		return "unknown"
	}
	return fmt.Sprintf("%T", err)
}

func run(cfg *config.ChromeServiceConfig) (result error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			stack := debug.Stack()
			if len(stack) > 8192 {
				stack = stack[:8192]
			}
			result = &lifecycleError{component: "startup", panicType: fmt.Sprintf("%T", recovered), stack: string(stack)}
		}
	}()

	flag.Parse()
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		logrus.WithField("error_type", fmt.Sprintf("%T", err)).Warn("environment file could not be loaded")
	}
	if err := database.Init(); err != nil {
		return &lifecycleError{component: "database", err: err}
	}
	logrus.Info("database initialized")
	util.InitUserIdentitiesCache()
	util.CreateChromeConfiguration()
	service.LoadBaseLayout()
	featureflags.Init(cfg)
	router := chi.NewRouter()
	metricsRouter := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.InjectLogger)
	router.Use(middleware.RequestLogger(logger.NewLogger(cfg, logrus.StandardLogger())))
	router.Use(logger.Recoverer)
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
			// AllowedOrigins must stay in sync with checkOrigin in rest/routes/websocket.go.
			subrouter.Use(cors.Handler(cors.Options{
				AllowedOrigins: []string{
					"https://console.redhat.com",
					"https://*.console.redhat.com",
					"https://console.stage.redhat.com",
					"https://*.console.stage.redhat.com",
					"https://console.dev.redhat.com",
					"https://*.console.dev.redhat.com",
					"https://*.foo.redhat.com",
					"https://stage.foo.redhat.com:1337",
					"https://prod.foo.redhat.com:1337",
				},
			}))
			subrouter.Use(m.ParseHeaders)
			subrouter.Use(logger.EnrichLoggerWithIdentity)
			subrouter.Route("/ws", routes.MakeWsRoute)
		})
	} else {
		logrus.Infoln("WebSockets are currently disabled")
	}

	metricsRouter.Handle("/metrics", promhttp.Handler())

	securitylog.LogStartup("chrome-service-backend", cfg.WebPort)

	serverStringAddr := fmt.Sprintf(":%s", strconv.Itoa(cfg.WebPort))
	server := &http.Server{Addr: serverStringAddr, Handler: router}
	metricsStringAddr := fmt.Sprintf(":%s", strconv.Itoa(cfg.MetricsPort))
	metricsServer := &http.Server{Addr: metricsStringAddr, Handler: metricsRouter}
	type serveResult struct {
		component string
		err       error
	}
	serveErrors := make(chan serveResult, 2)
	serve := func(component string, server *http.Server) {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErrors <- serveResult{component: component, err: err}
		}
	}
	// Handle SIGTERM/SIGINT for graceful shutdown with security logging.
	// Signal wait is synchronous in main so shutdown completes before exit.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigChan)
	go serve("http", server)
	go serve("metrics", metricsServer)

	var shutdownReason string
	var shutdownFailure error
	shutdownComponent := "shutdown"
	select {
	case sig := <-sigChan:
		shutdownReason = fmt.Sprintf("received %s", sig)
	case failed := <-serveErrors:
		shutdownReason = failed.component + " server failed"
		shutdownFailure = failed.err
		shutdownComponent = failed.component
	}

	// Graceful shutdown with bounded timeout to drain active connections.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		if shutdownFailure == nil {
			shutdownFailure = err
			shutdownComponent = "http_shutdown"
		}
	}
	if err := metricsServer.Shutdown(ctx); err != nil {
		if shutdownFailure == nil {
			shutdownFailure = err
			shutdownComponent = "metrics_shutdown"
		}
	}
	securitylog.LogShutdownWithOutcome("chrome-service-backend", shutdownReason, shutdownFailure != nil)
	if shutdownFailure != nil {
		return &lifecycleError{component: shutdownComponent, err: shutdownFailure}
	}
	return nil
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
