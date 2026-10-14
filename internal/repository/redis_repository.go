package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func (r *RedisRepository) GenerateID(ctx context.Context) (int64, error) {
	return r.client.Incr(ctx, "task:next_id").Result()
}
