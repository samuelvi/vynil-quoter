package ui_test

import (
	"bytes"
	"strings"
	"testing"
	"vinylquoter/internal/config"
	"vinylquoter/internal/ui"
)

func TestMenuChoiceFourMeansAllAndReplace(t *testing.T) {
	cfg, err := ui.ReadMenu(bytes.NewBufferString("3\n3\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AllImages || !cfg.Replace || cfg.Image != "" {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuCSVUpdateSubmenuMeansAllWithoutReplace(t *testing.T) {
	cfg, err := ui.ReadMenu(bytes.NewBufferString("3\n2\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AllImages || cfg.Replace || cfg.Image != "" {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuCSVBackReturnsNoAction(t *testing.T) {
	_, err := ui.ReadMenu(bytes.NewBufferString("3\n4\n"), &bytes.Buffer{})
	if err != ui.ErrNoAction {
		t.Fatalf("got %v", err)
	}
}

func TestMenuCanSelectGeminiProvider(t *testing.T) {
	cfg, err := ui.ReadMenu(bytes.NewBufferString("4\n3\n"), &bytes.Buffer{})
	if err != ui.ErrNoAction {
		t.Fatal(err)
	}
	if cfg.Provider != config.ProviderGemini || cfg.Model != config.DefaultGeminiModel {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuDefaultProviderIsLocalVisionModel(t *testing.T) {
	cfg, err := ui.ReadMenu(bytes.NewBufferString("2\n"), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != config.ProviderLMStudio || cfg.Model != config.DefaultLMStudioModel {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuExitUsesZero(t *testing.T) {
	_, err := ui.ReadMenu(bytes.NewBufferString("0\n"), &bytes.Buffer{})
	if err == nil || err.Error() != "EOF" {
		t.Fatalf("expected EOF for exit option 0, got %v", err)
	}
}

func TestMenuSevenIsInvalid(t *testing.T) {
	_, err := ui.ReadMenu(bytes.NewBufferString("7\n"), &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "invalid menu choice") {
		t.Fatalf("expected invalid choice for option 7, got %v", err)
	}
}

func TestMenuShowsExitAsZero(t *testing.T) {
	stdout := &bytes.Buffer{}
	_, _ = ui.ReadMenuWithState(bytes.NewBufferString("0\n"), stdout, config.DefaultRunConfig())
	output := stdout.String()
	if !strings.Contains(output, "0) Salir") {
		t.Fatalf("menu should show exit as zero, got %s", output)
	}
	if !strings.Contains(output, "Elige una opción [0-6]:") {
		t.Fatalf("menu should show [0-6] prompt, got %s", output)
	}
}

func TestMenuCanSelectAlternateLMStudioVisionModel(t *testing.T) {
	cfg, err := ui.ReadMenu(bytes.NewBufferString("4\n2\n"), &bytes.Buffer{})
	if err != ui.ErrNoAction {
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
	_, _ = ui.ReadMenuWithState(bytes.NewBufferString("0\n"), stdout, state)
	output := stdout.String()
	if !strings.Contains(output, "Guardado csv (custom/report.csv)") {
		t.Fatalf("menu should show current CSV path, got %s", output)
	}
	if !strings.Contains(output, "Modelo (gemini: gemini-2.5-flash-lite)") {
		t.Fatalf("menu should show current model, got %s", output)
	}
}

func TestMenuCanChangeCSVPathWithoutSelectingModel(t *testing.T) {
	cfg, err := ui.ReadMenuWithState(bytes.NewBufferString("3\n1\ncustom.csv\n"), &bytes.Buffer{}, config.DefaultRunConfig())
	if err != ui.ErrNoAction {
		t.Fatalf("got %v", err)
	}
	if cfg.ReportPath != "custom.csv" {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuCanChangeModelFromMainMenu(t *testing.T) {
	cfg, err := ui.ReadMenuWithState(bytes.NewBufferString("4\n3\n"), &bytes.Buffer{}, config.DefaultRunConfig())
	if err != ui.ErrNoAction {
		t.Fatalf("got %v", err)
	}
	if cfg.Provider != config.ProviderGemini || cfg.Model != config.DefaultGeminiModel {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuShowsCurrentConditionState(t *testing.T) {
	stdout := &bytes.Buffer{}
	state := config.DefaultRunConfig()
	state.MediaCondition = "VG+"
	state.SleeveCondition = "G+"
	_, _ = ui.ReadMenuWithState(bytes.NewBufferString("0\n"), stdout, state)
	output := stdout.String()
	if !strings.Contains(output, "Calidad carátula (G+)") {
		t.Fatalf("menu should show sleeve condition, got %s", output)
	}
	if !strings.Contains(output, "Calidad vinilo (VG+)") {
		t.Fatalf("menu should show media condition, got %s", output)
	}
}

func TestMenuCanSelectSleeveCondition(t *testing.T) {
	cfg, err := ui.ReadMenuWithState(bytes.NewBufferString("5\n3\n"), &bytes.Buffer{}, config.DefaultRunConfig())
	if err != ui.ErrNoAction {
		t.Fatalf("got %v", err)
	}
	if cfg.SleeveCondition != "VG+" {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuCanSelectMediaCondition(t *testing.T) {
	cfg, err := ui.ReadMenuWithState(bytes.NewBufferString("6\n5\n"), &bytes.Buffer{}, config.DefaultRunConfig())
	if err != ui.ErrNoAction {
		t.Fatalf("got %v", err)
	}
	if cfg.MediaCondition != "G+" {
		t.Fatalf("got %#v", cfg)
	}
}
