package ai

import (
	"context"
	"testing"
	"time"

	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/incident"
	"cloudpulse/internal/service"
)

type fakeIncidentReader struct {
	found incident.Incident
}

func (reader *fakeIncidentReader) FindByID(
	ctx context.Context,
	id int64,
) (incident.Incident, error) {

	return reader.found, nil
}

type fakeServiceReader struct {
	found service.Service
}

func (reader *fakeServiceReader) FindByID(
	ctx context.Context,
	id int64,
) (service.Service, error) {

	return reader.found, nil
}

type fakeHealthCheckReader struct {
	results []healthcheck.Result
}

func (reader *fakeHealthCheckReader) FindRecentByServiceID(
	ctx context.Context,
	serviceID int64,
	limit int,
) ([]healthcheck.Result, error) {

	return reader.results, nil
}

func TestContextBuilderBuildsIncidentContext(
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

	result, err :=
		builder.Build(
			context.Background(),
			10,
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.IncidentID != 10 {
		t.Fatalf(
			"expected incident id 10, got %d",
			result.IncidentID,
		)
	}

	if result.ServiceID != 20 {
		t.Fatalf(
			"expected service id 20, got %d",
			result.ServiceID,
		)
	}

	if result.ServiceName !=
		"Payment API" {
		t.Fatalf(
			"expected Payment API, got %s",
			result.ServiceName,
		)
	}

	if result.RuntimeStatus !=
		service.RuntimeStatusDown {
		t.Fatalf(
			"expected DOWN, got %s",
			result.RuntimeStatus,
		)
	}

	if len(result.RecentChecks) != 1 {
		t.Fatalf(
			"expected 1 recent check, got %d",
			len(result.RecentChecks),
		)
	}

	if result.RecentChecks[0].Success {
		t.Fatal(
			"expected recent health check failure",
		)
	}
}
