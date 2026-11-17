package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrJobAlreadyExists = errors.New("job dengan response_id ini sudah ada")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create menyimpan job baru saat webhook pertama kali diterima (anti-duplikat via response_id)
func (r *Repository) Create(ctx context.Context, responseID, nama string, payloadMentah json.RawMessage) (*Job, error) {
	query := `
		INSERT INTO jobs (response_id, nama, payload_mentah, status, tahap_terakhir, dibuat_pada)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (response_id) DO NOTHING
		RETURNING id, response_id, nama, payload_mentah, status, tahap_terakhir, dibuat_pada
	`
	now := time.Now()
	var j Job
	err := r.pool.QueryRow(ctx, query, responseID, nama, payloadMentah, StatusDiterima, TahapWebhook, now).Scan(
		&j.ID, &j.ResponseID, &j.Nama, &j.PayloadMentah, &j.Status, &j.TahapTerakhir, &j.DibuatPada,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobAlreadyExists
		}
		return nil, fmt.Errorf("gagal insert job: %w", err)
	}
	return &j, nil
}

// UpdateProgress memperbarui tahap pengerjaan job
func (r *Repository) UpdateProgress(ctx context.Context, jobID, status, tahap string) error {
	query := `
		UPDATE jobs
		SET status = $2, tahap_terakhir = $3
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, jobID, status, tahap)
	if err != nil {
		return fmt.Errorf("gagal update progress job: %w", err)
	}
	return nil
}

// UpdateTranslation menyimpan data terjemahan dan info LLM
func (r *Repository) UpdateTranslation(ctx context.Context, jobID string, dataTerjemah json.RawMessage, provider, model, promptVer string) error {
	query := `
		UPDATE jobs
		SET data_terjemah = $2, llm_provider = $3, llm_model = $4, prompt_version = $5, tahap_terakhir = $6
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, jobID, dataTerjemah, provider, model, promptVer, TahapTerjemah)
	if err != nil {
		return fmt.Errorf("gagal update translation job: %w", err)
	}
	return nil
}

// UpdateTemplateGenerated menyimpan lokasi file Excel sementara yang berhasil digenerate
func (r *Repository) UpdateTemplateGenerated(ctx context.Context, jobID, filePath string) error {
	query := `
		UPDATE jobs
		SET file_path = $2, tahap_terakhir = $3
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, jobID, filePath, TahapIsiTemplate)
	if err != nil {
		return fmt.Errorf("gagal update file path job: %w", err)
	}
	return nil
}

// MarkCompleted menandai job sukses beserta ID & URL berkas Google Drive
func (r *Repository) MarkCompleted(ctx context.Context, jobID, driveFileID, driveURL string) error {
	query := `
		UPDATE jobs
		SET status = $2, tahap_terakhir = $3, drive_file_id = $4, drive_url = $5, selesai_pada = $6
		WHERE id = $1
	`
	now := time.Now()
	_, err := r.pool.Exec(ctx, query, jobID, StatusSelesai, TahapUploadDrive, driveFileID, driveURL, now)
	if err != nil {
		return fmt.Errorf("gagal mark completed job: %w", err)
	}
	return nil
}

// MarkFailed menandai job gagal dengan pesan error
func (r *Repository) MarkFailed(ctx context.Context, jobID, tahap, errMsg string, isUploadError bool) error {
	status := StatusGagal
	if isUploadError {
		status = StatusGagalUpload
	}
	query := `
		UPDATE jobs
		SET status = $2, tahap_terakhir = $3, error = $4, selesai_pada = $5
		WHERE id = $1
	`
	now := time.Now()
	_, err := r.pool.Exec(ctx, query, jobID, status, tahap, errMsg, now)
	if err != nil {
		return fmt.Errorf("gagal mark failed job: %w", err)
	}
	return nil
}

// GetByResponseID mencari job berdasarkan ID respon Google Form
func (r *Repository) GetByResponseID(ctx context.Context, responseID string) (*Job, error) {
	query := `
		SELECT id, response_id, nama, payload_mentah, data_terjemah, status, tahap_terakhir,
		       error, llm_provider, llm_model, prompt_version, file_path, drive_file_id, drive_url,
		       dibuat_pada, selesai_pada
		FROM jobs
		WHERE response_id = $1
	`
	var j Job
	err := r.pool.QueryRow(ctx, query, responseID).Scan(
		&j.ID, &j.ResponseID, &j.Nama, &j.PayloadMentah, &j.DataTerjemah, &j.Status, &j.TahapTerakhir,
		&j.Error, &j.LLMProvider, &j.LLMModel, &j.PromptVersion, &j.FilePath, &j.DriveFileID, &j.DriveURL,
		&j.DibuatPada, &j.SelesaiPada,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal get job by response_id: %w", err)
	}
	return &j, nil
}
