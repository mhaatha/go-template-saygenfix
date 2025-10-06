package handler

import (
	"html/template"
	"net/http"
)

func NewStudentHandler() StudentHandler {
	return &StudentHandlerImpl{
		Template: template.Must(template.ParseFiles(
			"../../internal/templates/views/student/dashboard.html",
		)),
	}
}

type StudentHandlerImpl struct {
	Template *template.Template
}

func (handler *StudentHandlerImpl) RoomUjianView(w http.ResponseWriter, r *http.Request) {
	handler.Template.ExecuteTemplate(w, "student-dashboard", nil)
}
