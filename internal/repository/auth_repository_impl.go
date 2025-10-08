package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

func NewAuthRepository() AuthRepository {
	return &AuthRepositoryImpl{}
}

type AuthRepositoryImpl struct{}

func (repository *AuthRepositoryImpl) Save(ctx context.Context, tx pgx.Tx, session domain.Session) (domain.Session, error) {
	sqlQuery := `
	INSERT INTO sessions (session_id, user_id)
	VALUES ($1, $2)
	RETURNING created_at
	`

	err := tx.QueryRow(
		ctx,
		sqlQuery,
		session.SessionId,
		session.UserId,
	).Scan(
		&session.CreatedAt,
	)
	if err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

func (repository *AuthRepositoryImpl) FindUserBySessionId(ctx context.Context, tx pgx.Tx, sessionId string) (domain.User, error) {
	sqlQuery := `
	SELECT u.id, u.email, u.full_name, u.password, u.role, u.created_at, u.updated_at
	FROM users u
	JOIN sessions s ON u.id = s.user_id
	WHERE s.session_id = $1
	`

	var user = domain.User{}

	err := tx.QueryRow(ctx, sqlQuery, sessionId).Scan(
		&user.Id,
		&user.Email,
		&user.FullName,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (repository *AuthRepositoryImpl) Delete(ctx context.Context, tx pgx.Tx, sessionId string) error {
	sqlQuery := `
	DELETE FROM sessions
	WHERE session_id = $1
	`

	_, err := tx.Exec(ctx, sqlQuery, sessionId)
	if err != nil {
		return err
	}

	return nil
}
