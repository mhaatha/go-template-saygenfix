package handler

import (
	"html/template"
	"net/http"
)

func NewTeacherHandler() TeacherHandler {
	return &TeacherHandlerImpl{
		Template: template.Must(template.ParseFiles("../../internal/templates/views/teacher/dashboard.html")),
	}
}

type TeacherHandlerImpl struct {
	Template *template.Template
}

func (handler *TeacherHandlerImpl) RoomUjianView(w http.ResponseWriter, r *http.Request) {
	handler.Template.ExecuteTemplate(w, "teacher-dashboard", nil)
}
