package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
)

type TeacherRepository interface {
	SaveExam(ctx context.Context, tx pgx.Tx, examData domain.Exam, teacherId string, examId string) error
	BulkSaveQuestionAnswer(ctx context.Context, tx pgx.Tx, questionsAndAnswers []domain.QAItem, examId string) (string, error)

	FindUserById(ctx context.Context, tx pgx.Tx, userId string) (domain.User, error)
	FindExamsByUserId(ctx context.Context, tx pgx.Tx, userId string) ([]domain.Exam, error)

	FindExamById(ctx context.Context, tx pgx.Tx, examId string) (domain.Exam, error)
	UpdateIsActiveExamById(ctx context.Context, tx pgx.Tx, examId string, currentIsActive bool) error

	FindQAByExamId(ctx context.Context, tx pgx.Tx, examId string) ([]domain.QAItem, error)

	UpdateExamById(ctx context.Context, tx pgx.Tx, examId, roomName string, yearInt, durationInt int) error
	UpdateQuestionById(ctx context.Context, tx pgx.Tx, questionId, questionText, answerText string) error

	FindBiggestAttemptsByExamId(ctx context.Context, tx pgx.Tx, examId string) ([]web.ExamAttempt, error)
	FindStudentFullNameByExamAttemptsId(ctx context.Context, tx pgx.Tx, examAttemptsId string) (string, error)
}
