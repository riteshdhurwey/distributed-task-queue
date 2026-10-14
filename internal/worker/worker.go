package worker

import (
	"context"
	"fmt"
	"taskqueue/internal/model"
	"taskqueue/internal/queue"
	"taskqueue/internal/repository"
	"time"
)

type Worker struct {
	ID    int
	Queue *queue.Queue
	repo  repository.TaskRepository
}

func New(workerId int, que *queue.Queue, repo repository.TaskRepository) *Worker {
	return &Worker{ID: workerId, Queue: que, repo: repo}
}

func ProcessTask(task model.Task) error {
	fmt.Println("Processing", "Task ID", task.ID, " ", task.Name)
	time.Sleep(time.Second * 5)
	if task.Name == "fail" {
		fmt.Println("failed ", "Task ID ", task.ID, " ", task.Name)
		return fmt.Errorf("failed")
	}
	fmt.Println("completed ", "Task ID ", task.ID, " ", task.Name)

	return nil
}

func (worker *Worker) Start(ctx context.Context) {

	for {

		task := worker.Queue.Dequeue()

		newTask, err := worker.repo.GetTask(ctx, int64(task.ID))

		if err != nil {
			continue
		}
		newTask.Status = "processing"
		newTask.UpdatedAt = time.Now()
		newTask.Attempts++
		worker.repo.UpdateTask(ctx, newTask)

		processErr := ProcessTask(*newTask)

		if processErr != nil {
			newTask.Status = "failed"
			newTask.UpdatedAt = time.Now()

		} else {
			newTask.Status = "completed"
			newTask.UpdatedAt = time.Now()

		}
		worker.repo.UpdateTask(ctx, newTask)

	}
}
