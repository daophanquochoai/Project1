package provider

import "github.com/google/wire"

var Set = wire.NewSet(ProvideGRPCConnection)
