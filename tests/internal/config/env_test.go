package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"vinylquoter/internal/config"
)

func TestLoadDefaultsReadsDotEnvFile(t *testing.T) {
	tmp := t.TempDir()
	envPath := filepath.Join(tmp, ".env")
	content := "" +
		"VINYLQUOTER_SOURCE_DIR=custom/src\n" +
		"VINYLQUOTER_DESTINATION_DIR=custom/dst\n" +
		"VINYLQUOTER_REPORT_PATH=custom/report.csv\n" +
		"VINYLQUOTER_PROVIDER=gemini\n" +
		"VINYLQUOTER_MODEL=gemini-custom\n" +
		"VINYLQUOTER_LM_STUDIO_BASE_URL=http://lm-studio.test/v1\n" +
		"VINYLQUOTER_TIMEOUT_SECONDS=12\n" +
		"VINYLQUOTER_MAX_RETRIES=5\n" +
		"VINYLQUOTER_RETRY_DELAY_SECONDS=1.5\n" +
		"VINYLQUOTER_MEDIA_CONDITION=VG+\n" +
		"VINYLQUOTER_SLEEVE_CONDITION=Generic\n"
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadDefaultsFromEnvFile(envPath, emptyEnv)

	if err != nil {
		t.Fatal(err)
	}
	if cfg.SourceDir != "custom/src" || cfg.DestinationDir != "custom/dst" || cfg.ReportPath != "custom/report.csv" {
		t.Fatalf("paths not loaded: %#v", cfg)
	}
	if cfg.Provider != config.ProviderGemini || cfg.Model != "gemini-custom" {
		t.Fatalf("provider/model not loaded: %#v", cfg)
	}
	if cfg.LMStudioBaseURL != "http://lm-studio.test/v1" || cfg.TimeoutSeconds != 12 || cfg.MaxRetries != 5 || cfg.RetryDelaySecs != 1.5 {
		t.Fatalf("runtime values not loaded: %#v", cfg)
	}
	if cfg.MediaCondition != "VG+" || cfg.SleeveCondition != "Generic" {
		t.Fatalf("conditions not loaded: %#v", cfg)
	}
}

func TestLoadDefaultsProcessEnvOverridesDotEnvFile(t *testing.T) {
	tmp := t.TempDir()
	envPath := filepath.Join(tmp, ".env")
	if err := os.WriteFile(envPath, []byte("VINYLQUOTER_REPORT_PATH=file.csv\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadDefaultsFromEnvFile(envPath, func(key string) (string, bool) {
		if key == "VINYLQUOTER_REPORT_PATH" {
			return "process.csv", true
		}
		return "", false
	})

	if err != nil {
		t.Fatal(err)
	}
	if cfg.ReportPath != "process.csv" {
		t.Fatalf("process env should override .env file, got %#v", cfg)
	}
}

func TestLoadDefaultsRejectsInvalidEnvValues(t *testing.T) {
	_, err := config.LoadDefaultsFromEnvFile("", func(key string) (string, bool) {
		if key == "VINYLQUOTER_MEDIA_CONDITION" {
			return "BAD", true
		}
		return "", false
	})

	if err == nil {
		t.Fatal("expected invalid media condition error")
	}
}

func emptyEnv(string) (string, bool) {
	return "", false
}
