package excel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"automate-gform-drive/internal/student"
	"github.com/xuri/excelize/v2"
)

// GenerateCV mengisi template resmi (output-template-cv.xlsx) dengan data CV yang sudah diterjemahkan
func GenerateCV(templatePath string, cv *student.CVData, outputPath string) error {
	f, err := excelize.OpenFile(templatePath)
	if err != nil {
		return fmt.Errorf("gagal membuka template excel: %w", err)
	}
	defer f.Close()

	sheetName := "Template"
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return fmt.Errorf("berkas template tidak memiliki sheet")
	}

	found := false
	for _, name := range sheetList {
		if strings.TrimSpace(name) == sheetName {
			found = true
			break
		}
	}
	if !found {
		sheetName = sheetList[0]
	}

	setVal := func(axis, val string) {
		if val != "" {
			_ = f.SetCellValue(sheetName, axis, val)
		}
	}

	// 0. Nomor Registrasi / Urut pada Kotak Kuning (G1:G2)
	if cv.NomorRegistrasi != "" {
		setVal("G1", cv.NomorRegistrasi)
	}

	// 1. Tanggal Pembuatan (作成日 - G4) sesuai waktu submit
	submitTime := cv.SubmitTime
	if submitTime.IsZero() {
		submitTime = time.Now()
	}
	formattedDate := fmt.Sprintf("%d年 %02d月 %02d日", submitTime.Year(), submitTime.Month(), submitTime.Day())
	setVal("G4", formattedDate)

	// 2. Foto Peserta (A5:B11)
	if len(cv.FotoBytes) > 0 {
		ext := cv.FotoExt
		if ext == "" {
			ext = ".jpg"
		}
		_ = f.AddPictureFromBytes(sheetName, "A5", &excelize.Picture{
			Extension: ext,
			File:      cv.FotoBytes,
			Format: &excelize.GraphicOptions{
				AutoFit: true,
				OffsetX: 8,
				OffsetY: 8,
			},
		})
	}

	// 3. Data Diri
	setVal("C7", cv.NamaKatakana)
	setVal("C8", cv.NamaRomaji)

	// Format Panggilan (Katakana atas, Romaji kapital bawah pada G8)
	if cv.Panggilan != "" {
		setVal("G8", cv.Panggilan)
	}

	// Format Tanggal Lahir (C11)
	if cv.TanggalLahir != "" {
		setVal("C11", cv.TanggalLahir)
	}

	// Usia (G11) - langsung tulis agar tidak terjadi error #VALUE!
	if cv.Usia != "" {
		setVal("G11", cv.Usia)
	}

	setVal("A13", cv.JenisKelamin)
	setVal("C13", cv.TempatLahir)
	setVal("D13", cv.StatusPernikahan)
	setVal("G13", cv.NoTelp)
	setVal("A15", cv.Alamat)
	setVal("G15", cv.Domisili)
	setVal("A17", cv.TinggiBadan)
	setVal("C17", cv.BeratBadan)
	setVal("D17", cv.MenyetirMobil)
	setVal("G17", cv.Agama)
	setVal("A19", cv.GolonganDarah)
	setVal("C19", cv.Alkohol)
	setVal("D19", cv.Merokok)
	setVal("G19", cv.Keahlian)
	setVal("C20", cv.Kelebihan)
	setVal("G20", cv.TinggalBerkelompok)
	setVal("C21", cv.Kekurangan)
	setVal("G21", cv.Hobi)
	setVal("D23", cv.TujuanJepang)
	setVal("D24", cv.TargetTabungan)
	setVal("D25", cv.RencanaPulang)

	// 4. Pendidikan (Maksimal 4 baris: SD 29, SMP 30, SMA 31, PT 32)
	eduRows := []int{29, 30, 31, 32}
	for i, edu := range cv.Pendidikan {
		if i >= len(eduRows) {
			break
		}
		row := eduRows[i]
		setVal(fmt.Sprintf("B%d", row), edu.Sekolah)
		setVal(fmt.Sprintf("D%d", row), edu.Jurusan)
		setVal(fmt.Sprintf("G%d", row), edu.Periode)
	}

	// Rapikan font style periode pekerjaan: samakan style G37 dan G38 dengan G36 agar tidak berantakan
	styleG36, err := f.GetCellStyle(sheetName, "G36")
	if err == nil && styleG36 != 0 {
		_ = f.SetCellStyle(sheetName, "G37", "G37", styleG36)
		_ = f.SetCellStyle(sheetName, "G38", "G38", styleG36)
	}

	// 5. Riwayat Pekerjaan (Maksimal 3 baris: 36, 37, 38)
	workRows := []int{36, 37, 38}
	for i, work := range cv.Pekerjaan {
		if i >= len(workRows) {
			break
		}
		row := workRows[i]
		setVal(fmt.Sprintf("A%d", row), work.Perusahaan)
		setVal(fmt.Sprintf("D%d", row), work.Rincian)
		setVal(fmt.Sprintf("G%d", row), work.Periode)
	}

	// 6. Susunan Keluarga (Maksimal 4 baris: 43, 44, 45, 46)
	famRows := []int{43, 44, 45, 46}
	for i, fam := range cv.Keluarga {
		if i >= len(famRows) {
			break
		}
		row := famRows[i]
		setVal(fmt.Sprintf("A%d", row), fam.Hubungan)
		setVal(fmt.Sprintf("C%d", row), fam.Nama)
		setVal(fmt.Sprintf("D%d", row), fam.Pekerjaan)
		setVal(fmt.Sprintf("G%d", row), fam.Usia)
	}

	// 7. Hasil Ujian Awal (事前試験結果 - Baris 50)
	setVal("A50", cv.HasilUjian.BahasaJepang)
	setVal("C50", cv.HasilUjian.Matematika)
	setVal("D50", cv.HasilUjian.PushUp)
	setVal("E50", cv.HasilUjian.SitUp)
	setVal("G50", cv.HasilUjian.Grip)

	// Pastikan direktori tujuan tersedia
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("gagal membuat folder output: %w", err)
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("gagal menyimpan berkas output excel: %w", err)
	}

	return nil
}

// ReadGeneratedBytes membaca byte dari file output yang telah dibuat
func ReadGeneratedBytes(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
