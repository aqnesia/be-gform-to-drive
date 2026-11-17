package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"automate-gform-drive/internal/config"
	"automate-gform-drive/internal/drive"
	"automate-gform-drive/internal/job"
	"automate-gform-drive/internal/notify"
	"automate-gform-drive/internal/platform/database"
	"automate-gform-drive/internal/platform/httpserver"
	"automate-gform-drive/internal/platform/logger"
	"automate-gform-drive/internal/translate"
)

func main() {
	logger.Init()

	cfg := config.Load()
	slog.Info("configuration loaded", "port", cfg.Port, "llm_provider", cfg.LLMProvider)

	// Inisialisasi Database
	var repo *job.Repository
	var db *database.DB
	if cfg.DatabaseURL != "" {
		var err error
		db, err = database.Connect(context.Background(), cfg.DatabaseURL)
		if err != nil {
			slog.Warn("database connection failed, running without database persistence", "error", err)
		} else {
			defer db.Close()
			repo = job.NewRepository(db.Pool)
		}
	} else {
		slog.Warn("DATABASE_URL is not set, running in memory-only mode")
	}

	// Inisialisasi Layanan Terjemahan AI & Glosarium
	translator, err := translate.NewService(cfg, "configs/glossary.yaml")
	if err != nil {
		slog.Warn("gagal inisialisasi translator, berjalan tanpa LLM", "error", err)
	}

	// Inisialisasi Google Drive Client
	var driveClient *drive.Client
	if cfg.DriveScriptURL != "" {
		driveClient = drive.NewClient(cfg.DriveScriptURL, cfg.DriveScriptSecret)
	} else {
		slog.Warn("DRIVE_SCRIPT_URL belum diset, hasil generate akan disimpan sementara di folder lokal")
	}

	// Inisialisasi WhatsApp Notifier
	var notifier *notify.Client
	if cfg.GoWAURL != "" && cfg.AdminWANumber != "" {
		notifier = notify.NewClient(cfg.GoWAURL, cfg.GoWAAuth, cfg.AdminWANumber)
	}

	// Inisialisasi Worker
	worker := job.NewWorker(repo, translator, driveClient, notifier, cfg.TemplatePath, cfg.TmpDir, 50)
	worker.Start(1)
	defer worker.Stop()

	server := httpserver.New(cfg.Port, cfg.WebhookSecret, repo, worker)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server exited cleanly")
}


