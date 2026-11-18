package job

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"automate-gform-drive/internal/excel"
	"automate-gform-drive/internal/student"
)

// Translator mengabstraksikan modul terjemahan LLM
type Translator interface {
	TranslateCV(ctx context.Context, cv *student.CVData) (*student.CVData, string, string, string, error)
}

// DriveUploader mengabstraksikan modul pengunggah ke Google Drive
type DriveUploader interface {
	Upload(ctx context.Context, fileName string, fileBytes []byte) (fileID string, fileURL string, err error)
}

// Notifier mengabstraksikan modul notifikasi WhatsApp
type Notifier interface {
	SendFailure(ctx context.Context, nama, tahap, errMsg string, t time.Time) error
}

type JobTask struct {
	JobID        string
	ResponseID   string
	SubmitTime   time.Time
	CVFileName   string
	CVFileBase64 string
	FormData     map[string]interface{}
}

type Worker struct {
	repo         *Repository
	translator   Translator
	drive        DriveUploader
	notifier     Notifier
	templatePath string
	tmpDir       string

	taskQueue chan *JobTask
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewWorker(
	repo *Repository,
	translator Translator,
	drive DriveUploader,
	notifier Notifier,
	templatePath string,
	tmpDir string,
	bufferSize int,
) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		repo:         repo,
		translator:   translator,
		drive:        drive,
		notifier:     notifier,
		templatePath: templatePath,
		tmpDir:       tmpDir,
		taskQueue:    make(chan *JobTask, bufferSize),
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (w *Worker) Start(concurrency int) {
	for i := 0; i < concurrency; i++ {
		w.wg.Add(1)
		go w.processLoop(i + 1)
	}
	slog.Info("job worker dimulai", "concurrency", concurrency)
}

func (w *Worker) Enqueue(task *JobTask) bool {
	select {
	case w.taskQueue <- task:
		return true
	default:
		slog.Warn("antrean job penuh", "job_id", task.JobID)
		return false
	}
}

func (w *Worker) Stop() {
	w.cancel()
	close(w.taskQueue)
	w.wg.Wait()
	slog.Info("job worker berhenti")
}

func (w *Worker) processLoop(workerID int) {
	defer w.wg.Done()
	for {
		select {
		case <-w.ctx.Done():
			return
		case task, ok := <-w.taskQueue:
			if !ok {
				return
			}
			w.handleTask(task)
		}
	}
}

func (w *Worker) handleTask(task *JobTask) {
	ctx := context.Background()
	slog.Info("memulai pemrosesan job", "job_id", task.JobID, "response_id", task.ResponseID)

	// 1. Validasi & Dekode berkas Excel isian peserta
	if w.repo != nil {
		_ = w.repo.UpdateProgress(ctx, task.JobID, StatusDiproses, TahapValidasi)
	}

	if task.CVFileBase64 == "" {
		w.failJob(ctx, task, TahapValidasi, "berkas CV tidak ditemukan pada payload", false)
		return
	}

	excelBytes, err := base64.StdEncoding.DecodeString(task.CVFileBase64)
	if err != nil {
		w.failJob(ctx, task, TahapValidasi, fmt.Sprintf("gagal decode base64 berkas CV: %v", err), false)
		return
	}

	cv, err := excel.ReadFromBytes(excelBytes)
	if err != nil {
		w.failJob(ctx, task, TahapValidasi, fmt.Sprintf("gagal membaca format excel peserta: %v", err), false)
		return
	}

	// Masukkan submit time sebagai 作成日
	cv.SubmitTime = task.SubmitTime
	cv.ResponseID = task.ResponseID

	// Normalisasi data baku
	cv.Usia = student.NormalizeUsia(cv.Usia, cv.TanggalLahir, time.Now())
	cv.JenisKelamin = student.NormalizeJenisKelamin(cv.JenisKelamin)
	cv.TinggiBadan = student.NormalizeTinggi(cv.TinggiBadan)
	cv.BeratBadan = student.NormalizeBerat(cv.BeratBadan)
	cv.GolonganDarah = student.NormalizeGolonganDarah(cv.GolonganDarah)
	cv.Agama = student.NormalizeAgama(cv.Agama)
	cv.StatusPernikahan = student.NormalizeStatusNikah(cv.StatusPernikahan)
	cv.MenyetirMobil = student.NormalizeMenyetir(cv.MenyetirMobil)
	cv.Alkohol = student.NormalizeYaTidak(cv.Alkohol)
	cv.Merokok = student.NormalizeYaTidak(cv.Merokok)
	cv.TinggalBerkelompok = student.NormalizeTinggalBerkelompok(cv.TinggalBerkelompok)

	for i := range cv.Pendidikan {
		cv.Pendidikan[i].Periode = student.NormalizePeriode(cv.Pendidikan[i].Periode)
	}

	for i := range cv.Pekerjaan {
		cv.Pekerjaan[i].Periode = student.NormalizePeriode(cv.Pekerjaan[i].Periode)
	}

	for i := range cv.Keluarga {
		cv.Keluarga[i].Hubungan = student.NormalizeHubunganKeluarga(cv.Keluarga[i].Hubungan)
		cv.Keluarga[i].Usia = student.NormalizeUsiaKeluarga(cv.Keluarga[i].Usia)
	}

	// 2. Terjemahan (LLM & Glosarium)
	if w.repo != nil {
		_ = w.repo.UpdateProgress(ctx, task.JobID, StatusDiproses, TahapTerjemah)
	}

	var provider, model, promptVer string
	if w.translator != nil {
		translatedCV, p, m, pv, err := w.translator.TranslateCV(ctx, cv)
		if err != nil {
			w.failJob(ctx, task, TahapTerjemah, fmt.Sprintf("gagal menerjemahkan dengan LLM: %v", err), false)
			return
		}
		cv = translatedCV
		provider = p
		model = m
		promptVer = pv
	}

	if w.repo != nil {
		translatedJSON, _ := json.Marshal(cv)
		_ = w.repo.UpdateTranslation(ctx, task.JobID, translatedJSON, provider, model, promptVer)
	}

	// 3. Isi Template Excel Resmi (output-template-cv.xlsx)
	if w.repo != nil {
		_ = w.repo.UpdateProgress(ctx, task.JobID, StatusDiproses, TahapIsiTemplate)
	}

	cleanNama := sanitizeFileName(cv.NamaRomaji)
	if cleanNama == "" {
		cleanNama = "PESERTA_" + task.ResponseID
	}
	outFileName := fmt.Sprintf("CV_%s.xlsx", cleanNama)
	outPath := filepath.Join(w.tmpDir, fmt.Sprintf("%s_%s", task.JobID, outFileName))

	if err := excel.GenerateCV(w.templatePath, cv, outPath); err != nil {
		w.failJob(ctx, task, TahapIsiTemplate, fmt.Sprintf("gagal mengisi template excel: %v", err), false)
		return
	}
	if w.repo != nil {
		_ = w.repo.UpdateTemplateGenerated(ctx, task.JobID, outPath)
	}

	// 4. Upload ke Google Drive
	if w.repo != nil {
		_ = w.repo.UpdateProgress(ctx, task.JobID, StatusDiproses, TahapUploadDrive)
	}

	if w.drive != nil {
		// Ambil bytes file yang sudah jadi
		outBytes, err := excel.ReadGeneratedBytes(outPath)
		if err != nil {
			w.failJob(ctx, task, TahapUploadDrive, fmt.Sprintf("gagal membaca output excel untuk upload: %v", err), true)
			return
		}

		fileID, fileURL, err := w.drive.Upload(ctx, outFileName, outBytes)
		if err != nil {
			w.failJob(ctx, task, TahapUploadDrive, fmt.Sprintf("gagal mengunggah ke Google Drive: %v", err), true)
			return
		}

		if w.repo != nil {
			_ = w.repo.MarkCompleted(ctx, task.JobID, fileID, fileURL)
		}
		slog.Info("job berhasil diselesaikan", "job_id", task.JobID, "drive_url", fileURL)
	} else {
		// Jika driver upload belum dipasang (tahap pengujian)
		if w.repo != nil {
			_ = w.repo.MarkCompleted(ctx, task.JobID, "local", outPath)
		}
		slog.Info("job berhasil diselesaikan (disimpan lokal)", "job_id", task.JobID, "file_path", outPath)
	}
}

func (w *Worker) failJob(ctx context.Context, task *JobTask, tahap, errMsg string, isUploadError bool) {
	slog.Error("job gagal", "job_id", task.JobID, "tahap", tahap, "error", errMsg)
	if w.repo != nil {
		_ = w.repo.MarkFailed(ctx, task.JobID, tahap, errMsg, isUploadError)
	}

	if w.notifier != nil {
		nama := ""
		if val, ok := task.FormData["NAMA"]; ok {
			nama = fmt.Sprint(val)
		}
		_ = w.notifier.SendFailure(ctx, nama, tahap, errMsg, time.Now())
	}
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToUpper(name)
	replacer := strings.NewReplacer(" ", "_", "/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	return replacer.Replace(name)
}
