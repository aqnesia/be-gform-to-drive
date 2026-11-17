package job

import (
	"encoding/json"
	"time"
)

const (
	StatusDiterima    = "diterima"
	StatusDiproses    = "diproses"
	StatusSelesai     = "selesai"
	StatusGagal       = "gagal"
	StatusGagalUpload = "gagal_upload"

	TahapWebhook     = "webhook"
	TahapValidasi    = "validasi"
	TahapTerjemah    = "terjemah"
	TahapIsiTemplate = "isi_template"
	TahapUploadDrive = "upload_drive"
)

type Job struct {
	ID            string          `json:"id"`
	ResponseID    string          `json:"response_id"`
	Nama          string          `json:"nama"`
	PayloadMentah json.RawMessage `json:"payload_mentah"`
	DataTerjemah  json.RawMessage `json:"data_terjemah,omitempty"`
	Status        string          `json:"status"`
	TahapTerakhir string          `json:"tahap_terakhir"`
	Error         *string         `json:"error,omitempty"`
	LLMProvider   *string         `json:"llm_provider,omitempty"`
	LLMModel      *string         `json:"llm_model,omitempty"`
	PromptVersion *string         `json:"prompt_version,omitempty"`
	FilePath      *string         `json:"file_path,omitempty"`
	DriveFileID   *string         `json:"drive_file_id,omitempty"`
	DriveURL      *string         `json:"drive_url,omitempty"`
	DibuatPada    time.Time       `json:"dibuat_pada"`
	SelesaiPada   *time.Time      `json:"selesai_pada,omitempty"`
}
