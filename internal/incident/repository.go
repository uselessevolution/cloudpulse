package incident

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("incident not found")

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

func (repository *Repository) CreateOpen(
	ctx context.Context,
	serviceID int64,
	failureCount int,
	lastErrorMessage *string,
) (Incident, error) {

	const query = `
		INSERT INTO incidents (
			service_id,
			status,
			failure_count,
			last_error_message
		)
		VALUES (
			$1,
			'OPEN',
			$2,
			$3
		)
		RETURNING
			id,
			service_id,
			status,
			started_at,
			resolved_at,
			failure_count,
			last_error_message,
			created_at,
			updated_at
	`

	var created Incident

	err := repository.pool.QueryRow(
		ctx,
		query,
		serviceID,
		failureCount,
		lastErrorMessage,
	).Scan(
		&created.ID,
		&created.ServiceID,
		&created.Status,
		&created.StartedAt,
		&created.ResolvedAt,
		&created.FailureCount,
		&created.LastErrorMessage,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		return Incident{}, fmt.Errorf(
			"create open incident: %w",
			err,
		)
	}

	return created, nil
}
func (repository *Repository) FindOpenByServiceID(
	ctx context.Context,
	serviceID int64,
) (Incident, error) {

	const query = `
		SELECT
			id,
			service_id,
			status,
			started_at,
			resolved_at,
			failure_count,
			last_error_message,
			created_at,
			updated_at
		FROM incidents
		WHERE
			service_id = $1
			AND status = 'OPEN'
		ORDER BY started_at DESC
		LIMIT 1
	`

	var found Incident

	err := repository.pool.QueryRow(
		ctx,
		query,
		serviceID,
	).Scan(
		&found.ID,
		&found.ServiceID,
		&found.Status,
		&found.StartedAt,
		&found.ResolvedAt,
		&found.FailureCount,
		&found.LastErrorMessage,
		&found.CreatedAt,
		&found.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Incident{}, ErrNotFound
	}

	if err != nil {
		return Incident{}, fmt.Errorf(
			"find open incident: %w",
			err,
		)
	}

	return found, nil
}
func (repository *Repository) UpdateOpen(
	ctx context.Context,
	serviceID int64,
	failureCount int,
	lastErrorMessage *string,
) (Incident, error) {

	const query = `
		UPDATE incidents
		SET
			failure_count = $2,
			last_error_message = $3,
			updated_at = NOW()
		WHERE
			service_id = $1
			AND status = 'OPEN'
		RETURNING
			id,
			service_id,
			status,
			started_at,
			resolved_at,
			failure_count,
			last_error_message,
			created_at,
			updated_at
	`

	var updated Incident

	err := repository.pool.QueryRow(
		ctx,
		query,
		serviceID,
		failureCount,
		lastErrorMessage,
	).Scan(
		&updated.ID,
		&updated.ServiceID,
		&updated.Status,
		&updated.StartedAt,
		&updated.ResolvedAt,
		&updated.FailureCount,
		&updated.LastErrorMessage,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Incident{}, ErrNotFound
	}

	if err != nil {
		return Incident{}, fmt.Errorf(
			"update open incident: %w",
			err,
		)
	}

	return updated, nil
}
func (repository *Repository) Resolve(
	ctx context.Context,
	id int64,
	resolvedAt time.Time,
) (Incident, error) {

	const query = `
		UPDATE incidents
		SET
			status = 'RESOLVED',
			resolved_at = $2,
			updated_at = NOW()
		WHERE
			id = $1
			AND status = 'OPEN'
		RETURNING
			id,
			service_id,
			status,
			started_at,
			resolved_at,
			failure_count,
			last_error_message,
			created_at,
			updated_at
	`

	var resolved Incident

	err := repository.pool.QueryRow(
		ctx,
		query,
		id,
		resolvedAt,
	).Scan(
		&resolved.ID,
		&resolved.ServiceID,
		&resolved.Status,
		&resolved.StartedAt,
		&resolved.ResolvedAt,
		&resolved.FailureCount,
		&resolved.LastErrorMessage,
		&resolved.CreatedAt,
		&resolved.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Incident{}, ErrNotFound
	}

	if err != nil {
		return Incident{}, fmt.Errorf(
			"resolve incident: %w",
			err,
		)
	}

	return resolved, nil
}
func (repository *Repository) FindAll(
	ctx context.Context,
) ([]Incident, error) {

	const query = `
		SELECT
			id,
			service_id,
			status,
			started_at,
			resolved_at,
			failure_count,
			last_error_message,
			created_at,
			updated_at
		FROM incidents
		ORDER BY started_at DESC
	`

	rows, err :=
		repository.pool.Query(
			ctx,
			query,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"query incidents: %w",
			err,
		)
	}

	defer rows.Close()

	incidents :=
		make(
			[]Incident,
			0,
		)

	for rows.Next() {
		var current Incident

		err := rows.Scan(
			&current.ID,
			&current.ServiceID,
			&current.Status,
			&current.StartedAt,
			&current.ResolvedAt,
			&current.FailureCount,
			&current.LastErrorMessage,
			&current.CreatedAt,
			&current.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"scan incident: %w",
				err,
			)
		}

		incidents =
			append(
				incidents,
				current,
			)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate incidents: %w",
			err,
		)
	}

	return incidents, nil
}
func (repository *Repository) FindByID(
	ctx context.Context,
	id int64,
) (Incident, error) {

	const query = `
		SELECT
			id,
			service_id,
			status,
			started_at,
			resolved_at,
			failure_count,
			last_error_message,
			created_at,
			updated_at
		FROM incidents
		WHERE id = $1
	`

	var found Incident

	err :=
		repository.pool.QueryRow(
			ctx,
			query,
			id,
		).Scan(
			&found.ID,
			&found.ServiceID,
			&found.Status,
			&found.StartedAt,
			&found.ResolvedAt,
			&found.FailureCount,
			&found.LastErrorMessage,
			&found.CreatedAt,
			&found.UpdatedAt,
		)

	if errors.Is(
		err,
		pgx.ErrNoRows,
	) {
		return Incident{},
			ErrNotFound
	}

	if err != nil {
		return Incident{},
			fmt.Errorf(
				"find incident by id: %w",
				err,
			)
	}

	return found, nil
}
func (repository *Repository) CountOpen(
	ctx context.Context,
) (int, error) {

	const query = `
		SELECT COUNT(*)
		FROM incidents
		WHERE status = 'OPEN'
	`

	var count int

	err := repository.pool.QueryRow(
		ctx,
		query,
	).Scan(
		&count,
	)

	if err != nil {
		return 0, fmt.Errorf(
			"count open incidents: %w",
			err,
		)
	}

	return count, nil
}
