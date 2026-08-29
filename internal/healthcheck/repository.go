package healthcheck

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
) (Result, error) {

	const query = `
		INSERT INTO health_check_results (
			service_id,
			status_code,
			latency_ms,
			success,
			error_message
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			service_id,
			checked_at,
			status_code,
			latency_ms,
			success,
			error_message
	`

	var created Result

	err := repository.pool.QueryRow(
		ctx,
		query,
		params.ServiceID,
		params.StatusCode,
		params.LatencyMS,
		params.Success,
		params.ErrorMessage,
	).Scan(
		&created.ID,
		&created.ServiceID,
		&created.CheckedAt,
		&created.StatusCode,
		&created.LatencyMS,
		&created.Success,
		&created.ErrorMessage,
	)

	if err != nil {
		return Result{}, fmt.Errorf(
			"create health check result: %w",
			err,
		)
	}

	return created, nil
}

func (repository *Repository) FindRecentByServiceID(
	ctx context.Context,
	serviceID int64,
	limit int,
) ([]Result, error) {

	const query = `
		SELECT
			id,
			service_id,
			checked_at,
			status_code,
			latency_ms,
			success,
			error_message
		FROM health_check_results
		WHERE service_id = $1
		ORDER BY checked_at DESC
		LIMIT $2
	`

	rows, err := repository.pool.Query(
		ctx,
		query,
		serviceID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query recent health checks: %w",
			err,
		)
	}

	defer rows.Close()

	results := make(
		[]Result,
		0,
	)

	for rows.Next() {
		var current Result

		err := rows.Scan(
			&current.ID,
			&current.ServiceID,
			&current.CheckedAt,
			&current.StatusCode,
			&current.LatencyMS,
			&current.Success,
			&current.ErrorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"scan health check result: %w",
				err,
			)
		}

		results = append(
			results,
			current,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate health check results: %w",
			err,
		)
	}

	return results, nil
}
