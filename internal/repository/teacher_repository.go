package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

type TeacherRepository interface {
	SaveExam(ctx context.Context, tx pgx.Tx, examData domain.Exam, teacherId string, examId string) error
	BulkSaveQuestionAnswer(ctx context.Context, tx pgx.Tx, questionsAndAnswers []domain.QAItem, examId string) (string, error)

	FindUserById(ctx context.Context, tx pgx.Tx, userId string) (domain.User, error)
	FindExamsByUserId(ctx context.Context, tx pgx.Tx, userId string) ([]domain.Exam, error)
}
