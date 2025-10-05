package router

import (
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/handler"
)

func UserRouter(handler handler.UserHandler, mux *http.ServeMux) {
	mux.HandleFunc("GET /register", handler.RegisterView)
	mux.HandleFunc("POST /register", handler.RegisterAction)
}
