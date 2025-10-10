package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/generative-ai-go/genai"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mhaatha/go-template-saygenfix/internal/config"
	"github.com/mhaatha/go-template-saygenfix/internal/helper"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
	"github.com/mhaatha/go-template-saygenfix/internal/repository"
	"google.golang.org/api/option"
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

func (service *StudentServiceImpl) CreateExamAttempt(ctx context.Context, studentId, examId string) (string, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return "", err
	}
	defer helper.CommitOrRollback(ctx, tx)

	examAttemptId, err := service.StudentRepository.CreateExamAttempt(ctx, tx, studentId, examId)
	if err != nil {
		return "", err
	}

	return examAttemptId, nil
}

func (service *StudentServiceImpl) SaveAnswer(ctx context.Context, answer web.StudentAnswer) error {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return err
	}
	defer helper.CommitOrRollback(ctx, tx)

	err = service.StudentRepository.SaveAnswer(ctx, tx, answer)
	if err != nil {
		return err
	}

	return nil
}

func (service *StudentServiceImpl) CompleteExamAttempt(ctx context.Context, attemptId string) error {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return err
	}
	defer helper.CommitOrRollback(ctx, tx)

	err = service.StudentRepository.CompleteExamAttempt(ctx, tx, attemptId)
	if err != nil {
		return err
	}

	return nil
}

func (service *StudentServiceImpl) GetExamByAttempId(ctx context.Context, attemptId string) (domain.Exam, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return domain.Exam{}, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	exam, err := service.StudentRepository.FindExamByAttemptId(ctx, tx, attemptId)
	if err != nil {
		return domain.Exam{}, err
	}

	return exam, nil
}

func (service *StudentServiceImpl) GetAnswersByAttemptId(ctx context.Context, attemptId string) ([]web.StudentAnswer, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return nil, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	answers, err := service.StudentRepository.FindAnswersByAttemptId(ctx, tx, attemptId)
	if err != nil {
		return nil, err
	}

	return answers, nil
}

func (service *StudentServiceImpl) CalculateScore(ctx context.Context, attemptId string) ([]domain.EssayCorrection, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return nil, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	// Ambil exam by attempId
	exam, err := service.StudentRepository.FindExamByAttemptId(ctx, tx, attemptId)
	if err != nil {
		return nil, err
	}

	// Ambil questions by examId
	questions, err := service.StudentRepository.FindQuestionsByExamId(ctx, tx, exam.Id)
	if err != nil {
		return nil, err
	}

	type QuestionAnswer struct {
		Id            string
		Question      string
		CorrectAnswer string
		StudentAnswer string
	}

	studentAnswers, err := service.StudentRepository.FindAnswersByAttemptId(ctx, tx, attemptId)
	if err != nil {
		return nil, err
	}

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

	dataJSON, err := json.Marshal(questionsAndAnswers)
	if err != nil {
		return nil, err
	}

	// Handle Gemini API
	client, err := genai.NewClient(ctx, option.WithAPIKey(service.Config.GeminiAPIKey))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")

	promptTemplate := `Anda adalah sebuah API penilaian otomatis yang sangat akurat. Tugas Anda adalah memproses sebuah JSON array yang berisi data ujian. Untuk setiap objek dalam array input, Anda harus memberikan skor dan feedback. Untuk perhitungannya yaitu:
(100 / total-soal) * jumlah-benar. Satu jawaban benar misalnya itu nilainya 20, dan total soal itu ada 5, maka jika ada yang benar semua nilainya adalah 100. Jika soal essay nomor 4 nilainya cukup maka dia bisa dianggap nilainya 14. Maka total nilainya adalah 94. Gunakan metode sentence-BERT untuk membandingkan CorrectAnswer dan StudentAnswer lalu berikan nilai sesuai dengan ketentuan.

Berikut adalah daftar soal dan jawaban dalam format JSON Array:
%s

Instruksi Output:
Respons Anda HARUS berupa string JSON valid tanpa tambahan teks, komentar, atau markdown. Respons harus berupa JSON Array, di mana setiap objek cocok dengan satu objek input dan memiliki struktur: {"student_answer_id": "<Id>", "question": "<Question>", "student_answer": "<StudentAnswer>", "score": <nilai_angka>, "feedback": "<'Sangat Sesuai'|'Sesuai'|'Cukup'|'Tidak Sesuai'|'Sangat Tidak Sesuai'>", "max_score": "<nilai maksimal per soal>", "similarity": "<nilai similarity 0-1>"}. Jangan tambahkan format markdown atau teks lain di luar JSON tersebut. Gunakan plaintext tanpa format markdown dalam tiap value question dan answer.`

	prompt := fmt.Sprintf(
		promptTemplate,
		string(dataJSON),
	)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	rawResponse := ""
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			rawResponse += string(part.(genai.Text))
		}
	}

	// Clean the response
	cleanResponse := strings.TrimSpace(rawResponse)
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")

	var essayCorrections []domain.EssayCorrection
	err = json.Unmarshal([]byte(cleanResponse), &essayCorrections)
	if err != nil {
		slog.Error("error when unmarshaling the cleanResponse", "err", err)
		return nil, err
	}

	// Update student answers
	for _, essayCorrection := range essayCorrections {
		err := service.StudentRepository.UpdateAnswerById(ctx, tx, essayCorrection.StudentAnswerId, essayCorrection.Score, essayCorrection.Feedback, essayCorrection.MaxScore, essayCorrection.Similarity)

		if err != nil {
			slog.Error("error when updating student answer", "err", err)
			return nil, err
		}
	}

	// Update score in exam_attempts
	err = service.StudentRepository.UpdateScoresByAttemptId(ctx, tx, attemptId, essayCorrections)
	if err != nil {
		slog.Error("error when updating scores", "err", err)
		return nil, err
	}

	return essayCorrections, nil
}

func (service *StudentServiceImpl) GetExamAttemptsByExamIdAndStudentId(ctx context.Context, userId string, examId string) ([]web.ExamAttempt, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return nil, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	attempts, err := service.StudentRepository.FindAttemptsByExamIdAndStudentId(ctx, tx, userId, examId)
	if err != nil {
		return nil, err
	}

	return attempts, nil
}

func (service *StudentServiceImpl) GetBiggestExamAttemptsByStudentId(ctx context.Context, userId string) ([]web.ExamAttemptsCustom, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return nil, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	attempts, err := service.StudentRepository.FindBiggestAttemptsByStudentId(ctx, tx, userId)
	if err != nil {
		slog.Error("error when getting biggest attempts", "err", err)
		return nil, err
	}

	return attempts, nil
}

func (service *StudentServiceImpl) GetExamsWithScoreAndTeacherNameByExamId(ctx context.Context, examAttempts []web.ExamAttemptsCustom) ([]web.ExamWithScoreAndTeacherName, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return nil, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	scores, err := service.StudentRepository.FindExamsWithScoreAndTeacherNameByExamId(ctx, tx, examAttempts)
	if err != nil {
		return nil, err
	}

	return scores, nil
}

func (service *StudentServiceImpl) GetBiggestScoreByStudentIdAndExamId(ctx context.Context, userId string, examId string) (string, int, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return "", 0, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	examAttempId, score, err := service.StudentRepository.FindBiggestScoreByStudentIdAndExamId(ctx, tx, userId, examId)
	if err != nil {
		return "", 0, err
	}

	return examAttempId, score, nil
}

func (service *StudentServiceImpl) GetStudentAnswersByExamAttemptId(ctx context.Context, attemptId string) ([]web.StudentAnswer, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return nil, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	answer, err := service.StudentRepository.FindStudentAnswersByAttemptId(ctx, tx, attemptId)
	if err != nil {
		return nil, err
	}

	return answer, nil
}

func (service *StudentServiceImpl) FindQuestionById(ctx context.Context, questionId string) (web.QuestionAndRightAnswer, error) {
	// Open transaction
	tx, err := service.DB.Begin(ctx)
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
		return web.QuestionAndRightAnswer{}, err
	}
	defer helper.CommitOrRollback(ctx, tx)

	questionAndRightAnswer, err := service.StudentRepository.FindQuestionById(ctx, tx, questionId)
	if err != nil {
		return web.QuestionAndRightAnswer{}, err
	}

	return questionAndRightAnswer, nil
}
