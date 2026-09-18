package subscription

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("subscription not found")

type Subscription struct {
	ID        string    `json:"id"`
	TargetURL string    `json:"target_url"`
	Secret    string    `json:"secret"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
