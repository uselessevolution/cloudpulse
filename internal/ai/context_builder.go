package ai

import (
	"context"
	"fmt"

	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/incident"
	"cloudpulse/internal/service"
)

type IncidentReader interface {
	FindByID(
		ctx context.Context,
		id int64,
	) (incident.Incident, error)
}

type ServiceReader interface {
	FindByID(
		ctx context.Context,
		id int64,
	) (service.Service, error)
}

type HealthCheckReader interface {
	FindRecentByServiceID(
		ctx context.Context,
		serviceID int64,
		limit int,
	) ([]healthcheck.Result, error)
}

type ContextBuilder struct {
	incidentReader    IncidentReader
	serviceReader     ServiceReader
	healthCheckReader HealthCheckReader
	recentCheckLimit  int
}

func NewContextBuilder(
	incidentReader IncidentReader,
	serviceReader ServiceReader,
	healthCheckReader HealthCheckReader,
	recentCheckLimit int,
) *ContextBuilder {

	return &ContextBuilder{
		incidentReader:    incidentReader,
		serviceReader:     serviceReader,
		healthCheckReader: healthCheckReader,
		recentCheckLimit:  recentCheckLimit,
	}
}

func (builder *ContextBuilder) Build(
	ctx context.Context,
	incidentID int64,
) (IncidentContext, error) {

	foundIncident, err :=
		builder.incidentReader.FindByID(
			ctx,
			incidentID,
		)

	if err != nil {
		return IncidentContext{},
			fmt.Errorf(
				"load incident: %w",
				err,
			)
	}

	foundService, err :=
		builder.serviceReader.FindByID(
			ctx,
			foundIncident.ServiceID,
		)

	if err != nil {
		return IncidentContext{},
			fmt.Errorf(
				"load incident service: %w",
				err,
			)
	}

	recentChecks, err :=
		builder.healthCheckReader.
			FindRecentByServiceID(
				ctx,
				foundService.ID,
				builder.recentCheckLimit,
			)

	if err != nil {
		return IncidentContext{},
			fmt.Errorf(
				"load recent health checks: %w",
				err,
			)
	}

	return IncidentContext{
		IncidentID:     foundIncident.ID,
		ServiceID:      foundService.ID,
		ServiceName:    foundService.Name,
		ServiceURL:     foundService.URL,
		RuntimeStatus:  foundService.RuntimeStatus,
		IncidentStatus: foundIncident.Status,
		StartedAt:      foundIncident.StartedAt,
		ResolvedAt:     foundIncident.ResolvedAt,
		FailureCount:   foundIncident.FailureCount,
		LastError:      foundIncident.LastErrorMessage,
		RecentChecks:   recentChecks,
	}, nil
}
