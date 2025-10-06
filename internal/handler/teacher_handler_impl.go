package handler

import (
	"html/template"
	"net/http"
)

func NewTeacherHandler() TeacherHandler {
	return &TeacherHandlerImpl{
		Template: template.Must(template.ParseFiles(
			"../../internal/templates/views/teacher/dashboard.html",
			"../../internal/templates/views/teacher/upload.html",
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
