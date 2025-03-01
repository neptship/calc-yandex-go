package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/neptship/calc-yandex-go/internal/agent"
	"github.com/neptship/calc-yandex-go/internal/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("Starting %d agent workers", cfg.ComputingPower)
	agent.StartWorkers(ctx, cfg)

	<-ctx.Done()
	log.Println("Shutting down agent workers")
}
