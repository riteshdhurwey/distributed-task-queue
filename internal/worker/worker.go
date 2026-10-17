package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"taskqueue/internal/model"
	"taskqueue/internal/repository"
	"time"
)

type Worker struct {
	ID    int
	repo  repository.TaskRepository
	queue repository.QueueRepository
	Process func(model.Task) error
}

func New(workerId int, queue repository.QueueRepository, repo repository.TaskRepository,Process func(model.Task)error) *Worker {
	return &Worker{ID: workerId, queue: queue, repo: repo,Process: Process}
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

		if ctx.Err() != nil {
			return
		}

		taskID, err := worker.queue.Dequeue(ctx)

		if err != nil {
			if ctx.Err() != nil || errors.Is(err, context.Canceled) {
				return
			}

			log.Printf("dequeue failed %v", err)

			if !wait(ctx, 2*time.Second) {
				return
			}

			continue
		}

		if taskID == "" {
			if !wait(ctx, 2*time.Second) {
				return
			}
			continue
		}

		id, err := strconv.Atoi(taskID)

		if err != nil {
			log.Printf("invalid task ID %q: %v", taskID, err)
			continue
		}

		claimedID, err := worker.repo.ClaimTask(ctx, int64(id))

		if err != nil {
			log.Printf("failed to claim task %d , %v",claimedID,err)
			continue
		}

		claimedTask,err := worker.repo.GetTask(ctx,claimedID)

		if err != nil {
			log.Printf("failed to fetch task %s: %v", taskID, err)
			continue
		}

		if claimedTask == nil {
			log.Printf("task %s not found", taskID)
			continue
		}

		processErr := worker.Process(*claimedTask)

		if processErr != nil {
			claimedTask.Status = "failed"
			log.Printf("task %s failed: %v", taskID, processErr)

		} else {
			claimedTask.Status = "completed"

		}

		claimedTask.UpdatedAt = time.Now()
		claimedTask.LeaseExpiresAt = nil
		err = worker.repo.UpdateTask(ctx, claimedTask)

		if err != nil {
			log.Printf("failed to save task %s result: %v", taskID, err)
		}

	}
}

func wait(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
