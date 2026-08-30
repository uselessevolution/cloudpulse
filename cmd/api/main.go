package main

import (
	"cloudpulse/internal/ai"
	"cloudpulse/internal/config"
	"cloudpulse/internal/database"
	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/httpapi"
	"cloudpulse/internal/incident"
	"cloudpulse/internal/metrics"
	"cloudpulse/internal/monitoring"
	"cloudpulse/internal/service"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	metrics.Register()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	appContext, appCancel :=
		context.WithCancel(
			context.Background(),
		)

	defer appCancel()

	dbPool, err := database.NewPostgresPool(
		appContext,
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("connected to PostgreSQL")
	serviceRepository :=
		service.NewRepository(
			dbPool,
		)

	healthCheckRepository :=
		healthcheck.NewRepository(
			dbPool,
		)
	incidentRepository :=
		incident.NewRepository(
			dbPool,
		)
	aiContextBuilder :=
		ai.NewContextBuilder(
			incidentRepository,
			serviceRepository,
			healthCheckRepository,
			10,
		)
	var incidentAnalyzer ai.IncidentAnalyzer

	switch strings.ToLower(
		cfg.AIProvider,
	) {
	case "mock":
		incidentAnalyzer =
			ai.NewMockAnalyzer()

		log.Printf(
			"event=ai_provider_initialized provider=mock",
		)

	case "openai":
		if os.Getenv(
			"OPENAI_API_KEY",
		) == "" {
			log.Fatal(
				"OPENAI_API_KEY is required when AI_PROVIDER=openai",
			)
		}

		incidentAnalyzer =
			ai.NewOpenAIAnalyzer(
				cfg.OpenAIModel,
			)

		log.Printf(
			"event=ai_provider_initialized provider=openai model=%s",
			cfg.OpenAIModel,
		)

	default:
		log.Fatalf(
			"unsupported AI_PROVIDER: %s",
			cfg.AIProvider,
		)
	}
	aiHandler :=
		ai.NewHandler(
			aiContextBuilder,
			incidentAnalyzer,
		)
	openIncidentCount, err :=
		incidentRepository.CountOpen(
			appContext,
		)

	if err != nil {
		log.Fatalf(
			"failed to count open incidents: %v",
			err,
		)
	}

	metrics.OpenIncidents.Set(
		float64(
			openIncidentCount,
		),
	)

	log.Printf(
		"event=open_incident_metric_initialized count=%d",
		openIncidentCount,
	)
	incidentHandler :=
		incident.NewHandler(
			incidentRepository,
		)
	incidentEvaluator :=
		incident.NewEvaluator(
			serviceRepository,
			incidentRepository,
		)
	checker :=
		healthcheck.NewChecker()

	workerPool :=
		healthcheck.NewWorkerPool(
			5,
			checker,
			healthCheckRepository,
			incidentEvaluator,
		)

	scheduler :=
		monitoring.NewScheduler(
			serviceRepository,
			workerPool,
			5*time.Second,
		)

	serviceHandler :=
		service.NewHandler(
			serviceRepository,
		)

	router :=
		httpapi.NewRouter(
			serviceHandler,
			incidentHandler,
			aiHandler,
		)
	go scheduler.Run(
		appContext,
	)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErrors := make(
		chan error,
		1,
	)

	go func() {
		log.Printf(
			"CloudPulse API starting on :%s",
			cfg.Port,
		)

		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignals := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		shutdownSignals,
		os.Interrupt,
		syscall.SIGTERM,
	)

	select {
	case signalReceived := <-shutdownSignals:
		log.Printf(
			"shutdown signal received: %s",
			signalReceived,
		)

	case serverError := <-serverErrors:
		if !errors.Is(
			serverError,
			http.ErrServerClosed,
		) {
			log.Printf(
				"HTTP server error: %v",
				serverError,
			)
		}
	}
	appCancel()
	shutdownContext, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	if err := server.Shutdown(
		shutdownContext,
	); err != nil {
		log.Printf(
			"HTTP shutdown error: %v",
			err,
		)
	}

	dbPool.Close()

	log.Println(
		"CloudPulse shutdown complete",
	)
}
