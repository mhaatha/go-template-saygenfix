package service

import (
	"context"
	"mime/multipart"

	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

type TeacherService interface {
	GenerateQuestionAnswer(ctx context.Context, file multipart.File, totalQuestion int, examData domain.Exam)
}
