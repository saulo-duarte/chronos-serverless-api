package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/saulo-duarte/chronos-lambda/internal/project"
	"github.com/saulo-duarte/chronos-lambda/internal/user"
)

type RouterConfig struct {
	UserHandler    *user.Handler
	ProjectHandler *project.Handler
}

func New(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Group(func(r chi.Router) {
		r.Get("/register", cfg.UserHandler.Register)
		r.Get("/google/callback", cfg.UserHandler.GoogleCallback)
		r.Post("/login", cfg.UserHandler.Login)
		r.Post("/refresh", cfg.UserHandler.RefreshToken)
	})

	project.RegisterRoutes(r, cfg.ProjectHandler)

	return r
}
