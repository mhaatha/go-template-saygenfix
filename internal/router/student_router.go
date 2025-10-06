package router

import (
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/handler"
)

func StudentRouter(handler handler.StudentHandler, mux *http.ServeMux) {
	mux.HandleFunc("GET /room-ujian-student", handler.RoomUjianView)
}
