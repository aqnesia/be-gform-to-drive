package drive

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	scriptURL  string
	secret     string
	httpClient *http.Client
}

func NewClient(scriptURL, secret string) *Client {
	return &Client{
		scriptURL: scriptURL,
		secret:    secret,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			// Tangani redirect 302 dari Google Apps Script agar tetap mengirimkan request dengan benar
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("terlalu banyak redirect")
				}
				return nil
			},
		},
	}
}

type uploadRequest struct {
	Secret     string `json:"secret"`
	FileName   string `json:"file_name"`
	FileBase64 string `json:"file_base64"`
}

type uploadResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	FileID   string `json:"file_id,omitempty"`
	FileURL  string `json:"file_url,omitempty"`
	FileName string `json:"file_name,omitempty"`
}

func (c *Client) Upload(ctx context.Context, fileName string, fileBytes []byte) (string, string, error) {
	if c.scriptURL == "" {
		return "", "", fmt.Errorf("DRIVE_SCRIPT_URL belum dikonfigurasi")
	}

	reqBody := uploadRequest{
		Secret:     c.secret,
		FileName:   fileName,
		FileBase64: base64.StdEncoding.EncodeToString(fileBytes),
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("gagal marshal request upload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.scriptURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return "", "", fmt.Errorf("gagal membuat request upload: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("gagal mengirim berkas ke Apps Script: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("gagal membaca balasan Apps Script: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("Apps Script mengembalikan status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res uploadResponse
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		return "", "", fmt.Errorf("gagal membaca format json balasan Apps Script: %w (raw: %s)", err, string(bodyBytes))
	}

	if res.Status != "success" {
		return "", "", fmt.Errorf("Apps Script gagal menyimpan berkas: %s", res.Message)
	}

	return res.FileID, res.FileURL, nil
}
