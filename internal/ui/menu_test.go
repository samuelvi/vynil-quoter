package ui

import (
	"bytes"
	"testing"
	"vinylquoter/internal/config"
)

func TestMenuChoiceFourMeansAllAndReplace(t *testing.T) {
	cfg, err := ReadMenu(bytes.NewBufferString("4\n\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AllImages || !cfg.Replace || cfg.Image != "" {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuCanSelectGeminiProvider(t *testing.T) {
	cfg, err := ReadMenu(bytes.NewBufferString("2\n2\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != config.ProviderGemini || cfg.Model != config.DefaultGeminiModel {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuDefaultProviderIsLocalVisionModel(t *testing.T) {
	cfg, err := ReadMenu(bytes.NewBufferString("2\n\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != config.ProviderLMStudio || cfg.Model != config.DefaultLMStudioModel {
		t.Fatalf("got %#v", cfg)
	}
}
