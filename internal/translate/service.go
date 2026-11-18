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
		p = NewOpenAIProvider(cfg.LLMAPIKey, cfg.LLMModel, cfg.LLMBaseURL)
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

	// Siapkan payload esensial untuk LLM (sangat hemat token)
	payload := &TranslationPayload{
		NamaRomaji:     cv.NamaRomaji,
		Panggilan:      cv.PanggilanKatakana,
		TempatLahir:    cv.TempatLahir,
		Hobi:           cv.Hobi,
		TujuanJepang:   cv.TujuanJepang,
		TargetTabungan: cv.TargetTabungan,
		RencanaPulang:  cv.RencanaPulang,
	}

	for i, edu := range cv.Pendidikan {
		if strings.TrimSpace(edu.Jurusan) != "" {
			payload.Pendidikan = append(payload.Pendidikan, EduPayload{
				Index:   i,
				Jurusan: edu.Jurusan,
			})
		}
	}

	for i, job := range cv.Pekerjaan {
		if strings.TrimSpace(job.Rincian) != "" {
			payload.Pekerjaan = append(payload.Pekerjaan, JobPayload{
				Index:   i,
				Rincian: job.Rincian,
			})
		}
	}

	for i, fam := range cv.Keluarga {
		// Cek dulu apakah pekerjaan keluarga ada di kamus glosarium baku lokal
		if val, ok := s.glossary.Lookup(fam.Pekerjaan); ok {
			cv.Keluarga[i].Pekerjaan = val
		} else if strings.TrimSpace(fam.Pekerjaan) != "" {
			payload.Keluarga = append(payload.Keluarga, FamilyPayload{
				Index:     i,
				Pekerjaan: fam.Pekerjaan,
			})
		}
	}

	// Panggil LLM dengan prompt ringkas
	resp, err := s.provider.Translate(ctx, payload, "")
	if err != nil {
		return nil, "", "", "", fmt.Errorf("error terjemahan LLM (%s): %w", s.provider.Name(), err)
	}

	// Terapkan hasil terjemahan
	if resp.NamaKatakana != "" {
		cv.NamaKatakana = resp.NamaKatakana
	}

	// 1. Format nama panggilan disamakan dengan nama lengkap (Katakana atas, Romaji kapital bawah)
	katakana := strings.TrimSpace(resp.Panggilan)
	if katakana == "" {
		katakana = strings.TrimSpace(cv.PanggilanKatakana)
	}
	romaji := strings.TrimSpace(cv.PanggilanRomaji)
	if romaji == "" {
		fields := strings.Fields(cv.NamaRomaji)
		if len(fields) > 0 {
			romaji = fields[0]
		}
	}
	if katakana != "" && romaji != "" {
		cv.Panggilan = fmt.Sprintf("%s\n%s", katakana, strings.ToUpper(romaji))
	} else if katakana != "" {
		cv.Panggilan = katakana
	}

	if resp.TempatLahir != "" {
		cv.TempatLahir = resp.TempatLahir
	}

	// Sesuai requirement:
	// - Alamat (現住所): tetap gunakan bahasa Indonesia asli (jangan digenerate)
	// - Keahlian (得意・特技), Kelebihan (長所), Kekurangan (短所): samakan dengan CV upload (jangan digenerate)

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
			// Sesuai requirement: Nama perusahaan jangan digenerate oleh AI
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
