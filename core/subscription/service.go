package subscription

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(
	ctx context.Context,
	targetURL string,
	secret string,
) (*Subscription, error) {

	if targetURL == "" {
		return nil, fmt.Errorf("target URL is required")
	}

	subscription := &Subscription{
		ID:        uuid.New().String(),
		TargetURL: targetURL,
		Secret:    secret,
	}

	if err := s.repository.Create(ctx, subscription); err != nil {
		return nil, err

	}

	return subscription, nil
}



func (s *Service) GetByID(
	ctx context.Context,
	id string,
) (*Subscription, error) {

	if id == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	return s.repository.GetByID(ctx, id)
}

func (s *Service) GetAll(
	ctx context.Context,
) ([]Subscription, error) {

	return s.repository.GetAll(ctx)
}

func (s *Service) Update(
	ctx context.Context,
	id string,
	targetURL string,
	secret string,
) (*Subscription, error) {

	if id == "" {
		return nil, fmt.Errorf("subscription ID is required")
	}

	if targetURL == "" {
		return nil, fmt.Errorf("target URL is required")
	}

	subscription := &Subscription{
		ID:        id,
		TargetURL: targetURL,
		Secret:    secret,
	}

	if err := s.repository.Update(ctx, subscription); err != nil {
		return nil, err
	}

	return s.repository.GetByID(ctx, id)
}

func (s *Service) Delete(
	ctx context.Context,
	id string,
) error {

	if id == "" {
		return fmt.Errorf("subscription ID is required")
	}

	return s.repository.Delete(ctx, id)
}
