package task

import (
	"github.com/go-chi/chi/v5"
)

func Routes(h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.CreateTask)
	r.Get("/", h.GetTask)
	r.Get("/user", h.ListTasksByUser)
	r.Get("/project", h.ListTasksByProject)
	r.Put("/", h.UpdateTask)
	r.Delete("/", h.DeleteTask)

	return r
}
