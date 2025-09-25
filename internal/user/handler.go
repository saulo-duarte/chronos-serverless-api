package user

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
)

var FRONTEND_URL = os.Getenv("FRONTEND_URL")

type Handler struct {
	service UserService
}

func NewHandler(s UserService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	state := auth.GenerateState()
	url := auth.GetGoogleAuthURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Error("Código de autorização não encontrado")
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}

	_, jwtToken, err := h.service.HandleGoogleCallback(r.Context(), code)
	if err != nil {
		log.WithError(err).Error("Falha ao lidar com o callback do Google")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.JWT_COOKIE_NAME,
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})

	http.Redirect(w, r, FRONTEND_URL, http.StatusFound)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload struct {
		ProviderID string `json:"provider_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if payload.ProviderID == "" {
		http.Error(w, "provider_id is required", http.StatusBadRequest)
		return
	}

	user, jwtToken, refreshToken, err := h.service.Login(r.Context(), payload.ProviderID)
	if err != nil {
		if err == ErrUserNotFound {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.JWT_COOKIE_NAME,
		Value:    jwtToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     auth.REFRESH_TOKEN_COOKIE_NAME,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((14 * 24 * time.Hour).Seconds()),
	})

	config.JSON(w, http.StatusOK, map[string]any{
		"user":    user.ToResponse(),
		"message": "Login successful",
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	cookie, err := r.Cookie(auth.REFRESH_TOKEN_COOKIE_NAME)
	if err != nil {
		config.JSON(w, http.StatusUnauthorized, map[string]string{
			"error": "refresh token required",
		})
		return
	}

	newJWT, err := h.service.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		log.WithError(err).Error("Falha ao atualizar o token")
		config.JSON(w, http.StatusUnauthorized, map[string]string{
			"error": "failed to refresh token",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.JWT_COOKIE_NAME,
		Value:    newJWT,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   int((24 * time.Hour).Seconds()),
	})

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "token refreshed successfully",
	})
}
