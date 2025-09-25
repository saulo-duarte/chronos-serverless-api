package user

import (
	"context"
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

// newCookie cria o cookie e loga informações para debug
func newCookie(ctx context.Context, name, value string, maxAge int) *http.Cookie {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   maxAge,
	}

	log := config.WithContext(ctx)
	log.Infof("Criando cookie: %s=%s, MaxAge=%d, isProduction=%v", name, value, maxAge, isProduction)

	if isProduction {
		c.SameSite = http.SameSiteNoneMode
		c.Secure = true
		log.Infof("Cookie %s configurado como Secure=true, SameSite=None", name)
	} else {
		c.SameSite = http.SameSiteLaxMode
		c.Secure = false
		log.Infof("Cookie %s configurado como Secure=false, SameSite=Lax", name)
	}

	return c
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

	log.Infof("FRONTEND_URL=%s, Request Origin=%s", FRONTEND_URL, r.Header.Get("Origin"))

	http.SetCookie(w, newCookie(r.Context(), auth.JWT_COOKIE_NAME, jwtToken, int((24*time.Hour).Seconds())))
	http.Redirect(w, r, FRONTEND_URL, http.StatusFound)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())
	log.Infof("Login request Origin=%s", r.Header.Get("Origin"))

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

	// Setando cookies com logs
	http.SetCookie(w, newCookie(r.Context(), auth.JWT_COOKIE_NAME, jwtToken, int((24*time.Hour).Seconds())))
	http.SetCookie(w, newCookie(r.Context(), auth.REFRESH_TOKEN_COOKIE_NAME, refreshToken, int((14*24*time.Hour).Seconds())))

	config.JSON(w, http.StatusOK, map[string]any{
		"user":    user.ToResponse(),
		"message": "Login successful",
	})
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	log := config.WithContext(r.Context())
	log.Infof("RefreshToken request Origin=%s", r.Header.Get("Origin"))

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

	http.SetCookie(w, newCookie(r.Context(), auth.JWT_COOKIE_NAME, newJWT, int((24*time.Hour).Seconds())))

	config.JSON(w, http.StatusOK, map[string]string{
		"message": "token refreshed successfully",
	})
}
