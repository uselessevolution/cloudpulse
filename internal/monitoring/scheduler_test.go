package monitoring

import (
	"context"
	"sync"
	"testing"
	"time"

	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/service"
)

type fakeDueServiceFinder struct {
	mutex    sync.Mutex
	services []service.Service
	calls    int
}

func (finder *fakeDueServiceFinder) FindDueForChecks(
	ctx context.Context,
	now time.Time,
) ([]service.Service, error) {

	finder.mutex.Lock()
	defer finder.mutex.Unlock()

	finder.calls++

	return finder.services, nil
}

type fakeResultWriter struct {
	mutex sync.Mutex
	count int
}

func (writer *fakeResultWriter) Create(
	ctx context.Context,
	params healthcheck.CreateParams,
) (healthcheck.Result, error) {

	writer.mutex.Lock()
	defer writer.mutex.Unlock()

	writer.count++

	return healthcheck.Result{
		ServiceID: params.ServiceID,
		Success:   params.Success,
	}, nil
}

func TestSchedulerRunsMonitoringCycle(
	t *testing.T,
) {
	finder := &fakeDueServiceFinder{
		services: []service.Service{},
	}

	writer := &fakeResultWriter{}

	checker :=
		healthcheck.NewChecker()

	workerPool :=
		healthcheck.NewWorkerPool(
			2,
			checker,
			writer,
		)

	scheduler :=
		NewScheduler(
			finder,
			workerPool,
			20*time.Millisecond,
		)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			75*time.Millisecond,
		)

	defer cancel()

	scheduler.Run(ctx)

	finder.mutex.Lock()
	calls := finder.calls
	finder.mutex.Unlock()

	if calls < 2 {
		t.Fatalf(
			"expected scheduler to run multiple cycles, got %d",
			calls,
		)
	}
}
