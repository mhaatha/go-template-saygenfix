package router

import (
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/handler"
)

func StudentRouter(handler handler.StudentHandler, mux *http.ServeMux) {
	mux.HandleFunc("GET /room-ujian-student", handler.RoomUjianView)
	mux.HandleFunc("GET /take-exam/{examId}", handler.TakeExamView)
	mux.HandleFunc("GET /take-exam/{examId}/question/{qNum}", handler.HandleQuestionPartial)
	mux.HandleFunc("POST /submit-exam/{examId}", handler.CorrectExam)
	mux.HandleFunc("GET /submit-exam/{examId}", handler.CorrectExamView)
}
