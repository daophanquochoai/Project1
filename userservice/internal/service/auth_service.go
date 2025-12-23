package service

import (
	"context"
	"errors"
	"github.com/agris/user-service/internal/cache"
	"github.com/agris/user-service/internal/config"
	"github.com/agris/user-service/internal/dto"
	"github.com/agris/user-service/internal/repository"
	"github.com/agris/user-service/internal/utils"
	"github.com/agris/user-service/pkg/jwtMg"
	"github.com/gofiber/fiber/v2/log"
	"time"
)

type AuthService interface {
	Login(ctx context.Context, request *dto.LoginRequest) (*dto.AuthResponse, *dto.ServiceResponse)
	ValidateToken(token string) bool
}

type authService struct {
	userRepo   repository.UserRepository
	jwtManager *jwtMg.JWTManager
	cfg        *config.Config
	rdRepo     cache.RedisRepository
}

func NewAuthService(userRepo repository.UserRepository, jwtManager *jwtMg.JWTManager, cfg *config.Config, rdRepo cache.RedisRepository) AuthService {
	return &authService{userRepo: userRepo, jwtManager: jwtManager, cfg: cfg, rdRepo: rdRepo}
}

func (s *authService) ValidateToken(token string) bool {
	if token == "" {
		return false
	}
	_, err := s.jwtManager.ValidateToken(token)
	return err == nil
}

func (s *authService) Login(ctx context.Context, request *dto.LoginRequest) (*dto.AuthResponse, *dto.ServiceResponse) {
	// Validate input
	if request.Email == "" || request.Password == "" {
		response := dto.ServiceResponse{
			Status: 400,
			Err:    errors.New(ErrInvalidData),
		}
		return nil, &response
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, request.Email, true)
	if err != nil || user == nil {
		return nil, err
	}

	if user.DeletedAt.Valid {
		response := dto.ServiceResponse{
			Status: 403,
			Err:    errors.New(ErrLockedAccount),
		}
		return nil, &response
	}

	// Check password
	if !utils.CheckPasswordHash(request.Password, user.PasswordHash) {
		response := dto.ServiceResponse{
			Status: 403,
			Err:    errors.New(ErrPasswordNotMatch),
		}
		return nil, &response
	}

	// expire refresh
	expireRefresh := time.Now().Add(s.cfg.JWT.RefreshExpiry)
	// Generate JWT token
	accessToken, errAccess := s.jwtManager.GenerateAccessToken(user, s.cfg.JWT.AccessExpiry)
	refreshToken, errRefresh := s.jwtManager.GenerateAccessToken(user, s.cfg.JWT.RefreshExpiry)

	if errAccess != nil || errRefresh != nil {
		response := dto.ServiceResponse{
			Status: 403,
			Err:    errors.New(ErrInternalServerError),
		}
		return nil, &response
	}

	//  Build response

	userResponse := dto.UserTokenResponse{
		Id:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}
	resp := &dto.AuthResponse{
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken.Token,
		ExpiresIn:    expireRefresh.Unix(),
		User:         &userResponse,
	}

	errRedis := s.rdRepo.SaveRefreshToken(ctx, user.ID, refreshToken.Token, refreshToken.ExpiresAt)
	if errRedis != nil {
		log.Error("[ERROR] : ", errRedis.Error())
	}

	return resp, nil
}
