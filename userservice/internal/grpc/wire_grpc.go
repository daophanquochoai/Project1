package grpc

//
//import (
//	"github.com/agris/user-service/internal/cache"
//	"github.com/agris/user-service/internal/config"
//	"github.com/agris/user-service/internal/database"
//	"github.com/agris/user-service/internal/grpc/service_grpc"
//	"github.com/agris/user-service/internal/repository"
//	"github.com/agris/user-service/internal/service"
//	"github.com/agris/user-service/pkg/jwtMg"
//	"github.com/google/wire"
//)
//
//func InitGRPCServer() (*service_grpc.AuthGRPCService, error) {
//	wire.Build(
//		config.Set,
//		cache.Set,
//		database.Set,
//		repository.Set,
//		service.Set,
//		jwtMg.Set,
//		service_grpc.Set,
//	)
//	return nil, nil
//}
