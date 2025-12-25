package cache

import "github.com/google/wire"

var Set = wire.NewSet(NewRedisClient)
