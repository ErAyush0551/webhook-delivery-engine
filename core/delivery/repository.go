package delivery

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, log *Log) error
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