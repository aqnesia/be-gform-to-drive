package student

import "time"

// CVData merepresentasikan seluruh data CV peserta magang
type CVData struct {
	// Metadata
	SubmitTime       time.Time `json:"submit_time"`
	ResponseID       string    `json:"response_id"`
	CreatedDate      string    `json:"created_date"`      // 作成日 (G4)
	NomorRegistrasi  string    `json:"nomor_registrasi"`  // Kotak Kuning (G1:G2)

	// Foto Peserta
	FotoBytes []byte `json:"-"`
	FotoExt   string `json:"foto_ext"`

	// Data Diri
	NamaRomaji         string `json:"nama_romaji"`         // C8
	NamaKatakana       string `json:"nama_katakana"`       // C7
	Panggilan          string `json:"panggilan"`           // G8 (Katakana\nRomaji)
	PanggilanKatakana  string `json:"panggilan_katakana"`
	PanggilanRomaji    string `json:"panggilan_romaji"`
	TanggalLahir       string `json:"tanggal_lahir"`       // C11
	Usia               string `json:"usia"`                // G11 (contoh: 21歳)
	JenisKelamin       string `json:"jenis_kelamin"`       // A13 (男 / 女)
	TempatLahir        string `json:"tempat_lahir"`        // C13
	StatusPernikahan   string `json:"status_pernikahan"`   // D13 (未婚 / 既婚 / 離婚)
	NoTelp             string `json:"no_telp"`             // G13
	Alamat             string `json:"alamat"`              // A15 (Tetap Bahasa Indonesia)
	Domisili           string `json:"domisili"`            // G15
	TinggiBadan        string `json:"tinggi_badan"`        // A17 (165センチメートル)
	BeratBadan         string `json:"berat_badan"`         // C17 (50キログラム)
	MenyetirMobil      string `json:"menyetir_mobil"`      // D17 (できる / できない)
	Agama              string `json:"agama"`               // G17 (イスラム教, dll)
	GolonganDarah      string `json:"golongan_darah"`      // A19 (O型)
	Alkohol            string `json:"alkohol"`             // C19 (有 / 無)
	Merokok            string `json:"merokok"`             // D19 (有 / 無)
	Keahlian           string `json:"keahlian"`            // G19 (dari input, tidak digenerate AI)
	Kelebihan          string `json:"kelebihan"`           // C20 (dari input, tidak digenerate AI)
	TinggalBerkelompok string `json:"tinggal_berkelompok"` // G20 (有 / 無)
	Kekurangan         string `json:"kekurangan"`          // C21 (dari input, tidak digenerate AI)
	Hobi               string `json:"hobi"`                // G21
	TujuanJepang       string `json:"tujuan_jepang"`       // D23
	TargetTabungan     string `json:"target_tabungan"`     // D24
	RencanaPulang      string `json:"rencana_pulang"`      // D25

	// Pendidikan
	Pendidikan []PendidikanItem `json:"pendidikan"` // Baris 29 - 32

	// Riwayat Pekerjaan
	Pekerjaan []PekerjaanItem `json:"pekerjaan"` // Baris 36 - 38

	// Susunan Keluarga
	Keluarga []KeluargaItem `json:"keluarga"` // Baris 43 - 46

	// Hasil Ujian Awal (事前試験結果)
	HasilUjian HasilUjianData `json:"hasil_ujian"`
}

type PendidikanItem struct {
	Tingkat string `json:"tingkat"` // SD, SMP, SMA, PT
	Sekolah string `json:"sekolah"` // B
	Jurusan string `json:"jurusan"` // D
	Periode string `json:"periode"` // G
}

type PekerjaanItem struct {
	Perusahaan string `json:"perusahaan"` // A (asli, tidak digenerate jika nama asli)
	Rincian    string `json:"rincian"`    // D
	Periode    string `json:"periode"`    // G
}

type KeluargaItem struct {
	Hubungan  string `json:"hubungan"`  // A
	Nama      string `json:"nama"`      // C
	Pekerjaan string `json:"pekerjaan"` // D
	Usia      string `json:"usia"`      // G
}

type HasilUjianData struct {
	BahasaJepang string `json:"bahasa_jepang"` // A50 (54点)
	Matematika   string `json:"matematika"`    // C50 (86.6点)
	PushUp       string `json:"push_up"`       // D50 (48)
	SitUp        string `json:"sit_up"`        // E50 (44)
	Grip         string `json:"grip"`          // G50 (35.1)
}
