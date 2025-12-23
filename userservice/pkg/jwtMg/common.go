package jwtMg

import (
	"errors"
	models "github.com/agris/user-service/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID uuid.UUID   `json:"user_id"`
	Email  string      `json:"email"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

type TokenInfo struct {
	Token     string
	JTI       string
	ExpiresIn int64
	ExpiresAt time.Time
}
