package translate

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"automate-gform-drive/internal/config"
	"automate-gform-drive/internal/student"
)

type Service struct {
	provider Provider
	glossary *Glossary
}

func NewService(cfg *config.Config, glossaryPath string) (*Service, error) {
	var glossary *Glossary
	if glossaryPath != "" {
		g, err := LoadGlossary(glossaryPath)
		if err != nil {
			slog.Warn("gagal memuat glossary, melanjutkan tanpa glossary", "error", err)
		} else {
			glossary = g
		}
	}

	var p Provider
	switch strings.ToLower(cfg.LLMProvider) {
	case "openai":
		p = NewOpenAIProvider(cfg.LLMAPIKey, cfg.LLMModel)
	case "gemini":
		fallthrough
	default:
		p = NewGeminiProvider(cfg.LLMAPIKey, cfg.LLMModel)
	}

	return &Service{
		provider: p,
		glossary: glossary,
	}, nil
}

// TranslateCV menerjemahkan teks bebas pada data CV dan mengembalikan objek CVData yang telah diperbarui
func (s *Service) TranslateCV(ctx context.Context, cv *student.CVData) (*student.CVData, string, string, string, error) {
	if s.provider == nil {
		return cv, "none", "none", PromptVersion, nil
	}

	// Siapkan payload untuk LLM
	payload := &TranslationPayload{
		NamaRomaji:     cv.NamaRomaji,
		Panggilan:      cv.Panggilan,
		TempatLahir:    cv.TempatLahir,
		Alamat:         cv.Alamat,
		Domisili:       cv.Domisili,
		Keahlian:       cv.Keahlian,
		Kelebihan:      cv.Kelebihan,
		Kekurangan:     cv.Kekurangan,
		Hobi:           cv.Hobi,
		TujuanJepang:   cv.TujuanJepang,
		TargetTabungan: cv.TargetTabungan,
		RencanaPulang:  cv.RencanaPulang,
	}

	for i, edu := range cv.Pendidikan {
		payload.Pendidikan = append(payload.Pendidikan, EduPayload{
			Index:   i,
			Jurusan: edu.Jurusan,
		})
	}

	for i, job := range cv.Pekerjaan {
		payload.Pekerjaan = append(payload.Pekerjaan, JobPayload{
			Index:      i,
			Perusahaan: job.Perusahaan,
			Rincian:    job.Rincian,
		})
	}

	for i, fam := range cv.Keluarga {
		// Cek dulu apakah pekerjaan keluarga ada di kamus glosarium baku
		if val, ok := s.glossary.Lookup(fam.Pekerjaan); ok {
			cv.Keluarga[i].Pekerjaan = val
		} else {
			payload.Keluarga = append(payload.Keluarga, FamilyPayload{
				Index:     i,
				Pekerjaan: fam.Pekerjaan,
			})
		}
	}

	glossaryPrompt := ""
	if s.glossary != nil {
		glossaryPrompt = s.glossary.FormatPromptList()
	}

	// Panggil LLM
	resp, err := s.provider.Translate(ctx, payload, glossaryPrompt)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("error terjemahan LLM (%s): %w", s.provider.Name(), err)
	}

	// Terapkan hasil terjemahan
	if resp.NamaKatakana != "" {
		cv.NamaKatakana = resp.NamaKatakana
	}
	if resp.Panggilan != "" {
		cv.Panggilan = resp.Panggilan
	}
	if resp.TempatLahir != "" {
		cv.TempatLahir = resp.TempatLahir
	}
	if resp.Alamat != "" {
		cv.Alamat = resp.Alamat
	}
	if resp.Domisili != "" {
		cv.Domisili = resp.Domisili
	}
	if resp.Keahlian != "" {
		cv.Keahlian = resp.Keahlian
	}
	if resp.Kelebihan != "" {
		cv.Kelebihan = resp.Kelebihan
	}
	if resp.Kekurangan != "" {
		cv.Kekurangan = resp.Kekurangan
	}
	if resp.Hobi != "" {
		cv.Hobi = resp.Hobi
	}
	if resp.TujuanJepang != "" {
		cv.TujuanJepang = resp.TujuanJepang
	}
	if resp.TargetTabungan != "" {
		cv.TargetTabungan = resp.TargetTabungan
	}
	if resp.RencanaPulang != "" {
		cv.RencanaPulang = resp.RencanaPulang
	}

	for _, eduResp := range resp.Pendidikan {
		if eduResp.Index >= 0 && eduResp.Index < len(cv.Pendidikan) {
			cv.Pendidikan[eduResp.Index].Jurusan = eduResp.Jurusan
		}
	}

	for _, jobResp := range resp.Pekerjaan {
		if jobResp.Index >= 0 && jobResp.Index < len(cv.Pekerjaan) {
			if jobResp.Perusahaan != "" {
				cv.Pekerjaan[jobResp.Index].Perusahaan = jobResp.Perusahaan
			}
			if jobResp.Rincian != "" {
				cv.Pekerjaan[jobResp.Index].Rincian = jobResp.Rincian
			}
		}
	}

	for _, famResp := range resp.Keluarga {
		if famResp.Index >= 0 && famResp.Index < len(cv.Keluarga) {
			if cv.Keluarga[famResp.Index].Pekerjaan == "" || !s.isAlreadyJapanese(cv.Keluarga[famResp.Index].Pekerjaan) {
				cv.Keluarga[famResp.Index].Pekerjaan = famResp.Pekerjaan
			}
		}
	}

	return cv, s.provider.Name(), s.provider.Model(), s.provider.PromptVersion(), nil
}

func (s *Service) isAlreadyJapanese(str string) bool {
	if s.glossary == nil {
		return false
	}
	for _, v := range s.glossary.terms {
		if str == v {
			return true
		}
	}
	return false
}
