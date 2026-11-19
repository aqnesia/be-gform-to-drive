package translate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGoogleTranslateProvider(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/translate/map" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("Unexpected method: %s", r.Method)
			http.Error(w, "bad method", http.StatusBadRequest)
			return
		}

		var req gTranslateMapRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Decode request body failed: %v", err)
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		if req.Options.To != "ja" {
			t.Errorf("Expected option to=ja, got %s", req.Options.To)
		}

		resp := map[string]gTranslateMapResponseItem{
			"hobi": {
				Original: "Menggambar",
				Text:     "絵を描くこと",
			},
			"tujuan_jepang": {
				Original: "Mencari modal",
				Text:     "資金を貯めること",
			},
			"edu_0": {
				Original: "Teknik Otomotif",
				Text:     "自動車工学",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	p := NewGoogleTranslateProvider(ts.URL)
	if p.Name() != "google-translate" {
		t.Errorf("unexpected name: %s", p.Name())
	}

	payload := &TranslationPayload{
		NamaRomaji:   "GALANG",
		Hobi:         "Menggambar",
		TujuanJepang: "Mencari modal",
		Pendidikan: []EduPayload{
			{Index: 0, Jurusan: "Teknik Otomotif"},
		},
	}

	res, err := p.Translate(context.Background(), payload, "")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	if res.Hobi != "絵を描くこと" {
		t.Errorf("Expected hobi 絵を描くこと, got %s", res.Hobi)
	}
	if res.TujuanJepang != "資金を貯めること" {
		t.Errorf("Expected tujuan_jepang 資金を貯めること, got %s", res.TujuanJepang)
	}
	if len(res.Pendidikan) != 1 || res.Pendidikan[0].Jurusan != "自動車工学" {
		t.Errorf("Expected edu 自動車工学, got %+v", res.Pendidikan)
	}
}
