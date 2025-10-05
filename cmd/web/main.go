package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/mhaatha/go-template-saygenfix/internal/config"
	"github.com/mhaatha/go-template-saygenfix/internal/database"
	"github.com/mhaatha/go-template-saygenfix/internal/handler"
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
	authHandler := handler.NewAuthHandler(authService)

	// Authentication router
	router.AuthRouter(authHandler, mux)

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

// GEMINI API EXAMPLE
// apiKey := os.Getenv("GEMINI_API_KEY")
// if apiKey == "" {
// 	log.Fatal("GEMINI_API_KEY environment variable not set.")
// }

// ctx := context.Background()
// client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
// if err != nil {
// 	log.Fatal(err)
// }
// defer client.Close()

// func printResponse(resp *genai.GenerateContentResponse) {
// 	for _, cand := range resp.Candidates {
// 		if cand.Content != nil {
// 			for _, part := range cand.Content.Parts {
// 				fmt.Println(part)
// 			}
// 		}
// 	}
// }

// func listModels(client *genai.Client, ctx context.Context) error {
// 	models := client.ListModels(ctx)

// 	for {
// 		model, err := models.Next()
// 		if model == nil {
// 			break
// 		}
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		fmt.Printf("%s | ", model.Name)
// 		for i, action := range model.SupportedGenerationMethods {
// 			if i > 0 {
// 				fmt.Print(", ")
// 			}
// 			fmt.Print(action)
// 		}
// 		fmt.Println()
// 	}

// 	return nil
// }

// func generateEssayFromPDF(client *genai.Client, ctx context.Context) {
// 	f, err := os.Open("/home/notrhel/Downloads/Anarkisme.pdf")
// 	if err != nil {
// 		log.Fatalf("Gagal membuka file PDF: %v", err)
// 	}
// 	defer f.Close()

// 	opts := &genai.UploadFileOptions{
// 		DisplayName: "Anarkisme.pdf",
// 		MIMEType:    "application/pdf",
// 	}

// 	// Upload file PDF ke Google Cloud Storage
// 	// dan dapatkan URL-nya. Ganti "your-bucket-name" dan "path/to/your-file.pdf".
// 	pdfURL, err := client.UploadFile(ctx, "", f, opts)
// 	if err != nil {
// 		log.Fatalf("Gagal mengunggah file PDF: %v", err)
// 	}

// 	model := client.GenerativeModel("gemini-2.5-flash")
// 	prompt := []genai.Part{
// 		// Sertakan file yang sudah di-upload menggunakan URI-nya
// 		genai.FileData{
// 			URI:      pdfURL.URI,
// 			MIMEType: "application/pdf",
// 		},
// 		genai.Text("Berdasarkan dokumen PDF ini, buat soal-jawaban essay sebanyak 5 soal dengan jawaban yang sederhana untuk siswa SMP."),
// 	}

// 	resp, err := model.GenerateContent(ctx, prompt...)
// 	if err != nil {
// 		log.Fatalf("Gagal menghasilkan konten: %v", err)
// 	}

// 	printResponse(resp)
// }
