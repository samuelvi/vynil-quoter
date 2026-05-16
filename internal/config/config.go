package config

const (
	DefaultSourceDir       = "data/src"
	DefaultDestinationDir  = "data/dst"
	DefaultReportPath      = "data/report/album_catalog.csv"
	DefaultGeminiModel     = "gemini-2.5-flash-lite"
	DefaultLMStudioBaseURL = "http://localhost:1234/v1"
	DefaultLMStudioModel   = "qwen2.5-vl-7b-instruct"
	AlternateLMStudioModel = "gemma-3-4b-it"
	DefaultProvider        = "lm-studio"
	ProviderGemini         = "gemini"
	ProviderLMStudio       = "lm-studio"
)

type RunConfig struct {
	Image           string
	AllImages       bool
	Replace         bool
	SourceDir       string
	DestinationDir  string
	ReportPath      string
	Provider        string
	Model           string
	LMStudioBaseURL string
	TimeoutSeconds  int
	MaxRetries      int
	RetryDelaySecs  float64
}

func DefaultRunConfig() RunConfig {
	return RunConfig{
		SourceDir:       DefaultSourceDir,
		DestinationDir:  DefaultDestinationDir,
		ReportPath:      DefaultReportPath,
		Provider:        DefaultProvider,
		Model:           DefaultLMStudioModel,
		LMStudioBaseURL: DefaultLMStudioBaseURL,
		TimeoutSeconds:  60,
		MaxRetries:      3,
		RetryDelaySecs:  7,
	}
}
