package middleware

import (
	"github.com/gofiber/fiber/v2"
	"productservice/internal/grpc/client"
	"strings"
)

type AuthMiddleware struct {
	authClient *client.AuthClient
}

func NewAuthMiddleware(authClient *client.AuthClient) *AuthMiddleware {
	return &AuthMiddleware{authClient: authClient}
}

func (am *AuthMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": ErrTokenInvalid,
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": ErrTokenInvalid,
			})
		}

		token := parts[1]

		resp, err := am.authClient.Authenticate(c.Context(), token)
		if err != nil || !resp.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": ErrAuth,
			})
		}

		c.Locals("userId", resp.UserId)
		c.Locals("role", resp.Role)
		c.Locals("token", token)

		return c.Next()
	}
}
