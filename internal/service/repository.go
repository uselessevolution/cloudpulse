package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("service not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(
	pool *pgxpool.Pool,
) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (repository *Repository) Create(
	ctx context.Context,
	params CreateParams,
) (Service, error) {

	const query = `
		INSERT INTO services (
			name,
			url,
			expected_status,
			check_interval_seconds,
			timeout_seconds,
			enabled
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
		id,
		name,
		url,
		expected_status,
		check_interval_seconds,
		timeout_seconds,
		enabled,
		runtime_status,
		consecutive_failures,
		consecutive_successes,
		last_checked_at,
		created_at,
		updated_at
	`

	var created Service

	err := repository.pool.QueryRow(
		ctx,
		query,
		params.Name,
		params.URL,
		params.ExpectedStatus,
		params.CheckIntervalSeconds,
		params.TimeoutSeconds,
		params.Enabled,
	).Scan(
		&created.ID,
		&created.Name,
		&created.URL,
		&created.ExpectedStatus,
		&created.CheckIntervalSeconds,
		&created.TimeoutSeconds,
		&created.Enabled,
		&created.RuntimeStatus,
		&created.ConsecutiveFailures,
		&created.ConsecutiveSuccesses,
		&created.LastCheckedAt,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return Service{}, fmt.Errorf(
			"create service: %w",
			err,
		)
	}

	return created, nil
}

func (repository *Repository) FindAll(
	ctx context.Context,
) ([]Service, error) {

	const query = `
		SELECT
				id,
				name,
				url,
				expected_status,
				check_interval_seconds,
				timeout_seconds,
				enabled,
				runtime_status,
				consecutive_failures,
				consecutive_successes,
				last_checked_at,
				created_at,
				updated_at
			FROM services
		ORDER BY id ASC
	`

	rows, err := repository.pool.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query services: %w",
			err,
		)
	}

	defer rows.Close()

	services := make(
		[]Service,
		0,
	)

	for rows.Next() {
		var current Service

		err := rows.Scan(
			&current.ID,
			&current.Name,
			&current.URL,
			&current.ExpectedStatus,
			&current.CheckIntervalSeconds,
			&current.TimeoutSeconds,
			&current.Enabled,
			&current.RuntimeStatus,
			&current.ConsecutiveFailures,
			&current.ConsecutiveSuccesses,
			&current.LastCheckedAt,
			&current.CreatedAt,
			&current.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"scan service: %w",
				err,
			)
		}

		services = append(
			services,
			current,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate services: %w",
			err,
		)
	}

	return services, nil
}
func (repository *Repository) FindDueForChecks(
	ctx context.Context,
	now time.Time,
) ([]Service, error) {

	const query = `
		SELECT
			s.id,
			s.name,
			s.url,
			s.expected_status,
			s.check_interval_seconds,
			s.timeout_seconds,
			s.enabled,
			s.runtime_status,
			s.consecutive_failures,
			s.consecutive_successes,
			s.last_checked_at,
			s.created_at,
			s.updated_at
		FROM services s
		LEFT JOIN LATERAL (
			SELECT h.checked_at
			FROM health_check_results h
			WHERE h.service_id = s.id
			ORDER BY h.checked_at DESC
			LIMIT 1
		) latest_check ON TRUE
		WHERE
			s.enabled = TRUE
			AND (
    			latest_check.checked_at IS NULL
    			OR latest_check.checked_at
        		+ make_interval(
            	secs => s.check_interval_seconds
        	)
        	<= $1::timestamptz
			)
		ORDER BY s.id ASC
	`

	rows, err := repository.pool.Query(
		ctx,
		query,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query services due for checks: %w",
			err,
		)
	}

	defer rows.Close()

	services := make(
		[]Service,
		0,
	)

	for rows.Next() {
		var current Service

		err := rows.Scan(
			&current.ID,
			&current.Name,
			&current.URL,
			&current.ExpectedStatus,
			&current.CheckIntervalSeconds,
			&current.TimeoutSeconds,
			&current.Enabled,
			&current.RuntimeStatus,
			&current.ConsecutiveFailures,
			&current.ConsecutiveSuccesses,
			&current.LastCheckedAt,
			&current.CreatedAt,
			&current.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"scan due service: %w",
				err,
			)
		}

		services = append(
			services,
			current,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate due services: %w",
			err,
		)
	}

	return services, nil
}
func (repository *Repository) FindByID(
	ctx context.Context,
	id int64,
) (Service, error) {

	const query = `
		SELECT
			id,
			name,
			url,
			expected_status,
			check_interval_seconds,
			timeout_seconds,
			enabled,
			runtime_status,
			consecutive_failures,
			consecutive_successes,
			last_checked_at,
			created_at,
			updated_at
		FROM services
		WHERE id = $1
	`

	var found Service

	err := repository.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&found.ID,
		&found.Name,
		&found.URL,
		&found.ExpectedStatus,
		&found.CheckIntervalSeconds,
		&found.TimeoutSeconds,
		&found.Enabled,
		&found.RuntimeStatus,
		&found.ConsecutiveFailures,
		&found.ConsecutiveSuccesses,
		&found.LastCheckedAt,
		&found.CreatedAt,
		&found.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Service{}, ErrNotFound
	}

	if err != nil {
		return Service{}, fmt.Errorf(
			"find service by id: %w",
			err,
		)
	}

	return found, nil
}

func (repository *Repository) Delete(
	ctx context.Context,
	id int64,
) error {

	const query = `
		DELETE FROM services
		WHERE id = $1
	`

	result, err := repository.pool.Exec(
		ctx,
		query,
		id,
	)
	if err != nil {
		return fmt.Errorf(
			"delete service: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
func (repository *Repository) UpdateRuntimeState(
	ctx context.Context,
	id int64,
	runtimeStatus string,
	consecutiveFailures int,
	consecutiveSuccesses int,
	lastCheckedAt time.Time,
) (Service, error) {

	const query = `
		UPDATE services
		SET
			runtime_status = $2,
			consecutive_failures = $3,
			consecutive_successes = $4,
			last_checked_at = $5,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			name,
			url,
			expected_status,
			check_interval_seconds,
			timeout_seconds,
			enabled,
			runtime_status,
			consecutive_failures,
			consecutive_successes,
			last_checked_at,
			created_at,
			updated_at
	`

	var updated Service

	err := repository.pool.QueryRow(
		ctx,
		query,
		id,
		runtimeStatus,
		consecutiveFailures,
		consecutiveSuccesses,
		lastCheckedAt,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.URL,
		&updated.ExpectedStatus,
		&updated.CheckIntervalSeconds,
		&updated.TimeoutSeconds,
		&updated.Enabled,
		&updated.RuntimeStatus,
		&updated.ConsecutiveFailures,
		&updated.ConsecutiveSuccesses,
		&updated.LastCheckedAt,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Service{}, ErrNotFound
	}

	if err != nil {
		return Service{}, fmt.Errorf(
			"update service runtime state: %w",
			err,
		)
	}

	return updated, nil
}
