package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloudpulse/internal/service"
)

func TestCheckerSuccess(
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

	checker := NewChecker()

	target := service.Service{
		ID:             1,
		URL:            server.URL,
		ExpectedStatus: http.StatusOK,
		TimeoutSeconds: 2,
	}

	result :=
		checker.Check(
			context.Background(),
			target,
		)

	if !result.Success {
		t.Fatalf(
			"expected success, got failure: %v",
			result.ErrorMessage,
		)
	}

	if result.StatusCode == nil {
		t.Fatal(
			"expected status code",
		)
	}

	if *result.StatusCode !=
		http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			*result.StatusCode,
		)
	}
}

func TestCheckerUnexpectedStatus(
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
						http.StatusInternalServerError,
					)
				},
			),
		)

	defer server.Close()

	checker := NewChecker()

	target := service.Service{
		ID:             2,
		URL:            server.URL,
		ExpectedStatus: http.StatusOK,
		TimeoutSeconds: 2,
	}

	result :=
		checker.Check(
			context.Background(),
			target,
		)

	if result.Success {
		t.Fatal(
			"expected health check to fail",
		)
	}

	if result.StatusCode == nil {
		t.Fatal(
			"expected status code",
		)
	}

	if *result.StatusCode !=
		http.StatusInternalServerError {
		t.Fatalf(
			"expected 500, got %d",
			*result.StatusCode,
		)
	}

	if result.ErrorMessage == nil {
		t.Fatal(
			"expected error message",
		)
	}
}

func TestCheckerTimeout(
	t *testing.T,
) {
	server :=
		httptest.NewServer(
			http.HandlerFunc(
				func(
					writer http.ResponseWriter,
					request *http.Request,
				) {
					time.Sleep(
						1500 *
							time.Millisecond,
					)

					writer.WriteHeader(
						http.StatusOK,
					)
				},
			),
		)

	defer server.Close()

	checker := NewChecker()

	target := service.Service{
		ID:             3,
		URL:            server.URL,
		ExpectedStatus: http.StatusOK,
		TimeoutSeconds: 1,
	}

	result :=
		checker.Check(
			context.Background(),
			target,
		)

	if result.Success {
		t.Fatal(
			"expected timeout failure",
		)
	}

	if result.StatusCode != nil {
		t.Fatal(
			"expected no status code after timeout",
		)
	}

	if result.ErrorMessage == nil {
		t.Fatal(
			"expected timeout error message",
		)
	}
}
