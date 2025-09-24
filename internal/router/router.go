package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/middlewares"
	"github.com/saulo-duarte/chronos-lambda/internal/project"
	studysubject "github.com/saulo-duarte/chronos-lambda/internal/study_subject"
	studytopic "github.com/saulo-duarte/chronos-lambda/internal/study_topic"
	"github.com/saulo-duarte/chronos-lambda/internal/task"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

type RouterConfig struct {
	UserHandler         *user.Handler
	ProjectHandler      *project.Handler
	TaskHandler         *task.Handler
	StudySubjectHandler *studysubject.Handler
	StudyTopicHandler   *studytopic.Handler
}

func New(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.CorsMiddleware)

	r.Mount("/users", user.Routes(cfg.UserHandler))

	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)

		r.Mount("/projects", project.Routes(cfg.ProjectHandler))
		r.Mount("/tasks", task.Routes(cfg.TaskHandler))
		r.Mount("/study-subjects", studysubject.Routes(cfg.StudySubjectHandler))
		r.Mount("/study-topics", studytopic.Routes(cfg.StudyTopicHandler))

		r.Get("/study-subjects/{studySubjectId}/topics", cfg.StudyTopicHandler.ListStudyTopics)
		r.Get("/study-topics/{studyTopicId}/tasks", cfg.TaskHandler.ListTasksByStudyTopic)
	})
	return r
}
