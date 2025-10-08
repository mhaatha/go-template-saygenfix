package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/mhaatha/go-template-saygenfix/internal/config"
	"github.com/mhaatha/go-template-saygenfix/internal/database"
	"github.com/mhaatha/go-template-saygenfix/internal/handler"
	"github.com/mhaatha/go-template-saygenfix/internal/middleware"
	"github.com/mhaatha/go-template-saygenfix/internal/repository"
	"github.com/mhaatha/go-template-saygenfix/internal/router"
	"github.com/mhaatha/go-template-saygenfix/internal/service"
)

func main() {
	// Log init
	config.LogInit()

	// Config init
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("godotenv fails to load .env file", "err", err)
		os.Exit(1)
	}

	// Validator init
	validate := config.ValidatorInit()

	// Database init
	db, err := database.ConnectDB(cfg)
	if err != nil {
		slog.Error("failed connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Main ServeMux
	mux := http.NewServeMux()

	// Landing page resources
	landingPageHandler := handler.NewLandingPageHandler()

	// Landing page router
	router.LandingPageRouter(landingPageHandler, mux)

	// File server for static files
	fileServer := http.FileServer(http.Dir("../../internal/templates/public/assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets", fileServer))
	cssFileServer := http.FileServer(http.Dir("../../internal/templates/public/css"))
	mux.Handle("/css/", http.StripPrefix("/css", cssFileServer))

	// User resources
	userRepository := repository.NewUserRepository()
	userService := service.NewUserService(userRepository, db, validate)
	userHandler := handler.NewUserHandler(userService)

	// User router
	router.UserRouter(userHandler, mux)

	// Authentication resources
	authRepository := repository.NewAuthRepository()
	authService := service.NewAuthService(authRepository, db, validate)
	authHandler := handler.NewAuthHandler(authService, userService, cfg)

	// Authentication router
	router.AuthRouter(authHandler, mux)

	// Teacher resources
	teacherRepository := repository.NewTeacherRepository()
	teacherService := service.NewTeacherService(teacherRepository, db, validate, cfg)
	teacherHandler := handler.NewTeacherHandler(teacherService)

	// Teacher router with middleware
	teacherRouter := http.NewServeMux()
	router.TeacherRouter(teacherHandler, teacherRouter)

	// Middleware for teacher
	authMiddleware := middleware.NewAuthMiddleware(authService, cfg)
	mux.Handle("/teacher/", authMiddleware.Authenticate(authMiddleware.RequireRole("teacher")(teacherRouter)))

	// Student resources
	studentHandler := handler.NewStudentHandler()

	// Student router with middleware
	studentRouter := http.NewServeMux()
	router.StudentRouter(studentHandler, studentRouter)

	// Middleware for student
	mux.Handle("/student/", authMiddleware.Authenticate(authMiddleware.RequireRole("student")(studentRouter)))

	// Server
	server := http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: mux,
	}

	slog.Info("starting server on :" + cfg.AppPort)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("HTTP server error", "err", err)
		os.Exit(1)
	}
}
