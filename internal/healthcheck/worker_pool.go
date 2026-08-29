package healthcheck

import (
	"context"
	"log"
	"strconv"
	"sync"

	"cloudpulse/internal/metrics"
	"cloudpulse/internal/service"
)

type ResultWriter interface {
	Create(
		ctx context.Context,
		params CreateParams,
	) (Result, error)
}

type ResultEvaluator interface {
	Evaluate(
		ctx context.Context,
		target service.Service,
		result Result,
	) error
}

type WorkerPool struct {
	workerCount int
	checker     *Checker
	repository  ResultWriter
	evaluator   ResultEvaluator
}

func NewWorkerPool(
	workerCount int,
	checker *Checker,
	repository ResultWriter,
	evaluator ResultEvaluator,
) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		checker:     checker,
		repository:  repository,
		evaluator:   evaluator,
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

			checkResult :=
				pool.checker.Check(
					ctx,
					target,
				)

			savedResult, err :=
				pool.repository.Create(
					ctx,
					CreateParams{
						ServiceID:    target.ID,
						StatusCode:   checkResult.StatusCode,
						LatencyMS:    checkResult.LatencyMS,
						Success:      checkResult.Success,
						ErrorMessage: checkResult.ErrorMessage,
					},
				)

			if err != nil {
				log.Printf(
					"event=health_check_persist_failed worker=%d service_id=%d error=%q",
					workerID,
					target.ID,
					err,
				)

				continue
			}

			serviceID :=
				strconv.FormatInt(
					target.ID,
					10,
				)

			metrics.HealthChecksTotal.
				WithLabelValues(
					serviceID,
					strconv.FormatBool(
						savedResult.Success,
					),
				).
				Inc()

			metrics.HealthCheckLatency.
				WithLabelValues(
					serviceID,
				).
				Observe(
					float64(
						savedResult.LatencyMS,
					) / 1000,
				)

			if pool.evaluator != nil {
				err =
					pool.evaluator.Evaluate(
						ctx,
						target,
						savedResult,
					)

				if err != nil {
					log.Printf(
						"event=incident_evaluation_failed worker=%d service_id=%d error=%q",
						workerID,
						target.ID,
						err,
					)

					continue
				}
			}

			log.Printf(
				"event=health_check worker=%d service_id=%d success=%t latency_ms=%d",
				workerID,
				target.ID,
				savedResult.Success,
				savedResult.LatencyMS,
			)
		}
	}
}
