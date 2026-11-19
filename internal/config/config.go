package config

import (
	"os"
	"strings"
)

type Config struct {
	Port              string
	DatabaseURL       string
	WebhookSecret     string
	LLMProvider       string
	LLMModel          string
	LLMAPIKey         string
	LLMBaseURL        string
	GoogleTranslateURL string
	DriveScriptURL    string
	DriveScriptSecret string
	GoWAURL           string
	GoWAAuth          string
	AdminWANumber     string
	TemplatePath      string
	TmpDir            string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		WebhookSecret:     getEnv("WEBHOOK_SECRET", ""),
		LLMProvider:       getEnv("LLM_PROVIDER", "gemini"),
		LLMModel:          getEnv("LLM_MODEL", "gemini-2.5-flash"),
		LLMAPIKey:         getEnv("LLM_API_KEY", ""),
		LLMBaseURL:        getEnv("LLM_BASE_URL", ""),
		GoogleTranslateURL: getEnv("GOOGLE_TRANSLATE_URL", "http://lpk_aqnesia_g-translate-service:8080"),
		DriveScriptURL:    getEnv("DRIVE_SCRIPT_URL", ""),
		DriveScriptSecret: getEnv("DRIVE_SCRIPT_SECRET", ""),
		GoWAURL:           getEnv("GOWA_URL", ""),
		GoWAAuth:          getEnv("GOWA_AUTH", ""),
		AdminWANumber:     getEnv("ADMIN_WA_NUMBER", ""),
		TemplatePath:      getEnv("TEMPLATE_PATH", "./templates/output-template-cv.xlsx"),
		TmpDir:            getEnv("TMP_DIR", "./tmp"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		val = strings.TrimSpace(val)
		val = strings.Trim(val, "\"'")
		return val
	}
	return defaultVal
}

