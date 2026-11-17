package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"automate-gform-drive/internal/job"
)

type Payload struct {
	ResponseID   string                 `json:"response_id"`
	SubmitTime   string                 `json:"submit_time"`
	CVFileName   string                 `json:"cv_file_name"`
	CVFileBase64 string                 `json:"cv_file_base64"`
	Data         map[string]interface{} `json:"data"`
}

type Handler struct {
	secret string
	repo   *job.Repository
	worker *job.Worker
}

func NewHandler(secret string, repo *job.Repository, worker *job.Worker) *Handler {
	return &Handler{
		secret: secret,
		repo:   repo,
		worker: worker,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Secret validation
	authHeader := r.Header.Get("X-Webhook-Secret")
	if authHeader == "" {
		authHeader = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	}

	if h.secret != "" && authHeader != h.secret {
		slog.Warn("unauthorized webhook request")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var payload Payload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Error("failed to decode webhook payload", "error", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if payload.ResponseID == "" {
		http.Error(w, "response_id is required", http.StatusBadRequest)
		return
	}

	slog.Info("received webhook submission", "response_id", payload.ResponseID)

	nama := ""
	if val, ok := payload.Data["NAMA"]; ok {
		nama = fmt.Sprint(val)
	} else if val, ok := payload.Data["Nama Lengkap"]; ok {
		nama = fmt.Sprint(val)
	}

	rawBytes, _ := json.Marshal(payload)

	// Simpan ke database jika repo aktif
	var jobID string
	if h.repo != nil {
		j, err := h.repo.Create(r.Context(), payload.ResponseID, nama, rawBytes)
		if err != nil {
			if errors.Is(err, job.ErrJobAlreadyExists) {
				slog.Info("response_id sudah pernah diproses (anti-duplikat)", "response_id", payload.ResponseID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"status":  "ignored",
					"message": "duplicate response_id ignored",
				})
				return
			}
			slog.Error("gagal menyimpan job ke database", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		jobID = j.ID
	} else {
		jobID = payload.ResponseID
	}

	// Parse submit time
	var submitTime time.Time
	if payload.SubmitTime != "" {
		parsed, err := time.Parse(time.RFC3339, payload.SubmitTime)
		if err == nil {
			submitTime = parsed
		}
	}
	if submitTime.IsZero() {
		submitTime = time.Now()
	}

	// Masukkan ke antrean worker
	if h.worker != nil {
		h.worker.Enqueue(&job.JobTask{
			JobID:        jobID,
			ResponseID:   payload.ResponseID,
			SubmitTime:   submitTime,
			CVFileName:   payload.CVFileName,
			CVFileBase64: payload.CVFileBase64,
			FormData:     payload.Data,
		})
	}

	// Respond 202 Accepted immediately as per PRD
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":      "accepted",
		"job_id":      jobID,
		"response_id": payload.ResponseID,
	})
}


