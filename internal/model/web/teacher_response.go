package web

import "github.com/mhaatha/go-template-saygenfix/internal/model/domain"

type TeacherDashboardResponse struct {
	User  domain.User
	Exams []domain.Exam
}
