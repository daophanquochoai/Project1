package interceptor

import (
	"context"
	"github.com/agris/user-service/pkg/jwtMg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

type AuthInterceptor struct {
	jwtManager    *jwtMg.JWTManager
	publicMethods map[string]bool
}

func NewAuthInterceptor(jwtManager *jwtMg.JWTManager) *AuthInterceptor {
	return &AuthInterceptor{jwtManager: jwtManager, publicMethods: map[string]bool{
		"/user.UserService/GetCurrentUserInfo": false,
		"/user.UserService/Authenticate":       true,
	}}
}

func (i *AuthInterceptor) Handler() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// next public method
		if i.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		if token == authHeader[0] {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}

		claims, err := i.jwtManager.ValidateToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}

		ctx = context.WithValue(ctx, "userId", claims.UserID)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "email", claims.Email)

		// Tiếp tục xử lý request
		return handler(ctx, req)
	}
}
