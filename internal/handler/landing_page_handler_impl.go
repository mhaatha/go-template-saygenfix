package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	appError "github.com/mhaatha/go-template-saygenfix/internal/errors"
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
	if err := handler.Template.ExecuteTemplate(w, "index.html", nil); err != nil {
		slog.Error("error when executing index template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}
