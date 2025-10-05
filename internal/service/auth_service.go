package service

import (
	"context"

	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
)

type AuthService interface {
	// Login
	Login(ctx context.Context, request web.LoginRequest, email, userHashedPassword, userId string) (string, error)

	// Logout
}
