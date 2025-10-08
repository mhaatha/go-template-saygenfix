package service

import (
	"context"

	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

type StudentService interface {
	GetActiveExams(ctx context.Context) ([]domain.Exam, error)
	GetTeacherById(ctx context.Context, teacherId string) (domain.User, error)
	GetExamById(ctx context.Context, examId string) (domain.Exam, error)
	GetQuestionsByExamId(ctx context.Context, examId string) ([]domain.QAItem, error)
}
