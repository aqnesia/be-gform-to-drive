package translate

import (
	"context"
	"fmt"
)

const PromptVersion = "v2.0-lean"

// TranslationPayload adalah data teks bebas esensial yang dikirim ke LLM (hemat token)
type TranslationPayload struct {
	NamaRomaji     string          `json:"nama_romaji"`
	Panggilan      string          `json:"panggilan"`
	TempatLahir    string          `json:"tempat_lahir"`
	Hobi           string          `json:"hobi"`
	TujuanJepang   string          `json:"tujuan_jepang"`
	TargetTabungan string          `json:"target_tabungan"`
	RencanaPulang  string          `json:"rencana_pulang"`
	Pendidikan     []EduPayload    `json:"pendidikan,omitempty"`
	Pekerjaan      []JobPayload    `json:"pekerjaan,omitempty"`
	Keluarga       []FamilyPayload `json:"keluarga,omitempty"`
}

type EduPayload struct {
	Index   int    `json:"index"`
	Jurusan string `json:"jurusan"`
}

type JobPayload struct {
	Index   int    `json:"index"`
	Rincian string `json:"rincian"`
}

type FamilyPayload struct {
	Index     int    `json:"index"`
	Pekerjaan string `json:"pekerjaan"`
}

// TranslationResponse adalah hasil kembalian ringkas dari LLM dalam bahasa Jepang
type TranslationResponse struct {
	NamaKatakana   string           `json:"nama_katakana"`
	Panggilan      string           `json:"panggilan"`
	TempatLahir    string           `json:"tempat_lahir"`
	Hobi           string           `json:"hobi"`
	TujuanJepang   string           `json:"tujuan_jepang"`
	TargetTabungan string           `json:"target_tabungan"`
	RencanaPulang  string           `json:"rencana_pulang"`
	Pendidikan     []EduResponse    `json:"pendidikan"`
	Pekerjaan      []JobResponse    `json:"pekerjaan"`
	Keluarga       []FamilyResponse `json:"keluarga"`
}

type EduResponse struct {
	Index   int    `json:"index"`
	Jurusan string `json:"jurusan"`
}

type JobResponse struct {
	Index   int    `json:"index"`
	Rincian string `json:"rincian"`
}

type FamilyResponse struct {
	Index     int    `json:"index"`
	Pekerjaan string `json:"pekerjaan"`
}

type Provider interface {
	Name() string
	Model() string
	PromptVersion() string
	Translate(ctx context.Context, payload *TranslationPayload, glossaryPrompt string) (*TranslationResponse, error)
}

func BuildSystemPrompt(glossaryPrompt string) string {
	extra := ""
	if glossaryPrompt != "" {
		extra = "\n" + glossaryPrompt
	}
	return fmt.Sprintf(`Penerjemah formal CV Magang Jepang (技能実習履歴書). Terjemahkan data teks peserta ke bahasa Jepang resume standar:
1. "nama_katakana", "panggilan", "tempat_lahir": Transliterasi Katakana standar (tanda tengah ・).
2. "pendidikan.jurusan", "pekerjaan.rincian": Terjemahkan ke istilah industri Jepang yang lazim.
3. "keluarga.pekerjaan": Terjemahkan pekerjaan keluarga (misal: "Wiraswasta" -> "自営業", "Petani" -> "農業").
4. "hobi", "tujuan_jepang", "target_tabungan", "rencana_pulang": Terjemahkan formal, padat, dan wajar.%s
KEMBALIKAN HANYA FORMAT JSON MURNI sesuai skema tanpa markdown apapun.`, extra)
}
