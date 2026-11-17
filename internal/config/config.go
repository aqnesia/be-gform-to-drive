package config

import (
	"os"
)

type Config struct {
	Port              string
	DatabaseURL       string
	WebhookSecret     string
	LLMProvider       string
	LLMModel          string
	LLMAPIKey         string
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
		return val
	}
	return defaultVal
}

