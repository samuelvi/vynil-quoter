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
	ConditionMint          = "M"
	ConditionNearMint      = "NM/M-"
	ConditionVeryGoodPlus  = "VG+"
	DefaultCondition       = "VG"
	ConditionGoodPlus      = "G+"
	ConditionGood          = "G"
	ConditionFair          = "F"
	ConditionPoor          = "P"
	ConditionGeneric       = "Generic"
)

var MediaConditions = []string{
	ConditionMint,
	ConditionNearMint,
	ConditionVeryGoodPlus,
	DefaultCondition,
	ConditionGoodPlus,
	ConditionGood,
	ConditionFair,
	ConditionPoor,
}

var SleeveConditions = []string{
	ConditionMint,
	ConditionNearMint,
	ConditionVeryGoodPlus,
	DefaultCondition,
	ConditionGoodPlus,
	ConditionGood,
	ConditionFair,
	ConditionPoor,
	ConditionGeneric,
}

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
	MediaCondition  string
	SleeveCondition string
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
		MediaCondition:  DefaultCondition,
		SleeveCondition: DefaultCondition,
	}
}

func IsMediaCondition(condition string) bool {
	return containsCondition(MediaConditions, condition)
}

func IsSleeveCondition(condition string) bool {
	return containsCondition(SleeveConditions, condition)
}

func containsCondition(conditions []string, condition string) bool {
	for _, allowed := range conditions {
		if condition == allowed {
			return true
		}
	}
	return false
}
