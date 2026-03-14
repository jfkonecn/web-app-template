package main

import (
	"log"

	"github.com/jfkonecn/web-app-template/internal/config"
	"github.com/jfkonecn/web-app-template/internal/server"
)

func main() {
	cfg := config.Load()
	r := server.NewRouter()

	log.Printf("starting server on %s", cfg.Address())
	if err := r.Run(cfg.Address()); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
