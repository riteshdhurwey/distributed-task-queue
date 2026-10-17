package worker

import (
	"context"
	"fmt"
	"sync"
	"taskqueue/config"
	"taskqueue/internal/database"
	"taskqueue/internal/model"
	"taskqueue/internal/repository"
	"testing"
	"time"
)

func TestWorkersProcessTasksOnce(t *testing.T) {
    counts := make(map[int64]int)
    var mu sync.Mutex

    testProcess := func(task model.Task) error {
        mu.Lock()
        counts[int64(task.ID)]++
        mu.Unlock()
        return nil
    }

    conn,err := database.ConnectDB()

	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()

	rdb,err := config.NewRedisClient()

	if err != nil {
		t.Fatal(err)
	}

	defer rdb.Close()

	repo := repository.NewPostgesRepo(conn)
	redisRepo := repository.NewRedisRepository(rdb,"taskqueue:test")

	ctx, cancel := context.WithCancel(context.Background())
defer cancel()

const totalTasks = 100

taskIDs := make([]int64, 0, totalTasks)

defer func() {
    cleanupCtx := context.Background()

    for _, id := range taskIDs {
        _, err := rdb.ZRem(cleanupCtx,"taskqueue:test", fmt.Sprint(id)).Result()
        if err != nil {
            t.Errorf("failed to remove task %d from Redis: %v", id, err)
        }

        _, err = conn.Exec(cleanupCtx, "DELETE FROM tasks WHERE id = $1", id)
        if err != nil {
            t.Errorf("failed to delete task %d from PostgreSQL: %v", id, err)
        }
    }
}()

for i := 0; i < totalTasks; i++ {
    task := model.Task{
        Name:   fmt.Sprintf("test-task-%d", i),
        Status: "pending",
    }

    id, err := repo.CreateTask(ctx, &task)
    if err != nil {
        t.Fatal(err)
    }

    err = redisRepo.Enqueue(ctx, id, float64(time.Now().UnixMilli()))
    if err != nil {
        t.Fatal(err)
    }

    taskIDs = append(taskIDs, id)
}

	workers := make([]*Worker, 5)

	for i := range workers {
		workers[i] = New(i+1, redisRepo, repo, testProcess)
	}
	var wg sync.WaitGroup

	for _, w := range workers {
		wg.Add(1)

		go func(w *Worker) {
			defer wg.Done()
			w.Start(ctx)
		}(w)
	}

	deadline := time.Now().Add(30 * time.Second)
allCompleted := false

for time.Now().Before(deadline) {
    allCompleted = true

    for _, id := range taskIDs {
        task, err := repo.GetTask(ctx, id)
        if err != nil {
            t.Fatal(err)
        }

        if task.Status != "completed" {
            allCompleted = false
            break
        }
    }

    if allCompleted {
        break
    }

    time.Sleep(100 * time.Millisecond)
}

cancel()
wg.Wait()

if !allCompleted {
    t.Fatal("timed out waiting for tasks to complete")
}

mu.Lock()
defer mu.Unlock()

for _, id := range taskIDs {
    if counts[id] != 1 {
        t.Errorf("task %d executed %d times; expected 1", id, counts[id])
    }
}



}