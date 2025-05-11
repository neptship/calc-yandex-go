package auth

import (
	"github.com/gofiber/fiber/v2"
)

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func RegisterHandler(service *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Login == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Login and password are required",
			})
		}

		err := service.Register(req.Login, req.Password)
		if err != nil {
			switch err {
			case ErrUserExists:
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "User already exists",
				})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "User registered successfully",
		})
	}
}

func LoginHandler(service *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Login == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Login and password are required",
			})
		}

		token, err := service.Login(req.Login, req.Password)
		if err != nil {
			switch err {
			case ErrInvalidLogin:
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid login or password",
				})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}

		return c.Status(fiber.StatusOK).JSON(LoginResponse{
			Token: token,
		})
	}
}
