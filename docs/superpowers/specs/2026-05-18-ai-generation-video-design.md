# AI Generation Video Design

## Objective

Create an illustrative MP4 video that teaches how AI was used to generate the VinylQuoter project from `prompt.txt`. The video should feel like a professor explaining the workflow at a high level, using an opencode-inspired terminal interface, the open-bundle tooling context, and the main `My-Agent` builder workflow.

## User-approved decisions

- Visual style: guided class with opencode/terminal panels.
- Narration mode: Spanish text on screen, no voice-over.
- Context usage: show illustrative percentages labeled as `Contexto IA aprox.`.
- Recommended approach: step-by-step opencode lesson rather than product-only demo.

## Constraints

- Duration must be between 1 and 3 minutes; target duration is about 90 seconds.
- The video content must be based on `prompt.txt` and the existing project structure.
- The video must show the use of opencode, open-bundle, and the primary `My-Agent` builder agent.
- It must show how the agent detects models, agents, skills, and approximate AI context usage.
- It must be generated locally using the existing video generation path where practical:
  - `tools/video/generate_video.py`
  - `docker/video/Dockerfile`
  - `docker/video/docker-compose.yml`
  - output under `data/video/`
- Generated MP4 files remain ignored by Git; source changes and docs can be versioned.

## Narrative structure

The video uses Spanish educational text in eight scenes:

1. **Starting contract: `prompt.txt`** — introduces the prompt as the source of truth for the project.
2. **My-Agent enters** — shows `My-Agent / Builder Workflow Agent` reading objective, constraints, and acceptance criteria.
3. **Stack detection** — shows Go 1.23, Docker Compose, Makefile, CLI, data directories, and tests.
4. **Model detection** — shows `qwen2.5-vl-7b-instruct`, `gemma-3-4b-it`, and `gemini-2.5-flash-lite`.
5. **Agent and skill detection** — shows builder flow, Memory Bank, planning, TDD, verification, and clean-code skills.
6. **Build by phases** — shows plan, failing tests, implementation, refactor, and docs.
7. **Verification** — shows Docker tests, `make quality`, and strict quality gate.
8. **Final lesson** — shows `data/src → data/dst → model → CSV` and closes with the teaching point: AI works best with context, tools, phases, and verification.

## Visual design

Each scene uses a 16:9 dark terminal classroom layout:

- Top header: `VinylQuoter · AI Generation Lesson · My-Agent` plus current timestamp range.
- Left panel: key excerpts from `prompt.txt` or project requirements.
- Right panel: simulated opencode/My-Agent activity such as detected stack, skills, models, commands, and checks.
- Bottom caption: professor-style Spanish explanation.
- Context meter: `Contexto IA aprox.` with a colored progress bar and percentage.

The visual language should be educational, not flashy: high contrast, readable text, and short phrases.

## Context percentage policy

The context percentages are illustrative, not exact telemetry. Every percentage must be labeled as approximate:

- Scene 1: `12%` — prompt and initial objective loaded.
- Scene 2: `22%` — repo/tooling context added.
- Scene 3: `34%` — stack and acceptance criteria detected.
- Scene 4: `45%` — model/provider requirements identified.
- Scene 5: `58%` — agents, skills, and Memory Bank workflow identified.
- Scene 6: `71%` — implementation phases and tests in context.
- Scene 7: `83%` — verification evidence and quality gate context.
- Scene 8: `88%` — final summary and project output context.

## Acceptance criteria

- `data/video/vinylquoter-ai-demo.mp4` is generated successfully.
- Video duration is at least 60 seconds and at most 180 seconds.
- Video includes Spanish professor-style text explaining each phase.
- Video explicitly references `prompt.txt`, opencode, open-bundle, `My-Agent`, models, agents, skills, and `Contexto IA aprox.`.
- The project test suite remains green.
- The strict quick quality gate passes before closure.

## Verification plan

- Run the video generator and verify MP4 creation.
- Use `ffprobe` or equivalent `ffmpeg` tooling to confirm duration.
- Run `make test` inside Docker.
- Run `python3 .opencode/meta/hooks/quality_gate.py --workspace . --mode quick --strict` from the active workspace before final handoff.
