package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type UserDataKey string

const (
	UserDataKeyID   UserDataKey = "userID"
	UserDataKeyRole UserDataKey = "userRole"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := extractToken(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := ValidateJWT(tokenStr)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserDataKeyID, claims.UserID)
		ctx = context.WithValue(ctx, UserDataKeyRole, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1], nil
		}
	}

	cookie, err := r.Cookie("jwt")
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return "", errors.New("authorization token not found in header or cookie")
}
