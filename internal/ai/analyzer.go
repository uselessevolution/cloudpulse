package ai

import "context"

type IncidentAnalyzer interface {
	Analyze(
		ctx context.Context,
		incidentContext IncidentContext,
	) (Analysis, error)
}
