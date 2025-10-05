package service

import (
	"context"

	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
)

type UserService interface {
	RegisterNewUser(ctx context.Context, request web.RegisterUserRequest) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}
