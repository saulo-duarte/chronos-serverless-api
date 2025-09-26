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
var isProduction = os.Getenv("ENV") == "" || os.Getenv("ENV") == "production"

type Handler struct {
	service UserService
}

func NewHandler(s UserService) *Handler {
	return &Handler{service: s}
}

func newCookie(name, value string, maxAge int) *http.Cookie {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   maxAge,
	}
	if isProduction {
		c.SameSite = http.SameSiteNoneMode
		c.Secure = true
	}
	return c
}

func (h *Handler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())

	var payload struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if payload.Code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	user, jwtToken, refreshToken, err := h.service.LoginWithGoogleCode(r.Context(), payload.Code)
	if err != nil {
		log.WithError(err).Error("Falha no login via Google")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, newCookie(auth.JWT_COOKIE_NAME, jwtToken, int((24*time.Hour).Seconds())))
	http.SetCookie(w, newCookie(auth.REFRESH_TOKEN_COOKIE_NAME, refreshToken, int((14*24*time.Hour).Seconds())))

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

	http.SetCookie(w, newCookie(auth.JWT_COOKIE_NAME, newJWT, int((24*time.Hour).Seconds())))

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "token refreshed successfully",
	})
}
