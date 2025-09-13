package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/saulo-duarte/chronos-lambda/internal/project"
	"github.com/saulo-duarte/chronos-lambda/internal/task"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

type RouterConfig struct {
	UserHandler    *user.Handler
	ProjectHandler *project.Handler
	TaskHandler    *task.Handler
}

func New(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/users", user.Routes(cfg.UserHandler))
	r.Mount("/projects", project.Routes(cfg.ProjectHandler))
	r.Mount("/tasks", task.Routes(cfg.TaskHandler))

	return r
}
