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

	g4, _ := outF.GetCellValue("Template", "G4")
	if g4 == "" {
		t.Errorf("sel G4 (作成日) tidak boleh kosong")
	}

	t.Logf("Sukses generate! G4 = %s", g4)
}
