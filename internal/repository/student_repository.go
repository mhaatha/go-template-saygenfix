package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

type StudentRepository interface {
	FindActiveExams(ctx context.Context, tx pgx.Tx) ([]domain.Exam, error)
	FindTeacherById(ctx context.Context, tx pgx.Tx, teacherId string) (domain.User, error)
}
