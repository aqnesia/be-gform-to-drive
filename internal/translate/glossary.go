package translate

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type GlossaryConfig struct {
	Glossary map[string]string `yaml:"glossary"`
}

type Glossary struct {
	terms map[string]string
}

func LoadGlossary(path string) (*Glossary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca glossary: %w", err)
	}

	var cfg GlossaryConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("gagal unmarshal glossary: %w", err)
	}

	normalized := make(map[string]string)
	for k, v := range cfg.Glossary {
		normalized[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}

	return &Glossary{terms: normalized}, nil
}

func (g *Glossary) Lookup(indonesian string) (string, bool) {
	if g == nil || g.terms == nil {
		return "", false
	}
	val, ok := g.terms[strings.ToLower(strings.TrimSpace(indonesian))]
	return val, ok
}

func (g *Glossary) FormatPromptList() string {
	if g == nil || len(g.terms) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Gunakan daftar istilah baku Jepang berikut jika menemukan kata yang sesuai:\n")
	for k, v := range g.terms {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", k, v))
	}
	return sb.String()
}
