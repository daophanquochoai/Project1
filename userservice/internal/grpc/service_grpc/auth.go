package service_grpc

import (
	"context"
	"github.com/agris/user-service/internal/grpc/pb/userservicepb"
	"github.com/agris/user-service/internal/service"
	"github.com/agris/user-service/pkg/jwtMg"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type AuthGRPCService struct {
	userservicepb.UnimplementedUserServiceServer
	auth        service.AuthService
	jwtManager  *jwtMg.JWTManager
	userService service.UserService
}

func NewAuthGRPCService(auth service.AuthService, jwtManager *jwtMg.JWTManager, userService service.UserService) *AuthGRPCService {
	return &AuthGRPCService{auth: auth, jwtManager: jwtManager, userService: userService}
}

func (u *AuthGRPCService) GetCurrentUserInfo(ctx context.Context, empty *emptypb.Empty) (*userservicepb.UserResponse, error) {
	id := ctx.Value("userId").(uuid.UUID)
	if id == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, ErrInvalidData)
	}

	ct, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	users, err := u.userService.GetCurrentUser(ct, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, ErrNotFound)
	}

	response := &userservicepb.UserResponse{
		Id:    users.Id.String(),
		Name:  users.Name,
		Email: users.Email,
	}
	return response, nil
}

func (g *AuthGRPCService) Authenticate(ctx context.Context, authRequest *userservicepb.AuthRequest) (*userservicepb.AuthResponse, error) {
	if authRequest.Token == "" {
		return nil, status.Error(codes.Unauthenticated, ErrUnAuthenticated)
	}

	claims, err := g.jwtManager.ValidateToken(authRequest.Token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, ErrUnAuthenticated)
	}

	_, errUser := g.userService.GetCurrentUser(ctx, claims.UserID)
	if errUser != nil {
		return nil, status.Error(codes.Unauthenticated, ErrUnAuthenticated)
	}

	return &userservicepb.AuthResponse{
		Valid:  true,
		UserId: claims.UserID.String(),
		Role:   string(claims.Role),
	}, nil
}
