package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

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
		StatusCode:   params.StatusCode,
		LatencyMS:    params.LatencyMS,
		Success:      params.Success,
		ErrorMessage: params.ErrorMessage,
	}, nil
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

	checker := NewChecker()

	pool :=
		NewWorkerPool(
			2,
			checker,
			writer,
		)

	pool.Run(
		context.Background(),
		targets,
	)

	if len(writer.results) !=
		len(targets) {
		t.Fatalf(
			"expected %d results, got %d",
			len(targets),
			len(writer.results),
		)
	}

	for _, result := range writer.results {

		if !result.Success {
			t.Fatalf(
				"expected result for service %d to succeed",
				result.ServiceID,
			)
		}
	}
}
