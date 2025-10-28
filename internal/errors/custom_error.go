package errors

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
)

func RenderErrorPage(w http.ResponseWriter, template *template.Template, statusCode int, message string) {
	slog.Info("Render Error Page executed")

	w.WriteHeader(statusCode)

	data := web.ErrorPageData{
		StatusCode:   statusCode,
		StatusText:   http.StatusText(statusCode),
		ErrorMessage: message,
	}

	err := template.ExecuteTemplate(w, "error.html", data)
	fmt.Println(data)
	fmt.Println(err)
	if err != nil {
		slog.Error("error when executing 'error.html' template", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
