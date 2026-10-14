package repository

import (
	"errors"
	"sync"
	"taskqueue/internal/model"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type MemoryRepository struct {
	mutex  sync.RWMutex
	tasks  map[int]*model.Task
	TaskID int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tasks: make(map[int]*model.Task), TaskID: 1,
	}
}

func (repo *MemoryRepository) Create(task *model.Task) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.tasks[task.ID] = task

	return nil
}

func (repo *MemoryRepository) Get(id int) (*model.Task, error) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	task, ok := repo.tasks[id]

	if !ok {
		return &model.Task{}, ErrTaskNotFound
	}

	return task, nil
}

func (repo *MemoryRepository) Update(task *model.Task) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.tasks[task.ID] = task

	return nil
}

func (repo *MemoryRepository) GetTaskID() int {
	return repo.TaskID
}
