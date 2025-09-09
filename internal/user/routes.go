package user

import (
	"github.com/go-chi/chi/v5"
)

func Routes(h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Get("/register", h.Register)
	r.Get("/google/callback", h.GoogleCallback)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.RefreshToken)

	return r
}
