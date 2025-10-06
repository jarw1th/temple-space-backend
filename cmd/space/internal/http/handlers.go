package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"templespace/cmd/space/internal/domain"
	"templespace/cmd/space/internal/service"
)

type tokenVerifierStub struct{}

func (tokenVerifierStub) Verify(ctx context.Context, token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", http.ErrNoCookie
	}
	return "user-1", nil
}

// TokenVerifierStub returns a simple auth verifier for development wiring.
func TokenVerifierStub() service.TokenVerifier { return tokenVerifierStub{} }

type Handlers struct {
	svc *service.Service
}

func NewHandlers(svc *service.Service) *Handlers { return &Handlers{svc: svc} }

type createSpaceRequest struct {
	Name         string         `json:"name"`
	Location     string         `json:"location"`
	Tags         []string       `json:"tags"`
	Attributes   map[string]any `json:"attributes"`
	PricePerHour float64        `json:"price_per_hour"`
}

func (h *Handlers) handleCreateSpace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req createSpaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	access := r.Header.Get("Authorization")
	sp := &domain.Space{Name: req.Name, Location: req.Location, Tags: req.Tags, Attributes: req.Attributes, PricePerHour: req.PricePerHour}
	out, err := h.svc.CreateSpace(r.Context(), access, sp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(out)
}

func (h *Handlers) handleListSpaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	q := domain.Query{
		Name:        r.URL.Query().Get("name"),
		Location:    r.URL.Query().Get("location"),
		Tags:        splitCSV(r.URL.Query().Get("tags")),
		MinCapacity: atoiDefault(r.URL.Query().Get("min_capacity")),
		MinPrice:    atofDefault(r.URL.Query().Get("min_price")),
		MaxPrice:    atofDefault(r.URL.Query().Get("max_price")),
	}
	res, err := h.svc.ListSpaces(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *Handlers) handleUpdateSpace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/spaces/")
	var req createSpaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	access := r.Header.Get("Authorization")
	sp := &domain.Space{ID: id, Name: req.Name, Location: req.Location, Tags: req.Tags, Attributes: req.Attributes, PricePerHour: req.PricePerHour, UpdatedAt: time.Now().UTC()}
	out, err := h.svc.UpdateSpace(r.Context(), access, sp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func atoiDefault(s string) int {
	if s == "" {
		return 0
	}
	var n int
	_, _ = fmt.Sscanf(s, "%d", &n)
	return n
}

func atofDefault(s string) float64 {
	if s == "" {
		return 0
	}
	var n float64
	_, _ = fmt.Sscanf(s, "%f", &n)
	return n
}
