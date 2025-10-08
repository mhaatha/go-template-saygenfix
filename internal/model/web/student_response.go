package web

import "github.com/mhaatha/go-template-saygenfix/internal/model/domain"

type StudentDashboardResponse struct {
	User     domain.User
	Exams    []domain.Exam
	Teachers map[string]domain.User
}

type ExamPageData struct {
	ExamID                string
	ExamTitle             string
	Questions             []domain.QAItem
	CurrentQuestion       domain.QAItem
	CurrentQuestionNumber int
	TotalQuestions        int
	NextQuestionNumber    int
	PrevQuestionNumber    int
}

type Question struct {
	Number        int
	Text          string
	StudentAnswer string // Untuk menyimpan jawaban sementara
}
