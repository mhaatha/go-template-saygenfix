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

	// Rute untuk melihat soal ujian (sekarang menerima GET dan POST)
	mux.HandleFunc("GET /student/question/{examId}/{qNum}", handler.HandleQuestionPartial)
	// --- PERUBAHAN DI SINI ---
	// Menambahkan rute POST agar bisa menerima data jawaban saat navigasi
	mux.HandleFunc("POST /student/question/{examId}/{qNum}", handler.HandleQuestionPartial)

	// Submit exam
	mux.HandleFunc("POST /student/submit-exam/{examId}", handler.SubmitExam)

	// Rute untuk melihat hasil ujian setelah submit
	mux.HandleFunc("GET /student/exam-result/{examId}", handler.CorrectExamView)

	// Rute untuk melihat daftar semua hasil ujian
	mux.HandleFunc("GET /student/exam-result", handler.ExamResultView)
}
