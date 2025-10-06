package router

import (
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/handler"
)

func TeacherRouter(handler handler.TeacherHandler, mux *http.ServeMux) {
	mux.HandleFunc("GET /room-ujian-teacher", handler.RoomUjianView)
	mux.HandleFunc("GET /upload-teacher", handler.UploadView)
}
