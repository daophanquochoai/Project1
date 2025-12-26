package client

import (
	"context"
	"productservice/internal/grpc/pb/userservicepb"

	"github.com/gofiber/fiber/v2/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthClient struct {
	client userservicepb.UserServiceClient
}

func NewAuthClient(conn *grpc.ClientConn) *AuthClient {
	return &AuthClient{
		client: userservicepb.NewUserServiceClient(conn),
	}
}

func (a *AuthClient) Authenticate(ctx context.Context, token string) (*userservicepb.AuthResponse, error) {
	request := &userservicepb.AuthRequest{
		Token: token,
	}

	resp, err := a.client.Authenticate(ctx, request)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return resp, nil

}

func (a *AuthClient) GetCurrentUserInfo(ctx context.Context, token string) (*userservicepb.UserResponse, error) {
	md := metadata.Pairs(
		"authorization", "Bearer "+token,
	)

	// Thêm metadata vào context
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Gọi gRPC service với context đã có metadata
	response, err := a.client.GetCurrentUserInfo(ctx, &emptypb.Empty{})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return response, nil
}
