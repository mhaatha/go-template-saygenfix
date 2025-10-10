package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
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

func (r *teacherRepositoryImpl) FindExamById(ctx context.Context, tx pgx.Tx, examId string) (domain.Exam, error) {
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
func (r *teacherRepositoryImpl) UpdateIsActiveExamById(ctx context.Context, tx pgx.Tx, examId string, currentIsActive bool) error {
	sqlQuery := `
	UPDATE exams
	SET is_active = $1
	WHERE id = $2
	`

	isActive := !currentIsActive
	_, err := tx.Exec(ctx, sqlQuery, isActive, examId)
	if err != nil {
		return err
	}

	return nil
}

func (r *teacherRepositoryImpl) FindQAByExamId(ctx context.Context, tx pgx.Tx, examId string) ([]domain.QAItem, error) {
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

func (r *teacherRepositoryImpl) UpdateExamById(ctx context.Context, tx pgx.Tx, examId, roomName string, yearInt, durationInt int) error {
	sqlQuery := `
	UPDATE exams
	SET name = $1, year = $2, duration_in_minutes = $3, updated_at = now()
	WHERE id = $4
	`

	_, err := tx.Exec(ctx, sqlQuery, roomName, yearInt, durationInt, examId)
	if err != nil {
		return err
	}

	return nil
}

func (r *teacherRepositoryImpl) UpdateQuestionById(ctx context.Context, tx pgx.Tx, questionId, questionText, answerText string) error {
	sqlQuery := `
	UPDATE questions
	SET question = $1, correct_answer = $2
	WHERE id = $3
	`

	_, err := tx.Exec(ctx, sqlQuery, questionText, answerText, questionId)
	if err != nil {
		return err
	}

	return nil
}

func (r *teacherRepositoryImpl) FindBiggestAttemptsByExamId(ctx context.Context, tx pgx.Tx, examId string) ([]web.ExamAttempt, error) {
	// 1. Query tetap sama, ambil semua attempt untuk ujian ini.
	sqlQuery := `
    SELECT id, student_id, score, started_at, completed_at
    FROM exam_attempts
    WHERE exam_id = $1
    `

	rows, err := tx.Query(ctx, sqlQuery, examId)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Jangan lupa untuk selalu menutup rows.

	// 2. Buat map untuk menampung attempt dengan skor tertinggi per siswa.
	// Kunci map adalah student_id (string), nilainya adalah struct ExamAttempt itu sendiri.
	// Map ini diinisialisasi SATU KALI di luar loop.
	highestScoreAttempts := make(map[string]web.ExamAttempt)

	for rows.Next() {
		var currentAttempt web.ExamAttempt
		err := rows.Scan(
			&currentAttempt.ID,
			&currentAttempt.StudentID,
			&currentAttempt.Score,
			&currentAttempt.StartedAt,
			&currentAttempt.CompletedAt,
		)
		if err != nil {
			return nil, err
		}

		// 3. Logika untuk memfilter skor tertinggi.
		// Cek apakah sudah ada attempt untuk StudentID ini di map.
		existingAttempt, ok := highestScoreAttempts[currentAttempt.StudentID]

		// Jika belum ada (ok == false), atau jika skor attempt saat ini LEBIH BESAR
		// dari yang sudah ada, maka simpan/timpa data di map dengan attempt saat ini.
		if !ok || currentAttempt.Score > existingAttempt.Score {
			highestScoreAttempts[currentAttempt.StudentID] = currentAttempt
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// 4. Ubah map menjadi slice untuk hasil akhir.
	// Pada titik ini, `highestScoreAttempts` hanya berisi satu attempt per siswa,
	// yaitu yang memiliki skor tertinggi.
	finalAttempts := make([]web.ExamAttempt, 0, len(highestScoreAttempts))
	for _, attempt := range highestScoreAttempts {
		finalAttempts = append(finalAttempts, attempt)
	}

	return finalAttempts, nil
}

func (r *teacherRepositoryImpl) FindStudentFullNameByExamAttemptsId(ctx context.Context, tx pgx.Tx, examAttemptsId string) (string, string, error) {
	sqlQuery := `
	SELECT full_name, id
	FROM users
	WHERE id = (
		SELECT student_id
		FROM exam_attempts
		WHERE id = $1
	)
	`

	var fullName string
	var id string
	err := tx.QueryRow(ctx, sqlQuery, examAttemptsId).Scan(&fullName, &id)
	if err != nil {
		return "", "", err
	}

	return fullName, id, nil
}
