package web

import (
	"time"

	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

type StudentDashboardResponse struct {
	User     domain.User
	Exams    []domain.Exam
	Teachers map[string]domain.User
}

type ExamPageData struct {
	ExamID                string
	AttemptID             string
	ExamTitle             string
	Questions             []domain.QAItem
	CurrentQuestion       domain.QAItem
	CurrentQuestionNumber int
	TotalQuestions        int
	NextQuestionNumber    int
	PrevQuestionNumber    int
	SavedAnswer           map[string]string
}

type Question struct {
	Number        int
	Text          string
	StudentAnswer string // Untuk menyimpan jawaban sementara
}

type StudentAnswer struct {
	ID            string `json:"id"`
	ExamAttemptID string `json:"exam_attempt_id"`
	QuestionID    string `json:"question_id"`
	StudentAnswer string `json:"student_answer"`
	Score         int    `json:"score"`
	Feedback      string `json:"feedback"`
}

type ExamAttempt struct {
	ID          string    `json:"id"`
	StudentID   string    `json:"student_id"`
	ExamID      string    `json:"exam_id"`
	Score       int       `json:"score"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

type ExamResultData struct {
	TotalScore          int
	Corrections         []domain.EssayCorrection
	MaxScorePerQuestion int
}
