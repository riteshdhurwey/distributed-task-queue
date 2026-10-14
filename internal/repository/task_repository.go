package repository

import (
	"context"
	"taskqueue/internal/model"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *model.Task) (int64, error)
	GetTask(ctx context.Context, id int64) (*model.Task, error)
	UpdateTask(ctx context.Context, task *model.Task) error
	GetPendingTasks(ctx context.Context) (*[]model.Pair, error)
}
