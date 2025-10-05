package handler

import (
	"net/http"
)

type LandingPageHandler interface {
	Index(w http.ResponseWriter, r *http.Request)
}
