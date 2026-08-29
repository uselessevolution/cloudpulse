package service

import "time"

const (
	RuntimeStatusHealthy  = "HEALTHY"
	RuntimeStatusDegraded = "DEGRADED"
	RuntimeStatusDown     = "DOWN"
)

type Service struct {
	ID                   int64  `json:"id"`
	Name                 string `json:"name"`
	URL                  string `json:"url"`
	ExpectedStatus       int    `json:"expectedStatus"`
	CheckIntervalSeconds int    `json:"checkIntervalSeconds"`
	TimeoutSeconds       int    `json:"timeoutSeconds"`
	Enabled              bool   `json:"enabled"`

	RuntimeStatus        string     `json:"runtimeStatus"`
	ConsecutiveFailures  int        `json:"consecutiveFailures"`
	ConsecutiveSuccesses int        `json:"consecutiveSuccesses"`
	LastCheckedAt        *time.Time `json:"lastCheckedAt"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateParams struct {
	Name                 string
	URL                  string
	ExpectedStatus       int
	CheckIntervalSeconds int
	TimeoutSeconds       int
	Enabled              bool
}
