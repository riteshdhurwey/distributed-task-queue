package main

import (
	"taskqueue/internal/handler"
	"taskqueue/internal/queue"
	"taskqueue/internal/repository"
	"taskqueue/internal/service"
	"taskqueue/internal/worker"

	"github.com/gin-gonic/gin"
)


func main(){

	repo := repository.NewMemoryRepository()
	Queue := queue.New(10)
	Worker1 := worker.New(1,Queue,repo)
	Worker2 := worker.New(2,Queue,repo)

	go Worker1.Start()
	go Worker2.Start()

	router := gin.Default()
	service := service.New(repo,Queue)
	handler := handler.New(service)


	router.POST("/tasks",handler.CreateTask)
	router.GET("/tasks/:id",handler.GetTask)
	router.POST("/tasks/:id/retry",handler.RetryTask)

	router.Run(":8080")
}