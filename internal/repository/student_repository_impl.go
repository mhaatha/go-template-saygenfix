package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

func NewStudentRepository() StudentRepository {
	return &StudentRepositoryImpl{}
}

type StudentRepositoryImpl struct{}

func (repository *StudentRepositoryImpl) FindActiveExams(ctx context.Context, tx pgx.Tx) ([]domain.Exam, error) {
	sqlQuery := `
	SELECT id, name, year, teacher_id, duration_in_minutes, is_active, created_at, updated_at
	FROM exams
	WHERE is_active = true
	`

	rows, err := tx.Query(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}

	exams := []domain.Exam{}
	for rows.Next() {
		exam := domain.Exam{}
		err := rows.Scan(
			&exam.Id,
			&exam.RoomName,
			&exam.Year,
			&exam.TeacherId,
			&exam.Duration,
			&exam.IsActive,
			&exam.CreatedAt,
			&exam.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		exams = append(exams, exam)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return exams, nil
}

func (repository *StudentRepositoryImpl) FindTeacherById(ctx context.Context, tx pgx.Tx, teacherId string) (domain.User, error) {
	sqlQuery := `
	SELECT id, full_name, email, password, role, created_at, updated_at
	FROM users
	WHERE id = $1
	`

	user := domain.User{}
	err := tx.QueryRow(ctx, sqlQuery, teacherId).Scan(
		&user.Id,
		&user.FullName,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (repository *StudentRepositoryImpl) FindExamById(ctx context.Context, tx pgx.Tx, examId string) (domain.Exam, error) {
	sqlQuery := `
	SELECT id, name, year, teacher_id, duration_in_minutes, is_active, created_at, updated_at
	FROM exams
	WHERE id = $1
	`

	exam := domain.Exam{}
	err := tx.QueryRow(ctx, sqlQuery, examId).Scan(
		&exam.Id,
		&exam.RoomName,
		&exam.Year,
		&exam.TeacherId,
		&exam.Duration,
		&exam.IsActive,
		&exam.CreatedAt,
		&exam.UpdatedAt,
	)
	if err != nil {
		return domain.Exam{}, err
	}

	return exam, nil
}

func (repository *StudentRepositoryImpl) FindQuestionsByExamId(ctx context.Context, tx pgx.Tx, examId string) ([]domain.QAItem, error) {
	sqlQuery := `
	SELECT id, question, correct_answer, exam_id
	FROM questions
	WHERE exam_id = $1
	`

	rows, err := tx.Query(ctx, sqlQuery, examId)
	if err != nil {
		return nil, err
	}

	questions := []domain.QAItem{}
	for rows.Next() {
		question := domain.QAItem{}
		err := rows.Scan(
			&question.Id,
			&question.Question,
			&question.Answer,
			&question.ExamId,
		)
		if err != nil {
			return nil, err
		}
		questions = append(questions, question)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return questions, nil
}
