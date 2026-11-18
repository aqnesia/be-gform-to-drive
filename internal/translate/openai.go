package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type OpenAIProvider struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAIProvider(apiKey, model, baseURL string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o-mini"
	}
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &OpenAIProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (o *OpenAIProvider) Name() string          { return "openai" }
func (o *OpenAIProvider) Model() string         { return o.model }
func (o *OpenAIProvider) PromptVersion() string { return PromptVersion }

func (o *OpenAIProvider) Translate(ctx context.Context, payload *TranslationPayload, glossaryPrompt string) (*TranslationResponse, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("openai api key kosong")
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("gagal marshal payload: %w", err)
	}

	systemPrompt := BuildSystemPrompt(glossaryPrompt)
	userPrompt := fmt.Sprintf("Data peserta yang harus diterjemahkan ke bahasa Jepang:\n%s", string(payloadJSON))

	reqBody := map[string]interface{}{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{
			"type": "json_object",
		},
		"temperature": 0.1,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("gagal marshal request openai: %w", err)
	}

	baseURL := strings.TrimRight(strings.TrimSpace(o.baseURL), "/")
	baseURL = strings.TrimSuffix(baseURL, "/chat/completions")
	url := fmt.Sprintf("%s/chat/completions", baseURL)
	slog.Info("memanggil LLM API", "provider", o.Name(), "base_url", baseURL, "model", o.model)

	cleanKey := strings.TrimSpace(o.apiKey)
	cleanKey = strings.Trim(cleanKey, "\"'")
	cleanKey = strings.TrimPrefix(cleanKey, "Bearer ")
	cleanKey = strings.TrimSpace(cleanKey)

	const maxRetry = 3
	var bodyBytes []byte
	var lastStatus int

	for attempt := 1; attempt <= maxRetry; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBytes))
		if err != nil {
			return nil, fmt.Errorf("gagal membuat request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cleanKey)
		req.Header.Set("x-api-key", cleanKey)

		resp, err := o.httpClient.Do(req)
		if err != nil {
			if attempt < maxRetry {
				slog.Warn("LLM request gagal, mencoba ulang", "attempt", attempt, "error", err)
				time.Sleep(time.Duration(attempt*5) * time.Second)
				continue
			}
			return nil, fmt.Errorf("gagal memanggil openai api: %w", err)
		}

		bodyBytes, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		lastStatus = resp.StatusCode

		if resp.StatusCode == http.StatusOK {
			break
		}

		// 503 / 429 = overload, coba lagi
		if (resp.StatusCode == 503 || resp.StatusCode == 429) && attempt < maxRetry {
			waitSec := time.Duration(attempt*10) * time.Second
			slog.Warn("LLM sedang overload, menunggu dan mencoba ulang", "attempt", attempt, "status", resp.StatusCode, "wait", waitSec)
			time.Sleep(waitSec)
			continue
		}

		return nil, fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	if lastStatus != http.StatusOK {
		return nil, fmt.Errorf("openai api error setelah %d percobaan (status %d): %s", maxRetry, lastStatus, string(bodyBytes))
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(bodyBytes, &openAIResp); err != nil {
		return nil, fmt.Errorf("gagal decode response openai: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("tidak ada pilihan balasan dari openai")
	}

	rawText := cleanJSONMarkdown(openAIResp.Choices[0].Message.Content)

	var result TranslationResponse
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		return nil, fmt.Errorf("gagal unmarshal output json openai: %w (raw: %s)", err, rawText)
	}

	return &result, nil
}
