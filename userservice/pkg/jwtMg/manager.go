package jwtMg

import (
	"errors"
	"github.com/agris/user-service/internal/config"
	"time"

	"github.com/agris/user-service/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTManager(config *config.Config) *JWTManager {
	return &JWTManager{
		secretKey:     config.JWT.Secret,
		accessExpiry:  config.JWT.AccessExpiry,
		refreshExpiry: config.JWT.RefreshExpiry,
	}
}

// tạo JWT access token với JTI
func (j *JWTManager) GenerateAccessToken(user *models.User, timeEx time.Duration) (*TokenInfo, error) {
	now := time.Now()
	expiresAt := now.Add(timeEx)
	jti := uuid.New().String()

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "user-service",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return nil, err
	}

	return &TokenInfo{
		Token:     tokenString,
		JTI:       jti,
		ExpiresIn: int64(j.accessExpiry.Seconds()),
		ExpiresAt: expiresAt,
	}, nil
}

func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
