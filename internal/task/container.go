package task

import "gorm.io/gorm"

type TaskContainer struct {
	Handler *Handler
}

func NewTaskContainer(db *gorm.DB) *TaskContainer {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	return &TaskContainer{
		Handler: handler,
	}
}
