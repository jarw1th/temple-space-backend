package http

import (
	"encoding/json"
	"log"
	stdhttp "net/http"
	"time"

	"templespace/cmd/auth/internal/auth"
	"templespace/cmd/auth/internal/storage"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Routes() *stdhttp.ServeMux {
	mux := stdhttp.NewServeMux()
	mux.HandleFunc("/healthz", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// wire in-memory deps for bootstrap
	magic := storage.NewInMemoryMagic()
	refresh := storage.NewInMemoryRefresh()
	users := storage.NewInMemoryUsers()
	signer := &auth.JWTSigner{Issuer: "templespace", Secret: []byte("dev-secret-change-me")}
	svc := auth.NewService(signer, 10*time.Minute, 60*time.Minute, 720*time.Hour, magic, refresh)
	svc.Users = users

	mux.HandleFunc("/auth/login", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method != stdhttp.MethodPost {
			w.WriteHeader(stdhttp.StatusMethodNotAllowed)
			return
		}
		var body struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
			w.WriteHeader(stdhttp.StatusBadRequest)
			return
		}
		token, err := svc.StartLogin(body.Email)
		if err != nil {
			log.Println("start login error:", err)
			w.WriteHeader(stdhttp.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"magic_token": token})
	})

	mux.HandleFunc("/auth/verify", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method != stdhttp.MethodGet {
			w.WriteHeader(stdhttp.StatusMethodNotAllowed)
			return
		}
		token := r.URL.Query().Get("token")
		if token == "" {
			w.WriteHeader(stdhttp.StatusBadRequest)
			return
		}
		access, refreshToken, expiresIn, err := svc.VerifyMagicToken(token)
		if err != nil {
			w.WriteHeader(stdhttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  access,
			"refresh_token": refreshToken,
			"expires_in":    expiresIn,
		})
	})

	// JWKS for API Gateway verification if RSA is configured
	mux.HandleFunc("/.well-known/jwks.json", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method != stdhttp.MethodGet {
			w.WriteHeader(stdhttp.StatusMethodNotAllowed)
			return
		}
		if signer.RSAPublic != nil {
			jwks := auth.JWKS{Keys: []auth.JWK{auth.RSAPublicToJWK(signer.RSAPublic, signer.KeyID)}}
			_ = json.NewEncoder(w).Encode(jwks)
			return
		}
		_ = json.NewEncoder(w).Encode(auth.JWKS{Keys: []auth.JWK{}})
	})
	return mux
}
