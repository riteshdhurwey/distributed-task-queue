package main

import (
	"context"
	"fmt"
	"taskqueue/internal/database"
	"taskqueue/internal/handler"
	"taskqueue/internal/queue"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewPostgesRepo(conn)
	Queue := queue.New()
	Worker1 := worker.New(1, Queue, repo)
	Worker2 := worker.New(2, Queue, repo)

	go Worker1.Start(ctx)
	go Worker2.Start(ctx)

	router := gin.Default()
	service := service.New(repo, Queue)
	handler := handler.New(service)

	router.POST("/tasks", handler.CreateTask)
	router.GET("/tasks/:id", handler.GetTask)
	router.POST("/tasks/:id/retry", handler.RetryTask)

	router.Run(":8080")
}
