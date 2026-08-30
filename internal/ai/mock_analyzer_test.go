package ai

import (
	"context"
	"strings"
	"testing"

	"cloudpulse/internal/incident"
	"cloudpulse/internal/service"
)

func TestMockAnalyzerUsesIncidentContext(
	t *testing.T,
) {
	message :=
		"connection refused"

	analyzer :=
		NewMockAnalyzer()

	incidentContext :=
		IncidentContext{
			IncidentID:     10,
			ServiceID:      20,
			ServiceName:    "Payment API",
			RuntimeStatus:  service.RuntimeStatusDown,
			IncidentStatus: incident.StatusOpen,
			FailureCount:   4,
			LastError:      &message,
		}

	analysis, err :=
		analyzer.Analyze(
			context.Background(),
			incidentContext,
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if analysis.Summary == "" {
		t.Fatal(
			"expected non-empty summary",
		)
	}

	if !strings.Contains(
		analysis.Summary,
		"Payment API",
	) {
		t.Fatalf(
			"expected summary to contain service name, got %q",
			analysis.Summary,
		)
	}

	if len(
		analysis.PossibleCauses,
	) == 0 {
		t.Fatal(
			"expected possible causes",
		)
	}

	if len(
		analysis.RecommendedActions,
	) == 0 {
		t.Fatal(
			"expected recommended actions",
		)
	}
}
