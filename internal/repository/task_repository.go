package repository

import "taskqueue/internal/model"

type TaskRepository interface {
	Create(task *model.Task) error
	Get(id int) (*model.Task,error)
	Update(task *model.Task) error
	GetTaskID() int
}