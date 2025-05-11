package orchestrator

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/neptship/calc-yandex-go/internal/models"
)

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResponse struct {
	ID int `json:"id"`
}

type ExpressionsResponse struct {
	Expressions []*models.Expression `json:"expressions"`
}

type ExpressionResponse struct {
	Expression *ExpressionWithoutDuplication `json:"expression"`
}

type ExpressionWithoutDuplication struct {
	ID     int      `json:"id"`
	Status string   `json:"status"`
	Result *float64 `json:"result,omitempty"`
}

type TaskResponse struct {
	Task *models.Task `json:"task"`
}

type TaskResultRequest struct {
	ID      int     `json:"id"`
	Result  float64 `json:"result"`
	IsError bool    `json:"isError"`
}

func CalculateHandler(service *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int)

		var req CalculateRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request format",
			})
		}

		if req.Expression == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Expression cannot be empty",
			})
		}

		if matched, _ := regexp.MatchString(`^-?\d+(\.\d+)?$`, req.Expression); matched {
			id, err := service.AddSimpleExpression(userID, req.Expression)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": err.Error(),
				})
			}

			return c.Status(fiber.StatusCreated).JSON(fiber.Map{
				"id": id,
			})
		}

		id, err := service.AddExpression(userID, req.Expression)
		if err != nil {
			if errors.Is(err, ErrInvalidExpression) {
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
					"error": "invalid expression",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "internal server error",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": id,
		})
	}
}

func GetExpressionsHandler(service *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int)

		expressions, err := service.GetAllExpressions(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get expressions",
			})
		}

		cleanExpressions := make([]*ExpressionWithoutDuplication, len(expressions))
		for i, expr := range expressions {
			cleanExpressions[i] = &ExpressionWithoutDuplication{
				ID:     expr.ID,
				Status: string(expr.Status),
				Result: expr.Result,
			}
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"expressions": cleanExpressions,
		})
	}
}

func GetExpressionHandler(service *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int)

		idStr := c.Params("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "Invalid expression ID",
			})
		}

		expression, err := service.GetExpressionByID(userID, id)
		if err != nil {
			if err == ErrExpressionNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Expression not found",
				})
			} else if err == ErrUnauthorized {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Access denied",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		cleanExpression := &ExpressionWithoutDuplication{
			ID:     expression.ID,
			Status: string(expression.Status),
			Result: expression.Result,
		}

		return c.Status(fiber.StatusOK).JSON(ExpressionResponse{
			Expression: cleanExpression,
		})
	}
}

func GetTaskHandler(service *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		task, err := service.GetNextTask()
		if err != nil {
			if err == ErrTaskNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "No tasks available",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		return c.Status(fiber.StatusOK).JSON(TaskResponse{
			Task: task,
		})
	}
}

func SubmitTaskResultHandler(service *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req TaskResultRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		var err error
		if req.IsError {
			err = service.SetTaskError(req.ID, "Calculation error")
		} else {
			err = service.SetTaskResult(req.ID, req.Result)
		}

		if err != nil {
			if err == ErrTaskNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Task not found",
				})
			}
			if err == ErrInvalidTaskResult {
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
					"error": "Invalid task result",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
		})
	}
}
