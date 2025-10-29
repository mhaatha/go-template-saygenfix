package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mhaatha/go-template-saygenfix/internal/helper"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
	"github.com/mhaatha/go-template-saygenfix/internal/repository"
)

func NewUserService(userRepository repository.UserRepository, db *pgxpool.Pool, validate *validator.Validate) UserService {
	return &UserServiceImpl{
		UserRepository: userRepository,
		DB:             db,
		Validate:       validate,
	}
}

type UserServiceImpl struct {
	UserRepository repository.UserRepository
	DB             *pgxpool.Pool
	Validate       *validator.Validate
}

func (service *UserServiceImpl) RegisterNewUser(ctx context.Context, request web.RegisterUserRequest) error {
	// Validate request
	err := service.Validate.Struct(request)
	if err != nil {
		return fmt.Errorf("failed to validate request body: %w", err)
	}

	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check if email already exists
	existingUser, err := service.UserRepository.FindByEmail(ctx, tx, request.Email)
	if err != nil {
		return fmt.Errorf("failed when calling FindByEmail repository: %w", err)
	}
	if existingUser.Id != "" {
		return errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := helper.HashPassword(request.Password)
	if err != nil {
		return fmt.Errorf("failed when calling HashPassword: %w", err)
	}

	user := domain.User{
		Email:    request.Email,
		FullName: request.FullName,
		Password: hashedPassword,
		Role:     request.Role,
	}

	return service.UserRepository.Save(ctx, tx, user)
}

func (service *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Check is email exists
	user, err := service.UserRepository.FindByEmail(ctx, tx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed when calling FindByEmail repository: %w", err)
	}
	if user.Id == "" {
		return domain.User{}, errors.New("user not found")
	}

	return user, nil
}
