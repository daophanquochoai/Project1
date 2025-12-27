//go:build wireinject
// +build wireinject

package internal

import (
	"encoding/json"
	"github.com/agris/user-service/config"
	"github.com/agris/user-service/internal/cache"
	"github.com/agris/user-service/internal/database"
	"github.com/agris/user-service/internal/handler"
	"github.com/agris/user-service/internal/middleware"
	"github.com/agris/user-service/internal/repository"
	"github.com/agris/user-service/internal/router"
	"github.com/agris/user-service/internal/service"
	"github.com/agris/user-service/pkg/jwtMg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	recoverFiber "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/wire"
	"time"
)

func New() (*fiber.App, error) {
	panic(wire.Build(
		config.Set,
		database.Set,
		cache.Set,
		service.Set,
		handler.Set,
		repository.Set,
		router.Set,
		jwtMg.Set,
		middleware.Set,
		NewServer,
	))
}

func NewServer(router *router.RouterHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		JSONDecoder:  json.Unmarshal,
		JSONEncoder:  json.Marshal,
	})

	app.Use(logger.New())
	app.Use(cors.New())
	recoverConfig := recoverFiber.ConfigDefault
	app.Use(recoverFiber.New(recoverConfig))

	router.InitRouter(app)

	return app
}
