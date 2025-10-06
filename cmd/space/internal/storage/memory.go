package storage

import (
	"context"
	"errors"
	"sync"

	"templespace/cmd/space/internal/domain"
)

type InMemorySpaces struct {
	mu   sync.RWMutex
	byID map[string]*domain.Space
}

func NewInMemorySpaces() *InMemorySpaces {
	return &InMemorySpaces{byID: make(map[string]*domain.Space)}
}

func (m *InMemorySpaces) Create(ctx context.Context, s *domain.Space) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.byID[s.ID]; exists {
		return errors.New("exists")
	}
	cp := *s
	m.byID[s.ID] = &cp
	return nil
}

func (m *InMemorySpaces) Update(ctx context.Context, s *domain.Space) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[s.ID]; !ok {
		return errors.New("not found")
	}
	cp := *s
	m.byID[s.ID] = &cp
	return nil
}

func (m *InMemorySpaces) GetByID(ctx context.Context, id string) (*domain.Space, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.byID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *s
	return &cp, nil
}

func (m *InMemorySpaces) List(ctx context.Context) ([]*domain.Space, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*domain.Space, 0, len(m.byID))
	for _, s := range m.byID {
		cp := *s
		out = append(out, &cp)
	}
	return out, nil
}

type InMemoryPhotos struct {
	mu      sync.RWMutex
	bySpace map[string][]*domain.SpacePhoto
}

func NewInMemoryPhotos() *InMemoryPhotos {
	return &InMemoryPhotos{bySpace: make(map[string][]*domain.SpacePhoto)}
}

func (m *InMemoryPhotos) AddPhoto(ctx context.Context, p *domain.SpacePhoto) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *p
	m.bySpace[p.SpaceID] = append(m.bySpace[p.SpaceID], &cp)
	return nil
}

func (m *InMemoryPhotos) ListBySpace(ctx context.Context, spaceID string) ([]*domain.SpacePhoto, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := m.bySpace[spaceID]
	out := make([]*domain.SpacePhoto, 0, len(list))
	for _, p := range list {
		cp := *p
		out = append(out, &cp)
	}
	return out, nil
}
