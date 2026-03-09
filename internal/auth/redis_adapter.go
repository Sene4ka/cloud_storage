package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisAdapter struct {
	client *redis.Client
}

func NewRedisAdapter(client *redis.Client) *RedisAdapter {
	return &RedisAdapter{client: client}
}

func (a *RedisAdapter) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return a.client.Set(ctx, key, value, expiration).Err()
}

func (a *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	return a.client.Get(ctx, key).Result()
}

func (a *RedisAdapter) Del(ctx context.Context, key string) error {
	return a.client.Del(ctx, key).Err()
}

func (a *RedisAdapter) Exists(ctx context.Context, key string) (int64, error) {
	return a.client.Exists(ctx, key).Result()
}
