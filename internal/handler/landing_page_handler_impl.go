package handler

import (
	"html/template"
	"net/http"
)

func NewLandingPageHandler() LandingPageHandler {
	return &LandingPageImpl{
		Template: template.Must(template.ParseFiles("../../internal/templates/index.html")),
	}
}

type LandingPageImpl struct {
	Template *template.Template
}

func (handler *LandingPageImpl) Index(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("../../internal/templates/index.html"))
	tmpl.Execute(w, nil)
}
