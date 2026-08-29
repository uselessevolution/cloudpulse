package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"cloudpulse/internal/service"
)

type fakeResultWriter struct {
	mutex   sync.Mutex
	results []CreateParams
}

func (writer *fakeResultWriter) Create(
	ctx context.Context,
	params CreateParams,
) (Result, error) {

	writer.mutex.Lock()
	defer writer.mutex.Unlock()

	writer.results = append(
		writer.results,
		params,
	)

	return Result{
		ServiceID:    params.ServiceID,
		CheckedAt:    time.Now(),
		StatusCode:   params.StatusCode,
		LatencyMS:    params.LatencyMS,
		Success:      params.Success,
		ErrorMessage: params.ErrorMessage,
	}, nil
}

type fakeResultEvaluator struct {
	mutex sync.Mutex
	calls int
}

func (evaluator *fakeResultEvaluator) Evaluate(
	ctx context.Context,
	target service.Service,
	result Result,
) error {

	evaluator.mutex.Lock()
	defer evaluator.mutex.Unlock()

	evaluator.calls++

	return nil
}

func TestWorkerPoolProcessesAllServices(
	t *testing.T,
) {
	server :=
		httptest.NewServer(
			http.HandlerFunc(
				func(
					writer http.ResponseWriter,
					request *http.Request,
				) {
					writer.WriteHeader(
						http.StatusOK,
					)
				},
			),
		)

	defer server.Close()

	targets := []service.Service{
		{
			ID:             1,
			URL:            server.URL,
			ExpectedStatus: 200,
			TimeoutSeconds: 2,
		},
		{
			ID:             2,
			URL:            server.URL,
			ExpectedStatus: 200,
			TimeoutSeconds: 2,
		},
		{
			ID:             3,
			URL:            server.URL,
			ExpectedStatus: 200,
			TimeoutSeconds: 2,
		},
	}

	writer :=
		&fakeResultWriter{
			results: make(
				[]CreateParams,
				0,
			),
		}

	evaluator :=
		&fakeResultEvaluator{}

	checker := NewChecker()

	pool :=
		NewWorkerPool(
			2,
			checker,
			writer,
			evaluator,
		)

	pool.Run(
		context.Background(),
		targets,
	)

	writer.mutex.Lock()
	resultCount := len(writer.results)
	writer.mutex.Unlock()

	if resultCount != len(targets) {
		t.Fatalf(
			"expected %d results, got %d",
			len(targets),
			resultCount,
		)
	}

	writer.mutex.Lock()

	for _, result := range writer.results {

		if !result.Success {
			writer.mutex.Unlock()

			t.Fatalf(
				"expected result for service %d to succeed",
				result.ServiceID,
			)
		}
	}

	writer.mutex.Unlock()

	evaluator.mutex.Lock()
	evaluationCalls :=
		evaluator.calls
	evaluator.mutex.Unlock()

	if evaluationCalls != len(targets) {
		t.Fatalf(
			"expected evaluator to be called %d times, got %d",
			len(targets),
			evaluationCalls,
		)
	}
}
