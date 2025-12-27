package handler

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"net/http"
	"productservice/internal/dto"
	"productservice/internal/service"
	"strings"
	"time"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

func (h *ProductHandler) GetProductRelated(ctx *fiber.Ctx) error {
	// Lấy productId từ URL params
	productIdStr := ctx.Params("productId")
	productId, err := uuid.Parse(productIdStr)
	if err != nil {
		log.Error(ErrInvalidData)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData.Error(),
		})
	}

	// Lấy query parameters
	pageRequest := &dto.PageSimilarAndRelatedRequest{
		Limit: ctx.QueryInt("limit", 5),
		Page:  ctx.QueryInt("page", 1),
	}

	// Gọi service/repository
	result, serviceResp := h.productService.GetProductSimilar(ctx.Context(), productId, pageRequest, &dto.Relation_Related)
	if serviceResp != nil {
		return ctx.Status(serviceResp.Status).JSON(fiber.Map{
			"error": serviceResp.Err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(result)
}

func (h *ProductHandler) GetProductSimilar(ctx *fiber.Ctx) error {
	// Lấy productId từ URL params
	productIdStr := ctx.Params("productId")
	productId, err := uuid.Parse(productIdStr)
	if err != nil {
		log.Error(ErrInvalidData)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData.Error(),
		})
	}

	// Lấy query parameters
	pageRequest := &dto.PageSimilarAndRelatedRequest{
		Limit: ctx.QueryInt("limit", 5),
		Page:  ctx.QueryInt("page", 1),
	}

	// Gọi service/repository
	result, serviceResp := h.productService.GetProductSimilar(ctx.Context(), productId, pageRequest, &dto.Relation_Similar)
	if serviceResp != nil {
		return ctx.Status(serviceResp.Status).JSON(fiber.Map{
			"error": serviceResp.Err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(result)
}

func (h *ProductHandler) GetProductById(ctx *fiber.Ctx) error {
	productId := ctx.Params("productId")
	productUuid, err := uuid.Parse(productId)
	if productId == "" || err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData,
		})
	}

	ct, cencel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cencel()

	product, errProduct := h.productService.GetProductById(ct, productUuid)
	if errProduct != nil {
		return ctx.Status(errProduct.Status).JSON(fiber.Map{
			"error": errProduct.Err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(product)
}

func (h *ProductHandler) GetListProduct(ctx *fiber.Ctx) error {
	var request dto.PageProdRequest

	// Lấy page với default = 1
	request.Page = ctx.QueryInt("page", 1)

	// Lấy limit với default = 20
	request.Limit = ctx.QueryInt("limit", 20)

	// Lấy search với default = ""
	request.Search = ctx.Query("search", "")

	// Lấy sort với default = ""
	request.SortBy = ctx.Query("sort", "")

	// Lấy order với default = ""
	request.SortOrder = ctx.Query("order", "")

	// Lấy min_price với default = 0
	request.MinPrice = ctx.QueryInt("min_price", 0)

	// Lấy max_price với default = 0
	request.MaxPrice = ctx.QueryInt("max_price", 0)

	// Lấy min_rate với default = 0
	request.MinRate = ctx.QueryInt("min_rate", 0)

	// Lấy max_rate với default = 0
	request.MaxRate = ctx.QueryInt("max_rate", 0)

	// Lấy category_ids (có thể có nhiều category)
	categoryIdsStr := ctx.Query("category_ids", "")
	if categoryIdsStr != "" {
		categoryIdsList := strings.Split(categoryIdsStr, ",")
		for _, idStr := range categoryIdsList {
			id, err := uuid.Parse(strings.TrimSpace(idStr))
			if err == nil {
				request.CategoryIds = append(request.CategoryIds, id)
			}
		}
	}

	// Gọi service
	response, err := h.productService.GetList(ctx.Context(), &request)
	if err != nil {
		return ctx.Status(err.Status).JSON(fiber.Map{
			"error": err.Err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(response)
}
