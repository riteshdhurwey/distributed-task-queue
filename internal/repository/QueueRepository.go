package repository

import "context"

type QueueRepository interface {
	Enqueue(ctx context.Context, taskID int64, score float64) error
	Dequeue(ctx context.Context) (string, error)
}
