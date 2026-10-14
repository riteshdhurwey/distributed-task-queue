package queue

import "taskqueue/internal/model"

type Queue struct {
	TaskQueue chan model.Task
}

func New() *Queue {
	return &Queue{TaskQueue: make(chan model.Task)}
}

func (que *Queue) Enqueue(task model.Task) {
	que.TaskQueue <- task
}

func (que *Queue) Dequeue() model.Task {
	return <-que.TaskQueue
}
