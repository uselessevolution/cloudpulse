package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HealthChecksTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cloudpulse_health_checks_total",
			Help: "Total number of service health checks.",
		},
		[]string{
			"service_id",
			"success",
		},
	)

	HealthCheckLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cloudpulse_health_check_latency_seconds",
			Help:    "Health check latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{
			"service_id",
		},
	)

	OpenIncidents = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "cloudpulse_open_incidents",
			Help: "Current number of open incidents.",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		HealthChecksTotal,
		HealthCheckLatency,
		OpenIncidents,
	)
}
