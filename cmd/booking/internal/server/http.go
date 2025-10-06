package server

import (
	"encoding/json"
	"log"
	"net/http"

	"templespace/cmd/booking/internal/config"
)

type HTTPServer struct {
	cfg *config.Config
}

func NewHTTPServer(cfg *config.Config) *HTTPServer {
	return &HTTPServer{cfg: cfg}
}

func (s *HTTPServer) Listen(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/booking", s.handleCreateBooking)
	mux.HandleFunc("/booking/", s.handleBookingActions)
	log.Printf("http listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

type createBookingRequest struct {
	SpaceID   string `json:"space_id"`
	UserID    string `json:"user_id"`
	SlotStart string `json:"slot_start"`
	SlotEnd   string `json:"slot_end"`
}

func (s *HTTPServer) handleCreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: integrate with domain service
	_ = req
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "pending", "message": "stub"})
}

func (s *HTTPServer) handleBookingActions(w http.ResponseWriter, r *http.Request) {
	// naive routing for /booking/{id}/pay and /booking/{id}/cancel
	switch {
	case r.Method == http.MethodPost && hasSuffix(r.URL.Path, "/pay"):
		s.handlePay(w, r)
	case r.Method == http.MethodPost && hasSuffix(r.URL.Path, "/cancel"):
		s.handleCancel(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func hasSuffix(path, suffix string) bool {
	if len(path) < len(suffix) {
		return false
	}
	return path[len(path)-len(suffix):] == suffix
}

func (s *HTTPServer) handlePay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "paid", "message": "stub"})
}

func (s *HTTPServer) handleCancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "cancelled", "message": "stub"})
}
