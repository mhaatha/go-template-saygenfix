package service

import (
	"context"
	"mime/multipart"

	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
)

type TeacherService interface {
	GenerateQuestionAnswer(ctx context.Context, file multipart.File, totalQuestion int, examData domain.Exam, teacherId string)

	TeacherDashboard(ctx context.Context, userId string) (web.TeacherDashboardResponse, error)
	UpdateIsActiveExamById(ctx context.Context, userId, examId string) (domain.Exam, error)
	GetExamById(ctx context.Context, examId string) (domain.Exam, error)
	GetQAByExamId(ctx context.Context, examId string) ([]domain.QAItem, error)
	UpdateExamById(ctx context.Context, examId, roomName string, yearInt, durationInt int) error

	UpdateQuestionById(ctx context.Context, questionId, questionText, answerText string) error

	GetBiggestExamAttemptsScoreByExamId(ctx context.Context, examId string) ([]web.ExamAttempt, error)
	GetStudentFullNameByExamAttemptsId(ctx context.Context, examAttemptsId string) (string, error)
}
