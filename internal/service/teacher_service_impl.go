package service

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/go-playground/validator/v10"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mhaatha/go-template-saygenfix/internal/config"
	"github.com/mhaatha/go-template-saygenfix/internal/helper"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
	"github.com/mhaatha/go-template-saygenfix/internal/repository"
	"google.golang.org/api/option"
)

func NewTeacherService(teacherRepository repository.TeacherRepository, db *pgxpool.Pool, validate *validator.Validate, cfg *config.Config) TeacherService {
	return &TeacherServiceImpl{
		TeacherRepository: teacherRepository,
		DB:                db,
		Validate:          validate,
		Config:            cfg,
	}
}

type TeacherServiceImpl struct {
	TeacherRepository repository.TeacherRepository
	DB                *pgxpool.Pool
	Validate          *validator.Validate
	Config            *config.Config
}

func (service *TeacherServiceImpl) GenerateQuestionAnswer(ctx context.Context, file multipart.File, totalQuestion int, examData domain.Exam, teacherId string) error {
	// Handle Gemini API
	client, err := genai.NewClient(ctx, option.WithAPIKey(service.Config.GeminiAPIKey))
	if err != nil {
		return fmt.Errorf("error when calling NewClient: %w", err)
	}
	defer client.Close()

	fileURL, err := helper.UploadPDF(ctx, client, file)
	if err != nil {
		return fmt.Errorf("failed when calling UploadPDF helper: %w", err)
	}

	qaList, err := helper.GenerateQAFromPDF(ctx, client, fileURL, totalQuestion)
	if err != nil {
		return fmt.Errorf("failed when calling GenerateQAFromPDF helper: %w", err)
	}

	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Create new exam and save it to database
	examId := "EXAM-" + uuid.NewString()[:8]
	err = service.TeacherRepository.SaveExam(ctx, tx, examData, teacherId, examId)
	if err != nil {
		return fmt.Errorf("failed when SaveExam repository: %w", err)
	}

	// Save the qaList to the database
	_, err = service.TeacherRepository.BulkSaveQuestionAnswer(ctx, tx, qaList, examId)
	if err != nil {
		return fmt.Errorf("failed when calling BulkSaveQuestionAnswer repository: %w", err)
	}

	return nil
}

func (service *TeacherServiceImpl) TeacherDashboard(ctx context.Context, userId string) (web.TeacherDashboardResponse, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.TeacherDashboardResponse{}, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Get user by userId
	user, err := service.TeacherRepository.FindUserById(ctx, tx, userId)
	if err != nil {
		return web.TeacherDashboardResponse{}, fmt.Errorf("failed when calling FindUserById repository: %w", err)
	}

	if user.Role == "teacher" {
		user.Role = "Teacher"
	}

	// Get exams by userId
	exams, err := service.TeacherRepository.FindExamsByUserId(ctx, tx, userId)
	if err != nil {
		return web.TeacherDashboardResponse{}, fmt.Errorf("failed when calling FindExamsByUserId repository: %w", err)
	}

	dashboardData := web.TeacherDashboardResponse{
		User:  user,
		Exams: exams,
	}

	return dashboardData, nil
}

func (service *TeacherServiceImpl) UpdateIsActiveExamById(ctx context.Context, userId, examId string) (domain.Exam, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return domain.Exam{}, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	exam, err := service.TeacherRepository.FindExamById(ctx, tx, examId)
	if err != nil {
		return domain.Exam{}, fmt.Errorf("failed when calling FindExamById repository: %w", err)
	}

	err = service.TeacherRepository.UpdateIsActiveExamById(ctx, tx, examId, exam.IsActive)
	if err != nil {
		return domain.Exam{}, fmt.Errorf("failed when calling UpdateIsActiveExamById repository: %w", err)
	}

	updatedExam, err := service.TeacherRepository.FindExamById(ctx, tx, examId)
	if err != nil {
		return domain.Exam{}, fmt.Errorf("failed when calling FindExamById repository: %w", err)
	}

	return updatedExam, nil
}

func (service *TeacherServiceImpl) GetExamById(ctx context.Context, examId string) (domain.Exam, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return domain.Exam{}, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	exam, err := service.TeacherRepository.FindExamById(ctx, tx, examId)
	if err != nil {
		return domain.Exam{}, fmt.Errorf("failed when calling FindExamById repository: %w", err)
	}

	return exam, nil
}

func (service *TeacherServiceImpl) GetQAByExamId(ctx context.Context, examId string) ([]domain.QAItem, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	qaList, err := service.TeacherRepository.FindQAByExamId(ctx, tx, examId)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindQAByExamId repository: %w", err)
	}

	return qaList, nil
}

func (service *TeacherServiceImpl) UpdateExamById(ctx context.Context, examId, roomName string, yearInt, durationInt int) error {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	err = service.TeacherRepository.UpdateExamById(ctx, tx, examId, roomName, yearInt, durationInt)
	if err != nil {
		return fmt.Errorf("failed when calling UpdateExamById repository: %w", err)
	}

	return nil
}

func (service *TeacherServiceImpl) UpdateQuestionById(ctx context.Context, questionId, questionText, answerText string) error {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	err = service.TeacherRepository.UpdateQuestionById(ctx, tx, questionId, questionText, answerText)
	if err != nil {
		return fmt.Errorf("failed when calling UpdateQuestionById repository: %w", err)
	}

	return nil
}

func (service *TeacherServiceImpl) GetBiggestExamAttemptsScoreByExamId(ctx context.Context, examId string) ([]web.ExamAttempt, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	attempts, err := service.TeacherRepository.FindBiggestAttemptsByExamId(ctx, tx, examId)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindBiggestAttemptsByExamId repository: %w", err)
	}

	return attempts, nil
}

func (service *TeacherServiceImpl) GetStudentFullNameByExamAttemptsId(ctx context.Context, examAttemptsId string) (string, string, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	studentFullName, studentId, err := service.TeacherRepository.FindStudentFullNameByExamAttemptsId(ctx, tx, examAttemptsId)
	if err != nil {
		return "", "", fmt.Errorf("failed when calling FindStudentFullNameByExamAttemptsId repository: %w", err)
	}

	return studentFullName, studentId, nil
}
