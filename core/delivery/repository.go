package delivery

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, log *Log) error
	GetByWebhookID(ctx context.Context, webhookID string) ([]Log, error)
	ListBySubscription(ctx context.Context, subscriptionID string, limit int) ([]Log, error)
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, log *Log) error {
	if log.ID == "" {
		log.ID = uuid.NewString()
	}

	query := `
		INSERT INTO delivery_attempts (
			id, webhook_id, subscription_id, target_url,
			attempt, status, http_status, error
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at
	`

	err := r.db.QueryRow(
		ctx, query,
		log.ID, log.WebhookID, log.SubscriptionID, log.TargetURL,
		log.AttemptNumber, log.Status, log.HTTPStatus, log.Error,
	).Scan(&log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create delivery log: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByWebhookID(ctx context.Context, webhookID string) ([]Log, error) {
	query := `
		SELECT id, webhook_id, subscription_id, target_url, attempt, status, http_status, error, created_at
		FROM delivery_attempts
		WHERE webhook_id = $1
		ORDER BY attempt ASC
	`
	rows, err := r.db.Query(ctx, query, webhookID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery logs: %w", err)
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.ID, &l.WebhookID, &l.SubscriptionID, &l.TargetURL,
			&l.AttemptNumber, &l.Status, &l.HTTPStatus, &l.Error, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan delivery log: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading delivery logs: %w", err)
	}
	return logs, nil
}

func (r *PostgresRepository) ListBySubscription(ctx context.Context, subscriptionID string, limit int) ([]Log, error) {
	query := `
		SELECT id, webhook_id, subscription_id, target_url, attempt, status, http_status, error, created_at
		FROM delivery_attempts
		WHERE subscription_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, subscriptionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list delivery logs: %w", err)
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.ID, &l.WebhookID, &l.SubscriptionID, &l.TargetURL,
			&l.AttemptNumber, &l.Status, &l.HTTPStatus, &l.Error, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan delivery log: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading delivery logs: %w", err)
	}
	return logs, nil
}

func (r *PostgresRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	result, err := r.db.Exec(ctx, `DELETE FROM delivery_attempts WHERE created_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}
	return result.RowsAffected(), nil
}

//we have made this file to define the repository interface and its implementation for the delivery logs.
// we have used the pgxpool package to connect to the database and perform CRUD operations on the delivery_attempts table.
//we have defined methods to create a log, get logs by webhook ID, list logs by subscription ID, and delete old logs.
//we have used the context package to pass context to the database operations for better control over timeouts and cancellations.
//we have used the uuid package to generate unique IDs for the logs.
//we have used the time package to handle timestamps for the logs.
//postgresRepository struct implements the Repository interface and provides the actual database operations using SQL queries.
