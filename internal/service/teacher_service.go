package service

import (
	"context"
	"mime/multipart"

	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
)

type TeacherService interface {
	GenerateQuestionAnswer(ctx context.Context, file multipart.File, totalQuestion int, examData domain.Exam)

	TeacherDashboard(ctx context.Context, userId string) (web.TeacherDashboardResponse, error)
}
