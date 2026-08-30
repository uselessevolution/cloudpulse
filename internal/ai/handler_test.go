package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/incident"
	"cloudpulse/internal/service"

	"github.com/go-chi/chi/v5"
)

type fakeAnalyzer struct {
	analysis Analysis
	err      error
	calls    int
}

func (analyzer *fakeAnalyzer) Analyze(
	ctx context.Context,
	incidentContext IncidentContext,
) (Analysis, error) {
	analyzer.calls++

	return analyzer.analysis,
		analyzer.err
}

func TestAIHandlerAnalyzesIncident(
	t *testing.T,
) {
	now := time.Now()

	message :=
		"connection refused"

	incidentReader :=
		&fakeIncidentReader{
			found: incident.Incident{
				ID:               10,
				ServiceID:        20,
				Status:           incident.StatusOpen,
				StartedAt:        now,
				FailureCount:     3,
				LastErrorMessage: &message,
			},
		}

	serviceReader :=
		&fakeServiceReader{
			found: service.Service{
				ID:            20,
				Name:          "Payment API",
				URL:           "http://localhost:65530",
				RuntimeStatus: service.RuntimeStatusDown,
			},
		}

	healthCheckReader :=
		&fakeHealthCheckReader{
			results: []healthcheck.Result{
				{
					ServiceID:    20,
					CheckedAt:    now,
					Success:      false,
					ErrorMessage: &message,
				},
			},
		}

	builder :=
		NewContextBuilder(
			incidentReader,
			serviceReader,
			healthCheckReader,
			10,
		)

	analyzer :=
		&fakeAnalyzer{
			analysis: Analysis{
				Summary: "Payment API is experiencing repeated connection failures.",
				PossibleCauses: []string{
					"Service process may not be listening on the configured port.",
				},
				RecommendedActions: []string{
					"Verify that the service process is running.",
				},
			},
		}

	handler :=
		NewHandler(
			builder,
			analyzer,
		)

	router :=
		chi.NewRouter()

	router.Post(
		"/api/incidents/{id}/ai-summary",
		handler.AnalyzeIncident,
	)

	request :=
		httptest.NewRequest(
			http.MethodPost,
			"/api/incidents/10/ai-summary",
			nil,
		)

	response :=
		httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code !=
		http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d: %s",
			response.Code,
			response.Body.String(),
		)
	}

	if analyzer.calls != 1 {
		t.Fatalf(
			"expected analyzer to be called once, got %d",
			analyzer.calls,
		)
	}
}

func TestAIHandlerRejectsInvalidIncidentID(
	t *testing.T,
) {
	builder :=
		NewContextBuilder(
			&fakeIncidentReader{},
			&fakeServiceReader{},
			&fakeHealthCheckReader{},
			10,
		)

	analyzer :=
		&fakeAnalyzer{}

	handler :=
		NewHandler(
			builder,
			analyzer,
		)

	router :=
		chi.NewRouter()

	router.Post(
		"/api/incidents/{id}/ai-summary",
		handler.AnalyzeIncident,
	)

	request :=
		httptest.NewRequest(
			http.MethodPost,
			"/api/incidents/abc/ai-summary",
			nil,
		)

	response :=
		httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code !=
		http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			response.Code,
		)
	}

	if analyzer.calls != 0 {
		t.Fatal(
			"analyzer should not be called for invalid incident id",
		)
	}
}
