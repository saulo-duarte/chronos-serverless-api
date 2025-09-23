package task

import (
	"github.com/saulo-duarte/chronos-lambda/internal/project"
	studytopic "github.com/saulo-duarte/chronos-lambda/internal/study_topic"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
	"gorm.io/gorm"
)

type TaskContainer struct {
	Handler *Handler
}

func NewTaskContainer(
	db *gorm.DB,
	projectService project.ProjectService,
	studyTopicRepo studytopic.StudyTopicRepository,
	eventHandler EventHandler,
	userRepository user.UserRepository,
) *TaskContainer {
	repo := NewRepository(db)
	service := NewService(repo, projectService, userRepository, studyTopicRepo, eventHandler)
	handler := NewHandler(service)

	return &TaskContainer{
		Handler: handler,
	}
}
