package handler

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mhaatha/go-template-saygenfix/internal/middleware"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
	"github.com/mhaatha/go-template-saygenfix/internal/service"
)

func NewStudentHandler(studentService service.StudentService) StudentHandler {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	return &StudentHandlerImpl{
		Template: template.Must(
			template.New("base").Funcs(funcMap).ParseFiles(
				"../../internal/templates/views/student/dashboard.html",
				"../../internal/templates/views/student/take_exam.html",
				"../../internal/templates/views/partial/question_partial.html",
				"../../internal/templates/views/partial/question_form.html",
				"../../internal/templates/views/student/exam_result.html",
				"../../internal/templates/views/partial/student_dashboard_navbar.html",
				"../../internal/templates/views/student/score_list.html",
				"../../internal/templates/views/partial/student_exam_result_navbar.html",
			),
		),
		StudentService: studentService,
	}
}

type StudentHandlerImpl struct {
	Template       *template.Template
	StudentService service.StudentService
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

	teachersMap := make(map[string]domain.User)
	for _, exam := range exams {
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

// TakeExamView mempersiapkan ujian, membuat attempt, dan menampilkan soal pertama.
func (handler *StudentHandlerImpl) TakeExamView(w http.ResponseWriter, r *http.Request) {
	examId := r.PathValue("examId")
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)

	// Membuat attemptID di awal untuk digunakan saat submit nanti.
	attemptID, err := handler.StudentService.CreateExamAttempt(r.Context(), user.Id, examId)
	if err != nil {
		log.Printf("Error creating exam attempt: %v", err)
		http.Error(w, "Gagal memulai sesi ujian", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "exam_attempt_id",
		Value:    attemptID,
		Path:     "/",
		Expires:  time.Now().Add(3 * time.Hour),
		HttpOnly: true,
	})

	// Memulai ujian dengan map jawaban yang masih kosong.
	initialAnswers := make(map[string]string)
	handler.serveQuestion(w, r, examId, 1, attemptID, initialAnswers)
}

// HandleQuestionPartial TIDAK menyimpan ke DB. Ia hanya mengelola state jawaban
// dengan cara mengambil semua jawaban dari form dan meneruskannya kembali saat merender soal berikutnya.
func (handler *StudentHandlerImpl) HandleQuestionPartial(w http.ResponseWriter, r *http.Request) {
	// Dapatkan attemptID dari cookie untuk diteruskan ke serveQuestion.
	cookie, err := r.Cookie("exam_attempt_id")
	if err != nil {
		http.Error(w, "Sesi ujian tidak valid atau telah berakhir", http.StatusUnauthorized)
		return
	}
	attemptID := cookie.Value

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Ambil semua jawaban yang ada di form (baik dari textarea maupun hidden inputs).
	studentAnswers := make(map[string]string)
	for key, values := range r.Form {
		if len(values) > 0 && strings.HasPrefix(key, "answers[") {
			// Ekstrak ID dari nama field "answers[the-question-id]"
			id := strings.TrimSuffix(strings.TrimPrefix(key, "answers["), "]")
			studentAnswers[id] = values[0]
		}
	}

	examID := r.PathValue("examId")
	qNumStr := r.PathValue("qNum")
	qNum, _ := strconv.Atoi(qNumStr)

	// Teruskan map jawaban yang didapat dari form, BUKAN dari DB.
	handler.serveQuestion(w, r, examID, qNum, attemptID, studentAnswers)
}

// serveQuestion adalah fungsi presenter yang bertanggung jawab untuk merender halaman ujian.
// Fungsi ini sekarang menerima state jawaban langsung dari handler yang memanggilnya.
func (handler *StudentHandlerImpl) serveQuestion(w http.ResponseWriter, r *http.Request, examID string, qNum int, attemptID string, savedAnswers map[string]string) {
	exam, err := handler.StudentService.GetExamById(r.Context(), examID)
	if err != nil {
		log.Printf("Error getting exam: %v", err)
		http.Error(w, "Ujian tidak ditemukan", http.StatusNotFound)
		return
	}

	questionList, err := handler.StudentService.GetQuestionsByExamId(r.Context(), examID)
	if err != nil {
		log.Printf("Error getting questions: %v", err)
		http.Error(w, "Gagal memuat soal", http.StatusInternalServerError)
		return
	}

	if len(questionList) == 0 {
		http.Error(w, "Ujian ini belum memiliki soal.", http.StatusNotFound)
		return
	}
	if qNum < 1 || qNum > len(questionList) {
		http.Error(w, "Soal tidak ditemukan", http.StatusNotFound)
		return
	}

	// Siapkan data untuk di-pass ke template.
	data := web.ExamPageData{
		ExamID:                examID,
		AttemptID:             attemptID, // Teruskan attemptID ke template
		ExamTitle:             exam.RoomName,
		Questions:             questionList,
		CurrentQuestion:       questionList[qNum-1],
		CurrentQuestionNumber: qNum,
		TotalQuestions:        len(questionList),
		NextQuestionNumber:    qNum + 1,
		PrevQuestionNumber:    qNum - 1,
		SavedAnswer:           savedAnswers, // Gunakan map jawaban yang sudah di-pass
	}

	templateName := "student-take-exam"
	if r.Header.Get("HX-Request") == "true" {
		templateName = "question-partial"
	}

	err = handler.Template.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Printf("Error executing template %s: %v", templateName, err)
		http.Error(w, "Terjadi kesalahan saat merender halaman", http.StatusInternalServerError)
	}
}

// SubmitExam adalah satu-satunya fungsi yang menyimpan semua jawaban ke database.
func (handler *StudentHandlerImpl) SubmitExam(w http.ResponseWriter, r *http.Request) {
	examId := r.PathValue("examId")
	cookie, err := r.Cookie("exam_attempt_id")
	if err != nil {
		http.Error(w, "Sesi ujian tidak valid atau telah berakhir", http.StatusUnauthorized)
		return
	}
	attemptID := cookie.Value

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// 1. Ambil semua jawaban dari form untuk terakhir kalinya.
	studentAnswers := make(map[string]string)
	for key, values := range r.Form {
		if len(values) > 0 && strings.HasPrefix(key, "answers[") {
			id := strings.TrimSuffix(strings.TrimPrefix(key, "answers["), "]")
			studentAnswers[id] = values[0]
		}
	}

	// 2. Simpan semua jawaban dari map ke database dalam satu perulangan.
	for questionID, studentAnswer := range studentAnswers {
		// Opsional: hanya simpan jawaban yang tidak kosong.
		if studentAnswer != "" {
			answer := web.StudentAnswer{
				ExamAttemptID: attemptID,
				QuestionID:    questionID,
				StudentAnswer: studentAnswer,
			}
			if err := handler.StudentService.SaveAnswer(r.Context(), answer); err != nil {
				log.Printf("Gagal menyimpan jawaban untuk soal %s: %v", questionID, err)
				http.Error(w, "Gagal menyimpan semua jawaban.", http.StatusInternalServerError)
				return
			}
		}
	}

	// 3. Panggil service untuk menghitung skor dan menyelesaikan ujian.
	// (Asumsi: service ini menangani kalkulasi dan update status attempt)
	_, err = handler.StudentService.CalculateScore(r.Context(), attemptID)
	if err != nil {
		log.Printf("Error calculating score: %v", err)
		http.Error(w, "Gagal menghitung skor ujian", http.StatusInternalServerError)
		return
	}

	// 4. Hapus cookie sesi ujian karena sudah selesai.
	http.SetCookie(w, &http.Cookie{
		Name:     "exam_attempt_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Cara standar untuk menghapus cookie.
		HttpOnly: true,
	})

	// 5. Arahkan pengguna ke halaman hasil.
	resultURL := fmt.Sprintf("/student/exam-result/%s", examId)
	w.Header().Set("HX-Redirect", resultURL)
	w.WriteHeader(http.StatusOK)
}

type AnswerResult struct {
	QuestionNumber int
	QuestionText   string
	CorrectAnswer  string
	StudentAnswer  string
	Status         string
	Score          int
	MaxScore       int
}

type ExamResultData struct {
	TotalScore    int
	FeedbackText  string
	FeedbackColor string
	ScoreColor    string
	ScoreColorEnd string
	ScoreOffset   float64
	Answers       []AnswerResult
}

func (handler *StudentHandlerImpl) CorrectExam(w http.ResponseWriter, r *http.Request) {}

func (handler *StudentHandlerImpl) CorrectExamView(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)
	if user.Role == "student" {
		user.Role = "Student"
	}
	examId := r.PathValue("examId")

	// Get exam_attempts by examId and studentId
	examAttempts, err := handler.StudentService.GetExamAttemptsByExamIdAndStudentId(r.Context(), user.Id, examId)
	if err != nil {
		log.Printf("Error getting student answers: %v", err)
		http.Error(w, "Gagal mendapatkan jawaban siswa", http.StatusInternalServerError)
		return
	}

	// Get student answers by exam_attemptsId
	var examAttemptsId string
	if len(examAttempts) > 0 {
		examAttemptsId = examAttempts[0].ID
	}
	_, err = handler.StudentService.GetAnswersByAttemptId(r.Context(), examAttemptsId)
	if err != nil {
		log.Printf("Error getting student answers: %v", err)
		http.Error(w, "Gagal mendapatkan jawaban siswa", http.StatusInternalServerError)
		return
	}

	cookie, _ := r.Cookie("exam_attempt_id")
	if cookie != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "exam_attempt_id",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})
	}

	// Pass data to template
	AA := struct {
		User       domain.User
		TotalScore int
	}{
		User:       user,
		TotalScore: 90,
	}

	err = handler.Template.ExecuteTemplate(w, "student-exam-result", AA)
	if err != nil {
		slog.Error(err.Error())
	}
}

func (handler *StudentHandlerImpl) ExamResultView(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)
	if user.Role == "student" {
		user.Role = "Student"
	}

	examAttemptsCustom, err := handler.StudentService.GetBiggestExamAttemptsByStudentId(r.Context(), user.Id)
	if err != nil {
		log.Printf("Error when calling GetBiggestExamAttemptsByStudentId: %v", err)
		http.Error(w, "Gagal mendapatkan jawaban siswa", http.StatusInternalServerError)
		return
	}

	examsWithScoreAndTeacherName, err := handler.StudentService.GetExamsWithScoreAndTeacherNameByExamId(r.Context(), examAttemptsCustom)
	if err != nil {
		log.Printf("Error when calling GetExamsWithScoreAndTeacherNameByExamId: %v", err)
		http.Error(w, "Gagal mendapatkan jawaban siswa", http.StatusInternalServerError)
		return
	}

	dataResponse := web.ScoreListResponse{
		Exams: examsWithScoreAndTeacherName,
		User:  user,
	}

	handler.Template.ExecuteTemplate(w, "student-score-list", dataResponse)
}
