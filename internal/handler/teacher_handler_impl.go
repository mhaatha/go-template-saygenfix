package handler

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/mhaatha/go-template-saygenfix/internal/helper"
	"github.com/mhaatha/go-template-saygenfix/internal/middleware"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/service"
)

func NewTeacherHandler(teacherService service.TeacherService) TeacherHandler {
	return &TeacherHandlerImpl{
		TeacherService: teacherService,
		Template: template.Must(template.ParseFiles(
			"../../internal/templates/views/teacher/dashboard.html",
			"../../internal/templates/views/teacher/upload.html",
			"../../internal/templates/views/teacher/check_exam.html",
			"../../internal/templates/views/teacher/exam_result.html",
			"../../internal/templates/views/teacher/generate-result.html",
			"../../internal/templates/views/teacher/edit_exam.html",
			"../../internal/templates/views/partial/teacher_dashboard_navbar.html",
			"../../internal/templates/views/partial/teacher_upload_navbar.html",
		)),
	}
}

type TeacherHandlerImpl struct {
	TeacherService service.TeacherService
	Template       *template.Template
}

func (handler *TeacherHandlerImpl) TeacherDashboard(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)
	dashboardResponse, err := handler.TeacherService.TeacherDashboard(r.Context(), user.Id)
	if err != nil {
		log.Fatal(err)
	}

	if err := handler.Template.ExecuteTemplate(w, "teacher-dashboard", dashboardResponse); err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
}

func (handler *TeacherHandlerImpl) GenerateAndCreateExamRoom(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.CurrentUserKey).(domain.User)

	// Ambil jumlah soal
	quantity := r.FormValue("quantity")
	fmt.Println(quantity)

	// Ambil exam data
	roomName := r.FormValue("room_name")
	year := r.FormValue("year")
	duration := r.FormValue("duration")

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		log.Print(err)
		return
	}

	durationInt, err := strconv.Atoi(duration)
	if err != nil {
		log.Print(err)
		return
	}

	examData := domain.Exam{
		RoomName: roomName,
		Year:     yearInt,
		Duration: durationInt,
	}

	totalQuestion, err := strconv.Atoi(quantity)
	if err != nil {
		log.Printf("Error converting quantity to int: %v", err)
		http.Error(w, "Invalid quantity value", http.StatusBadRequest)
		return
	}

	// 1. Parse multipart form, dengan batas ukuran memori 10 MB
	// File yang lebih besar dari ini akan disimpan di file sementara di disk.
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Gagal mem-parsing form", http.StatusInternalServerError)
		return
	}

	// 2. Ambil file dari form data menggunakan 'name' dari input field
	// "myFile" harus sama dengan atribut 'name' pada <input type="file" name="myFile">
	file, _, err := r.FormFile("pdf_file")
	if err != nil {
		log.Printf("Error mengambil file dari form: %v", err)
		http.Error(w, "File tidak ditemukan di request", http.StatusBadRequest)
		return
	}
	defer file.Close() // Jangan lupa untuk selalu menutup file

	handler.TeacherService.GenerateQuestionAnswer(r.Context(), file, totalQuestion, examData, user.Id)

	// Redirect HTMX
	w.Header().Set("HX-Redirect", "/teacher/dashboard")
}

func (handler *TeacherHandlerImpl) CheckExamView(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("id")
	if roomId == "" {
		slog.Error("room id is empty")
		helper.RenderError(w, "room id is empty")
		return
	}

	// GET API exam room by id

	// GET API answers

	// Kumpulkan data lalu kirim ke FE

	// Data DUMMY
	exam := domain.Exam{
		Id:        "EXAM-12123",
		RoomName:  "UTS PBO Semester 4",
		Year:      2025,
		Duration:  60,
		TeacherId: "36748630-eea7-4eff-b92f-f00fd2630a5d",
		CreatedAt: time.Now(),
	}

	handler.Template.ExecuteTemplate(w, "teacher-check-exam", exam)
}

func (handler *TeacherHandlerImpl) EditExamView(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("id")
	if roomId == "" {
		slog.Error("room id is empty")
		helper.RenderError(w, "room id is empty")
		return
	}

	// GET API exam room by id

	// GET API answers

	// Kumpulkan data lalu kirim ke FE

	// Data DUMMY
	exam := domain.Exam{
		Id:        "EXAM-12123",
		RoomName:  "UTS PBO Semester 4",
		Year:      2025,
		Duration:  60,
		TeacherId: "36748630-eea7-4eff-b92f-f00fd2630a5d",
		CreatedAt: time.Now(),
	}

	handler.Template.ExecuteTemplate(w, "teacher-edit-exam", exam)
}

func (handler *TeacherHandlerImpl) ExamResultView(w http.ResponseWriter, r *http.Request) {
	examId := r.PathValue("id")
	if examId == "" {
		slog.Error("exam id is empty")
		helper.RenderError(w, "exam id is empty")
		return
	}

	if err := handler.Template.ExecuteTemplate(w, "exam-result", nil); err != nil {
		log.Fatal(err)
	}
}

func (handler *TeacherHandlerImpl) GenerateResultView(w http.ResponseWriter, r *http.Request) {
	handler.Template.ExecuteTemplate(w, "teacher-generate-result", User{
		FullName: "Fulan S.pd, M.pd",
		Role:     "Pengajar",
	})
}
