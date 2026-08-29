package monitoring

import (
	"context"
	"log"
	"time"

	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/service"
)

type DueServiceFinder interface {
	FindDueForChecks(
		ctx context.Context,
		now time.Time,
	) ([]service.Service, error)
}

type Scheduler struct {
	serviceRepository DueServiceFinder
	workerPool        *healthcheck.WorkerPool
	tickInterval      time.Duration
}

func NewScheduler(
	serviceRepository DueServiceFinder,
	workerPool *healthcheck.WorkerPool,
	tickInterval time.Duration,
) *Scheduler {
	return &Scheduler{
		serviceRepository: serviceRepository,
		workerPool:        workerPool,
		tickInterval:      tickInterval,
	}
}

func (scheduler *Scheduler) Run(
	ctx context.Context,
) {
	log.Printf(
		"monitoring scheduler started with tick interval %s",
		scheduler.tickInterval,
	)

	scheduler.runCycle(ctx)

	ticker := time.NewTicker(
		scheduler.tickInterval,
	)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println(
				"monitoring scheduler stopped",
			)
			return

		case <-ticker.C:
			scheduler.runCycle(ctx)
		}
	}
}

func (scheduler *Scheduler) runCycle(
	ctx context.Context,
) {
	services, err :=
		scheduler.serviceRepository.
			FindDueForChecks(
				ctx,
				time.Now(),
			)

	if err != nil {
		log.Printf(
			"monitoring cycle failed to load due services: %v",
			err,
		)
		return
	}

	if len(services) == 0 {
		return
	}

	log.Printf(
		"monitoring cycle found %d service(s) due for checking",
		len(services),
	)

	scheduler.workerPool.Run(
		ctx,
		services,
	)
}
