package incident

import (
	"context"
	"testing"
	"time"

	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/service"
)

type runtimeUpdate struct {
	ID                   int64
	RuntimeStatus        string
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	LastCheckedAt        time.Time
}

type fakeServiceRuntimeStore struct {
	lastUpdate runtimeUpdate
}

func (store *fakeServiceRuntimeStore) UpdateRuntimeState(
	ctx context.Context,
	id int64,
	runtimeStatus string,
	consecutiveFailures int,
	consecutiveSuccesses int,
	lastCheckedAt time.Time,
) (service.Service, error) {

	store.lastUpdate = runtimeUpdate{
		ID:                   id,
		RuntimeStatus:        runtimeStatus,
		ConsecutiveFailures:  consecutiveFailures,
		ConsecutiveSuccesses: consecutiveSuccesses,
		LastCheckedAt:        lastCheckedAt,
	}

	return service.Service{
		ID:                   id,
		RuntimeStatus:        runtimeStatus,
		ConsecutiveFailures:  consecutiveFailures,
		ConsecutiveSuccesses: consecutiveSuccesses,
	}, nil
}

type fakeIncidentStore struct {
	openIncident *Incident

	createCalls  int
	updateCalls  int
	resolveCalls int

	lastCreatedFailureCount int
	lastUpdatedFailureCount int
	lastResolvedID          int64
}

func (store *fakeIncidentStore) CreateOpen(
	ctx context.Context,
	serviceID int64,
	failureCount int,
	lastErrorMessage *string,
) (Incident, error) {

	store.createCalls++

	store.lastCreatedFailureCount =
		failureCount

	created := Incident{
		ID:           100,
		ServiceID:    serviceID,
		Status:       StatusOpen,
		FailureCount: failureCount,
	}

	store.openIncident =
		&created

	return created, nil
}

func (store *fakeIncidentStore) FindOpenByServiceID(
	ctx context.Context,
	serviceID int64,
) (Incident, error) {

	if store.openIncident == nil {
		return Incident{}, ErrNotFound
	}

	return *store.openIncident, nil
}

func (store *fakeIncidentStore) UpdateOpen(
	ctx context.Context,
	serviceID int64,
	failureCount int,
	lastErrorMessage *string,
) (Incident, error) {

	store.updateCalls++

	store.lastUpdatedFailureCount =
		failureCount

	updated := Incident{
		ID:           store.openIncident.ID,
		ServiceID:    serviceID,
		Status:       StatusOpen,
		FailureCount: failureCount,
	}

	store.openIncident =
		&updated

	return updated, nil
}

func (store *fakeIncidentStore) Resolve(
	ctx context.Context,
	id int64,
	resolvedAt time.Time,
) (Incident, error) {

	store.resolveCalls++
	store.lastResolvedID = id

	resolved := *store.openIncident

	resolved.Status =
		StatusResolved

	resolved.ResolvedAt =
		&resolvedAt

	store.openIncident = nil

	return resolved, nil
}

func TestEvaluatorFirstFailureMakesServiceDegraded(
	t *testing.T,
) {
	serviceStore :=
		&fakeServiceRuntimeStore{}

	incidentStore :=
		&fakeIncidentStore{}

	evaluator :=
		NewEvaluator(
			serviceStore,
			incidentStore,
		)

	now := time.Now()

	target := service.Service{
		ID:                  1,
		RuntimeStatus:       service.RuntimeStatusHealthy,
		ConsecutiveFailures: 0,
	}

	result := healthcheck.Result{
		ServiceID: 1,
		CheckedAt: now,
		Success:   false,
	}

	err :=
		evaluator.Evaluate(
			context.Background(),
			target,
			result,
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if serviceStore.lastUpdate.RuntimeStatus !=
		service.RuntimeStatusDegraded {
		t.Fatalf(
			"expected DEGRADED, got %s",
			serviceStore.lastUpdate.RuntimeStatus,
		)
	}

	if serviceStore.lastUpdate.ConsecutiveFailures != 1 {
		t.Fatalf(
			"expected 1 consecutive failure, got %d",
			serviceStore.lastUpdate.ConsecutiveFailures,
		)
	}

	if incidentStore.createCalls != 0 {
		t.Fatal(
			"incident should not be created after first failure",
		)
	}
}

func TestEvaluatorThirdFailureMakesServiceDownAndCreatesIncident(
	t *testing.T,
) {
	serviceStore :=
		&fakeServiceRuntimeStore{}

	incidentStore :=
		&fakeIncidentStore{}

	evaluator :=
		NewEvaluator(
			serviceStore,
			incidentStore,
		)

	message :=
		"connection refused"

	target := service.Service{
		ID:                  2,
		RuntimeStatus:       service.RuntimeStatusDegraded,
		ConsecutiveFailures: 2,
	}

	result := healthcheck.Result{
		ServiceID:    2,
		CheckedAt:    time.Now(),
		Success:      false,
		ErrorMessage: &message,
	}

	err :=
		evaluator.Evaluate(
			context.Background(),
			target,
			result,
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if serviceStore.lastUpdate.RuntimeStatus !=
		service.RuntimeStatusDown {
		t.Fatalf(
			"expected DOWN, got %s",
			serviceStore.lastUpdate.RuntimeStatus,
		)
	}

	if serviceStore.lastUpdate.ConsecutiveFailures != 3 {
		t.Fatalf(
			"expected 3 failures, got %d",
			serviceStore.lastUpdate.ConsecutiveFailures,
		)
	}

	if incidentStore.createCalls != 1 {
		t.Fatalf(
			"expected one incident creation, got %d",
			incidentStore.createCalls,
		)
	}

	if incidentStore.lastCreatedFailureCount != 3 {
		t.Fatalf(
			"expected incident failure count 3, got %d",
			incidentStore.lastCreatedFailureCount,
		)
	}
}

func TestEvaluatorFailureWhileDownKeepsServiceDown(
	t *testing.T,
) {
	serviceStore :=
		&fakeServiceRuntimeStore{}

	existing := Incident{
		ID:           200,
		ServiceID:    3,
		Status:       StatusOpen,
		FailureCount: 3,
	}

	incidentStore :=
		&fakeIncidentStore{
			openIncident: &existing,
		}

	evaluator :=
		NewEvaluator(
			serviceStore,
			incidentStore,
		)

	target := service.Service{
		ID:                   3,
		RuntimeStatus:        service.RuntimeStatusDown,
		ConsecutiveFailures:  0,
		ConsecutiveSuccesses: 1,
	}

	result := healthcheck.Result{
		ServiceID: 3,
		CheckedAt: time.Now(),
		Success:   false,
	}

	err :=
		evaluator.Evaluate(
			context.Background(),
			target,
			result,
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if serviceStore.lastUpdate.RuntimeStatus !=
		service.RuntimeStatusDown {
		t.Fatalf(
			"expected service to stay DOWN, got %s",
			serviceStore.lastUpdate.RuntimeStatus,
		)
	}

	if incidentStore.updateCalls != 1 {
		t.Fatalf(
			"expected open incident to be updated once, got %d",
			incidentStore.updateCalls,
		)
	}

	if incidentStore.lastUpdatedFailureCount != 4 {
		t.Fatalf(
			"expected incident failure count 4, got %d",
			incidentStore.lastUpdatedFailureCount,
		)
	}
}

func TestEvaluatorFirstRecoverySuccessKeepsServiceDown(
	t *testing.T,
) {
	serviceStore :=
		&fakeServiceRuntimeStore{}

	existing := Incident{
		ID:           300,
		ServiceID:    4,
		Status:       StatusOpen,
		FailureCount: 3,
	}

	incidentStore :=
		&fakeIncidentStore{
			openIncident: &existing,
		}

	evaluator :=
		NewEvaluator(
			serviceStore,
			incidentStore,
		)

	target := service.Service{
		ID:                   4,
		RuntimeStatus:        service.RuntimeStatusDown,
		ConsecutiveFailures:  3,
		ConsecutiveSuccesses: 0,
	}

	result := healthcheck.Result{
		ServiceID: 4,
		CheckedAt: time.Now(),
		Success:   true,
	}

	err :=
		evaluator.Evaluate(
			context.Background(),
			target,
			result,
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if serviceStore.lastUpdate.RuntimeStatus !=
		service.RuntimeStatusDown {
		t.Fatalf(
			"expected service to remain DOWN after one success, got %s",
			serviceStore.lastUpdate.RuntimeStatus,
		)
	}

	if serviceStore.lastUpdate.ConsecutiveSuccesses != 1 {
		t.Fatalf(
			"expected 1 consecutive success, got %d",
			serviceStore.lastUpdate.ConsecutiveSuccesses,
		)
	}

	if serviceStore.lastUpdate.ConsecutiveFailures != 0 {
		t.Fatalf(
			"expected failures to reset to 0, got %d",
			serviceStore.lastUpdate.ConsecutiveFailures,
		)
	}

	if incidentStore.resolveCalls != 0 {
		t.Fatal(
			"incident should not resolve after only one success",
		)
	}
}

func TestEvaluatorSecondRecoverySuccessResolvesIncident(
	t *testing.T,
) {
	serviceStore :=
		&fakeServiceRuntimeStore{}

	existing := Incident{
		ID:           400,
		ServiceID:    5,
		Status:       StatusOpen,
		FailureCount: 5,
	}

	incidentStore :=
		&fakeIncidentStore{
			openIncident: &existing,
		}

	evaluator :=
		NewEvaluator(
			serviceStore,
			incidentStore,
		)

	target := service.Service{
		ID:                   5,
		RuntimeStatus:        service.RuntimeStatusDown,
		ConsecutiveFailures:  0,
		ConsecutiveSuccesses: 1,
	}

	result := healthcheck.Result{
		ServiceID: 5,
		CheckedAt: time.Now(),
		Success:   true,
	}

	err :=
		evaluator.Evaluate(
			context.Background(),
			target,
			result,
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if serviceStore.lastUpdate.RuntimeStatus !=
		service.RuntimeStatusHealthy {
		t.Fatalf(
			"expected HEALTHY, got %s",
			serviceStore.lastUpdate.RuntimeStatus,
		)
	}

	if serviceStore.lastUpdate.ConsecutiveSuccesses != 2 {
		t.Fatalf(
			"expected 2 consecutive successes, got %d",
			serviceStore.lastUpdate.ConsecutiveSuccesses,
		)
	}

	if incidentStore.resolveCalls != 1 {
		t.Fatalf(
			"expected one resolve call, got %d",
			incidentStore.resolveCalls,
		)
	}

	if incidentStore.lastResolvedID != 400 {
		t.Fatalf(
			"expected incident 400 to be resolved, got %d",
			incidentStore.lastResolvedID,
		)
	}
}
