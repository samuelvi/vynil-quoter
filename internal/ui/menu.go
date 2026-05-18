package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"vinylquoter/internal/config"
)

var ErrNoAction = errors.New("no menu action selected")

func ReadMenu(in io.Reader, out io.Writer) (config.RunConfig, error) {
	return ReadMenuWithState(in, out, config.DefaultRunConfig())
}

func ReadMenuWithState(in io.Reader, out io.Writer, state config.RunConfig) (config.RunConfig, error) {
	cfg := state
	cfg.Image = ""
	cfg.AllImages = false
	cfg.Replace = false
	reader := readerFor(in)
	fmt.Fprintln(out, "\nVinyl Quoter")
	fmt.Fprintln(out, "1) Procesar una imagen concreta")
	fmt.Fprintln(out, "2) Procesar todas las imágenes de data/src")
	fmt.Fprintf(out, "3) Guardado csv (%s)\n", cfg.ReportPath)
	fmt.Fprintf(out, "4) Modelo (%s)\n", modelLabel(cfg))
	fmt.Fprintf(out, "5) Calidad carátula (%s)\n", cfg.SleeveCondition)
	fmt.Fprintf(out, "6) Calidad vinilo (%s)\n", cfg.MediaCondition)
	fmt.Fprintln(out, "0) Salir")
	fmt.Fprint(out, "Elige una opción [0-6]: ")
	choice, _ := reader.ReadString('\n')
	switch strings.TrimSpace(choice) {
	case "0":
		if confirmExit(reader, out) {
			return cfg, io.EOF
		}
		return cfg, ErrNoAction
	case "1":
		fmt.Fprint(out, "Ruta o nombre de imagen: ")
		image, _ := reader.ReadString('\n')
		cfg.Image = strings.TrimSpace(image)
	case "2":
		cfg.AllImages = true
	case "3":
		updated, err := ReadCSVMenu(reader, out, cfg)
		if err != nil {
			return updated, err
		}
		cfg = updated
		cfg.AllImages = true
	case "4":
		provider, model, err := ReadProvider(reader, out)
		if err != nil {
			return cfg, err
		}
		cfg.Provider = provider
		cfg.Model = model
		return cfg, ErrNoAction
	case "5":
		condition, err := ReadSleeveCondition(reader, out, cfg.SleeveCondition)
		if err != nil {
			return cfg, err
		}
		cfg.SleeveCondition = condition
		return cfg, ErrNoAction
	case "6":
		condition, err := ReadMediaCondition(reader, out, cfg.MediaCondition)
		if err != nil {
			return cfg, err
		}
		cfg.MediaCondition = condition
		return cfg, ErrNoAction
	default:
		return cfg, fmt.Errorf("invalid menu choice")
	}
	return cfg, nil
}

func confirmExit(in *bufio.Reader, out io.Writer) bool {
	fmt.Fprint(out, "¿Seguro que quieres salir? [s/N]: ")
	answer, _ := in.ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "s", "si", "sí", "y", "yes":
		return true
	default:
		return false
	}
}

func readerFor(in io.Reader) *bufio.Reader {
	if reader, ok := in.(*bufio.Reader); ok {
		return reader
	}
	return bufio.NewReader(in)
}

func ReadCSVMenu(in *bufio.Reader, out io.Writer, state config.RunConfig) (config.RunConfig, error) {
	fmt.Fprintf(out, "\nGuardado csv (%s)\n", state.ReportPath)
	fmt.Fprintln(out, "1) Cambiar ruta CSV actual")
	fmt.Fprintln(out, "2) Actualizar CSV final actual")
	fmt.Fprintln(out, "3) Machacar/regenerar CSV final actual")
	fmt.Fprintln(out, "4) Volver")
	fmt.Fprint(out, "Elige una opción [1-4]: ")
	choice, _ := in.ReadString('\n')
	switch strings.TrimSpace(choice) {
	case "1":
		fmt.Fprint(out, "Ruta CSV: ")
		reportPath, _ := in.ReadString('\n')
		state.ReportPath = strings.TrimSpace(reportPath)
		return state, ErrNoAction
	case "2":
		state.AllImages = true
		state.Replace = false
		return state, nil
	case "3":
		state.AllImages = true
		state.Replace = true
		return state, nil
	case "4":
		return state, ErrNoAction
	default:
		return state, fmt.Errorf("invalid CSV menu choice")
	}
}

func modelLabel(cfg config.RunConfig) string {
	return cfg.Provider + ": " + cfg.Model
}

func ReadMediaCondition(in *bufio.Reader, out io.Writer, current string) (string, error) {
	return readCondition(in, out, "Calidad vinilo", current, config.MediaConditions)
}

func ReadSleeveCondition(in *bufio.Reader, out io.Writer, current string) (string, error) {
	return readCondition(in, out, "Calidad carátula", current, config.SleeveConditions)
}

func readCondition(in *bufio.Reader, out io.Writer, title string, current string, values []string) (string, error) {
	fmt.Fprintf(out, "\n%s (%s)\n", title, current)
	for index, value := range values {
		defaultLabel := ""
		if value == config.DefaultCondition {
			defaultLabel = " [por defecto]"
		}
		fmt.Fprintf(out, "%d) %s%s\n", index+1, value, defaultLabel)
	}
	fmt.Fprintf(out, "Elige calidad [1-%d, Enter=%s]: ", len(values), current)
	choice, _ := in.ReadString('\n')
	trimmed := strings.TrimSpace(choice)
	if trimmed == "" {
		return current, nil
	}
	index, err := strconv.Atoi(trimmed)
	if err != nil || index < 1 || index > len(values) {
		return current, fmt.Errorf("invalid condition choice")
	}
	return values[index-1], nil
}

func ReadProvider(in *bufio.Reader, out io.Writer) (string, string, error) {
	fmt.Fprintln(out, "\nModelo de reconocimiento")
	fmt.Fprintf(out, "1) LM Studio local - %s [por defecto]\n", config.DefaultLMStudioModel)
	fmt.Fprintf(out, "2) LM Studio local - %s\n", config.AlternateLMStudioModel)
	fmt.Fprintf(out, "3) Gemini - %s\n", config.DefaultGeminiModel)
	fmt.Fprint(out, "Elige modelo [1-3, Enter=1]: ")
	choice, _ := in.ReadString('\n')
	switch strings.TrimSpace(choice) {
	case "2":
		return config.ProviderLMStudio, config.AlternateLMStudioModel, nil
	case "3":
		return config.ProviderGemini, config.DefaultGeminiModel, nil
	default:
		return config.ProviderLMStudio, config.DefaultLMStudioModel, nil
	}
}
