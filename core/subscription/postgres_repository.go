package subscription

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
	subscription *Subscription,
) error {
	// Keep the repository safe for callers that construct subscriptions directly.
	if subscription.ID == "" {
		subscription.ID = uuid.NewString()
	}

	query := `
		INSERT INTO subscriptions (
			id,
			target_url,
			secret,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING
			id,
			target_url,
			secret,
			is_active,
			created_at,
			updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		subscription.ID,
		subscription.TargetURL,
		subscription.Secret,
	).Scan(
		&subscription.ID,
		&subscription.TargetURL,
		&subscription.Secret,
		&subscription.IsActive,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}



func (r *PostgresRepository) GetByID(
	ctx context.Context,
	id string,
) (*Subscription, error) {

	query := `
		SELECT
			id,
			target_url,
			secret,
			is_active,
			created_at,
			updated_at
		FROM subscriptions
		WHERE id = $1
	`

	var subscription Subscription

	err := r.db.QueryRow(ctx, query, id).Scan(
		&subscription.ID,
		&subscription.TargetURL,
		&subscription.Secret,
		&subscription.IsActive,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &subscription, nil
}

func (r *PostgresRepository) GetAll(
	ctx context.Context,
) ([]Subscription, error) {

	query := `
		SELECT
			id,
			target_url,
			secret,
			is_active,
			created_at,
			updated_at
		FROM subscriptions
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	defer rows.Close()

	var subscriptions []Subscription

	for rows.Next() {

		var subscription Subscription

		err := rows.Scan(
			&subscription.ID,
			&subscription.TargetURL,
			&subscription.Secret,
			&subscription.IsActive,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (r *PostgresRepository) Update(
	ctx context.Context,
	subscription *Subscription,
) error {

	query := `
		UPDATE subscriptions
		SET
			target_url = $2,
			secret = $3,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx,
		query,
		subscription.ID,
		subscription.TargetURL,
		subscription.Secret,
	)

	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) Delete(
	ctx context.Context,
	id string,
) error {

	query := `
		DELETE FROM subscriptions
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id)

	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
