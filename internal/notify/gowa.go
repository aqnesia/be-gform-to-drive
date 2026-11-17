package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL     string
	authToken   string
	adminNumber string
	httpClient  *http.Client
}

func NewClient(baseURL, authToken, adminNumber string) *Client {
	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		authToken:   authToken,
		adminNumber: adminNumber,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) SendFailure(ctx context.Context, nama, tahap, errMsg string, t time.Time) error {
	if c.baseURL == "" || c.adminNumber == "" {
		slog.Debug("GoWA notifikasi diabaikan karena URL atau nomor admin belum diset")
		return nil
	}

	pesan := fmt.Sprintf(
		"⚠️ *[GAGAL] Otomasi Generate CV Jepang*\n\n"+
			"• *Nama*: %s\n"+
			"• *Tahap*: %s\n"+
			"• *Waktu*: %s\n"+
			"• *Pesan Error*: %s\n\n"+
			"_Mohon periksa log atau database untuk detail lebih lanjut._",
		nama, tahap, t.Format("02/01/2006 15:04:05 WIB"), errMsg,
	)

	reqBody := map[string]interface{}{
		"phone":   c.adminNumber,
		"message": pesan,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("gagal marshal pesan gowa: %w", err)
	}

	url := fmt.Sprintf("%s/send-message", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("gagal membuat request notifikasi gowa: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Warn("gagal mengirim notifikasi whatsapp GoWA", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("GoWA mengembalikan response non-2xx", "status", resp.StatusCode)
		return fmt.Errorf("gowa error status: %d", resp.StatusCode)
	}

	slog.Info("notifikasi kegagalan berhasil dikirim ke whatsapp admin", "target", c.adminNumber)
	return nil
}
