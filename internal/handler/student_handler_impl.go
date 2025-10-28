package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	appError "github.com/mhaatha/go-template-saygenfix/internal/errors"
	"github.com/mhaatha/go-template-saygenfix/internal/middleware"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
	"github.com/mhaatha/go-template-saygenfix/internal/service"
)

func tojson(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return template.JS("null")
	}
	return template.JS(b)
}

func NewStudentHandler(studentService service.StudentService) StudentHandler {
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"tojson": tojson,
		"div": func(a, b int) int {
			if b == 0 {
				return 0 // Hindari pembagian dengan nol
			}
			return a / b
		},
		"ge": func(a, b int) bool {
			return a >= b // ge = Greater than or Equal
		},
		"greater": func(a, b float64) bool {
			return a > b // ge = Greater than or Equal
		},
		"floatConvert": func(a int) float64 {
			return float64(a)
		},
		"getNormalize": func(a, b float64) float64 {
			hasil := a * b
			return float64(hasil)
		},
		"toInt": func(a float64) int {
			return int(a)
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
				"../../internal/templates/views/error.html",
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
		slog.Error("failed to get active exams", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	teachersMap := make(map[string]domain.User)
	for _, exam := range exams {
		if _, found := teachersMap[exam.TeacherId]; !found {
			teacher, err := handler.StudentService.GetTeacherById(r.Context(), exam.TeacherId)
			if err != nil {
				slog.Error("failed to get teacher by id", "err", err)

				if errors.Is(err, sql.ErrNoRows) {
					appError.RenderErrorPage(w, handler.Template, http.StatusNotFound, fmt.Sprintf("Teacher with id %s is not found", exam.TeacherId))
					return
				}

				appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
				return
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
		slog.Error("failed to execute student-dasboard template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

// TakeExamView mempersiapkan ujian, membuat attempt, dan menampilkan soal pertama.
func (handler *StudentHandlerImpl) TakeExamView(w http.ResponseWriter, r *http.Request) {
	examId := r.PathValue("examId")
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)

	// Membuat attemptID di awal untuk digunakan saat submit nanti.
	attemptID, err := handler.StudentService.CreateExamAttempt(r.Context(), user.Id, examId)
	if err != nil {
		slog.Error("error when calling create exam attempt service", "err", err)

		if errors.Is(err, sql.ErrNoRows) {
			appError.RenderErrorPage(w, handler.Template, http.StatusNotFound, fmt.Sprintf("Exam with id %s is not found", examId))
			return
		}

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
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
		slog.Error("failed to get exam_attempt_id cookie", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusUnauthorized, "Sesi ujian tidak valid atau telah berakhir")
		return
	}
	attemptID := cookie.Value

	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse form", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
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
	qNum, err := strconv.Atoi(qNumStr)
	if err != nil {
		slog.Error("failed to convert qNum to int", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Teruskan map jawaban yang didapat dari form, BUKAN dari DB.
	handler.serveQuestion(w, r, examID, qNum, attemptID, studentAnswers)
}

// serveQuestion adalah fungsi presenter yang bertanggung jawab untuk merender halaman ujian.
// Fungsi ini sekarang menerima state jawaban langsung dari handler yang memanggilnya.
func (handler *StudentHandlerImpl) serveQuestion(w http.ResponseWriter, r *http.Request, examID string, qNum int, attemptID string, savedAnswers map[string]string) {
	exam, err := handler.StudentService.GetExamById(r.Context(), examID)
	if err != nil {
		slog.Error("error getting exam", "err", err)

		if errors.Is(err, sql.ErrNoRows) {
			appError.RenderErrorPage(w, handler.Template, http.StatusNotFound, fmt.Sprintf("Exam with id %s is not found", examID))
			return
		}

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	questionList, err := handler.StudentService.GetQuestionsByExamId(r.Context(), examID)
	if err != nil {
		slog.Error("error getting question", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if len(questionList) == 0 {
		slog.Info("this exam has no questions")

		appError.RenderErrorPage(w, handler.Template, http.StatusNotFound, "Ujian tidak memiliki soal")
		return
	}

	if qNum < 1 || qNum > len(questionList) {
		slog.Info("questions is not found")

		appError.RenderErrorPage(w, handler.Template, http.StatusNotFound, "Soal tidak ditemukan")
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
		slog.Error("failed to execute template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

// SubmitExam adalah satu-satunya fungsi yang menyimpan semua jawaban ke database.
func (handler *StudentHandlerImpl) SubmitExam(w http.ResponseWriter, r *http.Request) {
	examId := r.PathValue("examId")
	cookie, err := r.Cookie("exam_attempt_id")
	if err != nil {
		slog.Error("failed to get exam_attempt_id cookie", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusUnauthorized, "Sesi ujian tidak valid atau telah berakhir")
		return
	}
	attemptID := cookie.Value

	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse form", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
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
				slog.Error("failed to save answer", "err", err)

				appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
				return
			}
		}
	}

	// 3. Panggil service untuk menghitung skor dan menyelesaikan ujian.
	// (Asumsi: service ini menangani kalkulasi dan update status attempt)
	_, err = handler.StudentService.CalculateScore(r.Context(), attemptID)
	if err != nil {
		slog.Error("error calculating score", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "exam_attempt_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// 4. Arahkan pengguna ke halaman hasil.
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

	// Get exam_attempts.score by student_id and exam_id
	examAttempId, totalScore, err := handler.StudentService.GetBiggestScoreByStudentIdAndExamId(r.Context(), user.Id, examId)
	if err != nil {
		slog.Error("error when calling GetBiggestExamAttemptsByStudentId", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	type CorrectionResult struct {
		Question         string
		RightAnswer      string
		StudentAnswer    string
		Score            int
		QuestionMaxScore int
		Similarity       float64
	}

	// Get student_answers by examAttemptId di mana akan mendapatkan data questionId untuk mendapatkan Question dan RightAnswer
	// StudentAnswer, Score, QuestionMaxScor
	studentAnswers, err := handler.StudentService.GetStudentAnswersByExamAttemptId(r.Context(), examAttempId)
	if err != nil {
		slog.Error("error when calling GetBiggestExamAttemptsByStudentId", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	corretionsResult := []CorrectionResult{}
	for _, studentAnswer := range studentAnswers {
		questionAndRightAnswer, err := handler.StudentService.FindQuestionById(r.Context(), studentAnswer.QuestionID)
		if err != nil {
			slog.Error("error when calling find question by id service", "err", err)

			appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		corretionsResult = append(corretionsResult, CorrectionResult{
			Question:         questionAndRightAnswer.Question,
			RightAnswer:      questionAndRightAnswer.RightAnswer,
			StudentAnswer:    studentAnswer.StudentAnswer,
			Score:            studentAnswer.Score,
			QuestionMaxScore: studentAnswer.QuestionMaxScore,
			Similarity:       studentAnswer.Similarity,
		})
	}

	type Result struct {
		TotalScore        int
		CorrectionResults []CorrectionResult
	}

	dataResponse := struct {
		User   domain.User
		Result Result
	}{
		User: user,
		Result: Result{
			TotalScore:        totalScore,
			CorrectionResults: corretionsResult,
		},
	}

	if err := handler.Template.ExecuteTemplate(w, "student-exam-result", dataResponse); err != nil {
		slog.Error("failed to execute template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (handler *StudentHandlerImpl) ExamResultView(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)
	if user.Role == "student" {
		user.Role = "Student"
	}

	examAttemptsCustom, err := handler.StudentService.GetBiggestExamAttemptsByStudentId(r.Context(), user.Id)
	if err != nil {
		slog.Error("error when calling GetBiggestExamAttemptsByStudentId", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	examsWithScoreAndTeacherName, err := handler.StudentService.GetExamsWithScoreAndTeacherNameByExamId(r.Context(), examAttemptsCustom)
	if err != nil {
		slog.Error("error when calling GetExamsWithScoreAndTeacherNameByExamId", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	dataResponse := web.ScoreListResponse{
		Exams: examsWithScoreAndTeacherName,
		User:  user,
	}

	handler.Template.ExecuteTemplate(w, "student-score-list", dataResponse)
}
