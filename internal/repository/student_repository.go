package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
)

type StudentRepository interface {
	FindActiveExams(ctx context.Context, tx pgx.Tx) ([]domain.Exam, error)
	FindTeacherById(ctx context.Context, tx pgx.Tx, teacherId string) (domain.User, error)
	FindExamById(ctx context.Context, tx pgx.Tx, examId string) (domain.Exam, error)
	FindQuestionsByExamId(ctx context.Context, tx pgx.Tx, examId string) ([]domain.QAItem, error)
	CreateExamAttempt(ctx context.Context, tx pgx.Tx, studentId, examId string) (string, error)
	SaveAnswer(ctx context.Context, tx pgx.Tx, answer web.StudentAnswer) error
	CompleteExamAttempt(ctx context.Context, tx pgx.Tx, attemptId string) error

	FindExamByAttemptId(ctx context.Context, tx pgx.Tx, attemptId string) (domain.Exam, error)
	FindAnswersByAttemptId(ctx context.Context, tx pgx.Tx, attemptId string) ([]web.StudentAnswer, error)
}
