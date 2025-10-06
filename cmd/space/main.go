package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	spaceHttp "templespace/cmd/space/internal/http"
	"templespace/cmd/space/internal/queue"
	"templespace/cmd/space/internal/readmodel"
	spaceServer "templespace/cmd/space/internal/server"
	"templespace/cmd/space/internal/service"
	"templespace/cmd/space/internal/storage"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// In-memory dependencies and handlers
	repo := storage.NewInMemorySpaces()
	photos := storage.NewInMemoryPhotos()
	rm := readmodel.NewInMemoryReadModel()
	events := queue.NewInMemoryPublisher()
	auth := spaceHttp.TokenVerifierStub()
	svc := service.New(repo, photos, rm, events, auth)
	handlers := spaceHttp.NewHandlers(svc)
	router := spaceHttp.NewRouter(handlers)

	srv := &http.Server{
		Addr:         ":8082",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("space http server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// Placeholder for gRPC server wiring
	_ = spaceServer.GRPCPlaceholder()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http server shutdown error: %v", err)
		os.Exit(1)
	}
	log.Println("space service gracefully stopped")
}
