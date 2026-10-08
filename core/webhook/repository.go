

package webhook

import "context"

type Repository interface {
	Create(ctx context.Context, webhook *Webhook) error
	GetByID(ctx context.Context, id string) (*Webhook, error)
	UpdateStatus(ctx context.Context, id string, status string) error
}