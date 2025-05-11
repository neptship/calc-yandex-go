package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/neptship/calc-yandex-go/internal/auth"
	"github.com/neptship/calc-yandex-go/internal/config"
	"github.com/neptship/calc-yandex-go/internal/database"
	"github.com/neptship/calc-yandex-go/internal/grpc"
	"github.com/neptship/calc-yandex-go/internal/orchestrator"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewDatabase(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	authService, err := auth.NewService(db.GetDB())
	if err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}

	service := orchestrator.NewService(cfg, db)

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		log.Printf("Incoming request: Method=%s, Path=%s, OriginalURL=%s", c.Method(), c.Path(), c.OriginalURL())
		return c.Next()
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	api := app.Group("/api/v1")
	api.Post("/register", auth.RegisterHandler(authService))
	api.Post("/login", auth.LoginHandler(authService))

	apiProtected := api.Group("/")
	apiProtected.Use(auth.AuthMiddleware(authService))
	apiProtected.Post("/calculate", orchestrator.CalculateHandler(service))
	apiProtected.Get("/expressions", orchestrator.GetExpressionsHandler(service))
	apiProtected.Get("/expressions/:id", orchestrator.GetExpressionHandler(service))

	internal := app.Group("/internal")
	internal.Get("/task", orchestrator.GetTaskHandler(service))
	internal.Post("/task", orchestrator.SubmitTaskResultHandler(service))

	go func() {
		grpcAddr := fmt.Sprintf("%s:%d", "0.0.0.0", cfg.GRPCPort)
		log.Printf("Starting gRPC server on %s", grpcAddr)
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("Failed to listen for gRPC: %v", err)
		}
		if err := grpc.StartGRPCServer(service, lis); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	log.Printf("Starting HTTP server on port %d", cfg.Port)
	log.Fatal(app.Listen(":" + strconv.Itoa(cfg.Port)))
}
