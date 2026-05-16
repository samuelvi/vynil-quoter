package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"vinylquoter/internal/config"
)

func ReadMenu(in io.Reader, out io.Writer) (config.RunConfig, error) {
	cfg := config.DefaultRunConfig()
	reader := bufio.NewReader(in)
	fmt.Fprintln(out, "\nVinyl Quoter")
	fmt.Fprintln(out, "1) Procesar una imagen concreta")
	fmt.Fprintln(out, "2) Procesar todas las imágenes de data/src")
	fmt.Fprintln(out, "3) Actualizar CSV final por defecto")
	fmt.Fprintln(out, "4) Machacar/regenerar CSV final por defecto")
	fmt.Fprintln(out, "5) Salir")
	fmt.Fprint(out, "Elige una opción [1-5]: ")
	choice, _ := reader.ReadString('\n')
	switch strings.TrimSpace(choice) {
	case "1":
		fmt.Fprint(out, "Ruta o nombre de imagen: ")
		image, _ := reader.ReadString('\n')
		cfg.Image = strings.TrimSpace(image)
	case "2", "3":
		cfg.AllImages = true
	case "4":
		cfg.AllImages = true
		cfg.Replace = true
	case "5":
		return cfg, io.EOF
	default:
		return cfg, fmt.Errorf("invalid menu choice")
	}
	provider, model, err := ReadProvider(reader, out)
	if err != nil {
		return cfg, err
	}
	cfg.Provider = provider
	cfg.Model = model
	return cfg, nil
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
