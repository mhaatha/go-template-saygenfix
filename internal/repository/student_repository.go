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

	UpdateAnswerById(ctx context.Context, tx pgx.Tx, answerId string, answerScore int, answerFeedback string, maxScore int, similarity float64) error
	FindAttemptsByExamIdAndStudentId(ctx context.Context, tx pgx.Tx, userId, examId string) ([]web.ExamAttempt, error)
	UpdateScoresByAttemptId(ctx context.Context, tx pgx.Tx, attemptId string, essayCorrections []domain.EssayCorrection) error
	FindBiggestAttemptsByStudentId(ctx context.Context, tx pgx.Tx, userId string) ([]web.ExamAttemptsCustom, error)
	FindExamsWithScoreAndTeacherNameByExamId(ctx context.Context, tx pgx.Tx, examAttempts []web.ExamAttemptsCustom) ([]web.ExamWithScoreAndTeacherName, error)
	FindBiggestScoreByStudentIdAndExamId(ctx context.Context, tx pgx.Tx, userId string, examId string) (string, int, error)
	FindStudentAnswersByAttemptId(ctx context.Context, tx pgx.Tx, attemptId string) ([]web.StudentAnswer, error)
	FindQuestionById(ctx context.Context, tx pgx.Tx, questionId string) (web.QuestionAndRightAnswer, error)
}
