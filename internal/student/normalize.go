package student

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var digitRegex = regexp.MustCompile(`\d+`)

var monthMap = map[string]string{
	"januari": "01", "jan": "01",
	"februari": "02", "feb": "02",
	"maret": "03", "mar": "03",
	"april": "04", "apr": "04",
	"mei": "05", "may": "05",
	"juni": "06", "jun": "06",
	"juli": "07", "jul": "07",
	"agustus": "08", "agu": "08", "agt": "08", "aug": "08",
	"september": "09", "sep": "09",
	"oktober": "10", "okt": "10", "oct": "10",
	"november": "11", "nov": "11",
	"desember": "12", "des": "12", "dec": "12",
}

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

// NormalizeStatusNikah converts marital status to Japanese kanji (未婚 / 既婚 / 離婚)
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
	case strings.Contains(lower, "konghucu"):
		return "儒教"
	default:
		return raw
	}
}

// NormalizeMenyetir converts driving ability to pure Japanese (できる / できない)
func NormalizeMenyetir(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if strings.Contains(lower, "tidak") || strings.Contains(lower, "belum") || strings.Contains(lower, "bukan") || strings.Contains(lower, "できない") {
		return "できない"
	}
	if strings.Contains(lower, "bisa") || strings.Contains(lower, "ya") || strings.Contains(lower, "sim") || strings.Contains(lower, "できる") {
		return "できる"
	}
	if raw == "" {
		return ""
	}
	return "できない"
}

// NormalizeTinggalBerkelompok converts group living experience to pure Japanese (有 / 無)
func NormalizeTinggalBerkelompok(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if strings.Contains(lower, "tidak") || strings.Contains(lower, "belum") || strings.Contains(lower, "bukan") || strings.Contains(lower, "無") {
		return "無"
	}
	if strings.Contains(lower, "pernah") || strings.Contains(lower, "ya") || strings.Contains(lower, "ada") || strings.Contains(lower, "bisa") || strings.Contains(lower, "有") {
		return "有"
	}
	return raw
}

// NormalizeYaTidak converts affirmative/negative values to 有 / 無 (Alkohol & Merokok)
func NormalizeYaTidak(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if strings.Contains(lower, "tidak") || strings.Contains(lower, "bukan") || strings.Contains(lower, "belum") || strings.Contains(lower, "tanpa") || strings.Contains(lower, "無") || strings.Contains(lower, "いいえ") {
		return "無"
	}
	if strings.Contains(lower, "ya") || strings.Contains(lower, "ada") || strings.Contains(lower, "pernah") || strings.Contains(lower, "bisa") || strings.Contains(lower, "punya") || strings.Contains(lower, "merokok") || strings.Contains(lower, "sering") || strings.Contains(lower, "kadang") || strings.Contains(lower, "有") || strings.Contains(lower, "はい") {
		return "有"
	}
	return raw
}

// NormalizeUsia ensures the age is always in "XX歳" format without formula errors
func NormalizeUsia(rawUsia, rawTglLahir string, now time.Time) string {
	rawUsia = strings.TrimSpace(rawUsia)
	if rawUsia != "" && !strings.Contains(rawUsia, "#") {
		digits := digitRegex.FindString(rawUsia)
		if digits != "" {
			return digits + "歳"
		}
	}

	// Calculate from birth date if rawUsia is missing or formula error
	re := regexp.MustCompile(`(\d{4})[^\d]+(\d{1,2})[^\d]+(\d{1,2})`)
	if m := re.FindStringSubmatch(rawTglLahir); len(m) == 4 {
		y, _ := strconv.Atoi(m[1])
		mth, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		birth := time.Date(y, time.Month(mth), d, 0, 0, 0, 0, time.UTC)
		age := now.Year() - birth.Year()
		if now.YearDay() < birth.YearDay() {
			age--
		}
		if age > 0 && age < 120 {
			return fmt.Sprintf("%d歳", age)
		}
	}

	return rawUsia
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

func parseDatePart(part string) string {
	part = strings.TrimSpace(part)
	lower := strings.ToLower(part)
	if lower == "sekarang" || lower == "saat ini" || lower == "present" || lower == "now" || lower == "現在" {
		return "現在"
	}

	// Check if already in Japanese format: 2023年 11月 or 2023年11月
	jpRe := regexp.MustCompile(`(\d{4})\s*年\s*(\d{1,2})\s*月`)
	if m := jpRe.FindStringSubmatch(part); len(m) > 0 {
		month, _ := strconv.Atoi(m[2])
		return fmt.Sprintf("%s年 %02d月", m[1], month)
	}

	// Month Name Year: "Maret 2025"
	for name, num := range monthMap {
		re := regexp.MustCompile(fmt.Sprintf(`(?i)\b%s\b[^\d]*(\d{4})`, name))
		if m := re.FindStringSubmatch(part); len(m) > 0 {
			return fmt.Sprintf("%s年 %s月", m[1], num)
		}
		reRev := regexp.MustCompile(fmt.Sprintf(`(?i)(\d{4})[^\d]*\b%s\b`, name))
		if m := reRev.FindStringSubmatch(part); len(m) > 0 {
			return fmt.Sprintf("%s年 %s月", m[1], num)
		}
	}

	// Numeric: MM/YYYY or MM-YYYY or YYYY/MM or YYYY-MM
	numRe1 := regexp.MustCompile(`(\d{1,2})[-/](\d{4})`)
	if m := numRe1.FindStringSubmatch(part); len(m) > 0 {
		month, _ := strconv.Atoi(m[1])
		return fmt.Sprintf("%s年 %02d月", m[2], month)
	}
	numRe2 := regexp.MustCompile(`(\d{4})[-/](\d{1,2})`)
	if m := numRe2.FindStringSubmatch(part); len(m) > 0 {
		month, _ := strconv.Atoi(m[2])
		return fmt.Sprintf("%s年 %02d月", m[1], month)
	}

	return part
}

// NormalizePeriode converts period strings (including "maret 2025 - sekarang") to standardized Japanese format
func NormalizePeriode(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var sep string
	for _, s := range []string{"～", "~", " - ", "-", " s/d ", " sampai ", " to "} {
		if strings.Contains(raw, s) {
			sep = s
			break
		}
	}

	if sep == "" {
		return parseDatePart(raw)
	}

	parts := strings.SplitN(raw, sep, 2)
	p1 := parseDatePart(parts[0])
	p2 := parseDatePart(parts[1])

	return fmt.Sprintf("%s ～ %s", p1, p2)
}
