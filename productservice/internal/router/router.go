package router

import (
	"productservice/internal/handler"
	"productservice/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

type RouterHandler struct {
	productApi *handler.ProductHandler
	rateApi    *handler.RateHandler
}

func NewRouterHandler(productApi *handler.ProductHandler, rateApi *handler.RateHandler) *RouterHandler {
	return &RouterHandler{
		productApi: productApi,
		rateApi:    rateApi,
	}
}

func (r *RouterHandler) InitRouter(root *fiber.App, md *middleware.Middleware) {
	root.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "productservice",
		})
	})

	productGroup := root.Group("/products")
	productGroup.Get("/list", r.productApi.GetListProduct)
	productGroup.Get("/ratings/:productId", r.rateApi.GetRateListOfProduct)
	productGroup.Get("/ratings/statistic/:productId", r.rateApi.GetRateStatisticOfProduct)
	productGroup.Get("/product/:productId", r.productApi.GetProductById)
	productGroup.Get("/product/:productId/similar", r.productApi.GetProductSimilar)
	productGroup.Get("/product/:productId/related", r.productApi.GetProductRelated)

	ratingProductGroup := root.Group("/products")
	ratingProductGroup.Use(md.Auth.Handler())
	ratingProductGroup.Post("/:productId/ratings", r.rateApi.RateProduct)
	ratingProductGroup.Put("/ratings/:ratingId", r.rateApi.UpdateRateProduct)
	ratingProductGroup.Delete("/ratings/:rateingId", r.rateApi.DeleteRateProduct)

	ratingGroup := root.Group("/ratings")
	ratingGroup.Use(md.Auth.Handler())
	ratingGroup.Get("/me", r.rateApi.GetMyRating)

}
