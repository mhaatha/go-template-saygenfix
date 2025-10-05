package service

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mhaatha/go-template-saygenfix/internal/helper"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
	"github.com/mhaatha/go-template-saygenfix/internal/repository"
)

func NewAuthService(authRepository repository.AuthRepository, db *pgxpool.Pool, validate *validator.Validate) AuthService {
	return &AuthServiceImpl{
		AuthRepository: authRepository,
		DB:             db,
		Validate:       validate,
	}
}

type AuthServiceImpl struct {
	AuthRepository repository.AuthRepository
	DB             *pgxpool.Pool
	Validate       *validator.Validate
}

func (service *AuthServiceImpl) Login(ctx context.Context, request web.LoginRequest, email, userHashedPassword, userId string) (string, error) {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return "", err
	}

	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check if request password matched the hashed password
	if !helper.CheckPasswordHash(userHashedPassword, request.Password) {
		return "", fmt.Errorf("invalid email or password")
	}

	// Save session to db
	session, err := service.AuthRepository.Save(ctx, tx, domain.Session{
		SessionId: helper.Base64SessionId(),
		UserId:    userId,
	})
	if err != nil {
		return "", err
	}

	return session.SessionId, nil
}
