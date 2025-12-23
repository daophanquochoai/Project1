package cache

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisRepository interface {
	SaveRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error
}

type redisRepository struct {
	rd *redis.Client
}

func NewRedisRepository(rd *redis.Client) RedisRepository {
	return &redisRepository{rd: rd}
}

func (r *redisRepository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	key := fmt.Sprintf("refresh_tokens:%s", userID)
	score := float64(expiresAt.Unix())

	// Use pipeline to execute both commands atomically
	pipe := r.rd.Pipeline()

	pipe.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: token,
	})

	pipe.Expire(ctx, key, 30*24*time.Hour)

	_, err := pipe.Exec(ctx)
	return err
}
