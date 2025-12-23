package cache

import (
	"context"
	"github.com/agris/user-service/internal/config"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewRedisClient(config *config.Config) (*redis.Client, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
