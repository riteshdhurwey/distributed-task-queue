package main

import (
	"context"
	"fmt"
	"log"
	"taskqueue/config"
	"taskqueue/internal/database"
	"taskqueue/internal/handler"
	"taskqueue/internal/repository"
	"taskqueue/internal/service"
	"taskqueue/internal/worker"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	conn, err := database.ConnectDB()

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	fmt.Println("Connected To DataBase")

	rdb, err := config.NewRedisClient()

	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewPostgesRepo(conn)
	redisRepo := repository.NewRedisRepository(rdb,"taskqueue")

	service := service.New(repo, redisRepo)

	err = service.GetPendingTasks(ctx)

	if err != nil {
		log.Fatal(err)
	}

	Worker1 := worker.New(1, redisRepo, repo,worker.ProcessTask)
	Worker2 := worker.New(2, redisRepo, repo,worker.ProcessTask)

	go Worker1.Start(ctx)
	go Worker2.Start(ctx)

	// Re-enqueue processing tasks whose leases have expired.
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := service.RecoverAndRequeueTasks(ctx); err != nil {
					log.Printf("task recovery failed: %v", err)
				}
			}
		}
	}()

	router := gin.Default()
	handler := handler.New(service)

	router.POST("/tasks", handler.CreateTask)
	router.GET("/tasks/:id", handler.GetTask)
	router.POST("/tasks/:id/retry", handler.RetryTask)

	router.Run(":8080")
}
