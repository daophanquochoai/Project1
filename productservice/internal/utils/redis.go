package utils

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func DeleteCacheByPattern(ctx context.Context, rd *redis.Client, pattern string) error {
	var cursor uint64
	var keys []string

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = rd.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		if err := rd.Del(ctx, keys...).Err(); err != nil {
			return err
		}
		log.Printf("Deleted %d keys matching pattern: %s", len(keys), pattern)
	}

	return nil
}
