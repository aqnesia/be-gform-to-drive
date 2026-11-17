package translate

import (
	"context"
	"path/filepath"
	"testing"

	"automate-gform-drive/internal/config"
	"automate-gform-drive/internal/student"
)

type MockProvider struct{}

func (m *MockProvider) Name() string          { return "mock" }
func (m *MockProvider) Model() string         { return "mock-model" }
func (m *MockProvider) PromptVersion() string { return "v1.0" }
func (m *MockProvider) Translate(ctx context.Context, payload *TranslationPayload, glossaryPrompt string) (*TranslationResponse, error) {
	return &TranslationResponse{
		NamaKatakana: "テスト・タロウ",
		Kelebihan:    "真面目で勤勉",
		Kekurangan:    "心配性",
		Hobi:         "読書",
		TujuanJepang: "スキルアップと貯金のため",
	}, nil
}

func TestGlossaryAndTranslationService(t *testing.T) {
	glossaryPath := filepath.Join("..", "..", "configs", "glossary.yaml")
	g, err := LoadGlossary(glossaryPath)
	if err != nil {
		t.Fatalf("gagal memuat glossary: %v", err)
	}

	val, ok := g.Lookup("Petani")
	if !ok || val != "農業" {
		t.Errorf("Lookup Petani gagal, expected 農業, got %s", val)
	}

	svc, err := NewService(&config.Config{LLMProvider: "gemini"}, glossaryPath)
	if err != nil {
		t.Fatalf("gagal init service: %v", err)
	}

	// Use mock provider
	svc.provider = &MockProvider{}

	cv := &student.CVData{
		NamaRomaji: "Test Taro",
		Keluarga: []student.KeluargaItem{
			{Hubungan: "Ayah", Pekerjaan: "Petani"},
		},
	}

	res, provider, model, ver, err := svc.TranslateCV(context.Background(), cv)
	if err != nil {
		t.Fatalf("TranslateCV failed: %v", err)
	}

	if res.NamaKatakana != "テスト・タロウ" {
		t.Errorf("Expected katakana テスト・タロウ, got %s", res.NamaKatakana)
	}
	if res.Keluarga[0].Pekerjaan != "農業" {
		t.Errorf("Expected glossary translation 農業, got %s", res.Keluarga[0].Pekerjaan)
	}
	t.Logf("Sukses test translate service! provider=%s model=%s ver=%s", provider, model, ver)
}
