//go:build wireinject
// +build wireinject

package internal

import (
	"encoding/json"
	"productservice/config"
	"productservice/internal/cache"
	"productservice/internal/database"
	"productservice/internal/grpc/client"
	"productservice/internal/grpc/provider"
	"productservice/internal/handler"
	"productservice/internal/middleware"
	"productservice/internal/repository"
	"productservice/internal/router"
	"productservice/internal/service"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	recoverFiber "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/wire"
)

func New() (*fiber.App, func(), error) {
	panic(wire.Build(
		config.Set,
		database.Set,
		cache.Set,
		provider.Set, // Cung cấp *grpc.ClientConn và cleanup
		client.Set,   // Cung cấp *client.AuthClient
		repository.Set,
		service.Set,
		handler.Set,
		router.Set,
		middleware.Set, // Cung cấp *middleware.Middleware
		NewServer,
	))
}

func NewServer(router *router.RouterHandler, md *middleware.Middleware) *fiber.App {
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

	router.InitRouter(app, md)

	return app
}
