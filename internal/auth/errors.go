package auth

import "errors"

var (
	ErrMissingToken            = errors.New("missing authorization token")
	ErrInvalidToken            = errors.New("invalid token")
	ErrExpiredToken            = errors.New("token has expired")
	ErrInvalidClaims           = errors.New("invalid token claims")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)
