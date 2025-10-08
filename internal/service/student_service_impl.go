package service

import (
	"context"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mhaatha/go-template-saygenfix/internal/config"
	"github.com/mhaatha/go-template-saygenfix/internal/helper"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/repository"
)

func NewStudentService(studentRepository repository.StudentRepository, db *pgxpool.Pool, validate *validator.Validate, cfg *config.Config) *StudentServiceImpl {
	return &StudentServiceImpl{
		StudentRepository: studentRepository,
		DB:                db,
		Validate:          validate,
		Config:            cfg,
	}
}

type StudentServiceImpl struct {
	StudentRepository repository.StudentRepository
	DB                *pgxpool.Pool
	Validate          *validator.Validate
	Config            *config.Config
}

func (service *StudentServiceImpl) GetActiveExams(ctx context.Context) ([]domain.Exam, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return nil, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	exams, err := service.StudentRepository.FindActiveExams(ctx, tx)
	if err != nil {
		return nil, err
	}

	return exams, nil
}

func (service *StudentServiceImpl) GetTeacherById(ctx context.Context, teacherId string) (domain.User, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return domain.User{}, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	teacher, err := service.StudentRepository.FindTeacherById(ctx, tx, teacherId)
	if err != nil {
		return domain.User{}, err
	}

	return teacher, nil
}

func (service *StudentServiceImpl) GetExamById(ctx context.Context, examId string) (domain.Exam, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return domain.Exam{}, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	exam, err := service.StudentRepository.FindExamById(ctx, tx, examId)
	if err != nil {
		return domain.Exam{}, err
	}

	return exam, nil
}

func (service *StudentServiceImpl) GetQuestionsByExamId(ctx context.Context, examId string) ([]domain.QAItem, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return nil, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	qaList, err := service.StudentRepository.FindQuestionsByExamId(ctx, tx, examId)
	if err != nil {
		return nil, err
	}

	return qaList, nil
}
