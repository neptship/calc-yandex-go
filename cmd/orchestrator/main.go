package main

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/neptship/calc-yandex-go/internal/config"
	"github.com/neptship/calc-yandex-go/internal/orchestrator"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	service := orchestrator.NewService(cfg)

	app := fiber.New()

	api := app.Group("/api/v1")
	api.Post("/calculate", orchestrator.CalculateHandler(service))
	api.Get("/expressions", orchestrator.GetExpressionsHandler(service))
	api.Get("/expressions/:id", orchestrator.GetExpressionHandler(service))

	internal := app.Group("/internal")
	internal.Get("/task", orchestrator.GetTaskHandler(service))
	internal.Post("/task", orchestrator.SubmitTaskResultHandler(service))

	log.Printf("Starting orchestrator on port %d", cfg.Port)
	log.Fatal(app.Listen(":" + strconv.Itoa(cfg.Port)))
}
