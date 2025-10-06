package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"templespace/cmd/space/internal/domain"
)

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (userID string, err error)
}

type Service struct {
	repo      domain.SpaceRepository
	photos    domain.PhotoRepository
	readModel domain.ReadModel
	events    domain.EventPublisher
	auth      TokenVerifier
}

func New(repo domain.SpaceRepository, photos domain.PhotoRepository, readModel domain.ReadModel, events domain.EventPublisher, auth TokenVerifier) *Service {
	return &Service{repo: repo, photos: photos, readModel: readModel, events: events, auth: auth}
}

func (s *Service) CreateSpace(ctx context.Context, accessToken string, sp *domain.Space) (*domain.Space, error) {
	if _, err := s.auth.Verify(ctx, accessToken); err != nil {
		return nil, err
	}
	if sp.Name == "" {
		return nil, errors.New("name required")
	}
	sp.ID = generateID()
	sp.CreatedAt = time.Now().UTC()
	sp.UpdatedAt = sp.CreatedAt
	sp.Version = 1
	if err := s.repo.Create(ctx, sp); err != nil {
		return nil, err
	}
	_ = s.readModel.Index(ctx, sp)
	payload, _ := json.Marshal(sp)
	_ = s.events.Publish("space_created", sp.ID, payload)
	return sp, nil
}

func (s *Service) UpdateSpace(ctx context.Context, accessToken string, sp *domain.Space) (*domain.Space, error) {
	if _, err := s.auth.Verify(ctx, accessToken); err != nil {
		return nil, err
	}
	existing, err := s.repo.GetByID(ctx, sp.ID)
	if err != nil {
		return nil, err
	}
	existing.Name = sp.Name
	existing.Location = sp.Location
	existing.Tags = sp.Tags
	existing.Attributes = sp.Attributes
	existing.PricePerHour = sp.PricePerHour
	existing.Version++
	existing.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	_ = s.readModel.Index(ctx, existing)
	payload, _ := json.Marshal(existing)
	_ = s.events.Publish("space_updated", existing.ID, payload)
	return existing, nil
}

func (s *Service) GetSpace(ctx context.Context, id string) (*domain.Space, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListSpaces(ctx context.Context, q domain.Query) ([]*domain.Space, error) {
	return s.readModel.Search(ctx, q)
}

func generateID() string {
	return time.Now().UTC().Format("20060102150405.000000000")
}
