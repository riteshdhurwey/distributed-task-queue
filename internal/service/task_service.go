package service

import (
	"taskqueue/internal/model"
	"taskqueue/internal/queue"
	"taskqueue/internal/repository"
	"time"
	"errors"
)

var (
    ErrTaskNotFound = errors.New("task not found")
    ErrTaskNotFailed = errors.New("only failed tasks can be retried")
)

type TaskService struct {
	repo repository.TaskRepository
	queue *queue.Queue
}


func New(repo repository.TaskRepository,queue *queue.Queue) *TaskService{
	return &TaskService{repo: repo,queue: queue}
}

func (s *TaskService) CreateTask(taskName string)(*model.Task,error){
	
	task := model.Task{
		Name: taskName,
		ID: s.repo.GetTaskID(),
		Status: "pending",
		Attempts: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		}

    err := s.repo.Create(&task)

	if err != nil {
		return &model.Task{},err
	}
	s.queue.Enqueue(task)
	return &task,nil
}


func ( s *TaskService) GetTask(taskID int) (*model.Task,error){
	task,err := s.repo.Get(taskID)
	return task,err
}

func (s *TaskService) RetryTask(taskID int)(*model.Task,error){
	task,err := s.repo.Get(taskID)
	if err != nil {
		return &model.Task{},ErrTaskNotFound
	}

	if task.Status == "failed" {
		task.Status = "pending"
		task.Error = ""
		task.UpdatedAt = time.Now()
		s.repo.Update(task)
		s.queue.Enqueue(*task)

		return task,nil

	}else{
		return &model.Task{},ErrTaskNotFailed
	}
}