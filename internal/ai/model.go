package ai

import (
	"time"

	"cloudpulse/internal/healthcheck"
)

type IncidentContext struct {
	IncidentID     int64                `json:"incidentId"`
	ServiceID      int64                `json:"serviceId"`
	ServiceName    string               `json:"serviceName"`
	ServiceURL     string               `json:"serviceUrl"`
	RuntimeStatus  string               `json:"runtimeStatus"`
	IncidentStatus string               `json:"incidentStatus"`
	StartedAt      time.Time            `json:"startedAt"`
	ResolvedAt     *time.Time           `json:"resolvedAt"`
	FailureCount   int                  `json:"failureCount"`
	LastError      *string              `json:"lastError"`
	RecentChecks   []healthcheck.Result `json:"recentChecks"`
}
type Analysis struct {
	Summary            string   `json:"summary"`
	PossibleCauses     []string `json:"possibleCauses"`
	RecommendedActions []string `json:"recommendedActions"`
}
