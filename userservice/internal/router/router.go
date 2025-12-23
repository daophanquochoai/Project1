package router

import (
	"github.com/agris/user-service/internal/handler"
	"github.com/agris/user-service/internal/middleware"
	models "github.com/agris/user-service/internal/model"
	"github.com/gofiber/fiber/v2"
)

type RouterHandler struct {
	authApi *handler.AuthHandler
	userApi *handler.UserHandler
	md      *middleware.Middleware
}

func NewRouterHandler(authApi *handler.AuthHandler, userApi *handler.UserHandler, md *middleware.Middleware) *RouterHandler {
	return &RouterHandler{authApi: authApi, userApi: userApi, md: md}
}

func (r *RouterHandler) InitRouter(root *fiber.App) {
	root.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "userservice",
		})
	})

	authGroup := (*root).Group("/users")
	authGroup.Use(r.md.Auth.Optional())
	authGroup.Post("/login", r.authApi.Login)
	authGroup.Post("/register", r.userApi.CreateUser)

	userGroup := (*root).Group("/users")
	userGroup.Use(r.md.Auth.Authorize())
	userGroup.Get("/me", r.userApi.GetCurrentUserInfo)

	userGroupWithAdminRole := (*root).Group("/users")
	userGroupWithAdminRole.Use(r.md.Auth.RequireRole(models.RoleAdmin))
	userGroupWithAdminRole.Get("/list", r.userApi.GetListUser)
	userGroupWithAdminRole.Patch("/:userId/role", r.userApi.UpdateUserRole)
}
