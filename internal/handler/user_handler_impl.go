package handler

import (
	"html/template"
	"log/slog"
	"net/http"

	appError "github.com/mhaatha/go-template-saygenfix/internal/errors"
	"github.com/mhaatha/go-template-saygenfix/internal/helper"
	"github.com/mhaatha/go-template-saygenfix/internal/model/web"
	"github.com/mhaatha/go-template-saygenfix/internal/service"
)

func NewUserHandler(userService service.UserService) UserHandler {
	return &UserHandlerImpl{
		UserService: userService,
		Template: template.Must(template.ParseFiles(
			"../../internal/templates/views/register.html",
			"../../internal/templates/views/success_register.html",
		)),
	}
}

type UserHandlerImpl struct {
	UserService service.UserService
	Template    *template.Template
}

func (handler *UserHandlerImpl) RegisterAction(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse form", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Check password confirmation
	password := r.PostFormValue("password")
	confirmPassword := r.PostFormValue("confirm_password")

	isPasswordValid := helper.CheckPasswordConfirmation(password, confirmPassword)
	if !isPasswordValid {
		slog.Error("password and confirm password do not match")

		appError.RenderErrorPage(w, handler.Template, http.StatusUnauthorized, "Password and confirm password do not match")
		return
	}

	userRequest := web.RegisterUserRequest{
		Email:    r.PostFormValue("email"),
		FullName: r.PostFormValue("full_name"),
		Password: password,
		Role:     r.PostFormValue("role"),
	}

	// Call service
	err := handler.UserService.RegisterNewUser(r.Context(), userRequest)
	if err != nil {
		slog.Error("failed to register new user", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := handler.Template.ExecuteTemplate(w, "success-register", nil); err != nil {
		slog.Error("failed to execute success-register template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (handler *UserHandlerImpl) RegisterView(w http.ResponseWriter, r *http.Request) {
	if err := handler.Template.ExecuteTemplate(w, "register", nil); err != nil {
		slog.Error("failed to execute register template", "err", err)

		appError.RenderErrorPage(w, handler.Template, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}
