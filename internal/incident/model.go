package incident

import "time"

const (
	StatusOpen     = "OPEN"
	StatusResolved = "RESOLVED"
)

type Incident struct {
	ID               int64      `json:"id"`
	ServiceID        int64      `json:"serviceId"`
	Status           string     `json:"status"`
	StartedAt        time.Time  `json:"startedAt"`
	ResolvedAt       *time.Time `json:"resolvedAt"`
	FailureCount     int        `json:"failureCount"`
	LastErrorMessage *string    `json:"lastErrorMessage"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}
