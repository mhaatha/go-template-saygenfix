package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

func NewTeacherRepository() TeacherRepository {
	return &teacherRepositoryImpl{}
}

type teacherRepositoryImpl struct{}

func (r *teacherRepositoryImpl) SaveExam(ctx context.Context, tx pgx.Tx, examData domain.Exam, teacherId string, examId string) error {
	sqlQuery := `
	INSERT INTO exams (id, name, year, duration_in_minutes, teacher_id)
	VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tx.Exec(
		ctx,
		sqlQuery,
		examId,
		examData.RoomName,
		examData.Year,
		examData.Duration,
		teacherId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *teacherRepositoryImpl) BulkSaveQuestionAnswer(ctx context.Context, tx pgx.Tx, questionsAndAnswers []domain.QAItem, examId string) (string, error) {
	sqlQuery := `
	INSERT INTO questions (id, question, correct_answer, exam_id)
	VALUES ($1, $2, $3, $4)
	`

	stmt, err := tx.Prepare(ctx, "question_answer", sqlQuery)
	if err != nil {
		return "", nil
	}

	for _, item := range questionsAndAnswers {
		questionId := uuid.New()
		_, err := tx.Exec(
			ctx,
			stmt.Name,
			questionId,
			item.Question,
			item.Answer,
			examId,
		)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

func (r *teacherRepositoryImpl) FindUserById(ctx context.Context, tx pgx.Tx, userId string) (domain.User, error) {
	sqlQuery := `
	SELECT id, email, full_name, password, role, created_at, updated_at
	FROM users
	WHERE id = $1
	`

	user := domain.User{}

	err := tx.QueryRow(ctx, sqlQuery, userId).Scan(
		&user.Id,
		&user.Email,
		&user.FullName,
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

func (r *teacherRepositoryImpl) FindExamsByUserId(ctx context.Context, tx pgx.Tx, userId string) ([]domain.Exam, error) {
	sqlQuery := `
	SELECT id, name, year, teacher_id, duration_in_minutes, is_active, created_at, updated_at
	FROM exams
	WHERE teacher_id = $1
	`

	rows, err := tx.Query(ctx, sqlQuery, userId)
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
