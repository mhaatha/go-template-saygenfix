package handler

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"strconv"
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

func (handler *StudentHandlerImpl) TakeExamView(w http.ResponseWriter, r *http.Request) {
	examId := r.PathValue("examId")
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)

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
		Expires:  time.Now().Add(3 * time.Hour), // Sesuaikan durasi ujian
		HttpOnly: true,
	})

	handler.serveQuestion(w, r, examId, 1, attemptID)
}

func (handler *StudentHandlerImpl) HandleQuestionPartial(w http.ResponseWriter, r *http.Request) {
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
	studentAnswer := r.FormValue("answer")
	questionID := r.FormValue("questionId")

	if questionID != "" && studentAnswer != "" {
		answer := web.StudentAnswer{
			ExamAttemptID: attemptID,
			QuestionID:    questionID,
			StudentAnswer: studentAnswer,
		}
		if err := handler.StudentService.SaveAnswer(r.Context(), answer); err != nil {
			log.Printf("Error saving answer: %v", err)
		}
	}

	examID := r.PathValue("examId")
	qNumStr := r.PathValue("qNum")
	qNum, _ := strconv.Atoi(qNumStr)

	handler.serveQuestion(w, r, examID, qNum, attemptID)
}

func (handler *StudentHandlerImpl) serveQuestion(w http.ResponseWriter, r *http.Request, examID string, qNum int, attemptID string) {
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

	savedAnswers, _ := handler.StudentService.GetAnswersByAttemptId(r.Context(), attemptID)

	savedAnswersMap := make(map[string]string)
	for _, ans := range savedAnswers {
		savedAnswersMap[ans.QuestionID] = ans.StudentAnswer
	}

	data := web.ExamPageData{
		ExamID:                examID,
		ExamTitle:             exam.RoomName,
		Questions:             questionList,
		CurrentQuestion:       questionList[qNum-1],
		CurrentQuestionNumber: qNum,
		TotalQuestions:        len(questionList),
		NextQuestionNumber:    qNum + 1,
		PrevQuestionNumber:    qNum - 1,
		SavedAnswer:           savedAnswersMap,
	}

	templateName := "student-take-exam"
	if r.Header.Get("HX-Request") == "true" {
		templateName = "question-partial"
	}

	err = handler.Template.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Printf("Error executing template %s: %v", templateName, err)
		http.Error(w, "Terjadi kesalahan saat merender halaman", http.StatusInternalServerError)
		return
	}
}

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
	studentAnswer := r.FormValue("answer")
	questionID := r.FormValue("questionId")

	if questionID != "" {
		answer := web.StudentAnswer{
			ExamAttemptID: attemptID,
			QuestionID:    questionID,
			StudentAnswer: studentAnswer,
		}
		if err := handler.StudentService.SaveAnswer(r.Context(), answer); err != nil {
			log.Printf("Error saving final answer: %v", err)
		}
	}

	if err := handler.StudentService.CompleteExamAttempt(r.Context(), attemptID); err != nil {
		log.Printf("Error completing exam attempt: %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "exam_attempt_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	_, err = handler.StudentService.CalculateScore(r.Context(), attemptID)
	if err != nil {
		log.Printf("Error calculating score: %v", err)
		http.Error(w, "Gagal menghitung skor ujian", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "exam_attempt_id",
		Value:    attemptID,
		Path:     "/",
		Expires:  time.Now().Add(3 * time.Hour),
		HttpOnly: true,
	})

	// redirect to result page
	w.Header().Set("HX-Redirect", "/student/exam-result/"+examId)
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

	// Get attempt ID
	cookie, err := r.Cookie("exam_attempt_id")
	if err != nil {
		http.Error(w, "Sesi ujian tidak valid atau telah berakhir", http.StatusUnauthorized)
		return
	}
	attemptID := cookie.Value

	// Get student answers by exam attempt id
	answers, err := handler.StudentService.GetAnswersByAttemptId(r.Context(), attemptID)
	if err != nil {
		log.Printf("Error getting student answers: %v", err)
		http.Error(w, "Gagal mendapatkan jawaban siswa", http.StatusInternalServerError)
		return
	}

	fmt.Println("%+v", answers)

	AA := struct {
		User       domain.User
		TotalScore int
	}{
		User:       user,
		TotalScore: 90,
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "exam_attempt_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	err = handler.Template.ExecuteTemplate(w, "student-exam-result", AA)
	if err != nil {
		slog.Error(err.Error())
	}
}

func (handler *StudentHandlerImpl) ExamResultView(w http.ResponseWriter, r *http.Request) {
	handler.Template.ExecuteTemplate(w, "student-score-list", nil)
}
