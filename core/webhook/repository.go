// package webhook

// import "context"

// Repository interface batata hai ki webhook ke liye
// koi bhi storage implementation (Postgres, Mongo, ya mock)
// ye methods zaroor implement kare.
// type Repository interface {
// 	Create(ctx context.Context, webhook *Webhook) error
// 	GetByID(ctx context.Context, id string) (*Webhook, error)
// 	UpdateStatus(ctx context.Context, id string, status string) error
// 	ListBySubscription(ctx context.Context, subscriptionID string, limit int) ([]Webhook, error)
// }


package webhook

import "context"

type Repository interface {
	Create(ctx context.Context, webhook *Webhook) error
	GetByID(ctx context.Context, id string) (*Webhook, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}