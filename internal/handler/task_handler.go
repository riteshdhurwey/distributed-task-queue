package handler

import (
	"errors"
	"net/http"
	"strconv"
	"taskqueue/internal/service"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	service *service.TaskService
}

func New(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

func (h *TaskHandler) CreateTask(ctx *gin.Context) {
	type CreateTaskRequest struct {
		Name string `json:"name"`
	}

	var taskRequest CreateTaskRequest

	bindingErr := ctx.ShouldBindJSON(&taskRequest)

	if bindingErr != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": bindingErr})
		return
	}

	task, err := h.service.CreateTask(ctx, taskRequest.Name)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusCreated, task)

}

func (h *TaskHandler) GetTask(ctx *gin.Context) {
	idstr := ctx.Param("id")

	id, err := strconv.Atoi(idstr)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, InternalErr := h.service.GetTask(ctx, id)

	if InternalErr != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": InternalErr.Error()})
		return
	}

	ctx.JSON(http.StatusAccepted, task)

}

func (h *TaskHandler) RetryTask(ctx *gin.Context) {
	idstr := ctx.Param("id")

	id, err := strconv.Atoi(idstr)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, serviceErr := h.service.RetryTask(ctx, id)

	if serviceErr != nil {
		switch {
		case errors.Is(serviceErr, service.ErrTaskNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		case errors.Is(serviceErr, service.ErrTaskNotFailed):
			ctx.JSON(http.StatusConflict, gin.H{"error": "Only failed tasks can be retried"})
			return
		case errors.Is(serviceErr, service.ErrMaxAttemptLimitExceeded):
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "retry limit exceeded. only 5 requests allowed"})
			return
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry task"})
			return
		}

	}

	ctx.JSON(http.StatusCreated, gin.H{"data": task, "response": "Task queued for retry"})
}
