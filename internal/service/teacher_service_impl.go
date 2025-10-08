package service

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"mime/multipart"

	"github.com/go-playground/validator/v10"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (service *TeacherServiceImpl) GenerateQuestionAnswer(ctx context.Context, file multipart.File, totalQuestion int, examData domain.Exam, teacherId string) {
	// Handle Gemini API
	client, err := genai.NewClient(ctx, option.WithAPIKey(service.Config.GeminiAPIKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fileURL, err := helper.UploadPDF(ctx, client, file)
	if err != nil {
		log.Fatal(err)
	}

	qaList, err := helper.GenerateQAFromPDF(ctx, client, fileURL, totalQuestion)
	if err != nil {
		log.Fatal(err)
	}

	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Create new exam and save it to database
	examId := "EXAM-" + uuid.NewString()[:8]
	err = service.TeacherRepository.SaveExam(ctx, tx, examData, teacherId, examId)
	if err != nil {
		log.Fatal(err)
	}

	// Save the qaList to the database
	_, err = service.TeacherRepository.BulkSaveQuestionAnswer(ctx, tx, qaList, examId)
	if err != nil {
		log.Fatalf("Gagal menyimpan ke database: %v", err)
	}
}

func (service *TeacherServiceImpl) TeacherDashboard(ctx context.Context, userId string) (web.TeacherDashboardResponse, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Get user by userId
	user, err := service.TeacherRepository.FindUserById(ctx, tx, userId)
	if err != nil {
		slog.Error("failed to find user by id", "err", err)

		if errors.Is(err, pgx.ErrNoRows) {
			return web.TeacherDashboardResponse{}, errors.New("user not found")
		}
		return web.TeacherDashboardResponse{}, err
	}

	if user.Role == "teacher" {
		user.Role = "Teacher"
	}

	// Get exams by userId
	exams, err := service.TeacherRepository.FindExamsByUserId(ctx, tx, userId)
	if err != nil {
		slog.Error("failed to find exams by user id", "err", err)

		if errors.Is(err, pgx.ErrNoRows) {
			return web.TeacherDashboardResponse{}, nil
		}
		return web.TeacherDashboardResponse{}, err
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
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return domain.Exam{}, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	exam, err := service.TeacherRepository.FindExamById(ctx, tx, examId)
	if err != nil {
		return domain.Exam{}, err
	}

	err = service.TeacherRepository.UpdateIsActiveExamById(ctx, tx, examId, exam.IsActive)
	if err != nil {
		return domain.Exam{}, err
	}

	updatedExam, err := service.TeacherRepository.FindExamById(ctx, tx, examId)
	if err != nil {
		return domain.Exam{}, err
	}

	return updatedExam, nil
}
