package webhook

import (
	"context"
	"fmt"

	"hookflow/core/queue"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	repository  Repository
	redisClient *redis.Client
}

func NewService(repository Repository, redisClient *redis.Client) *Service {
	return &Service{
		repository:  repository,
		redisClient: redisClient,
	}
}

func (s *Service) Create(ctx context.Context, subscriptionID string, payload []byte) (*Webhook, error) {
	if subscriptionID == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	if len(payload) == 0 {
		return nil, fmt.Errorf("payload is required")
	}

	webhook := &Webhook{
		SubscriptionID: subscriptionID,
		Payload:        payload,
		Status:         "pending",
	}

	if err := s.repository.Create(ctx, webhook); err != nil {
		return nil, err
	}

	// Webhook DB me save ho gaya — ab delivery task ko Redis queue me daalo,
	// taaki background worker isse async process kar sake.
	task := queue.DeliveryTask{
		WebhookID:      webhook.ID,
		SubscriptionID: webhook.SubscriptionID,
		AttemptNumber:  1,
	}

	if err := queue.Enqueue(ctx, s.redisClient, task); err != nil {
		return nil, fmt.Errorf("webhook saved but failed to enqueue delivery: %w", err)
	}

	return webhook, nil
}