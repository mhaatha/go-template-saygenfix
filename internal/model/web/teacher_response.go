package web

import "github.com/mhaatha/go-template-saygenfix/internal/model/domain"

type TeacherDashboardResponse struct {
	User  domain.User
	Exams []domain.Exam
}

type ExamAttemptsWithStudentName struct {
	Id          string
	StudentName string
	Score       int
}

type TeacherCheckExamResponse struct {
	User         domain.User
	Exam         domain.Exam
	FlashMessage string
	ExamAttempts []ExamAttemptsWithStudentName
}

type TeacherEditExamResponse struct {
	User               domain.User
	Exam               domain.Exam
	QuestionAndAnswers []domain.QAItem
}
