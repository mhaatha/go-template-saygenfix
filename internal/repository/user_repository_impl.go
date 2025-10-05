package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

func NewUserRepository() UserRepository {
	return &UserRepositoryImpl{}
}

type UserRepositoryImpl struct{}

func (repository *UserRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, user domain.User) error {
	sqlQuery := `
	INSERT INTO users (id, email, full_name, password, role)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, updated_at
	`

	err := tx.QueryRow(
		ctx,
		sqlQuery,
		uuid.NewString(),
		user.Email,
		user.FullName,
		user.Password,
		user.Role,
	).Scan(
		&user.Id,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (repository *UserRepositoryImpl) FindByEmail(ctx context.Context, tx pgx.Tx, email string) (domain.User, error) {
	sqlQuery := `
	SELECT id, email, full_name, password, role
	FROM users
	WHERE email = $1
	`

	user := domain.User{}

	err := tx.QueryRow(
		ctx,
		sqlQuery,
		email,
	).Scan(
		&user.Id,
		&user.Email,
		&user.FullName,
		&user.Password,
		&user.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, nil
		}

		return domain.User{}, err
	}

	return user, nil
}
