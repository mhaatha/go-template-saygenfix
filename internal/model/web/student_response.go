package web

import "github.com/mhaatha/go-template-saygenfix/internal/model/domain"

type StudentDashboardResponse struct {
	User     domain.User
	Exams    []domain.Exam
	Teachers map[string]domain.User
}
