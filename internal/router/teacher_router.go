package router

import (
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/handler"
)

func TeacherRouter(handler handler.TeacherHandler, mux *http.ServeMux) {
	mux.HandleFunc("GET /teacher/dashboard", handler.TeacherDashboard)

	mux.HandleFunc("GET /teacher/upload", handler.UploadView)

	mux.HandleFunc("POST /teacher/generate-and-create-exam-room", handler.GenerateAndCreateExamRoom)

	mux.HandleFunc("GET /teacher/check-exam/{id}", handler.CheckExamView)

	mux.HandleFunc("GET /teacher/edit-exam/{id}", handler.EditExamView)

	mux.HandleFunc("GET /teacher/exam-result/{id}", handler.ExamResultView)

	// mux.HandleFunc("GET /teacher/generate-result", handler.GenerateResultView)
}
