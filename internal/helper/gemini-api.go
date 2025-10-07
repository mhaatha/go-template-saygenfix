package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"github.com/mhaatha/go-template-saygenfix/internal/model/domain"
)

func BuildResponse(resp *genai.GenerateContentResponse) string {
	var rawResponse string

	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					rawResponse += string(txt)
				}
			}
		}
	}

	return rawResponse
}

func UploadPDF(ctx context.Context, client *genai.Client, file multipart.File) (string, error) {
	// Generate random file name
	fileName := uuid.NewString() + ".pdf"

	opts := &genai.UploadFileOptions{
		DisplayName: fileName,
		MIMEType:    "application/pdf",
	}

	// Upload PDF files to Google Cloud Storage
	pdfURL, err := client.UploadFile(ctx, "", file, opts)
	if err != nil {
		slog.Error("error when uploading file to google cloud", "err", err)
		return "", err
	}

	return pdfURL.URI, nil
}

func GenerateQAFromPDF(ctx context.Context, client *genai.Client, fileURL string, totalQuestion int) ([]domain.QAItem, error) {
	model := client.GenerativeModel("gemini-2.5-pro")
	prompt := []genai.Part{
		genai.FileData{
			URI:      fileURL,
			MIMEType: "application/pdf",
		},
		genai.Text(fmt.Sprintf(`Berdasarkan dokumen PDF ini, buat %d soal esai beserta jawabannya. Buat jawabannya singkat namun cocok untuk koreksi essay, untuk soal dan jawabannya mengikuti isi dari dokumen PDF tersebut, teruntuk referensi soal dan jawaban diambil dari dokumen PDF, kemudian untuk kombinasi soal dan jawabannya menggunakan model Sentence-BERT. Format respons Anda WAJIB sebagai JSON array. Contoh: [{"question": "Apa itu...", "answer": "Jawabannya adalah..."}, {"question": "Siapa...", "answer": "Dia adalah..."}] Jangan tambahkan format markdown atau teks lain di luar JSON tersebut. Gunakan plaintext tanpa format markdown dalam tiap value question dan answer.`, totalQuestion)),
	}

	resp, err := model.GenerateContent(ctx, prompt...)
	if err != nil {
		slog.Error("error producing content", "err", err)
		return nil, err
	}

	rawResponse := BuildResponse(resp)

	// Clean the response
	cleanResponse := strings.TrimSpace(rawResponse)
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")

	var qaList []domain.QAItem
	err = json.Unmarshal([]byte(cleanResponse), &qaList)
	if err != nil {
		slog.Error("error when unmarshaling the cleanResponse", "err", err)
		return nil, err
	}

	return qaList, nil
}
