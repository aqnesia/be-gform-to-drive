# Otomasi Generate CV Jepang (技能実習履歴書)

Sistem backend otomatis untuk menerima jawaban Google Form, menerjemahkan isian ke bahasa Jepang, mengisi template Excel CV baku, dan menyimpannya langsung ke Google Drive.

## Struktur Folder

```
automate-gform-drive/
├── cmd/
│   └── api/
│       └── main.go              # Entry point server backend
├── internal/
│   ├── config/                  # Pembacaan environment variable
│   ├── platform/                # Infrastruktur umum (logger, httpserver)
│   ├── webhook/                 # Handler endpoint webhook GForm
│   └── health/                  # Endpoint /health
├── configs/
│   ├── cell_map.yaml            # Pemetaan field ke sel template Excel
│   ├── gform_map.yaml           # Pemetaan pertanyaan GForm ke field
│   └── glossary.yaml            # Istilah baku (glossary) terjemahan
├── templates/
│   └── cv_template.xlsx         # File template Excel CV Jepang (sheet "CV Kosong")
├── migrations/                  # File SQL migrasi PostgreSQL
├── scripts/
│   └── apps-script/
│       ├── gform_trigger.gs     # Script pemicu onFormSubmit
│       └── drive_upload.gs      # Script Web App pengunggah ke Drive
├── Dockerfile                   # Dockerfile untuk deployment Easypanel
├── .env.example                 # Template konfigurasi environment
└── README.md
```

## Menjalankan Server Lokal

```bash
cp .env.example .env
go run ./cmd/api
```

Health check:
```bash
curl http://localhost:8080/health
```

