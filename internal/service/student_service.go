package service

import (
	"context"

	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

type StudentService interface {
	GetActiveExams(ctx context.Context) ([]domain.Exam, error)
	GetTeacherById(ctx context.Context, teacherId string) (domain.User, error)
}
