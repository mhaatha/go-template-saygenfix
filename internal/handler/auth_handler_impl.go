package handler

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mhaatha/go-template-saygenfix/internal/config"
	appError "github.com/mhaatha/go-template-saygenfix/internal/errors"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
	"github.com/mhaatha/go-template-saygenfix/internal/service"
)

func NewAuthHandler(authService service.AuthService, userService service.UserService, cfg *config.Config) AuthHandler {
	return &AuthHandlerImpl{
		AuthService: authService,
		UserService: userService,
		Template: template.Must(template.ParseFiles(
			"../../internal/templates/views/login.html",
		)),
		Cfg: cfg,
	}
}

type AuthHandlerImpl struct {
	AuthService service.AuthService
	UserService service.UserService
	Template    *template.Template
	Cfg         *config.Config
}

func (handler *AuthHandlerImpl) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse form", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	userRequest := web.LoginRequest{
		Email:    r.PostFormValue("email"),
		Password: r.PostFormValue("password"),
	}

	// Get user
	user, err := handler.UserService.GetUserByEmail(r.Context(), userRequest.Email)
	if err != nil {
		slog.Error("failed to get user by email", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	// Set max age and session name
	sessionName := handler.Cfg.SessionName
	maxAge, _ := strconv.Atoi(handler.Cfg.SessionMaxAge)

	// Call teacher service
	sessionId, errr := handler.AuthService.Login(r.Context(), userRequest, user.Email, user.Password, user.Id)
	if errr != nil {
		slog.Error("failed to when calling Login service", "err", errr)

		appError.RenderErrorPage(w, handler.Template, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionName,
		Value:    sessionId,
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		Path:     "/",
	})

	// Redirect to teacher or student dashboard, depends on the what user role
	switch user.Role {
	case "teacher":
		w.Header().Set("HX-Redirect", "/teacher/dashboard")
	case "student":
		w.Header().Set("HX-Redirect", "/student/dashboard")
	}
}

func (handler *AuthHandlerImpl) LoginView(w http.ResponseWriter, r *http.Request) {
	if err := handler.Template.ExecuteTemplate(w, "login", nil); err != nil {
		slog.Error("failed to execute login template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}
