package user

import (
	"github.com/go-chi/chi/v5"
)

func Routes(h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Post("/login", h.GoogleLogin)
	r.Post("/refresh", h.RefreshToken)

	return r
}
