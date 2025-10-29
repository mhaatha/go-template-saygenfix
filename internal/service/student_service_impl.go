package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mhaatha/go-template-saygenfix/internal/config"
	"github.com/mhaatha/go-template-saygenfix/internal/helper"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
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
		return nil, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	exams, err := service.StudentRepository.FindActiveExams(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FincActiveExams repository: %w", err)
	}

	return exams, nil
}

func (service *StudentServiceImpl) GetTeacherById(ctx context.Context, teacherId string) (domain.User, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	teacher, err := service.StudentRepository.FindTeacherById(ctx, tx, teacherId)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed when calling FindTeacherById: %w", err)
	}

	return teacher, nil
}

func (service *StudentServiceImpl) GetExamById(ctx context.Context, examId string) (domain.Exam, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return domain.Exam{}, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	exam, err := service.StudentRepository.FindExamById(ctx, tx, examId)
	if err != nil {
		return domain.Exam{}, fmt.Errorf("failed when calling FindExamById repository: %w", err)
	}

	return exam, nil
}

func (service *StudentServiceImpl) GetQuestionsByExamId(ctx context.Context, examId string) ([]domain.QAItem, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	qaList, err := service.StudentRepository.FindQuestionsByExamId(ctx, tx, examId)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindQuestionsByExamId repository: %w", err)
	}

	return qaList, nil
}

func (service *StudentServiceImpl) CreateExamAttempt(ctx context.Context, studentId, examId string) (string, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	examAttemptId, err := service.StudentRepository.CreateExamAttempt(ctx, tx, studentId, examId)
	if err != nil {
		return "", fmt.Errorf("failed when calling CreateExamAttempt repository: %w", err)
	}

	return examAttemptId, nil
}

func (service *StudentServiceImpl) SaveAnswer(ctx context.Context, answer web.StudentAnswer) error {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	err = service.StudentRepository.SaveAnswer(ctx, tx, answer)
	if err != nil {
		return fmt.Errorf("failed when calling SaveAnswer repository: %w", err)
	}

	return nil
}

func (service *StudentServiceImpl) CompleteExamAttempt(ctx context.Context, attemptId string) error {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	err = service.StudentRepository.CompleteExamAttempt(ctx, tx, attemptId)
	if err != nil {
		return fmt.Errorf("failed when calling CompleteExamAttempt repository: %w", err)
	}

	return nil
}

func (service *StudentServiceImpl) GetExamByAttempId(ctx context.Context, attemptId string) (domain.Exam, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return domain.Exam{}, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	exam, err := service.StudentRepository.FindExamByAttemptId(ctx, tx, attemptId)
	if err != nil {
		return domain.Exam{}, fmt.Errorf("failed when calling FindExamByAttemptId repository: %w", err)
	}

	return exam, nil
}

func (service *StudentServiceImpl) GetAnswersByAttemptId(ctx context.Context, attemptId string) ([]web.StudentAnswer, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	answers, err := service.StudentRepository.FindAnswersByAttemptId(ctx, tx, attemptId)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindAnswersByAttemptId repository: %w", err)
	}

	return answers, nil
}

func (service *StudentServiceImpl) CalculateScore(ctx context.Context, attemptId string) ([]domain.EssayCorrection, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Get exam by attempId
	exam, err := service.StudentRepository.FindExamByAttemptId(ctx, tx, attemptId)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindExamByAttemptId repository: %w", err)
	}

	// Get questions by examId
	questions, err := service.StudentRepository.FindQuestionsByExamId(ctx, tx, exam.Id)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindQuestionsByExamId repository: %w", err)
	}

	type QuestionAnswer struct {
		Id            string `json:"id"`
		Question      string `json:"question"`
		CorrectAnswer string `json:"correct_answer"`
		StudentAnswer string `json:"student_answer"`
	}

	// Get student answers to struct the QuestionAnswer struct
	studentAnswers, err := service.StudentRepository.FindAnswersByAttemptId(ctx, tx, attemptId)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindAnswersByAttemptId repository: %w", err)
	}

	// Merge questions and answers
	questionsAndAnswers := []QuestionAnswer{}
	for _, question := range questions {
		for _, answer := range studentAnswers {
			if question.Id == answer.QuestionID {
				questionsAndAnswers = append(questionsAndAnswers, QuestionAnswer{
					Id:            answer.ID,
					Question:      question.Question,
					CorrectAnswer: question.Answer,
					StudentAnswer: answer.StudentAnswer,
				})
			}
		}
	}

	// Marshal questions and answers
	dataJSON, err := json.Marshal(questionsAndAnswers)
	if err != nil {
		return nil, fmt.Errorf("failed when marshal questions and answers: %w", err)
	}

	scoringAPIURL := service.Config.ScoringAPIURL
	if scoringAPIURL == "" {
		scoringAPIURL = "http://localhost:5000/score"
		slog.Warn("SCORING_API_URL tidak diset, menggunakan default fallback: " + scoringAPIURL)
	}

	// Request body
	requestBody := bytes.NewBuffer(dataJSON)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", scoringAPIURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request to scoring API URL: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", service.Config.ScoringAPIKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send the request to scoring API URL: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scoring API return error status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Unmarshal response
	var essayCorrections []domain.EssayCorrection
	err = json.Unmarshal(responseBody, &essayCorrections)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w. response: %s", err, string(responseBody))
	}

	// Update student answers
	for _, essayCorrection := range essayCorrections {
		err := service.StudentRepository.UpdateAnswerById(ctx, tx, essayCorrection.StudentAnswerId, essayCorrection.Score, essayCorrection.Feedback, essayCorrection.MaxScore, essayCorrection.Similarity)

		if err != nil {
			return nil, fmt.Errorf("failed when calling UpdateAnswerById repository: %w", err)
		}
	}

	// Update score in exam_attempts
	err = service.StudentRepository.UpdateScoresByAttemptId(ctx, tx, attemptId, essayCorrections)
	if err != nil {
		return nil, fmt.Errorf("failed when calling UpdateScoresByAttemptId repository: %w", err)
	}

	return essayCorrections, nil
}

func (service *StudentServiceImpl) GetExamAttemptsByExamIdAndStudentId(ctx context.Context, userId string, examId string) ([]web.ExamAttempt, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	attempts, err := service.StudentRepository.FindAttemptsByExamIdAndStudentId(ctx, tx, userId, examId)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindAttemptsByExamIdAndStudentId repository: %w", err)
	}

	return attempts, nil
}

func (service *StudentServiceImpl) GetBiggestExamAttemptsByStudentId(ctx context.Context, userId string) ([]web.ExamAttemptsCustom, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	attempts, err := service.StudentRepository.FindBiggestAttemptsByStudentId(ctx, tx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindBiggestAttemptsByStudentId repository: %w", err)
	}

	return attempts, nil
}

func (service *StudentServiceImpl) GetExamsWithScoreAndTeacherNameByExamId(ctx context.Context, examAttempts []web.ExamAttemptsCustom) ([]web.ExamWithScoreAndTeacherName, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	scores, err := service.StudentRepository.FindExamsWithScoreAndTeacherNameByExamId(ctx, tx, examAttempts)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindExamsWithScoreAndTeacherNameByExamId repository: %w", err)
	}

	return scores, nil
}

func (service *StudentServiceImpl) GetBiggestScoreByStudentIdAndExamId(ctx context.Context, userId string, examId string) (string, int, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	examAttempId, score, err := service.StudentRepository.FindBiggestScoreByStudentIdAndExamId(ctx, tx, userId, examId)
	if err != nil {
		return "", 0, fmt.Errorf("failed when calling FindBiggestScoreByStudentIdAndExamId repository: %w", err)
	}

	return examAttempId, score, nil
}

func (service *StudentServiceImpl) GetStudentAnswersByExamAttemptId(ctx context.Context, attemptId string) ([]web.StudentAnswer, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	answer, err := service.StudentRepository.FindStudentAnswersByAttemptId(ctx, tx, attemptId)
	if err != nil {
		return nil, fmt.Errorf("failed when calling FindStudentAnswersByAttemptId repository: %w", err)
	}

	return answer, nil
}

func (service *StudentServiceImpl) FindQuestionById(ctx context.Context, questionId string) (web.QuestionAndRightAnswer, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		return web.QuestionAndRightAnswer{}, fmt.Errorf("failed to open db transaction: %w", err)
	}
	defer helper.CommitOrRollback(ctx, tx)

	questionAndRightAnswer, err := service.StudentRepository.FindQuestionById(ctx, tx, questionId)
	if err != nil {
		return web.QuestionAndRightAnswer{}, fmt.Errorf("failed when calling FindQuestionById repository: %w", err)
	}

	return questionAndRightAnswer, nil
}
