package readmodel

import (
	"context"
	"strings"
	"sync"

	"templespace/cmd/space/internal/domain"
)

type InMemoryReadModel struct {
	mu   sync.RWMutex
	byID map[string]*domain.Space
}

func NewInMemoryReadModel() *InMemoryReadModel {
	return &InMemoryReadModel{byID: make(map[string]*domain.Space)}
}

func (m *InMemoryReadModel) Index(ctx context.Context, s *domain.Space) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *s
	m.byID[s.ID] = &cp
	return nil
}

func (m *InMemoryReadModel) Search(ctx context.Context, q domain.Query) ([]*domain.Space, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*domain.Space
	for _, s := range m.byID {
		if !matchSpace(s, q) {
			continue
		}
		cp := *s
		out = append(out, &cp)
	}
	return out, nil
}

func matchSpace(s *domain.Space, q domain.Query) bool {
	if q.Name != "" && !strings.Contains(strings.ToLower(s.Name), strings.ToLower(q.Name)) {
		return false
	}
	if q.Location != "" && !strings.Contains(strings.ToLower(s.Location), strings.ToLower(q.Location)) {
		return false
	}
	if q.MinCapacity > 0 {
		if capVal, ok := intFromAttributes(s.Attributes, "capacity"); !ok || capVal < q.MinCapacity {
			return false
		}
	}
	if q.MinPrice > 0 && s.PricePerHour < q.MinPrice {
		return false
	}
	if q.MaxPrice > 0 && s.PricePerHour > q.MaxPrice {
		return false
	}
	if len(q.Tags) > 0 && !hasAllTags(s.Tags, q.Tags) {
		return false
	}
	return true
}

func hasAllTags(have []string, need []string) bool {
	set := make(map[string]struct{}, len(have))
	for _, t := range have {
		set[strings.ToLower(t)] = struct{}{}
	}
	for _, want := range need {
		if _, ok := set[strings.ToLower(want)]; !ok {
			return false
		}
	}
	return true
}

func intFromAttributes(attrs map[string]any, key string) (int, bool) {
	if attrs == nil {
		return 0, false
	}
	if v, ok := attrs[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n), true
		case int:
			return n, true
		case int32:
			return int(n), true
		case int64:
			return int(n), true
		}
	}
	return 0, false
}
