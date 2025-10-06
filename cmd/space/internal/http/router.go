package http

import (
	stdhttp "net/http"
)

func NewRouter(h *Handlers) *stdhttp.ServeMux {
	mux := stdhttp.NewServeMux()

	mux.HandleFunc("/healthz", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/spaces", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		switch r.Method {
		case stdhttp.MethodPost:
			h.handleCreateSpace(w, r)
		case stdhttp.MethodGet:
			h.handleListSpaces(w, r)
		default:
			w.WriteHeader(stdhttp.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/spaces/", h.handleUpdateSpace) // expects PUT /spaces/{id}

	return mux
}
