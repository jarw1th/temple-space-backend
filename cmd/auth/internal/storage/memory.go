package storage

import (
	"sync"
	"time"
)

type MagicTokenStore interface {
	Save(token string, email string, ttl time.Duration)
	Get(token string) (string, bool)
	Delete(token string)
}

type RefreshTokenStore interface {
	Save(token string, userID string, ttl time.Duration)
	Get(token string) (string, bool)
	Delete(token string)
}

type InMemoryKV struct {
	mu sync.RWMutex
	m  map[string]entry
}

type entry struct {
	v      string
	expiry time.Time
}

func NewInMemoryKV() *InMemoryKV { return &InMemoryKV{m: make(map[string]entry)} }

func (s *InMemoryKV) set(key, val string, ttl time.Duration) {
	s.mu.Lock()
	exp := time.Time{}
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	s.m[key] = entry{v: val, expiry: exp}
	s.mu.Unlock()
}

func (s *InMemoryKV) get(key string) (string, bool) {
	s.mu.RLock()
	e, ok := s.m[key]
	s.mu.RUnlock()
	if !ok {
		return "", false
	}
	if !e.expiry.IsZero() && time.Now().After(e.expiry) {
		s.mu.Lock()
		delete(s.m, key)
		s.mu.Unlock()
		return "", false
	}
	return e.v, true
}

func (s *InMemoryKV) del(key string) {
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()
}

// Implement interfaces

type InMemoryMagic InMemoryKV

type InMemoryRefresh InMemoryKV

func NewInMemoryMagic() *InMemoryMagic     { return (*InMemoryMagic)(NewInMemoryKV()) }
func NewInMemoryRefresh() *InMemoryRefresh { return (*InMemoryRefresh)(NewInMemoryKV()) }

func (s *InMemoryMagic) Save(token string, email string, ttl time.Duration) {
	(*InMemoryKV)(s).set(token, email, ttl)
}
func (s *InMemoryMagic) Get(token string) (string, bool) { return (*InMemoryKV)(s).get(token) }
func (s *InMemoryMagic) Delete(token string)             { (*InMemoryKV)(s).del(token) }

func (s *InMemoryRefresh) Save(token string, userID string, ttl time.Duration) {
	(*InMemoryKV)(s).set(token, userID, ttl)
}
func (s *InMemoryRefresh) Get(token string) (string, bool) { return (*InMemoryKV)(s).get(token) }
func (s *InMemoryRefresh) Delete(token string)             { (*InMemoryKV)(s).del(token) }
