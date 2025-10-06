package main

import (
	"log"
	"os"

	"templespace/cmd/booking/internal/config"
	httpserver "templespace/cmd/booking/internal/server"
)

func main() {
	cfg := config.FromEnv()

	// Start HTTP server (gRPC server can be added similarly via build tags like in Auth)
	srv := httpserver.NewHTTPServer(cfg)
	if err := srv.Listen(cfg.HTTPAddr); err != nil {
		log.Println("http server error:", err)
		os.Exit(1)
	}
}
