package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/sirupsen/logrus"
)

const (
	FRONTEND_URL = "http://localhost:3001"
)

type Handler struct {
	service UserService
}

func NewHandler(s UserService) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Handle(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{
		"Access-Control-Allow-Origin":      FRONTEND_URL,
		"Access-Control-Allow-Credentials": "true",
		"Content-Type":                     "application/json",
	}

	if req.HTTPMethod == "OPTIONS" {
		headers["Access-Control-Allow-Methods"] = "GET,POST,OPTIONS"
		headers["Access-Control-Allow-Headers"] = "Content-Type,Authorization"
		return events.APIGatewayProxyResponse{StatusCode: http.StatusOK, Headers: headers}, nil
	}

	log := config.WithContext(ctx).WithFields(logrus.Fields{"path": req.Path, "method": req.HTTPMethod})

	switch req.Path {
	case "/register":
		if req.HTTPMethod == "GET" {
			log.Info("Processando rota de registro")
			return h.handleRegister(ctx, req, headers)
		}
	case "/google/callback":
		if req.HTTPMethod == "GET" {
			log.Info("Processando callback do Google")
			return h.handleGoogleCallback(ctx, req, headers)
		}
	case "/login":
		if req.HTTPMethod == "POST" {
			log.Info("Processando login")
			return h.handleLogin(ctx, req, headers)
		}
	case "/refresh":
		if req.HTTPMethod == "POST" {
			log.Info("Processando refresh token")
			return h.handleRefreshToken(ctx, req, headers)
		}
	}
	log.Warn("Rota não encontrada")
	return config.NotFound(ctx, headers)
}

func (h *Handler) handleRegister(ctx context.Context, req events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	state := auth.GenerateState()
	url := auth.GetGoogleAuthURL(state)

	headers["Location"] = url
	return events.APIGatewayProxyResponse{StatusCode: http.StatusFound, Headers: headers}, nil
}

func (h *Handler) handleGoogleCallback(ctx context.Context, req events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	log := config.WithContext(ctx)

	code, ok := req.QueryStringParameters["code"]
	if !ok || code == "" {
		log.Error("Código de autorização não encontrado")
		return config.BadRequest(ctx, "code not found", headers)
	}

	_, jwtToken, err := h.service.HandleGoogleCallback(ctx, code)
	if err != nil {
		log.WithError(err).Error("Falha ao lidar com o callback do Google")
		return config.InternalError(ctx, headers)
	}

	headers["Set-Cookie"] = auth.NewJWTCookie(jwtToken, 24*time.Hour)
	headers["Location"] = FRONTEND_URL
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusFound,
		Headers:    headers,
	}, nil
}

func (h *Handler) handleLogin(ctx context.Context, req events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	log := config.WithContext(ctx)

	var payload struct {
		ProviderID string `json:"provider_id"`
	}

	if err := json.Unmarshal([]byte(req.Body), &payload); err != nil {
		log.WithError(err).Error("Corpo da requisição inválido")
		return config.BadRequest(ctx, "invalid request body", headers)
	}

	if payload.ProviderID == "" {
		log.Warn("provider_id não fornecido")
		return config.BadRequest(ctx, "provider_id is required", headers)
	}

	user, jwtToken, refreshToken, err := h.service.Login(ctx, payload.ProviderID)
	if err != nil {
		if err == ErrUserNotFound {
			log.Warn("Usuário não encontrado durante o login")
			return config.Unauthorized(ctx, "user not found", headers)
		}
		log.WithError(err).Error("Erro interno durante o login")
		return config.InternalError(ctx, headers)
	}

	headers["Set-Cookie"] = auth.NewJWTCookie(jwtToken, 24*time.Hour)
	headers["Set-Cookie"] += fmt.Sprintf(", %s", auth.NewRefreshTokenCookie(refreshToken, 14*24*time.Hour))

	return config.APIGateway(ctx, http.StatusOK, map[string]interface{}{"user": user.ToResponse(), "message": "Login successful"}, headers)
}

func (h *Handler) handleRefreshToken(ctx context.Context, req events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	log := config.WithContext(ctx)
	refreshToken, err := getCookie(req.Headers, auth.REFRESH_TOKEN_COOKIE_NAME)
	if err != nil {
		log.WithError(err).Warn("Token de atualização não encontrado no cookie")
		return config.Unauthorized(ctx, "refresh token required", headers)
	}

	newJWT, err := h.service.RefreshToken(ctx, refreshToken)
	if err != nil {
		log.WithError(err).Error("Falha ao atualizar o token")
		return config.Unauthorized(ctx, "failed to refresh token", headers)
	}
	headers["Set-Cookie"] = auth.NewJWTCookie(newJWT, 24*time.Hour)

	return config.APIGateway(ctx, http.StatusOK, map[string]string{"message": "token refreshed successfully"}, headers)
}

func getCookie(headers map[string]string, name string) (string, error) {
	cookieHeader, ok := headers["Cookie"]
	if !ok {
		cookieHeader, ok = headers["cookie"]
		if !ok {
			return "", fmt.Errorf("cookie header not found")
		}
	}
	req := http.Request{Header: http.Header{"Cookie": {cookieHeader}}}
	cookies := req.Cookies()
	for _, c := range cookies {
		if c.Name == name {
			return c.Value, nil
		}
	}
	return "", fmt.Errorf("cookie not found: %s", name)
}
