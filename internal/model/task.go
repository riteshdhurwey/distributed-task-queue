package model

import "time"

type Task struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Attempts  int       `json:"attempts"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
