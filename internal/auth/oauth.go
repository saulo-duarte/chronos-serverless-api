package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOauthConfig *oauth2.Config

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

type AuthResult struct {
	ProviderID   string
	Username     string
	Email        string
	Picture      string
	AccessToken  string
	RefreshToken string
}

func InitOauth() {
	GoogleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes:       []string{"https://www.googleapis.com/auth/calendar.readonly", "https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

func GetGoogleAuthURL(state string) string {
	return GoogleOauthConfig.AuthCodeURL(state)
}

func HandleGoogleCallback(ctx context.Context, code string) (*AuthResult, error) {
	log := config.WithContext(ctx)

	token, err := GoogleOauthConfig.Exchange(ctx, code)
	if err != nil {
		log.WithError(err).Error("Falha ao trocar o código de autorização por um token")
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	userInfo, err := getGoogleUserInfo(ctx, token)
	if err != nil {
		log.WithError(err).Error("Falha ao obter informações do usuário do Google")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	log.WithFields(logrus.Fields{
		"email":   userInfo.Email,
		"user_id": userInfo.ID,
	}).Info("Informações do usuário obtidas com sucesso")

	encryptedAccessToken := ""
	encryptedRefreshToken := ""

	if token.AccessToken != "" {
		if enc, err := config.Encrypt(token.AccessToken); err == nil {
			encryptedAccessToken = enc
		}
	}

	if token.RefreshToken != "" {
		if enc, err := config.Encrypt(token.RefreshToken); err == nil {
			encryptedRefreshToken = enc
		}
	}

	return &AuthResult{
		ProviderID:   userInfo.ID,
		Username:     userInfo.Name,
		Email:        userInfo.Email,
		Picture:      userInfo.Picture,
		AccessToken:  encryptedAccessToken,
		RefreshToken: encryptedRefreshToken,
	}, nil
}

func getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	log := config.WithContext(ctx)
	client := GoogleOauthConfig.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.WithError(err).Error("Erro ao fazer a requisição de informações do usuário")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("status_code", resp.StatusCode).Error("Falha na requisição de informações do usuário")
		return nil, fmt.Errorf("failed to get user info, status: %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		log.WithError(err).Error("Falha ao decodificar a resposta da API de informações do usuário")
		return nil, err
	}

	return &userInfo, nil
}

func GenerateState() string {
	return "secure_random_state_" + time.Now().Format("20060102150405")
}
