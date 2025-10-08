package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/mhaatha/go-template-saygenfix/internal/config"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
	"github.com/mhaatha/go-template-saygenfix/internal/service"
)

// ContextKey is a custom type to avoid key collisions in context
type ContextKey string

const CurrentUserKey ContextKey = "currentUser"

func NewAuthMiddleware(authService service.AuthService, cfg *config.Config) AuthMiddleware {
	return &AuthMiddlewareImpl{
		AuthService: authService,
		Config:      cfg,
	}
}

type AuthMiddleware interface {
	Authenticate(next http.Handler) http.Handler
	RequireRole(role string) func(next http.Handler) http.Handler
}

type AuthMiddlewareImpl struct {
	AuthService service.AuthService
	Config      *config.Config
}

// Authenticate checks for a valid session cookie and adds user info to the request context.
func (m *AuthMiddlewareImpl) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(m.Config.SessionName)
		if err != nil || cookie.Value == "" {
			slog.Error("cookie not found", "err", err)

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := m.AuthService.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			slog.Error("failed to validate session", "err", err)
			http.SetCookie(w, &http.Cookie{Name: m.Config.SessionName, Value: "", Path: "/", MaxAge: -1})

			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Send session data through context
		ctx := context.WithValue(r.Context(), CurrentUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole checks if the user in the context has the required role.
// This middleware MUST run AFTER the Authenticate middleware.
func (m *AuthMiddlewareImpl) RequireRole(role string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user is in context
			user := r.Context().Value(CurrentUserKey)

			if user == nil {
				slog.Error("user not found in context")

				// No user in context, redirect to login
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Check user role
			userData, ok := user.(domain.User)
			if !ok || userData.Role != role {
				// Role does not match, redirect based on role
				// For example, a student trying to access a teacher page is redirected to the student dashboard
				currentUserRole := userData.Role
				if currentUserRole != "" {
					if currentUserRole == "teacher" {
						http.Redirect(w, r, "/teacher/dashboard", http.StatusForbidden)
						return
					}
					if currentUserRole == "student" {
						http.Redirect(w, r, "/student/dashboard", http.StatusForbidden)
						return
					}
				}

				// Default redirect if role is unknown
				http.Redirect(w, r, "/login", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), CurrentUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
