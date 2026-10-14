package config

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := client.Ping(ctx).Err()

	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
