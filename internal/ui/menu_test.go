package ui

import (
	"bytes"
	"strings"
	"testing"
	"vinylquoter/internal/config"
)

func TestMenuChoiceFourMeansAllAndReplace(t *testing.T) {
	cfg, err := ReadMenu(bytes.NewBufferString("3\n3\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AllImages || !cfg.Replace || cfg.Image != "" {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuCSVUpdateSubmenuMeansAllWithoutReplace(t *testing.T) {
	cfg, err := ReadMenu(bytes.NewBufferString("3\n2\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AllImages || cfg.Replace || cfg.Image != "" {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuCSVBackReturnsNoAction(t *testing.T) {
	_, err := ReadMenu(bytes.NewBufferString("3\n4\n"), &bytes.Buffer{})
	if err != ErrNoAction {
		t.Fatalf("got %v", err)
	}
}

func TestMenuCanSelectGeminiProvider(t *testing.T) {
	cfg, err := ReadMenu(bytes.NewBufferString("4\n3\n"), &bytes.Buffer{})
	if err != ErrNoAction {
		t.Fatal(err)
	}
	if cfg.Provider != config.ProviderGemini || cfg.Model != config.DefaultGeminiModel {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuDefaultProviderIsLocalVisionModel(t *testing.T) {
	cfg, err := ReadMenu(bytes.NewBufferString("2\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != config.ProviderLMStudio || cfg.Model != config.DefaultLMStudioModel {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuCanSelectAlternateLMStudioVisionModel(t *testing.T) {
	cfg, err := ReadMenu(bytes.NewBufferString("4\n2\n"), &bytes.Buffer{})
	if err != ErrNoAction {
		t.Fatal(err)
	}
	if cfg.Provider != config.ProviderLMStudio || cfg.Model != config.AlternateLMStudioModel {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuShowsCurrentCSVAndModelState(t *testing.T) {
	stdout := &bytes.Buffer{}
	state := config.DefaultRunConfig()
	state.ReportPath = "custom/report.csv"
	state.Provider = config.ProviderGemini
	state.Model = config.DefaultGeminiModel
	_, _ = ReadMenuWithState(bytes.NewBufferString("5\n"), stdout, state)
	output := stdout.String()
	if !strings.Contains(output, "Guardado csv (custom/report.csv)") {
		t.Fatalf("menu should show current CSV path, got %s", output)
	}
	if !strings.Contains(output, "Modelo (gemini: gemini-2.5-flash-lite)") {
		t.Fatalf("menu should show current model, got %s", output)
	}
}

func TestMenuCanChangeCSVPathWithoutSelectingModel(t *testing.T) {
	cfg, err := ReadMenuWithState(bytes.NewBufferString("3\n1\ncustom.csv\n"), &bytes.Buffer{}, config.DefaultRunConfig())
	if err != ErrNoAction {
		t.Fatalf("got %v", err)
	}
	if cfg.ReportPath != "custom.csv" {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuCanChangeModelFromMainMenu(t *testing.T) {
	cfg, err := ReadMenuWithState(bytes.NewBufferString("4\n3\n"), &bytes.Buffer{}, config.DefaultRunConfig())
	if err != ErrNoAction {
		t.Fatalf("got %v", err)
	}
	if cfg.Provider != config.ProviderGemini || cfg.Model != config.DefaultGeminiModel {
		t.Fatalf("got %#v", cfg)
	}
}
