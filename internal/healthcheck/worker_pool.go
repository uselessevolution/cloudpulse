package healthcheck

import (
	"context"
	"log"
	"sync"

	"cloudpulse/internal/service"
)

type ResultWriter interface {
	Create(
		ctx context.Context,
		params CreateParams,
	) (Result, error)
}

type WorkerPool struct {
	workerCount int
	checker     *Checker
	repository  ResultWriter
}

func NewWorkerPool(
	workerCount int,
	checker *Checker,
	repository ResultWriter,
) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		checker:     checker,
		repository:  repository,
	}
}

func (pool *WorkerPool) Run(
	ctx context.Context,
	services []service.Service,
) {
	jobs := make(
		chan service.Service,
	)

	var waitGroup sync.WaitGroup

	for workerID := 1; workerID <= pool.workerCount; workerID++ {

		waitGroup.Add(1)

		go pool.worker(
			ctx,
			workerID,
			jobs,
			&waitGroup,
		)
	}

	for _, target := range services {
		select {
		case jobs <- target:

		case <-ctx.Done():
			close(jobs)
			waitGroup.Wait()
			return
		}
	}

	close(jobs)

	waitGroup.Wait()
}

func (pool *WorkerPool) worker(
	ctx context.Context,
	workerID int,
	jobs <-chan service.Service,
	waitGroup *sync.WaitGroup,
) {
	defer waitGroup.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case target, ok := <-jobs:
			if !ok {
				return
			}

			result :=
				pool.checker.Check(
					ctx,
					target,
				)

			_, err :=
				pool.repository.Create(
					ctx,
					CreateParams{
						ServiceID:    target.ID,
						StatusCode:   result.StatusCode,
						LatencyMS:    result.LatencyMS,
						Success:      result.Success,
						ErrorMessage: result.ErrorMessage,
					},
				)

			if err != nil {
				log.Printf(
					"worker %d failed to save health check for service %d: %v",
					workerID,
					target.ID,
					err,
				)

				continue
			}

			log.Printf(
				"worker %d checked service %d: success=%t latency=%dms",
				workerID,
				target.ID,
				result.Success,
				result.LatencyMS,
			)
		}
	}
}
