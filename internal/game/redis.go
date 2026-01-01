package game

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func RedisContext() context.Context {
	return context.Background()
}
