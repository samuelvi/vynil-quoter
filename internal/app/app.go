package app

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"vinylquoter/internal/catalog"
	"vinylquoter/internal/config"
	"vinylquoter/internal/imageinput"
	"vinylquoter/internal/provider"
	"vinylquoter/internal/provider/gemini"
	"vinylquoter/internal/provider/lmstudio"
	"vinylquoter/internal/ui"
)

func ParseArgs(args []string) (config.RunConfig, error) {
	cfg := config.DefaultRunConfig()
	set := flag.NewFlagSet("vinyl-quoter", flag.ContinueOnError)
	set.StringVar(&cfg.SourceDir, "src", cfg.SourceDir, "source image directory")
	set.StringVar(&cfg.ReportPath, "report", cfg.ReportPath, "CSV report path")
	set.StringVar(&cfg.Image, "image", "", "process one image path or filename from data/src")
	set.BoolVar(&cfg.AllImages, "all", false, "process all supported images from data/src")
	set.BoolVar(&cfg.Replace, "replace", false, "regenerate the final CSV instead of updating it")
	set.StringVar(&cfg.Provider, "provider", cfg.Provider, "vision provider: lm-studio or gemini")
	set.StringVar(&cfg.Model, "model", cfg.Model, "vision model")
	set.StringVar(&cfg.LMStudioBaseURL, "lm-studio-base-url", cfg.LMStudioBaseURL, "LM Studio OpenAI-compatible base URL")
	set.IntVar(&cfg.TimeoutSeconds, "timeout", cfg.TimeoutSeconds, "request timeout seconds")
	set.IntVar(&cfg.MaxRetries, "max-retries", cfg.MaxRetries, "Gemini retry count")
	set.Float64Var(&cfg.RetryDelaySecs, "retry-delay", cfg.RetryDelaySecs, "fallback retry delay seconds")
	if err := set.Parse(args); err != nil {
		return cfg, err
	}
	if cfg.Provider == config.ProviderGemini && cfg.Model == config.DefaultLMStudioModel {
		cfg.Model = config.DefaultGeminiModel
	}
	if cfg.Provider == config.ProviderLMStudio && cfg.Model == "" {
		cfg.Model = config.DefaultLMStudioModel
	}
	if cfg.Image != "" {
		cfg.Replace = false
	}
	return cfg, nil
}

func Run(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	cfg, err := ParseArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	return runWithRecognizerFactory(ctx, cfg, stdin, stdout, stderr, recognizerFor)
}

func runWithRecognizerFactory(ctx context.Context, cfg config.RunConfig, stdin io.Reader, stdout io.Writer, stderr io.Writer, factory func(config.RunConfig) (provider.Recognizer, error)) int {
	if cfg.Image != "" || cfg.AllImages {
		return runOnce(ctx, cfg, stdout, stderr, factory)
	}
	reader := bufio.NewReader(stdin)
	state := cfg
	for {
		menuCfg, err := ui.ReadMenuWithState(reader, stdout, state)
		state = menuCfg
		if errors.Is(err, io.EOF) {
			return 0
		}
		if errors.Is(err, ui.ErrNoAction) {
			continue
		}
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			continue
		}
		menuCfg.SourceDir = state.SourceDir
		menuCfg.ReportPath = state.ReportPath
		menuCfg.LMStudioBaseURL = state.LMStudioBaseURL
		menuCfg.TimeoutSeconds = cfg.TimeoutSeconds
		menuCfg.MaxRetries = cfg.MaxRetries
		menuCfg.RetryDelaySecs = cfg.RetryDelaySecs
		_ = runOnce(ctx, menuCfg, stdout, stderr, factory)
	}
}

func runOnce(ctx context.Context, cfg config.RunConfig, stdout io.Writer, stderr io.Writer, factory func(config.RunConfig) (provider.Recognizer, error)) int {
	recognizer, err := factory(cfg)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	images, err := imageinput.Collect(cfg.SourceDir, cfg.Image, cfg.AllImages)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	rows, err := Process(ctx, images, cfg.ReportPath, cfg.Replace, recognizer)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 2
	}
	fmt.Fprintf(stdout, "CSV generado: %s (%d filas)\n", cfg.ReportPath, len(rows))
	return 0
}

func Process(ctx context.Context, images []string, reportPath string, replace bool, recognizer provider.Recognizer) ([]catalog.Row, error) {
	rows := []catalog.Row{}
	if !replace {
		existing, err := catalog.Read(reportPath)
		if err != nil {
			return nil, err
		}
		rows = existing
	}
	pending, err := catalog.Pending(images, reportPath, replace)
	if err != nil {
		return nil, err
	}
	for _, image := range pending {
		identification, err := recognizer.Identify(ctx, image)
		if err != nil {
			identification = catalog.Identification{Artist: "Unknown", Title: "Unknown", IdentificationConfidence: "manual-review", PriceConfidence: "manual-review", Notes: "identification failed: " + err.Error()}
		}
		rows = append(rows, catalog.Row{SourceImage: catalog.ImageID(image), Artist: identification.Artist, Title: identification.Title, IdentificationConfidence: identification.IdentificationConfidence, RecommendedPriceEUR: identification.RecommendedPriceEUR, PriceConfidence: identification.PriceConfidence, PriceBasis: identification.PriceBasis, Notes: identification.Notes})
		if err := catalog.Write(reportPath, rows); err != nil {
			return nil, err
		}
	}
	if len(pending) == 0 {
		if err := catalog.Write(reportPath, rows); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

func recognizerFor(cfg config.RunConfig) (provider.Recognizer, error) {
	client := &http.Client{Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second}
	switch cfg.Provider {
	case config.ProviderGemini:
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY is required")
		}
		return gemini.Client{APIKey: apiKey, Model: cfg.Model, HTTPClient: client, MaxRetries: cfg.MaxRetries, RetryDelay: time.Duration(cfg.RetryDelaySecs * float64(time.Second))}, nil
	case config.ProviderLMStudio:
		return lmstudio.Client{BaseURL: cfg.LMStudioBaseURL, Model: cfg.Model, HTTPClient: client}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}
