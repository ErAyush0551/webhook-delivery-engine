package subscription

import "context"

type Repository interface {
	Create(ctx context.Context, subscription *Subscription) error
	GetByID(ctx context.Context, id string) (*Subscription, error)
	GetAll(ctx context.Context) ([]Subscription, error)
	Update(ctx context.Context, subscription *Subscription) error
	Delete(ctx context.Context, id string) error
}
