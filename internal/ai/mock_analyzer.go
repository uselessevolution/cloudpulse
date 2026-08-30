package ai

import (
	"context"
	"fmt"
)

type MockAnalyzer struct {
}

func NewMockAnalyzer() *MockAnalyzer {
	return &MockAnalyzer{}
}

func (analyzer *MockAnalyzer) Analyze(
	ctx context.Context,
	incidentContext IncidentContext,
) (Analysis, error) {

	summary :=
		fmt.Sprintf(
			"Service %s experienced an incident with %d recorded failures. Current service status is %s and incident status is %s.",
			incidentContext.ServiceName,
			incidentContext.FailureCount,
			incidentContext.RuntimeStatus,
			incidentContext.IncidentStatus,
		)

	possibleCauses := []string{
		"The monitored service may have been temporarily unavailable.",
		"Network connectivity or endpoint availability may have contributed to the failed health checks.",
	}

	if incidentContext.LastError != nil {
		possibleCauses =
			append(
				possibleCauses,
				fmt.Sprintf(
					"Recent monitoring reported the error: %s",
					*incidentContext.LastError,
				),
			)
	}

	recommendedActions := []string{
		"Verify that the service process is running and listening on the configured endpoint.",
		"Review recent application and deployment logs around the incident start time.",
		"Check network connectivity between CloudPulse and the monitored service.",
	}

	return Analysis{
		Summary:            summary,
		PossibleCauses:     possibleCauses,
		RecommendedActions: recommendedActions,
	}, nil
}
