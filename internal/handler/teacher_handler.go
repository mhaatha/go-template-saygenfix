package handler

import "net/http"

type TeacherHandler interface {
	RoomUjianView(w http.ResponseWriter, r *http.Request)
}
