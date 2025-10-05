package router

import (
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/handler"
)

func AuthRouter(handler handler.AuthHandler, mux *http.ServeMux) {
	mux.HandleFunc("POST /login", handler.Login)
	mux.HandleFunc("GET /login", handler.LoginView)
}
