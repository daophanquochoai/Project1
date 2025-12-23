package dto

import (
	models "github.com/agris/user-service/internal/model"
	"github.com/google/uuid"
	"time"
)

type UserRequest struct {
	Name         string `json:"name" validate:"required,min=6"`
	Email        string `json:"email" validate:"required,email"`
	PasswordHash string `json:"password" validate:"required,min=6"`
}

type UpdateRoleRequest struct {
	AccountId uuid.UUID   `json:"accountId" validate:"required"`
	UserId    uuid.UUID   `json:"userId" validate:"required"`
	Role      models.Role `json:"role" validate:"required"`
}

type UpdateRoleResponse struct {
	Id        uuid.UUID   `json:"id" validate:"required"`
	Email     string      `json:"email" validate:"required,email"`
	Role      models.Role `json:"role" validate:"required"`
	UpdatedAt time.Time   `json:"updated_at" validate:"required"`
}

type RegisterUserRequest struct {
	Id         uuid.UUID   `json:"id"`
	Name       string      `json:"name"`
	Email      string      `json:"email"`
	Role       models.Role `json:"role"`
	Created_at time.Time   `json:"created_at"`
}

// status
type Status string

var (
	StatusActive   Status = "ACTIVE"
	StatusInactive Status = "INACTIVE"
)

type GetUserResponse struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Status     Status    `json:"status"`
	Created_at time.Time `json:"created_at"`
}
