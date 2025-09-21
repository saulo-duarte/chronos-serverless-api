package auth

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

type contextKey string

const claimsKey contextKey = "claims"

func Init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("Error")
	}
	jwtSecret = []byte(secret)
}

type Claims struct {
	UserID            string `json:"user_id"`
	Role              string `json:"role"`
	GoogleAccessToken string `json:"google_access_token,omitempty"`
	jwt.RegisteredClaims
}

func GenerateJWTWithGoogleToken(userID, role, googleAccessToken string, duration time.Duration) (string, error) {
	claims := Claims{
		UserID:            userID,
		Role:              role,
		GoogleAccessToken: googleAccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateJWT(userID, role string, duration time.Duration) (string, error) {
	return GenerateJWTWithGoogleToken(userID, role, "", duration)
}

func ValidateJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func GetAccessTokenFromContext(ctx context.Context) (string, error) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	if !ok {
		return "", errors.New("claims not found in context")
	}

	if claims.GoogleAccessToken == "" {
		return "", errors.New("google access token not found in claims")
	}

	return claims.GoogleAccessToken, nil
}
