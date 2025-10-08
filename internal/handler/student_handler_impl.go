package handler

import (
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mhaatha/go-template-saygenfix/internal/middleware"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
	"github.com/mhaatha/go-template-saygenfix/internal/service"
)

func NewStudentHandler(studentService service.StudentService) StudentHandler {
	return &StudentHandlerImpl{
		Template: template.Must(template.ParseFiles(
			"../../internal/templates/views/student/dashboard.html",
			"../../internal/templates/views/student/take_exam.html",
			"../../internal/templates/views/partial/question_partial.html",
			"../../internal/templates/views/student/exam_result.html",
			"../../internal/templates/views/partial/student_dashboard_navbar.html",
			"../../internal/templates/views/student/score_list.html",
			"../../internal/templates/views/partial/student_exam_result_navbar.html",
		)),
		StudentService: studentService,
	}
}

type StudentHandlerImpl struct {
	Template       *template.Template
	StudentService service.StudentService
}

// Question merepresentasikan satu soal
type Question struct {
	Number        int
	Text          string
	StudentAnswer string // Untuk menyimpan jawaban sementara
}

// ExamPageData adalah data yang dikirim ke template
type ExamPageData struct {
	ExamID                string
	ExamTitle             string
	Questions             []Question
	CurrentQuestion       Question
	CurrentQuestionNumber int
	TotalQuestions        int
	NextQuestionNumber    int
	PrevQuestionNumber    int
}

// --- DATA DUMMY (Ganti dengan logika database Anda) ---
var exams = map[string][]Question{
	"123": {
		{Number: 1, Text: "Apa itu Cloud Computing?"},
		{Number: 2, Text: "Jelaskan konsep PaaS."},
		{Number: 3, Text: "Sebutkan 3 penyedia layanan cloud utama."},
		{Number: 4, Text: "Apa perbedaan IaaS dan SaaS?"},
		{Number: 5, Text: "Apa itu arsitektur serverless?"},
	},
}

func (handler *StudentHandlerImpl) DashboardView(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)
	if user.Role == "student" {
		user.Role = "Student"
	}

	exams, err := handler.StudentService.GetActiveExams(r.Context())
	if err != nil {
		log.Fatal(err)
	}

	// Gunakan map untuk memastikan setiap guru hanya diambil datanya sekali
	teachersMap := make(map[string]domain.User)

	for _, exam := range exams {
		// Hanya ambil data guru jika belum ada di dalam map
		if _, found := teachersMap[exam.TeacherId]; !found {
			teacher, err := handler.StudentService.GetTeacherById(r.Context(), exam.TeacherId)
			if err != nil {
				log.Fatal(err)
			}
			teachersMap[exam.TeacherId] = teacher
		}
	}

	dashboardData := web.StudentDashboardResponse{
		User:     user,
		Exams:    exams,
		Teachers: teachersMap,
	}

	if err := handler.Template.ExecuteTemplate(w, "student-dashboard", dashboardData); err != nil {
		log.Fatal(err)
	}
}

func (handler *StudentHandlerImpl) TakeExamView(w http.ResponseWriter, r *http.Request) {
	examId := r.PathValue("examId")

	handler.serveQuestion(w, r, examId, 1)
}

// Handler untuk HTMX (memuat soal tertentu)
func (handler *StudentHandlerImpl) HandleQuestionPartial(w http.ResponseWriter, r *http.Request) {
	examID := r.PathValue("examId")
	qNumStr := r.PathValue("qNum")

	qNum, err := strconv.Atoi(qNumStr)
	if err != nil {
		http.Error(w, "Nomor soal tidak valid", http.StatusBadRequest)
		return
	}

	handler.serveQuestion(w, r, examID, qNum)
}

// Fungsi helper untuk menyiapkan data dan merender template
func (handler *StudentHandlerImpl) serveQuestion(w http.ResponseWriter, r *http.Request, examID string, qNum int) {
	questionList, ok := exams[examID]
	if !ok {
		http.Error(w, "Ujian tidak ditemukan", http.StatusNotFound)
		return
	}

	if qNum < 1 || qNum > len(questionList) {
		http.Error(w, "Soal tidak ditemukan", http.StatusNotFound)
		return
	}

	data := ExamPageData{
		ExamID:                examID,
		ExamTitle:             "UTS Semester 6", // Ambil dari database
		Questions:             questionList,
		CurrentQuestion:       questionList[qNum-1],
		CurrentQuestionNumber: qNum,
		TotalQuestions:        len(questionList),
		NextQuestionNumber:    qNum + 1,
		PrevQuestionNumber:    qNum - 1,
	}

	// Cek apakah ini request dari HTMX atau bukan
	// Jika BUKAN, render seluruh halaman. Jika YA, render hanya bagian soal.
	if r.Header.Get("HX-Request") == "true" {
		err := handler.Template.ExecuteTemplate(w, "question-partial", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		err := handler.Template.ExecuteTemplate(w, "student-take-exam", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// AnswerResult merepresentasikan hasil dari satu jawaban
type AnswerResult struct {
	QuestionNumber int
	QuestionText   string
	CorrectAnswer  string
	StudentAnswer  string
	Status         string // "Sesuai" atau "Cukup"
	Score          int
	MaxScore       int
}

// ExamResultData adalah data utama yang dikirim ke template
type ExamResultData struct {
	TotalScore    int
	FeedbackText  string
	FeedbackColor string
	ScoreColor    string // Warna awal gradien
	ScoreColorEnd string // Warna akhir gradien
	ScoreOffset   float64
	Answers       []AnswerResult
}

func (handler *StudentHandlerImpl) CorrectExamView(w http.ResponseWriter, r *http.Request) {
	// --- PEMBUATAN DATA DUMMY ---

	// Buat daftar hasil jawaban
	answers := []AnswerResult{
		{
			QuestionNumber: 1,
			QuestionText:   "Apa itu Cloud Computing ?",
			CorrectAnswer:  "Cloud computing adalah pengiriman sumber daya komputasi seperti server, penyimpanan, database, dan perangkat lunak melalui internet, yang memungkinkan pengguna untuk mengakses layanan ini sesuai permintaan dan hanya membayar apa yang mereka gunakan",
			StudentAnswer:  "Kita bisa simpan dan akses data secara online",
			Status:         "Sesuai",
			Score:          20,
			MaxScore:       20,
		},
		{
			QuestionNumber: 2,
			QuestionText:   "Apa itu Cloud Computing ?",
			CorrectAnswer:  "Cloud computing adalah pengiriman sumber daya komputasi seperti server, penyimpanan, database, dan perangkat lunak melalui internet...",
			StudentAnswer:  "Kita bisa simpan dan akses data secara online",
			Status:         "Sesuai",
			Score:          20,
			MaxScore:       20,
		},
		{
			QuestionNumber: 3,
			QuestionText:   "Apa itu Cloud Computing ?",
			CorrectAnswer:  "Cloud computing adalah pengiriman sumber daya komputasi seperti server, penyimpanan, database, dan perangkat lunak melalui internet...",
			StudentAnswer:  "Kita bisa simpan dan akses data secara online",
			Status:         "Cukup",
			Score:          10,
			MaxScore:       20,
		},
		{
			QuestionNumber: 4,
			QuestionText:   "Apa itu Cloud Computing ?",
			CorrectAnswer:  "Cloud computing adalah pengiriman sumber daya komputasi seperti server, penyimpanan, database, dan perangkat lunak melalui internet...",
			StudentAnswer:  "Kita bisa simpan dan akses data secara online",
			Status:         "Sesuai",
			Score:          20,
			MaxScore:       20,
		},
		{
			QuestionNumber: 5,
			QuestionText:   "Apa itu Cloud Computing ?",
			CorrectAnswer:  "Cloud computing adalah pengiriman sumber daya komputasi seperti server, penyimpanan, database, dan perangkat lunak melalui internet...",
			StudentAnswer:  "Kita bisa simpan dan akses data secara online",
			Status:         "Sesuai",
			Score:          20,
			MaxScore:       20,
		},
	}

	// Hitung total skor
	totalScore := 0
	for _, a := range answers {
		totalScore += a.Score
	}

	// Kalkulasi untuk lingkaran skor SVG
	circumference := 2 * 3.14159 * 42 // 2 * pi * radius
	scorePercentage := float64(totalScore) / 100.0
	scoreOffset := circumference * (1 - scorePercentage)

	// Siapkan data lengkap untuk dikirim ke template
	data := ExamResultData{
		TotalScore:    totalScore,
		FeedbackText:  "Bagus !",
		FeedbackColor: "#00FF90", // Kode warna hijau
		ScoreColor:    "#04FDFF", // Biru muda
		ScoreColorEnd: "#393FEF", // Biru tua
		ScoreOffset:   scoreOffset,
		Answers:       answers,
	}

	err := handler.Template.ExecuteTemplate(w, "student-exam-result", data)
	if err != nil {
		slog.Error(err.Error())
	}
}

func (handler *StudentHandlerImpl) CorrectExam(w http.ResponseWriter, r *http.Request) {
	// Panggil API untuk koreksi exam

	// Redirect HTMX to /submit-exam/123
	w.Header().Set("HX-Redirect", "/submit-exam/123")
	w.WriteHeader(http.StatusOK)
}

func (handler *StudentHandlerImpl) ExamResultView(w http.ResponseWriter, r *http.Request) {
	handler.Template.ExecuteTemplate(w, "student-score-list", nil)
}
