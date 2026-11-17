package student

import "time"

// CVData merepresentasikan seluruh data CV peserta magang
type CVData struct {
	// Metadata
	SubmitTime  time.Time `json:"submit_time"`
	ResponseID  string    `json:"response_id"`
	CreatedDate string    `json:"created_date"` // 作成日 (G4)

	// Data Diri
	NamaRomaji        string `json:"nama_romaji"`         // C8
	NamaKatakana      string `json:"nama_katakana"`       // C7
	Panggilan         string `json:"panggilan"`           // G8
	TanggalLahir      string `json:"tanggal_lahir"`       // C11
	Usia              string `json:"usia"`                // G11 (jika diisi teks, atau dihitung formula)
	JenisKelamin      string `json:"jenis_kelamin"`       // A13 (男 / 女)
	TempatLahir       string `json:"tempat_lahir"`        // C13
	StatusPernikahan  string `json:"status_pernikahan"`   // D13 (未婚 / 既婚)
	NoTelp            string `json:"no_telp"`             // G13
	Alamat            string `json:"alamat"`              // A15
	Domisili          string `json:"domisili"`            // G15
	TinggiBadan       string `json:"tinggi_badan"`        // A17 (165センチメートル)
	BeratBadan        string `json:"berat_badan"`         // C17 (50キログラム)
	MenyetirMobil     string `json:"menyetir_mobil"`      // D17
	Agama             string `json:"agama"`               // G17 (イスラム教, dll)
	GolonganDarah     string `json:"golongan_darah"`      // A19 (O型)
	Alkohol           string `json:"alkohol"`             // C19
	Merokok           string `json:"merokok"`             // D19
	Keahlian          string `json:"keahlian"`            // G19
	Kelebihan         string `json:"kelebihan"`           // C20
	TinggalBerkelompok string `json:"tinggal_berkelompok"` // G20
	Kekurangan        string `json:"kekurangan"`          // C21
	Hobi              string `json:"hobi"`                // G21
	TujuanJepang      string `json:"tujuan_jepang"`       // D23
	TargetTabungan    string `json:"target_tabungan"`     // D24
	RencanaPulang     string `json:"rencana_pulang"`      // D25

	// Pendidikan
	Pendidikan []PendidikanItem `json:"pendidikan"` // Baris 29 - 32

	// Riwayat Pekerjaan
	Pekerjaan []PekerjaanItem `json:"pekerjaan"` // Baris 36 - 38

	// Susunan Keluarga
	Keluarga []KeluargaItem `json:"keluarga"` // Baris 43 - 46
}

type PendidikanItem struct {
	Tingkat string `json:"tingkat"` // SD, SMP, SMA, PT
	Sekolah string `json:"sekolah"` // B
	Jurusan string `json:"jurusan"` // D
	Periode string `json:"periode"` // G
}

type PekerjaanItem struct {
	Perusahaan string `json:"perusahaan"` // A
	Rincian    string `json:"rincian"`    // D
	Periode    string `json:"periode"`    // G
}

type KeluargaItem struct {
	Hubungan  string `json:"hubungan"`  // A
	Nama      string `json:"nama"`      // C
	Pekerjaan string `json:"pekerjaan"` // D
	Usia      string `json:"usia"`      // G
}
