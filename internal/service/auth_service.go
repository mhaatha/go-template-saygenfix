package service

import (
	"context"

	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
)

type AuthService interface {
	// Login
	Login(ctx context.Context, request web.LoginRequest, email, userHashedPassword, userId string) (string, error)

	// ValidateSession
	ValidateSession(ctx context.Context, sessionId string) (domain.User, error)

	// Logout
}
