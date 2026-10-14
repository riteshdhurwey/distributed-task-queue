package worker

import (
	"context"
	"fmt"
	"strconv"
	"taskqueue/internal/model"
	"taskqueue/internal/repository"
	"time"
)

type Worker struct {
	ID    int
	repo  repository.TaskRepository
	queue repository.QueueRepository
}

func New(workerId int, queue repository.QueueRepository, repo repository.TaskRepository) *Worker {
	return &Worker{ID: workerId, queue: queue, repo: repo}
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

		taskID, err := worker.queue.Dequeue(ctx)

		id, err := strconv.Atoi(taskID)

		newTask, err := worker.repo.GetTask(ctx, int64(id))

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
