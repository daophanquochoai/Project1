package middleware

import (
	models "github.com/agris/user-service/internal/model"
	"github.com/agris/user-service/pkg/jwtMg"
	"github.com/gofiber/fiber/v2"
	"strings"
)

type AuthMiddleware struct {
	jwtManager *jwtMg.JWTManager
}

func NewAuthMiddleware(jwtManager *jwtMg.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (atw *AuthMiddleware) Authorize() fiber.Handler {
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

		claims, err := atw.jwtManager.ValidateToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": ErrAuth,
			})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("claims", claims)
		return c.Next()
	}
}

func (atw *AuthMiddleware) RequireRole(roles ...models.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {

		userRole, ok := c.Locals("role").(models.Role)

		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": ok,
			})
		}

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": ErrPermission,
		})
	}
}

func (atw *AuthMiddleware) Optional() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}

		token := parts[1]
		claims, err := atw.jwtManager.ValidateToken(token)
		if err == nil {
			c.Locals("userID", claims.UserID)
			c.Locals("email", claims.Email)
			c.Locals("role", claims.Role)
			c.Locals("claims", claims)
		}

		return c.Next()
	}
}
