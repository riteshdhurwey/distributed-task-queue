package repository

import (
	"context"
	"errors"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func (r *RedisRepository) Enqueue(ctx context.Context, taskID int64, score float64) error {
	_, err := r.client.ZAdd(ctx, "taskqueue", redis.Z{
		Score:  float64(score),
		Member: strconv.FormatInt(taskID, 10),
	}).Result()
	return err
}

func (r *RedisRepository) Dequeue(ctx context.Context) (string, error) {
	result, err := r.client.ZPopMin(ctx, "taskqueue", 1).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}

	if len(result) == 0 {
		return "", nil
	}

	return result[0].Member.(string), nil
}
