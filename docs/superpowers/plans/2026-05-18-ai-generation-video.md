# AI Generation Video Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Generate a 1–3 minute Spanish educational MP4 showing how `My-Agent` uses `prompt.txt`, opencode, open-bundle, models, agents, skills, and approximate AI context to build VinylQuoter.

**Architecture:** Keep the existing local video generator in `tools/video/generate_video.py`, but upgrade its slide data and renderer into an opencode-inspired classroom layout. Use Go project-file tests to enforce the requested terms and duration bounds because the repository's normal verification path is `make test` inside Docker.

**Tech Stack:** Python 3.12, Pillow, ffmpeg/ffprobe via `docker/video`, Go tests via `make test`, opencode quality gate.

---

## Scope check

This plan covers one subsystem: the local illustrative video generator and its generated MP4 artifact. It does not change VinylQuoter application behavior, provider clients, CSV logic, or Makefile targets.

## File structure

- Modify: `tests/internal/projectfiles/project_test.go` — add regression tests that verify the generator script contains the requested AI workflow concepts and that configured slide duration remains within 60–180 seconds.
- Modify: `tools/video/generate_video.py` — replace the current generic 60-second slide deck with a 90-second opencode/My-Agent lesson layout.
- Generate locally, not committed: `data/video/vinylquoter-ai-demo.mp4` — ignored MP4 output.
- Existing runtime kept unchanged: `docker/video/Dockerfile` and `docker/video/docker-compose.yml`.

## Commit policy

Do not commit during execution unless the user explicitly authorizes it. Use `git status --short` and `git diff` for handoff instead of automatic commit steps.

### Task 1: Add failing generator guardrail tests

**Files:**
- Modify: `tests/internal/projectfiles/project_test.go`

- [ ] **Step 1: Extend imports for duration parsing**

Replace the import block at the top of `tests/internal/projectfiles/project_test.go` with:

```go
import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)
```

- [ ] **Step 2: Add workflow content regression test**

Append this test after `TestVideoGeneratorProjectFiles`:

```go
func TestVideoGeneratorDocumentsAIGenerationWorkflow(t *testing.T) {
	generator := read(t, "tools/video/generate_video.py")
	for _, want := range []string{
		"prompt.txt",
		"opencode",
		"open-bundle",
		"My-Agent",
		"Contexto IA aprox.",
		"qwen2.5-vl-7b-instruct",
		"gemma-3-4b-it",
		"gemini-2.5-flash-lite",
		"brainstorming",
		"writing-plans",
		"test-driven-development",
		"verification-before-completion",
	} {
		if !strings.Contains(generator, want) {
			t.Fatalf("video generator missing %q", want)
		}
	}
}
```

- [ ] **Step 3: Add duration bounds regression test**

Append this test after `TestVideoGeneratorDocumentsAIGenerationWorkflow`:

```go
func TestVideoGeneratorDurationStaysWithinRequestedBounds(t *testing.T) {
	generator := read(t, "tools/video/generate_video.py")
	matches := regexp.MustCompile(`"duration":\s*(\d+)`).FindAllStringSubmatch(generator, -1)
	if len(matches) < 8 {
		t.Fatalf("expected at least 8 video scenes, got %d", len(matches))
	}

	totalSeconds := 0
	for _, match := range matches {
		seconds, err := strconv.Atoi(match[1])
		if err != nil {
			t.Fatalf("invalid duration %q: %v", match[1], err)
		}
		totalSeconds += seconds
	}

	if totalSeconds < 60 || totalSeconds > 180 {
		t.Fatalf("video duration should be 60-180 seconds, got %d", totalSeconds)
	}
}
```

- [ ] **Step 4: Run the new tests and verify they fail before implementation**

Run:

```bash
make test
```

Expected: FAIL in `vinylquoter/tests/internal/projectfiles` because the current generator does not include all required strings such as `open-bundle`, `My-Agent`, and `Contexto IA aprox.`.

### Task 2: Implement the opencode classroom video generator

**Files:**
- Modify: `tools/video/generate_video.py`

- [ ] **Step 1: Replace `tools/video/generate_video.py` with the approved generator**

Replace the whole file with:

```python
#!/usr/bin/env python3
"""Generate a local MP4 lesson about AI-assisted VinylQuoter generation."""

from __future__ import annotations

import shutil
import subprocess
from pathlib import Path
from textwrap import wrap

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parents[2]
PROMPT = ROOT / "prompt.txt"
FRAME_DIR = ROOT / ".cache" / "video" / "frames"
OUTPUT = ROOT / "data" / "video" / "vinylquoter-ai-demo.mp4"
WIDTH = 1280
HEIGHT = 720
FPS = 24

BG = "#08111f"
PANEL = "#0f172a"
PANEL_2 = "#111c33"
BORDER = "#30405f"
TEXT = "#f8fafc"
MUTED = "#b6c2d9"
BLUE = "#60a5fa"
GREEN = "#86efac"
YELLOW = "#fde68a"
PURPLE = "#c4b5fd"
RED = "#fca5a5"

SLIDES = [
    {
        "time": "00:00–00:10",
        "duration": 10,
        "context": 12,
        "title": "1. El contrato inicial está en prompt.txt",
        "subtitle": "La IA no empieza desde cero: recibe un contrato de producto.",
        "prompt": [
            "Proyecto: VinylQuoter",
            "CLI en Go 1.23",
            "Docker + Makefile obligatorios",
            "Pipeline: data/src → data/dst → modelo → CSV",
        ],
        "agent": [
            "$ open prompt.txt",
            "✓ objetivo y restricciones detectadas",
            "✓ criterios de aceptación localizados",
            "✓ producto separado de tooling externo",
        ],
        "caption": "Como profesor, aquí remarcaría que el prompt funciona como contrato: define producto, límites y criterios antes de programar.",
    },
    {
        "time": "00:10–00:22",
        "duration": 12,
        "context": 22,
        "title": "2. My-Agent prepara la sesión",
        "subtitle": "El agente builder restablece objetivo, restricciones y evidencias esperadas.",
        "prompt": [
            "Agente: My-Agent / Builder Workflow Agent",
            "Reglas: pasos pequeños, reversibles y verificables",
            "No cerrar sin pruebas frescas",
            "Quality gate obligatorio",
        ],
        "agent": [
            "$ opencode start --agent My-Agent",
            "✓ rol builder cargado",
            "✓ open-bundle disponible como tooling",
            "✓ Memory Bank consultado",
        ],
        "caption": "My-Agent no improvisa: primero carga contexto operativo, memoria del proyecto y reglas de entrega.",
    },
    {
        "time": "00:22–00:34",
        "duration": 12,
        "context": 34,
        "title": "3. Detección de stack y rutas",
        "subtitle": "El agente detecta qué herramientas usar antes de tocar código.",
        "prompt": [
            "go.mod → Go",
            "docker/test/docker-compose.yml → runtime de tests",
            "Makefile → interfaz operativa",
            "README/docs → comportamiento esperado",
        ],
        "agent": [
            "$ git status && git branch",
            "$ make test",
            "✓ Docker-first confirmado",
            "✓ data/video reservado para material audiovisual",
        ],
        "caption": "La detección de stack evita comandos inventados: en este proyecto todo se valida con Makefile y Docker.",
    },
    {
        "time": "00:34–00:44",
        "duration": 10,
        "context": 45,
        "title": "4. Modelos de visión detectados",
        "subtitle": "El prompt lista proveedores y modelos que la app debe saber manejar.",
        "prompt": [
            "LM Studio por defecto: qwen2.5-vl-7b-instruct",
            "LM Studio alternativo: gemma-3-4b-it",
            "Gemini opcional: gemini-2.5-flash-lite",
            "No imprimir secretos ni llamar red en tests",
        ],
        "agent": [
            "✓ provider/Recognizer como frontera",
            "✓ lmstudio client",
            "✓ gemini client",
            "✓ visionpayload compartido",
        ],
        "caption": "Aquí My-Agent separa modelos de arquitectura: los proveedores cambian, pero el contrato interno permanece estable.",
    },
    {
        "time": "00:44–00:56",
        "duration": 12,
        "context": 58,
        "title": "5. Agentes y skills activados",
        "subtitle": "El flujo combina planificación, TDD y verificación.",
        "prompt": [
            "brainstorming → entender y aprobar diseño",
            "writing-plans → plan ejecutable",
            "test-driven-development → rojo, verde, refactor",
            "verification-before-completion → evidencia antes de cerrar",
        ],
        "agent": [
            "✓ My-Agent coordina",
            "✓ tester/checker como revisores implícitos",
            "✓ clean-code para mantenibilidad",
            "✓ Memory Bank guarda decisiones",
        ],
        "caption": "La clave docente: la IA mejora cuando usa skills concretas en lugar de responder de memoria.",
    },
    {
        "time": "00:56–01:09",
        "duration": 13,
        "context": 71,
        "title": "6. Construcción por fases pequeñas",
        "subtitle": "Cada fase produce algo comprobable.",
        "prompt": [
            "1. Imagen de entrada",
            "2. Preparación local en data/dst",
            "3. Reconocimiento por modelo",
            "4. Upsert en album_catalog.csv",
        ],
        "agent": [
            "FAIL test esperado",
            "implementación mínima",
            "PASS paquete afectado",
            "refactor + docs",
        ],
        "caption": "El proyecto se genera como lo haría un equipo: pruebas primero, implementación pequeña y documentación sincronizada.",
    },
    {
        "time": "01:09–01:19",
        "duration": 10,
        "context": 83,
        "title": "7. Evidencia antes de cerrar",
        "subtitle": "La IA debe demostrar que lo construido funciona.",
        "prompt": [
            "make test pasa dentro de Docker",
            "make quality pasa",
            "docker compose config válido",
            "quality_gate.py --strict",
        ],
        "agent": [
            "$ make test",
            "ok vinylquoter/tests/internal/app",
            "ok vinylquoter/tests/internal/projectfiles",
            "quality gate: ok true",
        ],
        "caption": "No basta con decir 'he terminado': My-Agent necesita pruebas, salida de comandos y una puerta de calidad.",
    },
    {
        "time": "01:19–01:30",
        "duration": 11,
        "context": 88,
        "title": "8. Resultado: IA guiada por contexto",
        "subtitle": "VinylQuoter queda como una CLI Go reproducible y documentada.",
        "prompt": [
            "Entrada: data/src",
            "Preparación: data/dst",
            "Modelo: LM Studio o Gemini",
            "Salida: data/report/album_catalog.csv",
        ],
        "agent": [
            "✓ opencode organiza el trabajo",
            "✓ open-bundle aporta tooling",
            "✓ My-Agent mantiene disciplina",
            "✓ contexto + fases + tests = entrega fiable",
        ],
        "caption": "Lección final: la IA no sustituye el método; lo acelera cuando el prompt, los agentes y la verificación están bien definidos.",
    },
]


def font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont:
    candidates = [
        "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf" if bold else "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
        "/Library/Fonts/Arial.ttf",
    ]
    for candidate in candidates:
        try:
            return ImageFont.truetype(candidate, size)
        except OSError:
            continue
    return ImageFont.load_default()


def draw_wrapped(draw: ImageDraw.ImageDraw, text: str, xy: tuple[int, int], max_chars: int, fill: str, text_font: ImageFont.FreeTypeFont, line_gap: int = 6) -> int:
    x, y = xy
    for line in wrap(text, width=max_chars):
        draw.text((x, y), line, fill=fill, font=text_font)
        y += text_font.size + line_gap
    return y


def draw_header(draw: ImageDraw.ImageDraw, slide: dict[str, object]) -> None:
    header_font = font(22, bold=True)
    small_font = font(18)
    draw.rounded_rectangle((42, 28, 1238, 82), radius=14, fill=PANEL_2, outline=BORDER, width=2)
    draw.text((66, 44), "VinylQuoter · AI Generation Lesson · My-Agent", fill=TEXT, font=header_font)
    draw.text((1010, 46), str(slide["time"]), fill=BLUE, font=small_font)


def draw_context_meter(draw: ImageDraw.ImageDraw, percent: int) -> None:
    label_font = font(18, bold=True)
    draw.text((868, 104), f"Contexto IA aprox. {percent}%", fill=YELLOW, font=label_font)
    draw.rounded_rectangle((868, 134, 1196, 150), radius=8, fill="#1e293b")
    fill_width = int(328 * percent / 100)
    draw.rounded_rectangle((868, 134, 868 + fill_width, 150), radius=8, fill="#22c55e")


def draw_panel(draw: ImageDraw.ImageDraw, title: str, lines: list[str], box: tuple[int, int, int, int], accent: str) -> None:
    x1, y1, x2, y2 = box
    title_font = font(20, bold=True)
    body_font = font(22)
    draw.rounded_rectangle(box, radius=18, fill=PANEL, outline=BORDER, width=2)
    draw.rectangle((x1, y1, x2, y1 + 44), fill="#101a2f")
    draw.text((x1 + 18, y1 + 12), title, fill=accent, font=title_font)

    y = y1 + 70
    for line in lines:
        marker_color = GREEN if line.startswith("✓") else BLUE if line.startswith("$") else YELLOW
        draw.text((x1 + 24, y), "▸", fill=marker_color, font=body_font)
        y = draw_wrapped(draw, line, (x1 + 56, y), 36, TEXT, body_font, line_gap=4)
        y += 8
        if y > y2 - 38:
            break


def draw_slide(slide: dict[str, object], path: Path) -> None:
    img = Image.new("RGB", (WIDTH, HEIGHT), BG)
    draw = ImageDraw.Draw(img)
    title_font = font(40, bold=True)
    subtitle_font = font(24)
    caption_font = font(23)
    footer_font = font(17, bold=True)

    draw_header(draw, slide)
    draw_context_meter(draw, int(slide["context"]))
    draw.text((64, 108), str(slide["title"]), fill=TEXT, font=title_font)
    draw.text((66, 160), str(slide["subtitle"]), fill=MUTED, font=subtitle_font)

    draw_panel(draw, "PROMPT.TXT / REQUISITOS", list(slide["prompt"]), (64, 218, 604, 538), BLUE)
    draw_panel(draw, "OPENCODE + MY-AGENT", list(slide["agent"]), (636, 218, 1216, 538), PURPLE)

    draw.rounded_rectangle((64, 560, 1216, 646), radius=18, fill="#172554", outline="#3b82f6", width=2)
    draw_wrapped(draw, "Texto profesor: " + str(slide["caption"]), (88, 584), 94, TEXT, caption_font, line_gap=5)
    draw.text((66, 674), "Sin voz · texto narrativo en pantalla · generado localmente con Pillow + ffmpeg", fill="#7d8caf", font=footer_font)

    img.save(path)


def main() -> None:
    if not PROMPT.exists():
        raise SystemExit("prompt.txt not found")

    shutil.rmtree(FRAME_DIR, ignore_errors=True)
    FRAME_DIR.mkdir(parents=True, exist_ok=True)
    OUTPUT.parent.mkdir(parents=True, exist_ok=True)

    frame_index = 0
    for slide in SLIDES:
        slide_path = FRAME_DIR / f"slide-{frame_index:03d}.png"
        draw_slide(slide, slide_path)
        repeat = FPS * int(slide["duration"])
        for _ in range(repeat):
            frame_path = FRAME_DIR / f"frame-{frame_index:05d}.png"
            shutil.copy(slide_path, frame_path)
            frame_index += 1

    subprocess.run(
        [
            "ffmpeg",
            "-y",
            "-framerate",
            str(FPS),
            "-i",
            str(FRAME_DIR / "frame-%05d.png"),
            "-c:v",
            "libx264",
            "-pix_fmt",
            "yuv420p",
            "-movflags",
            "+faststart",
            str(OUTPUT),
        ],
        check=True,
    )
    print(f"Video generated: {OUTPUT}")


if __name__ == "__main__":
    main()
```

- [ ] **Step 2: Run the guardrail tests and verify they pass**

Run:

```bash
make test
```

Expected: PASS for `vinylquoter/tests/internal/projectfiles` and all existing packages.

### Task 3: Generate and verify the MP4 artifact

**Files:**
- Generate ignored output: `data/video/vinylquoter-ai-demo.mp4`

- [ ] **Step 1: Generate the video inside the existing video container**

Run:

```bash
docker compose -f docker/video/docker-compose.yml run --rm video-generator python3 tools/video/generate_video.py
```

Expected output includes:

```text
Video generated: /workspace/data/video/vinylquoter-ai-demo.mp4
```

- [ ] **Step 2: Verify the MP4 duration is within the accepted range**

Run:

```bash
docker compose -f docker/video/docker-compose.yml run --rm video-generator ffprobe -v error -show_entries format=duration -of default=nk=1:nw=1 data/video/vinylquoter-ai-demo.mp4
```

Expected: a numeric duration between `60.000000` and `180.000000`, approximately `90.000000`.

- [ ] **Step 3: Confirm generated MP4 remains ignored**

Run:

```bash
git status --short --ignored data/video/vinylquoter-ai-demo.mp4
```

Expected:

```text
!! data/video/vinylquoter-ai-demo.mp4
```

### Task 4: Final verification and handoff evidence

**Files:**
- Read-only verification of repository state.

- [ ] **Step 1: Run the full Docker test suite**

Run:

```bash
make test
```

Expected: all Go packages pass, including `vinylquoter/tests/internal/projectfiles`.

- [ ] **Step 2: Prepare opencode quality gate in the worktree if needed**

Run:

```bash
if [ ! -f .opencode/meta/hooks/quality_gate.py ]; then git submodule update --init opencode-bundle && make opencode.init; fi
```

Expected: `.opencode/meta/hooks/quality_gate.py` exists after the command. If submodule initialization is unavailable because of credentials or network access, report that blocker and run the same quality gate command from a workspace where `.opencode` is already initialized.

- [ ] **Step 3: Run the mandatory strict quick quality gate**

Run:

```bash
python3 .opencode/meta/hooks/quality_gate.py --workspace . --mode quick --strict
```

Expected: output indicates `ok: true` or equivalent successful strict quick gate status.

- [ ] **Step 4: Review diff and status for handoff**

Run:

```bash
git status --short && git diff -- tests/internal/projectfiles/project_test.go tools/video/generate_video.py docs/superpowers/specs/2026-05-18-ai-generation-video-design.md docs/superpowers/plans/2026-05-18-ai-generation-video.md
```

Expected: only the plan/spec and intended source/test files are modified or added; `data/video/vinylquoter-ai-demo.mp4` remains ignored.
