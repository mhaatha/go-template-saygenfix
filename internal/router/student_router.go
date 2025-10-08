package router

import (
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/handler"
)

func StudentRouter(handler handler.StudentHandler, mux *http.ServeMux) {
	// Dashboard
	mux.HandleFunc("GET /student/dashboard", handler.DashboardView)

	// Rute utama untuk memulai ujian (hanya untuk load awal)
	mux.HandleFunc("GET /student/take-exam/{examId}", handler.TakeExamView)

	// === PERUBAHAN DI SINI ===
	// Rute ini khusus untuk request HTMX saat berganti soal.
	// Path-nya diubah dari "/student/take-exam/{examId}/{qNum}" menjadi "/student/question/{examId}/{qNum}"
	// agar cocok dengan `hx-get` di template.
	mux.HandleFunc("GET /student/question/{examId}/{qNum}", handler.HandleQuestionPartial)

	// === PERUBAHAN DI SINI ===
	// Rute untuk submit keseluruhan jawaban ujian diaktifkan.
	mux.HandleFunc("POST /student/submit-exam/{examId}", handler.SubmitExam)

	// Rute untuk melihat hasil ujian setelah submit
	mux.HandleFunc("GET /student/exam-result/{examId}", handler.CorrectExamView)

	// Rute untuk melihat daftar semua hasil ujian
	mux.HandleFunc("GET /student/exam-result", handler.ExamResultView)
}
