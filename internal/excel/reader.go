package excel

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"automate-gform-drive/internal/student"
	"github.com/xuri/excelize/v2"
)

// ReadFromBytes membaca berkas template isian peserta dari byte slice (misal hasil decode base64)
func ReadFromBytes(data []byte) (*student.CVData, error) {
	reader := bytes.NewReader(data)
	return Read(reader)
}

// Read membaca berkas template isian peserta dari io.Reader
func Read(r io.Reader) (*student.CVData, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka berkas excel: %w", err)
	}
	defer f.Close()

	sheetName := "CV Kosong"
	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("berkas excel tidak memiliki sheet")
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

	getVal := func(axis string) string {
		val, _ := f.GetCellValue(sheetName, axis)
		return strings.TrimSpace(val)
	}

	valG7 := getVal("G7")
	valG8 := getVal("G8")
	var pKatakana, pRomaji string
	if valG7 != "" {
		pKatakana = valG7
		pRomaji = valG8
	} else if strings.Contains(valG8, "\n") {
		parts := strings.SplitN(valG8, "\n", 2)
		pKatakana = strings.TrimSpace(parts[0])
		pRomaji = strings.TrimSpace(parts[1])
	} else {
		pKatakana = valG8
	}

	namaRomaji := getVal("C8")
	if pRomaji == "" {
		fields := strings.Fields(namaRomaji)
		if len(fields) > 0 {
			pRomaji = fields[0]
		}
	}

	rawPanggilan := valG8
	if rawPanggilan == "" {
		rawPanggilan = valG7
	}

	cv := &student.CVData{
		NomorRegistrasi:    getVal("G1"),
		NamaRomaji:         namaRomaji,
		NamaKatakana:       getVal("C7"),
		Panggilan:          rawPanggilan,
		PanggilanKatakana:  pKatakana,
		PanggilanRomaji:    pRomaji,
		TanggalLahir:       getVal("C11"),
		Usia:               getVal("G11"),
		JenisKelamin:       getVal("A13"),
		TempatLahir:        getVal("C13"),
		StatusPernikahan:   getVal("D13"),
		NoTelp:             getVal("G13"),
		Alamat:             getVal("A15"),
		Domisili:           getVal("G15"),
		TinggiBadan:        getVal("A17"),
		BeratBadan:         getVal("C17"),
		MenyetirMobil:      getVal("D17"),
		Agama:              getVal("G17"),
		GolonganDarah:      getVal("A19"),
		Alkohol:            getVal("C19"),
		Merokok:            getVal("D19"),
		Keahlian:           getVal("G19"),
		Kelebihan:          getVal("C20"),
		TinggalBerkelompok: getVal("G20"),
		Kekurangan:         getVal("C21"),
		Hobi:               getVal("G21"),
		TujuanJepang:       getVal("D23"),
		TargetTabungan:     getVal("D24"),
		RencanaPulang:      getVal("D25"),
	}

	if cv.NomorRegistrasi == "" {
		cv.NomorRegistrasi = getVal("G2")
	}

	// Ekstrak Foto Peserta jika ada di dalam Excel
	for r := 1; r <= 15; r++ {
		for _, col := range []string{"A", "B", "C"} {
			cell := fmt.Sprintf("%s%d", col, r)
			pics, err := f.GetPictures(sheetName, cell)
			if err == nil && len(pics) > 0 {
				cv.FotoBytes = pics[0].File
				cv.FotoExt = pics[0].Extension
				break
			}
		}
		if len(cv.FotoBytes) > 0 {
			break
		}
	}

	// Pendidikan (SD: 29, SMP: 30, SMA: 31, PT: 32)
	eduRows := []struct {
		Tingkat string
		Row     int
	}{
		{"SD", 29},
		{"SMP", 30},
		{"SMA", 31},
		{"PT", 32},
	}
	for _, edu := range eduRows {
		sekolah := getVal(fmt.Sprintf("B%d", edu.Row))
		jurusan := getVal(fmt.Sprintf("D%d", edu.Row))
		periode := getVal(fmt.Sprintf("G%d", edu.Row))
		if sekolah != "" || periode != "" {
			cv.Pendidikan = append(cv.Pendidikan, student.PendidikanItem{
				Tingkat: edu.Tingkat,
				Sekolah: sekolah,
				Jurusan: jurusan,
				Periode: periode,
			})
		}
	}

	// Pekerjaan (Baris 36, 37, 38)
	for r := 36; r <= 38; r++ {
		perusahaan := getVal(fmt.Sprintf("A%d", r))
		rincian := getVal(fmt.Sprintf("D%d", r))
		periode := getVal(fmt.Sprintf("G%d", r))
		if perusahaan != "" || rincian != "" {
			cv.Pekerjaan = append(cv.Pekerjaan, student.PekerjaanItem{
				Perusahaan: perusahaan,
				Rincian:    rincian,
				Periode:    periode,
			})
		}
	}

	// Keluarga (Baris 43, 44, 45, 46)
	for r := 43; r <= 46; r++ {
		hubungan := getVal(fmt.Sprintf("A%d", r))
		nama := getVal(fmt.Sprintf("C%d", r))
		pekerjaan := getVal(fmt.Sprintf("D%d", r))
		usia := getVal(fmt.Sprintf("G%d", r))
		if nama != "" || hubungan != "" {
			cv.Keluarga = append(cv.Keluarga, student.KeluargaItem{
				Hubungan:  hubungan,
				Nama:      nama,
				Pekerjaan: pekerjaan,
				Usia:      usia,
			})
		}
	}

	// Hasil Ujian Awal (事前試験結果 - Baris 50-53 pada template input)
	cv.HasilUjian = student.HasilUjianData{
		Matematika:   getVal("A52"),
		BahasaJepang: getVal("C52"),
		PushUp:       getVal("D52"),
		SitUp:        getVal("E52"),
		Grip:         getVal("G52"),
	}
	// Fallback jika baris bergeser
	if cv.HasilUjian.Matematika == "" && cv.HasilUjian.BahasaJepang == "" {
		for r := 49; r <= 54; r++ {
			vA := getVal(fmt.Sprintf("A%d", r))
			if strings.Contains(vA, "点") || (len(vA) > 0 && vA[0] >= '0' && vA[0] <= '9') {
				cv.HasilUjian.Matematika = vA
				cv.HasilUjian.BahasaJepang = getVal(fmt.Sprintf("C%d", r))
				cv.HasilUjian.PushUp = getVal(fmt.Sprintf("D%d", r))
				cv.HasilUjian.SitUp = getVal(fmt.Sprintf("E%d", r))
				cv.HasilUjian.Grip = getVal(fmt.Sprintf("G%d", r))
				break
			}
		}
	}

	return cv, nil
}
