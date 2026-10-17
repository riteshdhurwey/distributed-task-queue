package service

import (
	"context"
	"time"
)

func (s *TaskService) GetPendingTasks(ctx context.Context) error {
	pendingTasks, err := s.repo.GetPendingTasks(ctx)

	if err != nil {
		return err
	}

	for _, task := range *pendingTasks {
		err := s.queue.Enqueue(ctx, int64(task.TaskID), float64(task.CreatedAt.UnixMilli()))

		if err != nil {
			return err
		}
	}

	return nil

}

func (s *TaskService) RecoverAndRequeueTasks(ctx context.Context) error {

	taskIDs, err := s.repo.RecoverExpiredTasks(ctx)

	if err != nil {
		return err
	}

	for _, id := range *taskIDs {
		err = s.queue.Enqueue(ctx, int64(id), float64(time.Now().UnixMilli()))

		if err != nil {
			return err
		}

	}

	return nil
}
