package helper

import (
	"html/template"
	"net/http"
)

func RenderError(w http.ResponseWriter, message string) {
	tmpl := `<div class="error-message" style="color:red;">{{.}}</div>`
	t, _ := template.New("error").Parse(tmpl)
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, message)
}
