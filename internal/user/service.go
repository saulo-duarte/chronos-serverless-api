package user

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/saulo-duarte/chronos-lambda/internal/auth"
	"github.com/saulo-duarte/chronos-lambda/internal/config"
	"github.com/sirupsen/logrus"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserService interface {
	HandleGoogleCallback(ctx context.Context, code string) (*User, string, error)
	Login(ctx context.Context, providerID string) (*User, string, string, error)
	RefreshToken(ctx context.Context, tokenString string) (string, error)
}

type userService struct {
	repo UserRepository
}

func NewService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) HandleGoogleCallback(ctx context.Context, code string) (*User, string, error) {
	log := config.WithContext(ctx)

	authResult, err := auth.HandleGoogleCallback(ctx, code)
	if err != nil {
		log.WithError(err).Error("Falha no callback de autenticação do Google")
		return nil, "", err
	}

	providerID := strings.TrimPrefix(authResult.ProviderID, "google-")
	log.WithField("provider_id", providerID).Info("Callback do Google processado com sucesso")

	user, err := s.repo.GetByProviderID(providerID)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		log.WithError(err).Error("Erro ao buscar usuário por provider ID")
		return nil, "", err
	}

	if user == nil {
		log.Info("Usuário não encontrado, criando novo usuário")
		user = &User{
			ID:                          uuid.New(),
			ProviderID:                  providerID,
			Username:                    authResult.Username,
			Email:                       authResult.Email,
			AvatarURL:                   authResult.Picture,
			Role:                        "USER",
			EncryptedGoogleAccessToken:  authResult.AccessToken,
			EncryptedGoogleRefreshToken: authResult.RefreshToken,
			CreatedAt:                   time.Now(),
			UpdatedAt:                   time.Now(),
		}
		if err := s.repo.Create(user); err != nil {
			log.WithError(err).Error("Falha ao criar novo usuário")
			return nil, "", err
		}
		log.WithField("user_id", user.ID).Info("Novo usuário criado com sucesso")
	} else {
		log.WithField("user_id", user.ID).Info("Usuário encontrado, atualizando informações")
		user.Username = authResult.Username
		user.Email = authResult.Email
		user.AvatarURL = authResult.Picture
		user.EncryptedGoogleAccessToken = authResult.AccessToken
		user.EncryptedGoogleRefreshToken = authResult.RefreshToken
		user.UpdatedAt = time.Now()
		if err := s.repo.Update(user); err != nil {
			log.WithError(err).Error("Falha ao atualizar usuário existente")
			return nil, "", err
		}
		log.WithField("user_id", user.ID).Info("Usuário atualizado com sucesso")
	}

	jwtToken, err := auth.GenerateJWTWithGoogleToken(user.ID.String(), user.Role, authResult.AccessToken, 24*time.Hour)
	if err != nil {
		log.WithError(err).Error("Falha ao gerar JWT para o usuário")
		return nil, "", err
	}
	log.WithField("user_id", user.ID).Info("JWT gerado com sucesso")

	return user, jwtToken, nil
}

func (s *userService) Login(ctx context.Context, providerID string) (*User, string, string, error) {
	log := config.WithContext(ctx)

	user, err := s.repo.GetByProviderID(providerID)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar usuário por ID")
		return nil, "", "", ErrUserNotFound
	}
	if user == nil {
		log.WithField("provider_id", providerID).Warn("Usuário não encontrado para login")
		return nil, "", "", ErrUserNotFound
	}

	log.WithFields(logrus.Fields{"user_id": user.ID, "provider_id": providerID}).Info("Login de usuário bem-sucedido")

	jwtToken, err := auth.GenerateJWT(user.ID.String(), user.Role, 24*time.Hour)
	if err != nil {
		log.WithError(err).Error("Falha ao gerar JWT para o login")
		return nil, "", "", err
	}

	refreshToken, err := auth.GenerateJWT(user.ID.String(), user.Role, 14*24*time.Hour)
	if err != nil {
		log.WithError(err).Error("Falha ao gerar refresh token para o login")
		return nil, "", "", err
	}

	return user, jwtToken, refreshToken, nil
}

func (s *userService) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	log := config.WithContext(ctx)

	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		log.WithError(err).Warn("Token de atualização inválido")
		return "", errors.New("invalid refresh token")
	}

	user, err := s.repo.GetByID(claims.UserID)
	if err != nil {
		log.WithError(err).Error("Erro ao buscar usuário por ID para o refresh token")
		return "", err
	}
	if user == nil {
		log.WithField("user_id", claims.UserID).Warn("Usuário não encontrado para o refresh token")
		return "", ErrUserNotFound
	}

	newJWT, err := auth.GenerateJWT(user.ID.String(), user.Role, 24*time.Hour)
	if err != nil {
		log.WithError(err).Error("Falha ao gerar novo JWT")
		return "", err
	}
	log.WithField("user_id", user.ID).Info("Token JWT atualizado com sucesso")

	return newJWT, nil
}
