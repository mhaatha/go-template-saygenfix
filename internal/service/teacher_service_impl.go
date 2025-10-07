package service

import (
	"context"
	"log"
	"mime/multipart"

	"github.com/go-playground/validator/v10"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mhaatha/go-template-saygenfix/internal/config"
	"github.com/mhaatha/go-template-saygenfix/internal/helper"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
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

func (service *TeacherServiceImpl) GenerateQuestionAnswer(ctx context.Context, file multipart.File, totalQuestion int, examData domain.Exam) {
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
	teacherId := "60548cf3-9624-4f00-865f-421a7f6922cf" // !!! Dapatkan dari session
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
