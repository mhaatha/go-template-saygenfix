package handler

import "net/http"

type UserHandler interface {
	RegisterAction(w http.ResponseWriter, r *http.Request)
	RegisterView(w http.ResponseWriter, r *http.Request)
}
