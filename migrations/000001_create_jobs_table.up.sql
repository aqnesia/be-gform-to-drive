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

