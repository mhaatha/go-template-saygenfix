package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

type UserRepository interface {
	Save(ctx context.Context, tx pgx.Tx, user domain.User) error
	FindByEmail(ctx context.Context, tx pgx.Tx, email string) (domain.User, error)
}
