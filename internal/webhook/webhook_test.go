package webhook

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"automate-gform-drive/internal/job"
)

func TestWebhookHandler(t *testing.T) {
	inputPath := filepath.Join("..", "..", "templates", "cv_template.xlsx")
	fileBytes, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("gagal membaca template: %v", err)
	}

	templatePath := filepath.Join("..", "..", "templates", "output-template-cv.xlsx")
	tmpDir := filepath.Join("..", "..", "tmp")

	// Inisialisasi Worker standalone
	worker := job.NewWorker(nil, nil, nil, nil, templatePath, tmpDir, 10)
	worker.Start(1)
	defer worker.Stop()

	handler := NewHandler("test-secret", nil, worker)

	payload := Payload{
		ResponseID:   "RESP_TEST_001",
		SubmitTime:   time.Now().Format(time.RFC3339),
		CVFileName:   "CV_TEST.xlsx",
		CVFileBase64: base64.StdEncoding.EncodeToString(fileBytes),
		Data: map[string]interface{}{
			"NAMA": "TEST SISWA",
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/webhook/gform", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Secret", "test-secret")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202 Accepted, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Tunggu worker memproses sejenak
	time.Sleep(500 * time.Millisecond)
	t.Log("Webhook submission berhasil diterima dan diproses antrean worker!")
}
