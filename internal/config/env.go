package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	EnvSourceDir       = "VINYLQUOTER_SOURCE_DIR"
	EnvDestinationDir  = "VINYLQUOTER_DESTINATION_DIR"
	EnvReportPath      = "VINYLQUOTER_REPORT_PATH"
	EnvProvider        = "VINYLQUOTER_PROVIDER"
	EnvModel           = "VINYLQUOTER_MODEL"
	EnvLMStudioBaseURL = "VINYLQUOTER_LM_STUDIO_BASE_URL"
	EnvTimeoutSeconds  = "VINYLQUOTER_TIMEOUT_SECONDS"
	EnvMaxRetries      = "VINYLQUOTER_MAX_RETRIES"
	EnvRetryDelaySecs  = "VINYLQUOTER_RETRY_DELAY_SECONDS"
	EnvMediaCondition  = "VINYLQUOTER_MEDIA_CONDITION"
	EnvSleeveCondition = "VINYLQUOTER_SLEEVE_CONDITION"
)

type LookupEnvFunc func(string) (string, bool)

func LoadDefaults() (RunConfig, error) {
	return LoadDefaultsFromEnvFile(".env", os.LookupEnv)
}

func LoadDefaultsFromEnvFile(path string, lookup LookupEnvFunc) (RunConfig, error) {
	cfg := DefaultRunConfig()
	fileValues, err := readEnvFile(path)
	if err != nil {
		return cfg, err
	}
	if err := applyEnvValues(&cfg, func(key string) (string, bool) {
		value, ok := fileValues[key]
		return value, ok
	}); err != nil {
		return cfg, err
	}
	if lookup != nil {
		if err := applyEnvValues(&cfg, lookup); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

func readEnvFile(path string) (map[string]string, error) {
	values := map[string]string{}
	if path == "" {
		return values, nil
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return values, nil
		}
		return values, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		values[strings.TrimSpace(key)] = trimEnvValue(value)
	}
	return values, scanner.Err()
}

func trimEnvValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 2 {
		first := trimmed[0]
		last := trimmed[len(trimmed)-1]
		if (first == '\'' && last == '\'') || (first == '"' && last == '"') {
			return trimmed[1 : len(trimmed)-1]
		}
	}
	return trimmed
}

func applyEnvValues(cfg *RunConfig, lookup LookupEnvFunc) error {
	if value, ok := lookup(EnvSourceDir); ok {
		cfg.SourceDir = value
	}
	if value, ok := lookup(EnvDestinationDir); ok {
		cfg.DestinationDir = value
	}
	if value, ok := lookup(EnvReportPath); ok {
		cfg.ReportPath = value
	}
	if value, ok := lookup(EnvProvider); ok {
		cfg.Provider = value
	}
	if value, ok := lookup(EnvModel); ok {
		cfg.Model = value
	}
	if value, ok := lookup(EnvLMStudioBaseURL); ok {
		cfg.LMStudioBaseURL = value
	}
	if err := applyIntEnv(lookup, EnvTimeoutSeconds, &cfg.TimeoutSeconds); err != nil {
		return err
	}
	if err := applyIntEnv(lookup, EnvMaxRetries, &cfg.MaxRetries); err != nil {
		return err
	}
	if err := applyFloatEnv(lookup, EnvRetryDelaySecs, &cfg.RetryDelaySecs); err != nil {
		return err
	}
	if value, ok := lookup(EnvMediaCondition); ok {
		if !IsMediaCondition(value) {
			return fmt.Errorf("invalid media condition: %s", value)
		}
		cfg.MediaCondition = value
	}
	if value, ok := lookup(EnvSleeveCondition); ok {
		if !IsSleeveCondition(value) {
			return fmt.Errorf("invalid sleeve condition: %s", value)
		}
		cfg.SleeveCondition = value
	}
	return nil
}

func applyIntEnv(lookup LookupEnvFunc, key string, target *int) error {
	value, ok := lookup(key)
	if !ok {
		return nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid %s: %s", key, value)
	}
	*target = parsed
	return nil
}

func applyFloatEnv(lookup LookupEnvFunc, key string, target *float64) error {
	value, ok := lookup(key)
	if !ok {
		return nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid %s: %s", key, value)
	}
	*target = parsed
	return nil
}
