package handler

import "net/http"

type TeacherHandler interface {
	TeacherDashboard(w http.ResponseWriter, r *http.Request)
	UploadView(w http.ResponseWriter, r *http.Request)
	CheckExamView(w http.ResponseWriter, r *http.Request)
	ExamResultView(w http.ResponseWriter, r *http.Request)
	GenerateAndCreateExamRoom(w http.ResponseWriter, r *http.Request)
	GenerateResultView(w http.ResponseWriter, r *http.Request)
}
