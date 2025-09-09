package project

import (
	"github.com/go-chi/chi/v5"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)

		r.Post("/projects", h.CreateProject)
		r.Get("/projects", h.ListProjects)
		r.Get("/projects/{id}", h.GetProject)
		r.Put("/projects/{id}", h.UpdateProject)
		r.Delete("/projects/{id}", h.DeleteProject)
	})
}
