package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	cfg "templespace/cmd/auth/internal/config"
	h "templespace/cmd/auth/internal/http"
	srv "templespace/cmd/auth/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c := cfg.Load()
	addr := net.JoinHostPort("0.0.0.0", c.HTTPPort)

	handler := h.NewHandler()
	mux := handler.Routes()

	httpSrv := srv.NewHTTP(mux, addr)
	go func() {
		if err := httpSrv.Listen(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Server.Shutdown(shutdownCtx)
}
