package router

import (
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/handler"
)

func StudentRouter(handler handler.StudentHandler, mux *http.ServeMux) {
	// Dashboard
	mux.HandleFunc("GET /student/dashboard", handler.DashboardView)

	mux.HandleFunc("GET /student/take-exam/{examId}", handler.TakeExamView)

	// mux.HandleFunc("GET /student/take-exam/{examId}/question/{qNum}", handler.HandleQuestionPartial)

	// mux.HandleFunc("POST /student/submit-exam/{examId}", handler.CorrectExam)

	mux.HandleFunc("GET /student/exam-result/{examId}", handler.CorrectExamView)

	mux.HandleFunc("GET /student/exam-result", handler.ExamResultView)

	// mux.HandleFunc("GET /student/exam-result/{examId}", handler.ResultExamView)
}
