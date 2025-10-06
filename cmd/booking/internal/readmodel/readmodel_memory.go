package readmodel

import (
	"sync"
	"time"
)

type MemoryReadModel struct {
	mu sync.Mutex
	// key: spaceID|start|end
	availability map[string]bool
}

func NewMemoryReadModel() *MemoryReadModel {
	return &MemoryReadModel{availability: map[string]bool{}}
}

func (m *MemoryReadModel) CacheAvailability(spaceID string, start, end time.Time, available bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := spaceID + "|" + start.UTC().Format(time.RFC3339) + "|" + end.UTC().Format(time.RFC3339)
	m.availability[key] = available
	return nil
}
