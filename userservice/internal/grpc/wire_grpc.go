package grpc

//
//import (
//	"github.com/agris/user-service/internal/cache"
//	"github.com/agris/user-service/internal/config"
//	"github.com/agris/user-service/internal/database"
//	"github.com/agris/user-service/internal/grpc/interceptor"
//	"github.com/agris/user-service/internal/grpc/service_grpc"
//	"github.com/agris/user-service/internal/repository"
//	"github.com/agris/user-service/internal/service"
//	"github.com/agris/user-service/pkg/jwtMg"
//	"github.com/google/wire"
//	"google.golang.org/grpc"
//)
//
//// InitGRPCServer khởi tạo TẤT CẢ dependencies và tạo gRPC server
//func InitGRPCServer() (*grpc.Server, func(), error) {
//	wire.Build(
//		config.Set,
//		cache.Set,
//		database.Set,
//		jwtMg.Set,
//		repository.Set,
//		service.Set,
//		interceptor.Set,
//		service_grpc.Set,
//		NewGRPCServer, // ← Provider tạo server
//	)
//	return nil, nil, nil
//}
