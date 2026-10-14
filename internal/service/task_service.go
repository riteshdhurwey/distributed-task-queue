package service

import (
	"context"
	"errors"
	"taskqueue/internal/model"
	"taskqueue/internal/queue"
	"taskqueue/internal/repository"
	"time"
)

var (
	ErrTaskNotFound  = errors.New("task not found")
	ErrTaskNotFailed = errors.New("only failed tasks can be retried")
)

type TaskService struct {
	repo  repository.TaskRepository
	queue *queue.Queue
}

func New(repo repository.TaskRepository, queue *queue.Queue) *TaskService {
	return &TaskService{repo: repo, queue: queue}
}

func (s *TaskService) CreateTask(ctx context.Context, taskName string) (*model.Task, error) {

	task := model.Task{
		Name:      taskName,
		Status:    "pending",
		Attempts:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	taskID, err := s.repo.CreateTask(ctx, &task)

	task.ID = int(taskID)

	if err != nil {
		return &model.Task{}, err
	}
	s.queue.Enqueue(task)
	return &task, nil
}

func (s *TaskService) GetTask(ctx context.Context, taskID int) (*model.Task, error) {
	task, err := s.repo.GetTask(ctx, int64(taskID))
	return task, err
}

func (s *TaskService) RetryTask(ctx context.Context, taskID int) (*model.Task, error) {
	task, err := s.repo.GetTask(ctx, int64(taskID))
	if err != nil {
		return &model.Task{}, ErrTaskNotFound
	}

	if task.Status == "failed" {
		task.Status = "pending"
		task.Error = ""
		task.UpdatedAt = time.Now()
		s.repo.UpdateTask(ctx,task)
		s.queue.Enqueue(*task)

		return task, nil

	} else {
		return &model.Task{}, ErrTaskNotFailed
	}
}
