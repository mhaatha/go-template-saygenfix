package handler

import "net/http"

type TeacherHandler interface {
	RoomUjianView(w http.ResponseWriter, r *http.Request)
	UploadView(w http.ResponseWriter, r *http.Request)
	CheckExamView(w http.ResponseWriter, r *http.Request)
	ExamResultView(w http.ResponseWriter, r *http.Request)
}
