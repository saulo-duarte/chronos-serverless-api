package task

import (
	"context"
	"errors"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrUnauthorized = errors.New("unauthorized")
)

type TaskService interface {
	CreateTask(ctx context.Context, t *Task) (*Task, error)
	GetTaskByID(ctx context.Context, id string) (*Task, error)
	ListTaskByUser(ctx context.Context, projectId string) (*Task, error)
	ListTaskByProject(ctx context.Context, t *Task) (*[]Task, error)
	UpdateTask(ctx context.Context, t *Task) (*Task, error)
	DeleteTask(ctx context.Context, id string) error
}
