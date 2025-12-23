package service_grpc

import (
	"context"
	"github.com/agris/user-service/internal/grpc/pb/userservicepb"
	"github.com/agris/user-service/internal/service"
	"github.com/agris/user-service/pkg/jwtMg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
