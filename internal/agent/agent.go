package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neptship/calc-yandex-go/internal/config"
	"github.com/neptship/calc-yandex-go/internal/models"
	"github.com/neptship/calc-yandex-go/pkg/calculation"
)

func StartWorkers(ctx context.Context, cfg *config.Config) {
	var wg sync.WaitGroup

	for i := 0; i < cfg.ComputingPower; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			runWorker(ctx, workerID, cfg)
		}(i + 1)
	}
}

func runWorker(ctx context.Context, id int, cfg *config.Config) {
	client := fiber.AcquireClient()
	defer fiber.ReleaseClient(client)

	baseURL := fmt.Sprintf("http://localhost:%d", cfg.Port)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return
		default:
			task, err := fetchTask(client, baseURL)
			if err != nil {
				time.Sleep(time.Duration(cfg.AgentPeriodicityMs) * time.Millisecond)
				continue
			}

			if task == nil {
				time.Sleep(time.Duration(cfg.AgentPeriodicityMs) * time.Millisecond)
				continue
			}

			result, isError := executeOperation(task)

			err = submitResult(client, baseURL, task.ID, result, isError)
			if err != nil {
				log.Printf("Worker %d failed to submit result: %v", id, err)
			}

			time.Sleep(time.Duration(cfg.AgentPeriodicityMs) * time.Millisecond)
		}
	}
}

func fetchTask(client *fiber.Client, baseURL string) (*models.Task, error) {
	agent := client.Get(fmt.Sprintf("%s/internal/task", baseURL))
	statusCode, body, errs := agent.String()
	if len(errs) > 0 {
		return nil, fmt.Errorf("error fetching task: %v", errs[0])
	}
	if statusCode == 404 {
		return nil, nil
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", statusCode)
	}

	var response struct {
		Task models.Task `json:"task"`
	}

	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return nil, fmt.Errorf("error unmarshalling task: %v", err)
	}

	return &response.Task, nil
}

func executeOperation(task *models.Task) (float64, bool) {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	var arg1, arg2 float64
	var err error

	switch v := task.Arg1.(type) {
	case float64:
		arg1 = v
	case int:
		arg1 = float64(v)
	case string:
		arg1, err = strconv.ParseFloat(v, 64)
		if err != nil {
			log.Printf("Error parsing arg1: %v", err)
			return 0, true
		}
	}

	switch v := task.Arg2.(type) {
	case float64:
		arg2 = v
	case int:
		arg2 = float64(v)
	case string:
		arg2, err = strconv.ParseFloat(v, 64)
		if err != nil {
			log.Printf("Error parsing arg2: %v", err)
			return 0, true
		}
	}

	result, err := calculation.EvaluateOperation(arg1, arg2, task.Operation)
	if err != nil {
		log.Printf("Error evaluating operation: %v", err)
		return 0, true
	}

	return result, false
}

func submitResult(client *fiber.Client, baseURL string, taskID int, result float64, isError bool) error {
	data := map[string]interface{}{
		"id":     taskID,
		"result": result,
	}

	if isError {
		data["isError"] = true
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling result: %v", err)
	}

	agent := client.Post(fmt.Sprintf("%s/internal/task", baseURL))
	agent.Body(body)
	agent.Set("Content-Type", "application/json")

	statusCode, _, errs := agent.String()
	if len(errs) > 0 {
		return fmt.Errorf("error submitting result: %v", errs[0])
	}

	if statusCode != 200 {
		return fmt.Errorf("unexpected status code: %d", statusCode)
	}

	return nil
}
