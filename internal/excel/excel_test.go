package excel

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"automate-gform-drive/internal/student"
	"github.com/xuri/excelize/v2"
)

func TestReadAndGenerate(t *testing.T) {
	inputPath := filepath.Join("..", "..", "templates", "cv_template.xlsx")
	outputPath := filepath.Join("..", "..", "tmp", "test_output.xlsx")
	templatePath := filepath.Join("..", "..", "templates", "output-template-cv.xlsx")

	fileData, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("gagal membaca input file: %v", err)
	}

	cv, err := ReadFromBytes(fileData)
	if err != nil {
		t.Fatalf("gagal parse CV: %v", err)
	}

	cv.SubmitTime = time.Date(2026, 11, 17, 18, 0, 0, 0, time.Local)
	// Normalisasi
	cv.NamaRomaji = "GALANG WAHYU SETIAWAN"
	cv.NamaKatakana = "ガラン・ワヒュ・セティアワン"
	cv.PanggilanKatakana = "ガラン"
	cv.PanggilanRomaji = "GALANG"
	cv.JenisKelamin = student.NormalizeJenisKelamin(cv.JenisKelamin)
	cv.TinggiBadan = student.NormalizeTinggi(cv.TinggiBadan)
	cv.BeratBadan = student.NormalizeBerat(cv.BeratBadan)
	cv.GolonganDarah = student.NormalizeGolonganDarah(cv.GolonganDarah)
	cv.Agama = student.NormalizeAgama(cv.Agama)
	cv.StatusPernikahan = student.NormalizeStatusNikah(cv.StatusPernikahan)

	err = GenerateCV(templatePath, cv, outputPath)
	if err != nil {
		t.Fatalf("gagal generate CV: %v", err)
	}

	// Verify generated file
	outF, err := excelize.OpenFile(outputPath)
	if err != nil {
		t.Fatalf("gagal membuka hasil generate: %v", err)
	}
	defer outF.Close()

	sheet := "Template"
	g4, _ := outF.GetCellValue(sheet, "G4")
	if g4 == "" {
		t.Errorf("sel G4 (作成日) tidak boleh kosong")
	}

	// 1. Kotak kuning G1 dan G2 harus kosong
	g1, _ := outF.GetCellValue(sheet, "G1")
	g2, _ := outF.GetCellValue(sheet, "G2")
	if g1 != "" || g2 != "" {
		t.Errorf("Kotak kuning G1/G2 harus kosong, tapi dapat G1=%q G2=%q", g1, g2)
	}

	// 2. Nama Lengkap vs Panggilan
	c7, _ := outF.GetCellValue(sheet, "C7")
	c8, _ := outF.GetCellValue(sheet, "C8")
	g7, _ := outF.GetCellValue(sheet, "G7")
	g8, _ := outF.GetCellValue(sheet, "G8")
	if c7 != "ガラン・ワヒュ・セティアワン" {
		t.Errorf("C7 salah: %q", c7)
	}
	if c8 != "GALANG WAHYU SETIAWAN" {
		t.Errorf("C8 salah: %q", c8)
	}
	if g7 != "ガラン" {
		t.Errorf("G7 (Panggilan Katakana) salah: %q", g7)
	}
	if g8 != "GALANG" {
		t.Errorf("G8 (Panggilan Romaji) salah: %q", g8)
	}

	// 3. Border garis putus-putus di bawah 集団生活経験 (D20:F20)
	sD20, _ := outF.GetCellStyle(sheet, "D20")
	stD20, _ := outF.GetStyle(sD20)
	hasBottomBorder := false
	for _, b := range stD20.Border {
		if b.Type == "bottom" && b.Style == 7 && b.Color == "000000" {
			hasBottomBorder = true
		}
	}
	if !hasBottomBorder {
		t.Errorf("D20 harus memiliki border bottom style 7 (dashed) hitam")
	}

	// 4. Border garis kiri 会社名 (A35) dan baris bawahnya (A36, A37, A38)
	for _, cell := range []string{"A35", "A36", "A37", "A38"} {
		sID, _ := outF.GetCellStyle(sheet, cell)
		st, _ := outF.GetStyle(sID)
		hasLeftBorder := false
		for _, b := range st.Border {
			if b.Type == "left" && b.Style == 2 && b.Color == "000000" {
				hasLeftBorder = true
			}
		}
		if !hasLeftBorder {
			t.Errorf("Sel %s harus memiliki border left solid medium (style 2) hitam", cell)
		}
	}

	t.Logf("Sukses generate dan verifikasi 4 poin revisi! G4 = %s", g4)
}
