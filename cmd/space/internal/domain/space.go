package domain

import (
	"context"
	"time"
)

type Space struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Location     string         `json:"location"`
	Tags         []string       `json:"tags"`
	Attributes   map[string]any `json:"attributes"`
	PricePerHour float64        `json:"price_per_hour"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Version      int64          `json:"version"`
}

type SpacePhoto struct {
	ID        string    `json:"id"`
	SpaceID   string    `json:"space_id"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type SpaceRepository interface {
	Create(ctx context.Context, s *Space) error
	Update(ctx context.Context, s *Space) error
	GetByID(ctx context.Context, id string) (*Space, error)
	List(ctx context.Context) ([]*Space, error)
}

type PhotoRepository interface {
	AddPhoto(ctx context.Context, p *SpacePhoto) error
	ListBySpace(ctx context.Context, spaceID string) ([]*SpacePhoto, error)
}

type ReadModel interface {
	Index(ctx context.Context, s *Space) error
	Search(ctx context.Context, q Query) ([]*Space, error)
}

type EventPublisher interface {
	Publish(topic, key string, payload []byte) error
}

type Query struct {
	Name        string
	Location    string
	Tags        []string
	MinCapacity int
	MinPrice    float64
	MaxPrice    float64
}
