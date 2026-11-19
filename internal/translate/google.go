package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GoogleTranslateProvider struct {
	baseURL    string
	httpClient *http.Client
}

func NewGoogleTranslateProvider(baseURL string) *GoogleTranslateProvider {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://lpk_aqnesia_g-translate-service:8080"
	}
	return &GoogleTranslateProvider{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *GoogleTranslateProvider) Name() string {
	return "google-translate"
}

func (p *GoogleTranslateProvider) Model() string {
	return "google-translate-api"
}

func (p *GoogleTranslateProvider) PromptVersion() string {
	return "v1.0-map"
}

type gTranslateQueryItem struct {
	Text string `json:"text"`
}

type gTranslateMapRequest struct {
	Queries map[string]gTranslateQueryItem `json:"queries"`
	Options gTranslateOptions              `json:"options"`
}

type gTranslateOptions struct {
	From                string `json:"from"`
	To                  string `json:"to"`
	RejectOnPartialFail bool   `json:"rejectOnPartialFail"`
}

type gTranslateMapResponseItem struct {
	Original string `json:"original"`
	Text     string `json:"text"`
}

func (p *GoogleTranslateProvider) Translate(ctx context.Context, payload *TranslationPayload, glossaryPrompt string) (*TranslationResponse, error) {
	queries := make(map[string]gTranslateQueryItem)

	addQuery := func(key, text string) {
		t := strings.TrimSpace(text)
		if t != "" {
			queries[key] = gTranslateQueryItem{Text: t}
		}
	}

	addQuery("hobi", payload.Hobi)
	addQuery("tujuan_jepang", payload.TujuanJepang)
	addQuery("target_tabungan", payload.TargetTabungan)
	addQuery("rencana_pulang", payload.RencanaPulang)
	addQuery("tempat_lahir", payload.TempatLahir)

	for _, edu := range payload.Pendidikan {
		addQuery(fmt.Sprintf("edu_%d", edu.Index), edu.Jurusan)
	}

	for _, job := range payload.Pekerjaan {
		addQuery(fmt.Sprintf("job_%d", job.Index), job.Rincian)
	}

	for _, fam := range payload.Keluarga {
		addQuery(fmt.Sprintf("fam_%d", fam.Index), fam.Pekerjaan)
	}

	resp := &TranslationResponse{
		NamaKatakana: payload.NamaRomaji,
		Panggilan:    payload.Panggilan,
	}

	if len(queries) == 0 {
		return resp, nil
	}

	reqBody := gTranslateMapRequest{
		Queries: queries,
		Options: gTranslateOptions{
			From:                "auto",
			To:                  "ja",
			RejectOnPartialFail: false,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("gagal marshal request google-translate: %w", err)
	}

	endpoint := p.baseURL + "/translate/map"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("koneksi ke google-translate-service (%s) gagal: %w", p.baseURL, err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		errBytes, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("google-translate-service status %d: %s", httpResp.StatusCode, string(errBytes))
	}

	var resultMap map[string]*gTranslateMapResponseItem
	if err := json.NewDecoder(httpResp.Body).Decode(&resultMap); err != nil {
		return nil, fmt.Errorf("gagal decode response json google-translate: %w", err)
	}

	getText := func(key string) string {
		if item, ok := resultMap[key]; ok && item != nil {
			return strings.TrimSpace(item.Text)
		}
		return ""
	}

	if val := getText("hobi"); val != "" {
		resp.Hobi = val
	}
	if val := getText("tujuan_jepang"); val != "" {
		resp.TujuanJepang = val
	}
	if val := getText("target_tabungan"); val != "" {
		resp.TargetTabungan = val
	}
	if val := getText("rencana_pulang"); val != "" {
		resp.RencanaPulang = val
	}
	if val := getText("tempat_lahir"); val != "" {
		resp.TempatLahir = val
	}

	for _, edu := range payload.Pendidikan {
		key := fmt.Sprintf("edu_%d", edu.Index)
		if val := getText(key); val != "" {
			resp.Pendidikan = append(resp.Pendidikan, EduResponse{
				Index:   edu.Index,
				Jurusan: val,
			})
		}
	}

	for _, job := range payload.Pekerjaan {
		key := fmt.Sprintf("job_%d", job.Index)
		if val := getText(key); val != "" {
			resp.Pekerjaan = append(resp.Pekerjaan, JobResponse{
				Index:   job.Index,
				Rincian: val,
			})
		}
	}

	for _, fam := range payload.Keluarga {
		key := fmt.Sprintf("fam_%d", fam.Index)
		if val := getText(key); val != "" {
			resp.Keluarga = append(resp.Keluarga, FamilyResponse{
				Index:     fam.Index,
				Pekerjaan: val,
			})
		}
	}

	return resp, nil
}
