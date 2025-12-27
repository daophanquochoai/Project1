package handler

import (
	"context"
	"fmt"
	"net/http"
	"productservice/internal/dto"
	"productservice/internal/service"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RateHandler struct {
	rateService service.RateService
}

func NewRateHandler(rateService service.RateService) *RateHandler {
	return &RateHandler{rateService: rateService}
}

func (h *RateHandler) GetMyRating(ctx *fiber.Ctx) error {
	fmt.Println(ctx.Locals("userId"))
	userId := ctx.Locals("userId").(string)
	userUuid, err := uuid.Parse(userId)
	if userId == "" || err != nil {
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": ErrUnAuthorized,
		})
	}

	var request dto.MyRatingsRequest
	request.Page = ctx.QueryInt("page", 1)
	request.Limit = ctx.QueryInt("limit", 5)
	request.SortBy = ctx.Query("sort", "created_at")

	ct, cancel := context.WithTimeout(ctx.Context(), 5*time.Minute)
	defer cancel()

	rateUser, errRate := h.rateService.GetMyRatings(ct, userUuid, &request)
	if errRate != nil {
		return ctx.Status(errRate.Status).JSON(fiber.Map{
			"error": errRate.Err.Error(),
		})
	}

	return ctx.JSON(rateUser)
}

func (h *RateHandler) GetRateStatisticOfProduct(ctx *fiber.Ctx) error {
	productId := ctx.Params("productId")
	if productId == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData.Error(),
		})
	}
	productUuuid, err := uuid.Parse(productId)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData.Error(),
		})
	}

	ct, cancel := context.WithTimeout(ctx.Context(), 5*time.Minute)
	defer cancel()

	statistic, errStatistic := h.rateService.GetRateStatisticOfProduct(ct, productUuuid)
	if errStatistic != nil {
		return ctx.Status(errStatistic.Status).JSON(fiber.Map{
			"error": errStatistic.Err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(statistic)
}

func (h *RateHandler) GetRateListOfProduct(ctx *fiber.Ctx) error {
	productId := ctx.Params("productId")
	if productId == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData.Error(),
		})
	}
	productUuuid, err := uuid.Parse(productId)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData.Error(),
		})
	}

	var request dto.RateListOfProduct
	request.Page = ctx.QueryInt("page", 1)

	request.Limit = ctx.QueryInt("limit", 5)

	request.SortBy = ctx.Query("sort", "created_at")

	request.ProductId = productUuuid

	ct, cancel := context.WithTimeout(ctx.Context(), 5*time.Minute)
	defer cancel()

	rateList, errRate := h.rateService.GetRateListOfProduct(ct, &request)
	if errRate != nil {
		return ctx.Status(errRate.Status).JSON(fiber.Map{
			"error": errRate.Err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(rateList)
}

func (h *RateHandler) DeleteRateProduct(ctx *fiber.Ctx) error {
	ratingId := ctx.Params("ratingId")
	ratingUuid, errParse := uuid.Parse(ratingId)
	if ratingId == "" || errParse != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData.Error(),
		})
	}

	ct, cancel := context.WithTimeout(ctx.Context(), 5*time.Minute)
	defer cancel()

	err := h.rateService.DeleteRatingProduct(ct, ratingUuid)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Err.Error(),
		})
	}
	return ctx.Status(http.StatusNoContent).JSON(fiber.Map{})
}

func (h *RateHandler) UpdateRateProduct(ctx *fiber.Ctx) error {
	rateId := ctx.Params("ratingId")
	if rateId == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData.Error(),
		})
	}
	rateUuid, err := uuid.Parse(rateId)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData.Error(),
		})
	}

	ct, cancel := context.WithTimeout(ctx.Context(), 5*time.Minute)
	defer cancel()

	var request dto.UpdateRatingProduct
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData.Error(),
		})
	}

	request.RateId = rateUuid
	rating, errRating := h.rateService.UpdateRatingProduct(ct, &request)
	if errRating != nil {
		return ctx.Status(errRating.Status).JSON(fiber.Map{
			"error": errRating.Err.Error(),
		})
	}

	return ctx.JSON(rating)
}

func (h *RateHandler) RateProduct(ctx *fiber.Ctx) error {
	var rateProduct dto.RateProduct
	if err := ctx.BodyParser(&rateProduct); err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrInvalidData,
		})
	}

	ct, cancel := context.WithTimeout(ctx.Context(), 5*time.Minute)
	defer cancel()

	productId := ctx.Params("productId")
	productUuid, err := uuid.Parse(productId)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrInvalidData,
		})
	}
	userId := ctx.Locals("userId").(string)
	userUuid, err := uuid.Parse(userId)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrInvalidData,
		})
	}

	rateProduct.ProductId = productUuid
	rateProduct.UserId = userUuid

	rateResponse, errRate := h.rateService.RatingProduct(ct, &rateProduct)
	if errRate != nil {
		return ctx.Status(errRate.Status).JSON(fiber.Map{
			"error": errRate.Err.Error(),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(rateResponse)
}
