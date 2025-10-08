package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

type AuthRepository interface {
	// Save session
	Save(ctx context.Context, tx pgx.Tx, session domain.Session) (domain.Session, error)
	// Find user by session id
	FindUserBySessionId(ctx context.Context, tx pgx.Tx, sessionId string) (domain.User, error)
	// Delete session
	Delete(ctx context.Context, tx pgx.Tx, sessionId string) error
}
