package healthcheck

import "time"

type Result struct {
	ID           int64     `json:"id"`
	ServiceID    int64     `json:"serviceId"`
	CheckedAt    time.Time `json:"checkedAt"`
	StatusCode   *int      `json:"statusCode"`
	LatencyMS    int64     `json:"latencyMs"`
	Success      bool      `json:"success"`
	ErrorMessage *string   `json:"errorMessage"`
}

type CreateParams struct {
	ServiceID    int64
	StatusCode   *int
	LatencyMS    int64
	Success      bool
	ErrorMessage *string
}
