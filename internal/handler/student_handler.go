package handler

import "net/http"

type StudentHandler interface {
	RoomUjianView(w http.ResponseWriter, r *http.Request)
}
