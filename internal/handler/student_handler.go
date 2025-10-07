package handler

import "net/http"

type StudentHandler interface {
	DashboardView(w http.ResponseWriter, r *http.Request)
	TakeExamView(w http.ResponseWriter, r *http.Request)
	HandleQuestionPartial(w http.ResponseWriter, r *http.Request)
	CorrectExam(w http.ResponseWriter, r *http.Request)
	CorrectExamView(w http.ResponseWriter, r *http.Request)
	ResultExamView(w http.ResponseWriter, r *http.Request)
}
