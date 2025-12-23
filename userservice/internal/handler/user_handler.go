package handler

import (
	"context"
	"github.com/agris/user-service/internal/dto"
	"github.com/agris/user-service/internal/service"
	"github.com/agris/user-service/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"time"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) UpdateUserRole(c *fiber.Ctx) error {

	account, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": ErrInternalServerError,
		})
	}

	userId := c.Params("userId")
	if userId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData,
		})
	}
	userIdConvert, err := uuid.Parse(userId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData,
		})
	}

	var request dto.UpdateRoleRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ErrInvalidData,
		})
	}

	request = dto.UpdateRoleRequest{
		AccountId: account,
		UserId:    userIdConvert,
		Role:      request.Role,
	}

	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userUpdated, errUpdated := h.userService.UpdateUserRole(ct, &request)
	if errUpdated != nil {
		return c.Status(errUpdated.Status).JSON(fiber.Map{
			"error": errUpdated.Err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(userUpdated)
}

func (h *UserHandler) GetListUser(c *fiber.Ctx) error {

	var pageRequest dto.PageRequest
	if err := c.BodyParser(&pageRequest); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	ct, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	users, err := h.userService.ListUsers(ct, &pageRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(users)
}

func (h *UserHandler) GetCurrentUserInfo(c *fiber.Ctx) error {

	id := c.Locals("userID").(uuid.UUID)

	ct, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	user, err := h.userService.GetCurrentUser(ct, id)
	if err != nil {
		return c.Status(err.Status).JSON(fiber.Map{
			"error": err.Err.Error(),
		})
	}

	return c.JSON(user)
}

func (h *UserHandler) CreateUser(ctx *fiber.Ctx) error {
	var userRequest dto.UserRequest
	if err := ctx.BodyParser(&userRequest); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := utils.ValidateStruct(&userRequest); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": utils.FormatValidationError(err),
		})
	}

	ct, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	user, err := h.userService.CreateUser(ct, &userRequest)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(user)
}
