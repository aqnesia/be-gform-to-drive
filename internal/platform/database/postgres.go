package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, connString string) (*DB, error) {
	if connString == "" {
		return nil, fmt.Errorf("DATABASE_URL tidak boleh kosong")
	}

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("gagal parsing konfigurasi database: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctxTimeout, config)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat connection pool: %w", err)
	}

	if err := pool.Ping(ctxTimeout); err != nil {
		return nil, fmt.Errorf("gagal ping database: %w", err)
	}

	slog.Info("koneksi database PostgreSQL berhasil terhubung")
	return &DB{Pool: pool}, nil
}

// AutoMigrate secara otomatis membuat tabel dan index jika belum tersedia
func (db *DB) AutoMigrate(ctx context.Context) error {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	CREATE TABLE IF NOT EXISTS jobs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		response_id TEXT NOT NULL UNIQUE,
		nama TEXT NOT NULL,
		payload_mentah JSONB NOT NULL,
		data_terjemah JSONB,
		status TEXT NOT NULL DEFAULT 'diterima',
		tahap_terakhir TEXT NOT NULL DEFAULT 'webhook',
		error TEXT,
		llm_provider TEXT,
		llm_model TEXT,
		prompt_version TEXT,
		file_path TEXT,
		drive_file_id TEXT,
		drive_url TEXT,
		dibuat_pada TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		selesai_pada TIMESTAMPTZ
	);

	CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
	CREATE INDEX IF NOT EXISTS idx_jobs_response_id ON jobs(response_id);
	`
	_, err := db.Pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("gagal menjalankan auto-migrate: %w", err)
	}
	slog.Info("auto-migrate database PostgreSQL berhasil dijalankan")
	return nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		slog.Info("koneksi database PostgreSQL ditutup")
	}
}
