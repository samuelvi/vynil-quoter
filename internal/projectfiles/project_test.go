package projectfiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func root(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func read(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(root(t), name))
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func TestMakefileRunsGoInsideTestContainer(t *testing.T) {
	makefile := read(t, "Makefile")
	for _, want := range []string{"$(DOCKER_RUN) $(GO) run $(APP)", "$(DOCKER_RUN) $(GO) run $(APP) --lm-studio-base-url \"$(LM_STUDIO_BASE_URL)\" $(MODEL_ARG) --all", "$(DOCKER_RUN) $(GO) run $(APP) --lm-studio-base-url \"$(LM_STUDIO_BASE_URL)\" $(MODEL_ARG) --all --replace", "$(DOCKER_RUN) $(GO) run $(APP) --all --provider gemini", "$(DOCKER_RUN) $(GO) build -o $(BIN) $(APP)", "$(DOCKER_RUN) $(GO) test ./..."} {
		if !strings.Contains(makefile, want) {
			t.Fatalf("Makefile missing %q", want)
		}
	}
}

func TestDockerVolumesAreExplicitBindMounts(t *testing.T) {
	compose := read(t, "docker/test/docker-compose.yml")
	if strings.Count(compose, "type: bind") != 3 {
		t.Fatalf("compose should define 3 explicit bind mounts:\n%s", compose)
	}
}

func TestGitignoreKeepsOnlyDataPlaceholders(t *testing.T) {
	gitignore := read(t, ".gitignore")
	for _, want := range []string{"data/*", "data/src/*", "!data/src/.gitkeep", "data/dst/*", "!data/dst/.gitkeep", "data/report/*", "!data/report/.gitkeep", "!data/video/", "data/video/*", "!data/video/.gitkeep"} {
		if !strings.Contains(gitignore, want) {
			t.Fatalf(".gitignore missing %q", want)
		}
	}
}

func TestVideoGeneratorProjectFiles(t *testing.T) {
	for _, path := range []string{"docker/video/Dockerfile", "docker/video/docker-compose.yml", "tools/video/generate_video.py", "data/video/.gitkeep"} {
		if _, err := os.Stat(filepath.Join(root(t), path)); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
}

func TestMakefileExposesDockerFirstAppTargets(t *testing.T) {
	makefile := read(t, "Makefile")
	for _, want := range []string{"build:", "run:", "run-cli:", "docker-build:", "docker-shell:", "$(DOCKER_RUN)"} {
		if !strings.Contains(makefile, want) {
			t.Fatalf("Makefile missing %q", want)
		}
	}
	for _, notWant := range []string{"run-local:", "video-build:", "video-generate:", "video-shell:", "$(VIDEO_COMPOSE)", "$(VIDEO_SERVICE)"} {
		if strings.Contains(makefile, notWant) {
			t.Fatalf("Makefile should not expose unsupported command %q", notWant)
		}
	}
}

func TestMakefileAppTargetsAutoPrepareDocker(t *testing.T) {
	makefile := read(t, "Makefile")
	for _, want := range []string{"build: docker-build", "run: docker-build", "run-cli: docker-build", "run-all: docker-build", "run-all-replace: docker-build", "run-gemini: docker-build", "test: docker-build", "docker-shell: docker-build"} {
		if !strings.Contains(makefile, want) {
			t.Fatalf("Makefile target should prepare Docker first: missing %q", want)
		}
	}
}

func TestReadmeDocumentsCLIAndDevelopmentDocs(t *testing.T) {
	readme := read(t, "README.md")
	for _, want := range []string{"make build", "make run", "make run-cli", "docs/QUICKSTART.md", "docs/DEVELOPMENT.md", "DSC01.jpg", "No manual rebuild is needed for normal use", "Go recompiles changed code automatically inside Docker", "Guardado csv", "Modelo (<current provider: model>)", "crops source images into `data/dst`", "analyzes the cropped image"} {
		if !strings.Contains(readme, want) {
			t.Fatalf("README missing %q", want)
		}
	}
	if strings.Contains(readme, "## Direct CLI usage inside the container") {
		t.Fatal("README should document make run-cli instead of direct Go commands")
	}
}

func TestQuickstartDocumentsInteractiveAndFlagUsage(t *testing.T) {
	quickstart := read(t, "docs/QUICKSTART.md")
	for _, want := range []string{"make run", "Make prepares Docker automatically", "No manual rebuild is needed", "Go recompiles changed code automatically inside Docker", "interactive menu", "Guardado csv", "Modelo (<current provider: model>)", "IMAGE=DSC01.jpg", "make run-all", "make run-all-replace", "make run-gemini", "make run-cli", "source_image", "data/src → data/dst → model → CSV"} {
		if !strings.Contains(quickstart, want) {
			t.Fatalf("QUICKSTART missing %q", want)
		}
	}
	for _, notWant := range []string{"go run ./cmd/vinyl-quoter", "bin/vinyl-quoter", "make docker-build", "make test-build", "make test-up", "make test-down"} {
		if strings.Contains(quickstart, notWant) {
			t.Fatalf("QUICKSTART should not include host/development command %q", notWant)
		}
	}
}

func TestDevelopmentDocDocumentsTestsAndDocker(t *testing.T) {
	development := read(t, "docs/DEVELOPMENT.md")
	for _, want := range []string{"make docker-build", "make test-build", "make test-up", "make test", "make quality", "make test-down", "auto-build Docker image", "No manual rebuild is needed for `make run`", "Use `make build` only when you need a binary"} {
		if !strings.Contains(development, want) {
			t.Fatalf("DEVELOPMENT missing %q", want)
		}
	}
}

func TestMakefileHelpDocumentsAutoRecompile(t *testing.T) {
	makefile := read(t, "Makefile")
	for _, want := range []string{"auto-builds Docker", "recompiles changed Go code", "make build", "only when you need bin/vinyl-quoter"} {
		if !strings.Contains(makefile, want) {
			t.Fatalf("Makefile help missing %q", want)
		}
	}
}

func TestPromptDocumentsDockerAutoPreparation(t *testing.T) {
	prompt := read(t, "prompt.txt")
	for _, want := range []string{"make run debe preparar Docker automáticamente", "make test debe preparar Docker automáticamente", "Menú principal exacto", "Guardado csv (<ruta actual>)", "Modelo (<proveedor: modelo actual>)", "El submenú Guardado csv no debe preguntar por modelo", "recortar cada imagen de data/src", "guardar el recorte en data/dst", "analizar la imagen recortada"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt.txt missing %q", want)
		}
	}
}
