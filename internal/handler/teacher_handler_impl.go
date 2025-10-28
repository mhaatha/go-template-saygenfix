package handler

import (
	"database/sql"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"slices"
	"strconv"

	appError "github.com/mhaatha/go-template-saygenfix/internal/errors"
	"github.com/mhaatha/go-template-saygenfix/internal/middleware"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
	"github.com/mhaatha/go-template-saygenfix/internal/service"
)

func NewTeacherHandler(teacherService service.TeacherService, studentService service.StudentService) TeacherHandler {
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
	}

	return &TeacherHandlerImpl{
		TeacherService: teacherService,
		StudentService: studentService,
		Template: template.Must(
			// 1. Mulai dengan membuat template baru. Nama "base" bisa apa saja.
			template.New("base").

				// 2. Tambahkan FuncMap Anda ke template yang baru dibuat.
				Funcs(funcMap).

				// 3. Baru parse semua file Anda seperti sebelumnya.
				ParseFiles(
					"../../internal/templates/views/teacher/dashboard.html",
					"../../internal/templates/views/teacher/upload.html",
					"../../internal/templates/views/teacher/check_exam.html",
					"../../internal/templates/views/teacher/exam_result.html",
					"../../internal/templates/views/teacher/generate-result.html",
					"../../internal/templates/views/teacher/edit_exam.html",
					"../../internal/templates/views/partial/teacher_dashboard_navbar.html",
					"../../internal/templates/views/partial/teacher_upload_navbar.html",
					"../../internal/templates/views/partial/teacher_check_exam_navbar.html",
					"../../internal/templates/views/partial/teacher_edit_exam_navbar.html",
					"../../internal/templates/views/partial/exam_card.html",
					"../../internal/templates/views/error.html",
				),
		),
	}
}

type TeacherHandlerImpl struct {
	TeacherService service.TeacherService
	StudentService service.StudentService
	Template       *template.Template
}

func (handler *TeacherHandlerImpl) TeacherDashboard(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)
	dashboardResponse, err := handler.TeacherService.TeacherDashboard(r.Context(), user.Id)
	if err != nil {
		slog.Error("error when calling teacher dashboard service", "err", err)

		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("error when calling teacher dashboard service", "err", err)

			appError.RenderErrorPage(w, handler.Template, http.StatusNotFound, "Teacher is not found")
			return
		}

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	years := []int{}
	for _, exam := range dashboardResponse.Exams {
		if !slices.Contains(years, exam.Year) {
			years = append(years, exam.Year)
		}
	}

	dashboardResponse.Years = years

	if err := handler.Template.ExecuteTemplate(w, "teacher-dashboard", dashboardResponse); err != nil {
		slog.Error("error when executing teacher dashboard template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

type User struct {
	FullName string
	Role     string
}

func (handler *TeacherHandlerImpl) UploadView(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)
	if user.Role == "teacher" {
		user.Role = "Teacher"
	}

	if err := handler.Template.ExecuteTemplate(w, "teacher-upload", user); err != nil {
		slog.Error("error when executing teacher upload template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (handler *TeacherHandlerImpl) GenerateAndCreateExamRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)

	// Ambil jumlah soal
	quantity := r.FormValue("quantity")

	// Ambil exam data
	roomName := r.FormValue("room_name")
	year := r.FormValue("year")
	duration := r.FormValue("duration")

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		slog.Error("error when converting year to int", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "year is not valid")
		return
	}

	durationInt, err := strconv.Atoi(duration)
	if err != nil {
		slog.Error("error when converting duration to int", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "duration is not valid")
		return
	}

	examData := domain.Exam{
		RoomName: roomName,
		Year:     yearInt,
		Duration: durationInt,
	}

	totalQuestion, err := strconv.Atoi(quantity)
	if err != nil {
		slog.Error("error when converting total question to int", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "total question is not valid")
		return
	}

	// 1. Parse multipart form, dengan batas ukuran memori 10 MB
	// File yang lebih besar dari ini akan disimpan di file sementara di disk.
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		slog.Error("error when parsing multipart form", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "file tidak valid")
		return
	}

	// 2. Ambil file dari form data menggunakan 'name' dari input field
	// "myFile" harus sama dengan atribut 'name' pada <input type="file" name="myFile">
	file, _, err := r.FormFile("pdf_file")
	if err != nil {
		slog.Error("error when getting file from form data", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "file tidak valid")
		return
	}
	defer file.Close() // Jangan lupa untuk selalu menutup file

	err = handler.TeacherService.GenerateQuestionAnswer(r.Context(), file, totalQuestion, examData, user.Id)
	if err != nil {
		slog.Error("error when calling generate question answer service", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Redirect HTMX
	w.Header().Set("HX-Redirect", "/teacher/dashboard")
}

func (handler *TeacherHandlerImpl) CheckExamView(w http.ResponseWriter, r *http.Request) {
	var successMessage string
	if r.URL.Query().Get("status") == "updated" {
		successMessage = "Data ujian berhasil diperbarui!"
	}

	roomId := r.PathValue("examId")
	if roomId == "" {
		slog.Info("exam id is not found")

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "exam id is not found")
		return
	}

	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)
	if user.Role == "teacher" {
		user.Role = "Teacher"
	}

	exam, err := handler.TeacherService.GetExamById(r.Context(), roomId)
	if err != nil {
		slog.Error("error when calling get exam by id service", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Get exam_attempts by examId
	examAttempts, err := handler.TeacherService.GetBiggestExamAttemptsScoreByExamId(r.Context(), roomId)
	if err != nil {
		slog.Error("error when calling get biggest exam attempts score by exam id service", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Get student users FullName by examAttemptsId
	var examAttemptsData []web.ExamAttemptsWithStudentName
	isInserted := make(map[string]bool)
	for _, attempt := range examAttempts {
		studentName, studentId, err := handler.TeacherService.GetStudentFullNameByExamAttemptsId(r.Context(), attempt.ID)
		if err != nil {
			slog.Error("error when calling get student full name by exam attempts id service", "err", err)

			appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		if !isInserted[studentName] {
			examAttemptsData = append(examAttemptsData, web.ExamAttemptsWithStudentName{
				Id:          attempt.ID,
				StudentId:   studentId,
				ExamId:      roomId,
				StudentName: studentName,
				Score:       attempt.Score,
			})
		}
	}

	examCheckResponse := web.TeacherCheckExamResponse{
		User:         user,
		Exam:         exam,
		FlashMessage: successMessage,
		ExamAttempts: examAttemptsData,
	}

	if err := handler.Template.ExecuteTemplate(w, "teacher-check-exam", examCheckResponse); err != nil {
		slog.Error("error when executing teacher-check-exam template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (handler *TeacherHandlerImpl) EditExamView(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("id")
	if roomId == "" {
		slog.Error("room id is empty")

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "room id is empty")
		return
	}

	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)
	if user.Role == "teacher" {
		user.Role = "Teacher"
	}

	exam, err := handler.TeacherService.GetExamById(r.Context(), roomId)
	if err != nil {
		slog.Error("error when calling get exam by id service", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	questionsAndAnswers, err := handler.TeacherService.GetQAByExamId(r.Context(), roomId)
	if err != nil {
		slog.Error("error when calling get qa by exam id service", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	examEditResponse := web.TeacherEditExamResponse{
		User:               user,
		Exam:               exam,
		QuestionAndAnswers: questionsAndAnswers,
	}

	if err := handler.Template.ExecuteTemplate(w, "teacher-edit-exam", examEditResponse); err != nil {
		slog.Error("error when execute teacher-edit-exam template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (handler *TeacherHandlerImpl) EditExam(w http.ResponseWriter, r *http.Request) {
	// 1. Parse form data
	if err := r.ParseForm(); err != nil {
		slog.Error("error parsing form data", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	examId := r.PathValue("id")

	// 2. Ambil data ujian utama
	roomName := r.FormValue("roomName")
	yearStr := r.FormValue("year")
	durationStr := r.FormValue("duration")

	// Lakukan konversi tipe data (string to int) dan validasi di sini
	yearInt, err := strconv.Atoi(yearStr)
	if err != nil {
		slog.Error("error when converting year to int", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "year is not valid")
		return
	}

	durationInt, err := strconv.Atoi(durationStr)
	if err != nil {
		slog.Error("error when converting duration to int", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "duration is not valid")
		return
	}

	// Panggil service untuk memperbarui data ujian
	if err := handler.TeacherService.UpdateExamById(r.Context(), examId, roomName, yearInt, durationInt); err != nil {
		slog.Error("error when calling update exam by id service", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// 3. Ambil dan proses data soal dan jawaban
	// r.Form["qa_ids"] akan berisi slice dari semua ID soal, contoh: ["id1", "id2", "id3"]
	qaIDs := r.Form["qa_ids"]

	for _, id := range qaIDs {
		// Bentuk nama field sesuai dengan yang ada di template
		questionFieldName := "question_" + id
		answerFieldName := "answer_" + id

		// Ambil nilainya
		questionText := r.FormValue(questionFieldName)
		answerText := r.FormValue(answerFieldName)

		// Panggil service Anda untuk mengupdate data soal ini di database.
		if err := handler.TeacherService.UpdateQuestionById(r.Context(), id, questionText, answerText); err != nil {
			slog.Error("error when calling update question by id service", "err", err)

			appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
			return
		}
	}

	// 4. Redirect pengguna kembali setelah selesai
	http.Redirect(w, r, "/teacher/check-exam/"+examId+"?status=updated", http.StatusSeeOther)
}

func (handler *TeacherHandlerImpl) ExamResultView(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)
	if user.Role == "teacher" {
		user.Role = "Teacher"
	}
	studentId := r.PathValue("id")
	if studentId == "" {
		slog.Error("path value student id is empty")

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "student id is empty")
		return
	}
	examId := r.URL.Query().Get("exam_id")
	if examId == "" {
		slog.Error("query exam_id is empty")

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "exam_id is empty")
		return
	}

	// Get exam_attempts.score by student_id and exam_id
	examAttempId, totalScore, err := handler.StudentService.GetBiggestScoreByStudentIdAndExamId(r.Context(), studentId, examId)
	if err != nil {
		slog.Error("error when calling GetBiggestExamAttemptsByStudentIdAndExamId service", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	type CorrectionResult struct {
		Question         string
		RightAnswer      string
		StudentAnswer    string
		Score            float64
		QuestionMaxScore float64
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
			slog.Error("error when calling FindQuestionById", "err", err)

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

	if err := handler.Template.ExecuteTemplate(w, "teacher-exam-result", dataResponse); err != nil {
		slog.Error("error when executing teacher-exam-result template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (handler *TeacherHandlerImpl) ExamToggleButton(w http.ResponseWriter, r *http.Request) {
	examId := r.PathValue("id")
	if examId == "" {
		slog.Error("path value exam id is empty")

		appError.RenderErrorPage(w, handler.Template, http.StatusBadRequest, "id is empty")
		return
	}

	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)

	// Update isActive
	updateToggle, err := handler.TeacherService.UpdateIsActiveExamById(r.Context(), user.Id, examId)
	if err != nil {
		slog.Error("error when calling update is active exam by id service", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	cardData := web.TeacherDashboardResponse{
		User:  user,
		Exams: []domain.Exam{updateToggle},
	}

	if err := handler.Template.ExecuteTemplate(w, "exam-card", cardData); err != nil {
		slog.Error("error when calling exam-card template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (handler *TeacherHandlerImpl) GenerateResultView(w http.ResponseWriter, r *http.Request) {
	if err := handler.Template.ExecuteTemplate(w, "teacher-generate-result", User{
		FullName: "Fulan S.pd, M.pd",
		Role:     "Pengajar",
	}); err != nil {
		slog.Error("error when executing teacher-generate-result template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}
