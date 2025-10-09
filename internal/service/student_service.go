package service

import (
	"context"

	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
)

type StudentService interface {
	GetActiveExams(ctx context.Context) ([]domain.Exam, error)
	GetTeacherById(ctx context.Context, teacherId string) (domain.User, error)
	GetExamById(ctx context.Context, examId string) (domain.Exam, error)
	GetQuestionsByExamId(ctx context.Context, examId string) ([]domain.QAItem, error)

	CreateExamAttempt(ctx context.Context, studentId, examId string) (string, error)
	SaveAnswer(ctx context.Context, answer web.StudentAnswer) error
	CompleteExamAttempt(ctx context.Context, attemptId string) error

	GetExamByAttempId(ctx context.Context, attemptId string) (domain.Exam, error)
	GetAnswersByAttemptId(ctx context.Context, attemptId string) ([]web.StudentAnswer, error)

	GetExamAttemptsByExamIdAndStudentId(ctx context.Context, userId string, examId string) ([]web.ExamAttempt, error)
	CalculateScore(ctx context.Context, attemptId string) ([]domain.EssayCorrection, error)

	GetBiggestExamAttemptsByStudentId(ctx context.Context, userId string) ([]web.ExamAttemptsCustom, error)
	GetExamsWithScoreAndTeacherNameByExamId(ctx context.Context, examAttempts []web.ExamAttemptsCustom) ([]web.ExamWithScoreAndTeacherName, error)
}
