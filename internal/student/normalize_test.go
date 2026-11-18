package student

import (
	"testing"
	"time"
)

func TestNormalizeMenyetir(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"bisa", "できる"},
		{"Tidak Bisa", "できない"},
		{"belum bisa", "できない"},
		{"punya sim", "できる"},
		{"できない Tidak Bisa", "できない"},
		{"できる", "できる"},
	}

	for _, tt := range tests {
		got := NormalizeMenyetir(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeMenyetir(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNormalizePeriode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"maret 2025 - sekarang", "2025年 03月 ～ 現在"},
		{"2019年 03月 ~ 2019年 04月", "2019年 03月 ～ 2019年 04月"},
		{"2023年 11月 ~ 2024年 04月", "2023年 11月 ～ 2024年 04月"},
		{"06/2020 - 06/2023", "2020年 06月 ～ 2023年 06月"},
		{"Januari 2021 sampai sekarang", "2021年 01月 ～ 現在"},
		{"2025-03 - present", "2025年 03月 ～ 現在"},
	}

	for _, tt := range tests {
		got := NormalizePeriode(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizePeriode(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNormalizeUsia(t *testing.T) {
	fixedNow := time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		rawUsia  string
		tglLahir string
		expected string
	}{
		{"21歳", "2005年7月7日", "21歳"},
		{"21", "2005年7月7日", "21歳"},
		{"#VALUE!", "2005年7月7日", "21歳"},
		{"", "2005年7月7日", "21歳"},
		{"", "2005-07-07", "21歳"},
	}

	for _, tt := range tests {
		got := NormalizeUsia(tt.rawUsia, tt.tglLahir, fixedNow)
		if got != tt.expected {
			t.Errorf("NormalizeUsia(%q, %q) = %q; want %q", tt.rawUsia, tt.tglLahir, got, tt.expected)
		}
	}
}
