package incident

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/metrics"
	"cloudpulse/internal/service"
)

const (
	failureThreshold  = 3
	recoveryThreshold = 2
)

type ServiceRuntimeStore interface {
	UpdateRuntimeState(
		ctx context.Context,
		id int64,
		runtimeStatus string,
		consecutiveFailures int,
		consecutiveSuccesses int,
		lastCheckedAt time.Time,
	) (service.Service, error)
}

type IncidentStore interface {
	CreateOpen(
		ctx context.Context,
		serviceID int64,
		failureCount int,
		lastErrorMessage *string,
	) (Incident, error)

	FindOpenByServiceID(
		ctx context.Context,
		serviceID int64,
	) (Incident, error)

	UpdateOpen(
		ctx context.Context,
		serviceID int64,
		failureCount int,
		lastErrorMessage *string,
	) (Incident, error)

	Resolve(
		ctx context.Context,
		id int64,
		resolvedAt time.Time,
	) (Incident, error)
}

type Evaluator struct {
	serviceStore  ServiceRuntimeStore
	incidentStore IncidentStore
}

func NewEvaluator(
	serviceStore ServiceRuntimeStore,
	incidentStore IncidentStore,
) *Evaluator {
	return &Evaluator{
		serviceStore:  serviceStore,
		incidentStore: incidentStore,
	}
}

func (evaluator *Evaluator) Evaluate(
	ctx context.Context,
	target service.Service,
	result healthcheck.Result,
) error {

	if result.Success {
		return evaluator.handleSuccess(
			ctx,
			target,
			result,
		)
	}

	return evaluator.handleFailure(
		ctx,
		target,
		result,
	)
}

func (evaluator *Evaluator) handleFailure(
	ctx context.Context,
	target service.Service,
	result healthcheck.Result,
) error {

	consecutiveFailures :=
		target.ConsecutiveFailures + 1

	consecutiveSuccesses := 0

	runtimeStatus :=
		service.RuntimeStatusDegraded

	if target.RuntimeStatus ==
		service.RuntimeStatusDown {

		runtimeStatus =
			service.RuntimeStatusDown

	} else if consecutiveFailures >=
		failureThreshold {

		runtimeStatus =
			service.RuntimeStatusDown
	}

	_, err :=
		evaluator.serviceStore.
			UpdateRuntimeState(
				ctx,
				target.ID,
				runtimeStatus,
				consecutiveFailures,
				consecutiveSuccesses,
				result.CheckedAt,
			)

	if err != nil {
		return fmt.Errorf(
			"update service after failed check: %w",
			err,
		)
	}

	if runtimeStatus !=
		service.RuntimeStatusDown {
		return nil
	}

	openIncident, err :=
		evaluator.incidentStore.
			FindOpenByServiceID(
				ctx,
				target.ID,
			)

	if errors.Is(
		err,
		ErrNotFound,
	) {
		_, createErr :=
			evaluator.incidentStore.
				CreateOpen(
					ctx,
					target.ID,
					consecutiveFailures,
					result.ErrorMessage,
				)

		if createErr != nil {
			return fmt.Errorf(
				"create incident after service went down: %w",
				createErr,
			)
		}

		metrics.OpenIncidents.Inc()

		return nil
	}

	if err != nil {
		return fmt.Errorf(
			"find open incident after failed check: %w",
			err,
		)
	}

	incidentFailureCount :=
		openIncident.FailureCount + 1

	_, err =
		evaluator.incidentStore.
			UpdateOpen(
				ctx,
				target.ID,
				incidentFailureCount,
				result.ErrorMessage,
			)

	if err != nil {
		return fmt.Errorf(
			"update open incident after failed check: %w",
			err,
		)
	}

	return nil
}

func (evaluator *Evaluator) handleSuccess(
	ctx context.Context,
	target service.Service,
	result healthcheck.Result,
) error {

	consecutiveFailures := 0

	consecutiveSuccesses :=
		target.ConsecutiveSuccesses + 1

	runtimeStatus :=
		target.RuntimeStatus

	if runtimeStatus == "" {
		runtimeStatus =
			service.RuntimeStatusHealthy
	}

	recovered :=
		target.RuntimeStatus !=
			service.RuntimeStatusHealthy &&
			consecutiveSuccesses >=
				recoveryThreshold

	if recovered {
		runtimeStatus =
			service.RuntimeStatusHealthy
	}

	_, err :=
		evaluator.serviceStore.
			UpdateRuntimeState(
				ctx,
				target.ID,
				runtimeStatus,
				consecutiveFailures,
				consecutiveSuccesses,
				result.CheckedAt,
			)

	if err != nil {
		return fmt.Errorf(
			"update service after successful check: %w",
			err,
		)
	}

	if !recovered {
		return nil
	}

	openIncident, err :=
		evaluator.incidentStore.
			FindOpenByServiceID(
				ctx,
				target.ID,
			)

	if errors.Is(
		err,
		ErrNotFound,
	) {
		return nil
	}

	if err != nil {
		return fmt.Errorf(
			"find open incident during recovery: %w",
			err,
		)
	}

	_, err =
		evaluator.incidentStore.
			Resolve(
				ctx,
				openIncident.ID,
				result.CheckedAt,
			)

	if err != nil {
		return fmt.Errorf(
			"resolve recovered incident: %w",
			err,
		)
	}

	metrics.OpenIncidents.Dec()

	return nil
}
