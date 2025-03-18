package redis_db

import (
	"context"
	"fmt"
	"task/config"

	"github.com/redis/go-redis/v9"
)

var RedisDb *redis.Client

func NewRedisClient(cfg config.RedisConfig) error {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	RedisDb = client
	return nil
}
