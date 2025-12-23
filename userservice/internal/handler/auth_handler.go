package handler

import (
	"context"
	"github.com/agris/user-service/internal/dto"
	"github.com/agris/user-service/internal/service"
	"github.com/agris/user-service/internal/utils"
	"github.com/gofiber/fiber/v2"
	"time"
)

type AuthHandler struct {
	s service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{s: s}
}

func (h *AuthHandler) Login(ctx *fiber.Ctx) error {
	var loginRequest dto.LoginRequest
	if err := ctx.BodyParser(&loginRequest); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := utils.ValidateStruct(&loginRequest); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": utils.FormatValidationError(err),
		})
	}

	ct, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	response, err := h.s.Login(ct, &loginRequest)
	if err != nil {
		return ctx.Status(err.Status).JSON(fiber.Map{
			"error": err.Err.Error(),
		})
	}

	return ctx.JSON(response)
}
