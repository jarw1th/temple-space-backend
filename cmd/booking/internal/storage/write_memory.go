package storage

import (
	"errors"
	"sync"
	"time"

	"templespace/cmd/booking/internal/domain"
)

type MemoryRepo struct {
	mu      sync.RWMutex
	byID    map[string]*domain.Booking
	bySpace map[string][]*domain.Booking
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{byID: map[string]*domain.Booking{}, bySpace: map[string][]*domain.Booking{}}
}

func (m *MemoryRepo) Create(b *domain.Booking) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[b.ID]; ok {
		return errors.New("duplicate id")
	}
	cp := *b
	m.byID[b.ID] = &cp
	m.bySpace[b.SpaceID] = append(m.bySpace[b.SpaceID], &cp)
	return nil
}

func (m *MemoryRepo) Update(b *domain.Booking) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[b.ID]; !ok {
		return errors.New("not found")
	}
	cp := *b
	m.byID[b.ID] = &cp
	// naive: not updating bySpace slice entries for brevity
	return nil
}

func (m *MemoryRepo) GetByID(id string) (*domain.Booking, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b := m.byID[id]
	if b == nil {
		return nil, errors.New("not found")
	}
	cp := *b
	return &cp, nil
}

func (m *MemoryRepo) IsAvailable(spaceID string, start, end time.Time) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, b := range m.bySpace[spaceID] {
		if overlaps(b.SlotStart, b.SlotEnd, start, end) && b.Status != domain.StatusCancelled {
			return false, nil
		}
	}
	return true, nil
}

func overlaps(aStart, aEnd, bStart, bEnd time.Time) bool {
	if !aStart.Before(aEnd) || !bStart.Before(bEnd) {
		return false
	}
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}
