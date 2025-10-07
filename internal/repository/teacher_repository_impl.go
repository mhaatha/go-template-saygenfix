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
