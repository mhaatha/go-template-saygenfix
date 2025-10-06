package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/helper"
)

func NewTeacherHandler() TeacherHandler {
	return &TeacherHandlerImpl{
		Template: template.Must(template.ParseFiles(
			"../../internal/templates/views/teacher/dashboard.html",
			"../../internal/templates/views/teacher/upload.html",
			"../../internal/templates/views/teacher/check_exam.html",
			"../../internal/templates/views/teacher/exam_result.html",
		)),
	}
}

type TeacherHandlerImpl struct {
	Template *template.Template
}

func (handler *TeacherHandlerImpl) RoomUjianView(w http.ResponseWriter, r *http.Request) {
	handler.Template.ExecuteTemplate(w, "teacher-dashboard", nil)
}

func (handler *TeacherHandlerImpl) UploadView(w http.ResponseWriter, r *http.Request) {
	handler.Template.ExecuteTemplate(w, "teacher-upload", nil)
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

	handler.Template.ExecuteTemplate(w, "teacher-check-exam", nil)
}

func (handler *TeacherHandlerImpl) ExamResultView(w http.ResponseWriter, r *http.Request) {
	examId := r.PathValue("id")
	if examId == "" {
		slog.Error("exam id is empty")
		helper.RenderError(w, "exam id is empty")
		return
	}

	handler.Template.ExecuteTemplate(w, "exam-result", nil)
}
