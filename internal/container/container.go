package container

import (
	"context"
	"log"
	"os"

	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/saulo-duarte/chronos-lambda/internal/project"
	studysubject "github.com/saulo-duarte/chronos-lambda/internal/study_subject"
	"github.com/saulo-duarte/chronos-lambda/internal/task"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

type Container struct {
	UserContainer         *user.UserContainer
	ProjectContainer      *project.ProjectContainer
	TaskContainer         *task.TaskContainer
	StudySubjectContainer *studysubject.StudySubjectContainer
}

func New() *Container {
	config.Init()
	auth.Init()
	config.InitCrypto()
	auth.InitOauth()

	dsn := os.Getenv("DATABASE_DSN")
	if err := config.Connect(context.Background(), dsn); err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	userContainer := user.NewUserContainer(config.DB)

	projectContainer := project.NewProjectContainer(config.DB)

	taskContainer := task.NewTaskContainer(config.DB)

	studySubjectContainer := studysubject.NewStudySubjectContainer(config.DB)

	return &Container{
		UserContainer:         userContainer,
		ProjectContainer:      projectContainer,
		TaskContainer:         taskContainer,
		StudySubjectContainer: studySubjectContainer,
	}
}
