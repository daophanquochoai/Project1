package service

import (
	"context"
	"errors"
	"github.com/agris/user-service/internal/dto"
	models "github.com/agris/user-service/internal/model"
	"github.com/agris/user-service/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type UserService interface {
	CreateUser(ctx context.Context, userRequest *dto.UserRequest) (*dto.RegisterUserRequest, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.GetUserResponse, *dto.ServiceResponse)
	ListUsers(ctx context.Context, pageRequest *dto.PageRequest) (*dto.PageResponse, error)
	UpdateUserRole(ctx context.Context, updateAccount *dto.UpdateRoleRequest) (*dto.UpdateRoleResponse, *dto.ServiceResponse)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (u *userService) UpdateUserRole(ctx context.Context, updateAccount *dto.UpdateRoleRequest) (*dto.UpdateRoleResponse, *dto.ServiceResponse) {

	if updateAccount.UserId == updateAccount.AccountId {
		response := dto.ServiceResponse{
			Status: http.StatusBadRequest,
			Err:    errors.New(ErrCantUpdateOwnAccount),
		}
		return nil, &response
	}

	user, err := u.userRepo.FindByID(ctx, updateAccount.UserId)
	if err != nil {
		response := dto.ServiceResponse{
			Status: http.StatusBadRequest,
			Err:    errors.New(ErrUserNotFound),
		}
		return nil, &response
	}

	// update
	user.Role = updateAccount.Role

	// save
	userUpdated, errUpdated := u.userRepo.Update(ctx, user)
	if errUpdated != nil {
		return nil, errUpdated
	}

	// response
	response := dto.UpdateRoleResponse{
		Id:        userUpdated.ID,
		Email:     userUpdated.Email,
		Role:      userUpdated.Role,
		UpdatedAt: userUpdated.UpdatedAt,
	}

	return &response, nil

}

func (u *userService) CreateUser(ctx context.Context, req *dto.UserRequest) (*dto.RegisterUserRequest, error) {
	// Validate input
	if req == nil || req.Email == "" || req.PasswordHash == "" || req.Name == "" {
		return nil, errors.New(ErrInvalidData)
	}

	// Check email exists
	existingUser, _ := u.userRepo.FindByEmail(ctx, req.Email, true)
	if existingUser != nil {
		return nil, errors.New(ErrEmailExists)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New(ErrHashPassword)
	}

	// Create user model
	user := &models.User{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         models.RoleUser,
	}

	// Save to database
	userCreated, err := u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	response := dto.RegisterUserRequest{
		Id:         userCreated.ID,
		Name:       userCreated.Name,
		Email:      userCreated.Email,
		Role:       userCreated.Role,
		Created_at: userCreated.CreatedAt,
	}

	return &response, nil
}

func (u *userService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*dto.GetUserResponse, *dto.ServiceResponse) {

	if userID == uuid.Nil {
		response := dto.ServiceResponse{
			Status: http.StatusNotFound,
			Err:    errors.New(ErrInvalidData),
		}
		return nil, &response
	}
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := dto.GetUserResponse{
		Id:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		Created_at: user.CreatedAt,
	}

	if !user.DeletedAt.Valid {
		response.Status = dto.StatusActive
	} else {
		response.Status = dto.StatusInactive
	}

	return &response, nil
}

func (s *userService) ListUsers(ctx context.Context, pageRequest *dto.PageRequest) (*dto.PageResponse, error) {
	// Validate input
	if pageRequest == nil {
		return nil, errors.New(ErrInvalidData)
	}

	// Set default values
	if pageRequest.Page < 1 {
		pageRequest.Page = 1
	}
	if pageRequest.Limit < 1 {
		pageRequest.Limit = 10
	}
	if pageRequest.Limit > 100 {
		pageRequest.Limit = 100 // Max limit
	}

	// Call repository
	response, err := s.userRepo.List(ctx, pageRequest)
	if err != nil {
		return nil, err
	}

	return response, nil
}
