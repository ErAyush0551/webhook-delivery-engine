package webhook

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	webhook *Webhook,
) error {
	// Defensive check: agar ID khaali hai to generate kar do
	// (repository khud responsible hai, kisi upar wali layer par bharosa nahi karta)
	if webhook.ID == "" {
		webhook.ID = uuid.NewString()
	}

	// Defensive check: agar status khaali hai to default "pending" set karo
	// warna khaali string ("") save ho jayegi, DB ka DEFAULT 'pending' override ho jayega
	if webhook.Status == "" {
		webhook.Status = "pending"
	}

	query := `
		INSERT INTO webhooks (
			id,
			subscription_id,
			payload,
			status
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			subscription_id,
			payload,
			status,
			created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		webhook.ID,
		webhook.SubscriptionID,
		webhook.Payload,
		webhook.Status,
	).Scan(
		&webhook.ID,
		&webhook.SubscriptionID,
		&webhook.Payload,
		&webhook.Status,
		&webhook.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetByID(
	ctx context.Context,
	id string,
) (*Webhook, error) {
	query := `
		SELECT
			id,
			subscription_id,
			payload,
			status,
			created_at
		FROM webhooks
		WHERE id = $1
	`

	var webhook Webhook
	err := r.db.QueryRow(ctx, query, id).Scan(
		&webhook.ID,
		&webhook.SubscriptionID,
		&webhook.Payload,
		&webhook.Status,
		&webhook.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("webhook not found: %w", err)
		}

		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	return &webhook, nil
}

func (r *PostgresRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status string,
) error {
	const query = `
		UPDATE webhooks
		SET status = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update webhook status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("webhook not found: %s", id)
	}

	return nil
}
