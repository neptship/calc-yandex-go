package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neptship/calc-yandex-go/internal/agent"
	"github.com/neptship/calc-yandex-go/internal/config"
	"github.com/neptship/calc-yandex-go/internal/grpc"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	grpcHost := os.Getenv("GRPC_HOST")
	if grpcHost == "" {
		grpcHost = cfg.GRPCHost
	}
	grpcAddr := fmt.Sprintf("%s:%d", grpcHost, cfg.GRPCPort)

	grpcClient, err := grpc.NewGRPCClient(grpcAddr)
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer grpcClient.Close()

	log.Printf("Starting %d agent workers with gRPC connection to %s", cfg.ComputingPower, grpcAddr)

	for i := 0; i < cfg.ComputingPower; i++ {
		go func(workerID int) {
			agent.RunWorker(ctx, workerID, cfg, grpcClient)
		}(i + 1)
	}

	<-ctx.Done()
	log.Println("Shutting down agent workers...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()
	<-shutdownCtx.Done()
}
