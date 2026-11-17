# PRD: Sistem Otomatis Generate CV Jepang (技能実習履歴書)

Versi: 1.0 (draft) · Tanggal: 17 November 2026

## 1. Ringkasan

Sistem otomatis yang menerima jawaban Google Form (GForm) peserta magang, menerjemahkan isian ke bahasa Jepang, mengisi template Excel CV baku, lalu menyimpan hasilnya langsung ke Google Drive. Tidak ada antarmuka (FE) dan tidak ada login. Pemantauan lewat database dan notifikasi WhatsApp.

**Alur utama**

```
Peserta isi GForm
   → Apps Script (trigger onFormSubmit) kirim data ke backend (webhook)
   → Backend: validasi → simpan job → terjemah (LLM) → isi template Excel
   → Kirim file ke Apps Script web app → simpan ke Google Drive
   → Status job "selesai" di DB
   → Jika gagal: status "gagal" di DB + notifikasi WhatsApp (GoWA)
```

## 2. Tujuan dan non-tujuan

**Tujuan**
- CV terbentuk otomatis begitu GForm disubmit, tanpa intervensi admin.
- Format output konsisten dengan template yang sudah dipakai.
- Setiap kegagalan tercatat di DB dan sampai ke admin lewat WhatsApp.
- Kode modular agar mudah dirawat orang lain.

**Non-tujuan (versi 1)**
- Antarmuka web / dashboard.
- Login dan manajemen user.
- Upload manual Excel atau jalur input selain webhook GForm.
- Foto peserta.
- Output PDF.
- Bulk generate dari file Excel/ZIP.
- Ekstraksi dari CV Indonesia bebas atau scan form.

## 3. Pengguna

| Peran | Kebutuhan |
|---|---|
| Peserta | Mengisi GForm. |
| Admin | Mengambil CV dari Drive dan menerima WA kalau ada error. Memantau lewat database. |

## 4. Skala

- Sekitar 50-90 CV per hari (kurang dari 10 per jam).
- Antrean cukup memakai tabel PostgreSQL + worker goroutine. Redis tidak diperlukan.

## 5. Keputusan yang sudah ditetapkan

| Hal | Keputusan |
|---|---|
| Sumber data | Hanya webhook GForm via Apps Script |
| Output | Excel (.xlsx), mengisi template `CV Kosong` |
| Foto | Tidak dipakai (kotak foto dibiarkan kosong) |
| Backend | Go, struktur modular |
| Database | PostgreSQL |
| Deploy | Easypanel |
| Penyimpanan hasil | Google Drive via Apps Script web app (akun Gmail biasa) |
| Folder Drive | `CV/YYYY-MM/` (dikelompokkan per bulan), langsung final |
| Penamaan file | `CV_NAMA_LENGKAP_SISWA.xlsx` |
| LLM | Provider dapat diganti lewat konfigurasi, API key lewat environment variable |
| Notifikasi error | WhatsApp via GoWA |
| Antarmuka | Tanpa FE |
| Pemantauan | Tabel `jobs` di DB + log stdout di Easypanel |

## 6. Kebutuhan fungsional

### 6.1 Penerimaan data (webhook)
- Endpoint `POST /api/webhook/gform` menerima jawaban GForm dalam JSON dari Apps Script.
- Autentikasi lewat secret token di header. Request tanpa token valid ditolak.
- Setiap jawaban punya `response_id` unik dari GForm. Jika sudah ada di DB, diabaikan (anti-duplikat).
- Backend langsung membalas cepat (202 Accepted) dan memproses di background.

### 6.2 Validasi dan normalisasi
- Field wajib diperiksa (nama, tanggal lahir, dan lainnya, daftar final ditentukan setelah pemetaan GForm).
- Format dinormalisasi sesuai aturan template, misalnya:
  - Tinggi: `165センチメートル`, berat: `49キログラム`
  - Usia keluarga: `40歳`, golongan darah: `O型`
  - Periode: `2011年　06月～2017年　07月` (spasi lebar)
- Opsi baku dipetakan lewat tabel tetap tanpa LLM (jenis kelamin, status nikah, agama, ya/tidak, dll.).
- Jika validasi gagal, job ditandai `gagal` dengan pesan error jelas.

### 6.3 Terjemahan
- Hanya field teks bebas yang diterjemahkan LLM, misalnya kelebihan, kekurangan, hobi, tujuan ke Jepang, rencana setelah pulang, rincian kerja, nama perusahaan, jurusan, pekerjaan keluarga.
- Nama orang ditulis dalam katakana dan romaji.
- Dukungan **glossary** (istilah baku) agar terjemahan konsisten.
- Provider LLM diakses lewat satu interface, sehingga model bisa diganti lewat konfigurasi tanpa mengubah modul lain.
- Setiap job mencatat provider, model, dan versi prompt yang dipakai.

### 6.4 Pengisian template Excel
- Template `CV Kosong` disimpan sebagai file di server.
- Backend mengisi sel sesuai **peta sel** (file konfigurasi), tanpa mengubah layout, merge cell, atau style.
- Formula usia (`INT((TODAY()-C11)/365)`) dipertahankan dan dihitung ulang saat file dibuka.
- Slot terbatas: pekerjaan 3 baris, keluarga 4 baris. Perilaku jika data melebihi slot perlu ditetapkan (lihat bagian 12).

### 6.5 Penyimpanan ke Google Drive
- Backend mengirim file ke **Apps Script web app** (dilindungi secret token). Script menyimpan ke `CV/YYYY-MM/` di Drive pemilik akun. Folder dibuat otomatis jika belum ada.
- Penamaan: `CV_NAMA_LENGKAP_SISWA.xlsx` (huruf kapital, spasi menjadi underscore, karakter tidak valid dihapus). Jika nama kembar di folder yang sama, ditambah akhiran `_2`, `_3`, dst.
- Link/ID file Drive disimpan di DB.
- Jika upload gagal, file tetap disimpan sementara di server, status `gagal_upload`, dan upload dapat diulang tanpa memanggil LLM lagi.

### 6.6 Notifikasi error (WhatsApp via GoWA)
- Dikirim setelah status gagal tersimpan di DB.
- Kegagalan kirim WA tidak boleh menghentikan proses utama.
- Isi pesan: nama, tahap gagal, pesan error, waktu.
- GoWA berjalan sebagai service terpisah di Easypanel dengan nomor WA khusus, tidak dibuka ke internet, dan data sesi disimpan di volume persisten.

### 6.7 Pemantauan
- Tabel `jobs` menjadi sumber utama status.
- Log aplikasi ditulis ke stdout (terlihat di Easypanel → service → Logs), mencakup `response_id`, nama, tahap, dan hasil.
- Endpoint `GET /health` untuk health check.

## 7. Status job

`diterima` → `diproses` → `selesai`

Gagal: `gagal` (umum), `gagal_upload` (file sudah jadi, upload Drive gagal)

Kolom `tahap_terakhir`: `webhook`, `validasi`, `terjemah`, `isi_template`, `upload_drive`.

## 8. Skema database (draft)

Tabel `jobs`

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | uuid | PK |
| response_id | text, unique | ID response GForm (anti-duplikat) |
| nama | text | Nama lengkap |
| payload_mentah | jsonb | Jawaban GForm asli |
| data_terjemah | jsonb | Hasil setelah terjemah dan normalisasi |
| status | text | Lihat bagian 7 |
| tahap_terakhir | text | |
| error | text | Pesan error jika gagal |
| llm_provider, llm_model, prompt_version | text | Jejak terjemahan |
| file_path | text | Lokasi file sementara di server |
| drive_file_id, drive_url | text | Hasil upload |
| dibuat_pada, selesai_pada | timestamptz | |

Tabel pendukung (opsional, ditentukan saat implementasi): `glossary` jika istilah ingin disimpan di DB, bukan file.

Catatan: `payload_mentah` berisi data pribadi, jadi perlu kebijakan retensi (lihat bagian 12).

## 9. Struktur folder backend (modular)

```
cv-generator/
├── cmd/
│   └── api/
│       └── main.go              # entry point, merakit semua modul
├── internal/
│   ├── config/                  # baca environment variable
│   ├── platform/                # infrastruktur umum
│   │   ├── database/            # koneksi PostgreSQL, migrasi
│   │   ├── logger/              # logger terstruktur (stdout)
│   │   └── httpserver/          # router, middleware
│   ├── webhook/                 # handler webhook GForm, cek token, anti-duplikat
│   ├── student/                 # model data CV, validasi, normalisasi format
│   ├── mapping/                 # peta kolom GForm → field, tabel opsi baku
│   ├── translate/               # interface Provider + glossary
│   │   ├── provider/            # implementasi per LLM (mudah tambah/ganti)
│   │   └── glossary/
│   ├── excel/                   # isi template (excelize), peta sel
│   ├── drive/                   # client ke Apps Script web app
│   ├── notify/                  # client GoWA (WhatsApp)
│   ├── job/                     # model, repository, antrean, worker, retry
│   └── health/                  # endpoint /health
├── configs/
│   ├── cell_map.yaml            # field → sel di template
│   ├── gform_map.yaml           # pertanyaan GForm → field
│   └── glossary.yaml            # istilah baku
├── templates/
│   └── cv_template.xlsx         # template Excel (sheet "CV Kosong")
├── migrations/                  # file SQL migrasi
├── scripts/
│   └── apps-script/
│       ├── gform_trigger.gs     # kirim jawaban GForm ke backend
│       └── drive_upload.gs      # web app penyimpan file ke Drive
├── docs/
│   └── PRD.md
├── Dockerfile
├── .env.example
└── README.md
```

**Aturan modul**
- Tiap modul punya tanggung jawab tunggal dan berkomunikasi lewat interface.
- Modul `job` mengorkestrasi urutan: validasi → terjemah → isi template → upload. Modul lain tidak saling memanggil langsung.
- Perubahan format template atau pertanyaan GForm cukup mengubah file di `configs/` dan `templates/`, tanpa mengubah kode.

## 10. Konfigurasi (environment variable)

| Variabel | Fungsi |
|---|---|
| `DATABASE_URL` | Koneksi PostgreSQL |
| `WEBHOOK_SECRET` | Token yang diverifikasi dari Apps Script |
| `LLM_PROVIDER`, `LLM_MODEL`, `LLM_API_KEY` | Pemilihan dan akses LLM |
| `DRIVE_SCRIPT_URL`, `DRIVE_SCRIPT_SECRET` | Alamat dan token Apps Script penyimpan Drive |
| `GOWA_URL`, `GOWA_AUTH`, `ADMIN_WA_NUMBER` | Pengiriman WA |
| `TEMPLATE_PATH`, `TMP_DIR` | Lokasi template dan file sementara |

Semua rahasia hanya lewat environment variable di Easypanel, tidak disimpan di kode.

## 11. Non-fungsional

- **Keamanan:** hanya endpoint webhook yang terbuka ke internet, dilindungi secret token. Data pribadi peserta tidak ditulis penuh ke log.
- **Ketahanan:** worker memproses satu job per satu waktu atau jumlah kecil secara paralel, aman terhadap restart (job yang menggantung dilanjutkan atau ditandai ulang).
- **Biaya:** catat provider, model, dan jumlah token per job untuk memantau biaya LLM.
- **Pemeliharaan:** log terstruktur, migrasi DB berversi, dan README untuk orang lain yang melanjutkan.

## 12. Pertanyaan terbuka (perlu keputusan)

1. **Daftar pertanyaan GForm** dan pemetaannya ke tiap sel template (diperlukan untuk `gform_map.yaml` dan `cell_map.yaml`).
2. **Field yang tidak ada di GForm** (事前試験結果: matematika, ujian Jepang, push-up, sit-up, grip, dan 評価): dikosongkan, atau diisi dengan cara lain?
3. **作成日 (tanggal dibuat):** memakai tanggal submit GForm atau tanggal generate?
4. **Provider LLM default:** ditentukan setelah pengujian pada data contoh.
5. **Data melebihi slot** (pekerjaan lebih dari 3, keluarga lebih dari 4): dipotong, digabung, atau ditandai gagal?
6. **Retry otomatis:** berapa kali percobaan ulang sebelum job dianggap gagal?
7. **Notifikasi tambahan:** perlu ringkasan harian lewat WA dan pemantauan down lewat Uptime Kuma?
8. **Retensi data:** berapa lama `payload_mentah` (data pribadi) disimpan di DB?
9. **Jika backend mati** saat peserta submit: perlu antrean ulang dari sisi Apps Script, atau diterima sebagai risiko?
10. **Kebijakan menimpa file** jika response yang sama diproses ulang (timpa, atau buat versi baru)?

## 13. Kriteria keberhasilan

- Submit GForm menghasilkan file Excel di `CV/YYYY-MM/` di Drive tanpa tindakan manual.
- Layout hasil identik dengan template.
- Semua job gagal tercatat di DB dengan tahap dan pesan error, dan admin menerima WA.
- Tidak ada file ganda untuk response yang sama.
- Ganti provider LLM cukup lewat konfigurasi.
