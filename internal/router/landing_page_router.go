package router

import (
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/handler"
)

func LandingPageRouter(handler handler.LandingPageHandler, mux *http.ServeMux) {
	mux.HandleFunc("/", handler.Index)
}
