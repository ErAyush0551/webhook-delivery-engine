package webhook

import "time"

type Webhook struct {
	ID             string    `json:"id"`
	SubscriptionID string    `json:"subscription_id"`
	Payload        []byte    `json:"payload"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
