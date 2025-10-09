package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
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

func (repository *StudentRepositoryImpl) CreateExamAttempt(ctx context.Context, tx pgx.Tx, studentId, examId string) (string, error) {
	sqlQuery := `
	INSERT INTO exam_attempts (student_id, exam_id)
	VALUES ($1, $2)
	RETURNING id
	`

	var examAttemptId string
	err := tx.QueryRow(ctx, sqlQuery, studentId, examId).Scan(&examAttemptId)
	if err != nil {
		return "", err
	}

	return examAttemptId, nil
}

func (repository *StudentRepositoryImpl) SaveAnswer(ctx context.Context, tx pgx.Tx, answer web.StudentAnswer) error {
	sqlQuery := `
	INSERT INTO student_answers (exam_attempt_id, question_id, student_answer)
	VALUES ($1, $2, $3)
	`

	_, err := tx.Exec(ctx, sqlQuery, answer.ExamAttemptID, answer.QuestionID, answer.StudentAnswer)

	return err
}

func (repository *StudentRepositoryImpl) CompleteExamAttempt(ctx context.Context, tx pgx.Tx, attemptId string) error {
	sqlQuery := `
	UPDATE exam_attempts
	SET completed_at = now()
	WHERE id = $1
	`

	_, err := tx.Exec(ctx, sqlQuery, attemptId)
	if err != nil {
		return err
	}

	return nil
}

func (repository *StudentRepositoryImpl) FindExamByAttemptId(ctx context.Context, tx pgx.Tx, attemptId string) (domain.Exam, error) {
	sqlQuery := `
	SELECT id, name, year, teacher_id, duration_in_minutes, is_active, created_at, updated_at
	FROM exams
	WHERE id = (
		SELECT exam_id
		FROM exam_attempts
		WHERE id = $1
	)
	`

	exam := domain.Exam{}
	err := tx.QueryRow(ctx, sqlQuery, attemptId).Scan(
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

func (repository *StudentRepositoryImpl) FindAnswersByAttemptId(ctx context.Context, tx pgx.Tx, attemptId string) ([]web.StudentAnswer, error) {
	sqlQuery := `
	SELECT id, question_id, student_answer, score, feedback
	FROM student_answers
	WHERE exam_attempt_id = $1
	`

	rows, err := tx.Query(ctx, sqlQuery, attemptId)
	if err != nil {
		return nil, err
	}

	answers := []web.StudentAnswer{}
	for rows.Next() {
		answer := web.StudentAnswer{}
		err := rows.Scan(
			&answer.ID,
			&answer.QuestionID,
			&answer.StudentAnswer,
			&answer.Score,
			&answer.Feedback,
		)
		if err != nil {
			return nil, err
		}
		answers = append(answers, answer)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return answers, nil
}

func (repository *StudentRepositoryImpl) UpdateAnswerById(ctx context.Context, tx pgx.Tx, answerId string, answerScore int, answerFeedback string) error {
	sqlQuery := `
	UPDATE student_answers
	SET score = $1, feedback = $2
	WHERE id = $3
	`

	_, err := tx.Exec(ctx, sqlQuery, answerScore, answerFeedback, answerId)
	if err != nil {
		return err
	}

	return nil
}

func (repository *StudentRepositoryImpl) FindAttemptsByExamIdAndStudentId(ctx context.Context, tx pgx.Tx, userId, examId string) ([]web.ExamAttempt, error) {
	sqlQuery := `
	SELECT id, student_id, exam_id, score, started_at, completed_at
	FROM exam_attempts
	WHERE student_id = $1 AND exam_id = $2
	`

	rows, err := tx.Query(ctx, sqlQuery, userId, examId)
	if err != nil {
		return nil, err
	}

	attempts := []web.ExamAttempt{}
	for rows.Next() {
		attempt := web.ExamAttempt{}
		err := rows.Scan(
			&attempt.ID,
			&attempt.StudentID,
			&attempt.ExamID,
			&attempt.Score,
			&attempt.StartedAt,
			&attempt.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return attempts, nil
}

func (repository *StudentRepositoryImpl) UpdateScoresByAttemptId(ctx context.Context, tx pgx.Tx, attemptId string, essayCorrections []domain.EssayCorrection) error {
	sqlQuery := `
	UPDATE exam_attempts
	SET score = $1
	WHERE id = $2
	`

	totalScore := 0
	for _, essayCorrection := range essayCorrections {
		totalScore += essayCorrection.Score
	}

	_, err := tx.Exec(ctx, sqlQuery, totalScore, attemptId)
	if err != nil {
		return err
	}

	return nil
}
