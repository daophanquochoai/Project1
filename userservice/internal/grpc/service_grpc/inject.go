package service_grpc

import "github.com/google/wire"

var Set = wire.NewSet(NewAuthGRPCService)
