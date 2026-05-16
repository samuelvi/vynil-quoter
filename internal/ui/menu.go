package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"vinylquoter/internal/config"
)

var ErrNoAction = errors.New("no menu action selected")

func ReadMenu(in io.Reader, out io.Writer) (config.RunConfig, error) {
	return ReadMenuWithState(in, out, config.DefaultRunConfig())
}

func ReadMenuWithState(in io.Reader, out io.Writer, state config.RunConfig) (config.RunConfig, error) {
	cfg := state
	reader := readerFor(in)
	fmt.Fprintln(out, "\nVinyl Quoter")
	fmt.Fprintln(out, "1) Procesar una imagen concreta")
	fmt.Fprintln(out, "2) Procesar todas las imágenes de data/src")
	fmt.Fprintf(out, "3) Guardado csv (%s)\n", cfg.ReportPath)
	fmt.Fprintf(out, "4) Modelo (%s)\n", modelLabel(cfg))
	fmt.Fprintln(out, "5) Salir")
	fmt.Fprint(out, "Elige una opción [1-5]: ")
	choice, _ := reader.ReadString('\n')
	switch strings.TrimSpace(choice) {
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
		return cfg, io.EOF
	default:
		return cfg, fmt.Errorf("invalid menu choice")
	}
	return cfg, nil
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
