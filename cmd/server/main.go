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
	redisRepo := repository.NewRedisRepository(rdb)

	pendingTasks, err := repo.GetPendingTasks(ctx)

	if err != nil {
		log.Fatal(err)
	}

	for _, task := range *pendingTasks {
		err := redisRepo.Enqueue(ctx, int64(task.TaskID), float64(task.CreatedAt.UnixMilli()))
		if err != nil {
			log.Fatal(err)
		}
	}

	Worker1 := worker.New(1, redisRepo, repo)
	Worker2 := worker.New(2, redisRepo, repo)

	go Worker1.Start(ctx)
	go Worker2.Start(ctx)

	router := gin.Default()
	service := service.New(repo, redisRepo)
	handler := handler.New(service)

	router.POST("/tasks", handler.CreateTask)
	router.GET("/tasks/:id", handler.GetTask)
	router.POST("/tasks/:id/retry", handler.RetryTask)

	router.Run(":8080")
}
