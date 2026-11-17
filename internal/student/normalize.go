package student

import (
	"fmt"
	"regexp"
	"strings"
)

var digitRegex = regexp.MustCompile(`\d+`)

// NormalizeTinggi converts e.g. "165", "165 cm", "165cm" into "165センチメートル"
func NormalizeTinggi(raw string) string {
	digits := digitRegex.FindString(raw)
	if digits == "" {
		return raw
	}
	return digits + "センチメートル"
}

// NormalizeBerat converts e.g. "50", "50 kg", "50kg" into "50キログラム"
func NormalizeBerat(raw string) string {
	digits := digitRegex.FindString(raw)
	if digits == "" {
		return raw
	}
	return digits + "キログラム"
}

// NormalizeGolonganDarah converts e.g. "O", "Golongan O", "O+" into "O型"
func NormalizeGolonganDarah(raw string) string {
	upper := strings.ToUpper(strings.TrimSpace(raw))
	switch {
	case strings.Contains(upper, "AB"):
		return "AB型"
	case strings.Contains(upper, "A"):
		return "A型"
	case strings.Contains(upper, "B"):
		return "B型"
	case strings.Contains(upper, "O"):
		return "O型"
	default:
		return raw
	}
}

// NormalizeJenisKelamin converts e.g. "Laki-laki", "L", "Pria" to "男" and "Perempuan", "P", "Wanita" to "女"
func NormalizeJenisKelamin(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if strings.Contains(lower, "laki") || strings.Contains(lower, "pria") || lower == "l" || strings.Contains(lower, "男") {
		return "男"
	}
	if strings.Contains(lower, "perempuan") || strings.Contains(lower, "wanita") || lower == "p" || strings.Contains(lower, "女") {
		return "女"
	}
	return raw
}

// NormalizeStatusNikah converts marital status to Japanese
func NormalizeStatusNikah(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if strings.Contains(lower, "belum") || strings.Contains(lower, "lajang") || strings.Contains(lower, "single") || strings.Contains(lower, "未婚") {
		return "未婚"
	}
	if strings.Contains(lower, "menikah") || strings.Contains(lower, "kawin") || strings.Contains(lower, "既婚") {
		return "既婚"
	}
	if strings.Contains(lower, "cerai") || strings.Contains(lower, "duda") || strings.Contains(lower, "janda") || strings.Contains(lower, "離婚") {
		return "離婚"
	}
	return raw
}

// NormalizeAgama converts standard Indonesian religions to Japanese
func NormalizeAgama(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(lower, "islam") || strings.Contains(lower, "muslim"):
		return "イスラム教"
	case strings.Contains(lower, "katolik"):
		return "カトリック"
	case strings.Contains(lower, "kristen") || strings.Contains(lower, "protestan"):
		return "キリスト教"
	case strings.Contains(lower, "hindu"):
		return "ヒンドゥー教"
	case strings.Contains(lower, "buddha") || strings.Contains(lower, "budha"):
		return "仏教"
	default:
		return raw
	}
}

// NormalizeYaTidak converts affirmative/negative values to 有 / 無
func NormalizeYaTidak(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if strings.Contains(lower, "ya") || strings.Contains(lower, "ada") || strings.Contains(lower, "pernah") || strings.Contains(lower, "bisa") || strings.Contains(lower, "punya") || strings.Contains(lower, "merokok") || strings.Contains(lower, "sering") || strings.Contains(lower, "kadang") {
		return "有"
	}
	if strings.Contains(lower, "tidak") || strings.Contains(lower, "bukan") || strings.Contains(lower, "belum") || strings.Contains(lower, "tanpa") || strings.Contains(lower, "mati") {
		return "無"
	}
	return raw
}

// NormalizeUsiaKeluarga converts e.g. "45", "45 tahun" to "45歳"
func NormalizeUsiaKeluarga(raw string) string {
	digits := digitRegex.FindString(raw)
	if digits == "" {
		return raw
	}
	return fmt.Sprintf("%s歳", digits)
}

// NormalizeHubunganKeluarga maps family relations to Japanese
func NormalizeHubunganKeluarga(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.Contains(lower, "ayah") || strings.Contains(lower, "bapak"):
		return "父"
	case strings.Contains(lower, "ibu") || strings.Contains(lower, "mama"):
		return "母"
	case strings.Contains(lower, "kakak") && (strings.Contains(lower, "laki") || strings.Contains(lower, "pria")):
		return "兄"
	case strings.Contains(lower, "kakak") && (strings.Contains(lower, "perempuan") || strings.Contains(lower, "wanita")):
		return "姉"
	case strings.Contains(lower, "kakak"):
		return "兄/姉"
	case strings.Contains(lower, "adik") && (strings.Contains(lower, "laki") || strings.Contains(lower, "pria")):
		return "弟"
	case strings.Contains(lower, "adik") && (strings.Contains(lower, "perempuan") || strings.Contains(lower, "wanita")):
		return "妹"
	case strings.Contains(lower, "adik"):
		return "弟/妹"
	case strings.Contains(lower, "suami"):
		return "夫"
	case strings.Contains(lower, "istri"):
		return "妻"
	case strings.Contains(lower, "anak"):
		return "子"
	default:
		return raw
	}
}
