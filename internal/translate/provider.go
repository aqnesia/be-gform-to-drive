package translate

import (
	"context"
	"fmt"
)

const PromptVersion = "v1.0"

// TranslationPayload adalah data teks bebas yang dikirim ke LLM
type TranslationPayload struct {
	NamaRomaji     string           `json:"nama_romaji"`
	Panggilan      string           `json:"panggilan"`
	TempatLahir    string           `json:"tempat_lahir"`
	Alamat         string           `json:"alamat"`
	Domisili       string           `json:"domisili"`
	Keahlian       string           `json:"keahlian"`
	Kelebihan      string           `json:"kelebihan"`
	Kekurangan     string           `json:"kekurangan"`
	Hobi           string           `json:"hobi"`
	TujuanJepang   string           `json:"tujuan_jepang"`
	TargetTabungan string           `json:"target_tabungan"`
	RencanaPulang  string           `json:"rencana_pulang"`
	Pendidikan     []EduPayload     `json:"pendidikan"`
	Pekerjaan      []JobPayload     `json:"pekerjaan"`
	Keluarga       []FamilyPayload  `json:"keluarga"`
}

type EduPayload struct {
	Index   int    `json:"index"`
	Jurusan string `json:"jurusan"`
}

type JobPayload struct {
	Index      int    `json:"index"`
	Perusahaan string `json:"perusahaan"`
	Rincian    string `json:"rincian"`
}

type FamilyPayload struct {
	Index     int    `json:"index"`
	Pekerjaan string `json:"pekerjaan"`
}

// TranslationResponse adalah hasil kembalian dari LLM dalam bahasa Jepang
type TranslationResponse struct {
	NamaKatakana   string              `json:"nama_katakana"`
	Panggilan      string              `json:"panggilan"`
	TempatLahir    string              `json:"tempat_lahir"`
	Alamat         string              `json:"alamat"`
	Domisili       string              `json:"domisili"`
	Keahlian       string              `json:"keahlian"`
	Kelebihan      string              `json:"kelebihan"`
	Kekurangan     string              `json:"kekurangan"`
	Hobi           string              `json:"hobi"`
	TujuanJepang   string              `json:"tujuan_jepang"`
	TargetTabungan string              `json:"target_tabungan"`
	RencanaPulang  string              `json:"rencana_pulang"`
	Pendidikan     []EduResponse       `json:"pendidikan"`
	Pekerjaan      []JobResponse       `json:"pekerjaan"`
	Keluarga       []FamilyResponse    `json:"keluarga"`
}

type EduResponse struct {
	Index   int    `json:"index"`
	Jurusan string `json:"jurusan"`
}

type JobResponse struct {
	Index      int    `json:"index"`
	Perusahaan string `json:"perusahaan"`
	Rincian    string `json:"rincian"`
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
	return fmt.Sprintf(`Anda adalah penerjemah profesional untuk formulir CV Magang Jepang (技能実習履歴書).
Tugas Anda adalah menerjemahkan dan mengonversi data teks peserta Indonesia ke dalam bahasa Jepang formal standar resume Jepang.

Aturan Penting:
1. "nama_katakana": Transliterasikan nama lengkap peserta ke Katakana standar Jepang (pisahkan kata dengan tanda tengah ・ misal: "ムハンマド・リズキ").
2. "tempat_lahir", "domisili": Transliterasikan nama kota/daerah ke Katakana atau gunakan ejaan Romaji bersih jika nama kota kecil.
3. "kelebihan", "kekurangan", "hobi", "tujuan_jepang", "rencana_pulang": Terjemahkan ke bahasa Jepang yang sopan, padat, dan wajar untuk resume magang/kerja.
4. "pekerjaan": Terjemahkan rincian kerja dan posisi ke istilah kerja Jepang yang lazim (misal: "Operator Mesin" -> "機械オペレーター", "Petani" -> "農業").
5. "keluarga": Terjemahkan pekerjaan keluarga ke istilah baku (misal: "Ibu Rumah Tangga" -> "主婦", "Petani" -> "農業", "Wiraswasta" -> "自営業").
6. %s
7. KEMBALIKAN HANYA FORMAT JSON MURNI yang sesuai dengan skema permintaan tanpa tambahan teks markdown apapun.`, glossaryPrompt)
}
