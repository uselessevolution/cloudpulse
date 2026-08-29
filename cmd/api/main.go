package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloudpulse/internal/config"
	"cloudpulse/internal/database"
	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/httpapi"
	"cloudpulse/internal/monitoring"
	"cloudpulse/internal/service"
)

func main() {
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

	checker :=
		healthcheck.NewChecker()

	workerPool :=
		healthcheck.NewWorkerPool(
			5,
			checker,
			healthCheckRepository,
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

	router := httpapi.NewRouter(
		serviceHandler,
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
