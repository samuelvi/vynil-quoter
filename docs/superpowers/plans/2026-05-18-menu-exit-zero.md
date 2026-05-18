# Menu Exit Zero Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Change the interactive main menu so `0` exits and `7` is invalid.

**Architecture:** Keep option numbers `1..6` unchanged and move only the exit command from `7` to `0`. Update tests first, then menu rendering/dispatch, then user-facing documentation and the regeneration prompt.

**Tech Stack:** Go 1.23, Docker Compose test runtime, Makefile.

---

### Task 1: Main Menu Exit Behavior

**Files:**
- Modify: `tests/internal/ui/menu_test.go`
- Modify: `internal/ui/menu.go`
- Modify: `docs/QUICKSTART.md`
- Modify: `docs/USAGE.md`
- Modify: `prompt.txt`

- [ ] **Step 1: Write failing tests**

Add tests asserting that `0` exits, `7` is invalid, and menu text shows `0) Salir` plus `[0-6]`.

- [ ] **Step 2: Run targeted UI tests to verify RED**

Run: `docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/ui -run 'TestMenuExit|TestMenuShowsExit' -count=1 -v`

Expected: FAIL because current code still exits on `7` and renders `[1-7]`.

- [ ] **Step 3: Implement minimal menu change**

Update `internal/ui/menu.go` to render `0) Salir`, prompt `[0-6]`, and return `io.EOF` for input `0`.

- [ ] **Step 4: Run targeted UI tests to verify GREEN**

Run: `docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/ui -count=1 -v`

Expected: PASS.

- [ ] **Step 5: Update docs and prompt**

Update `docs/QUICKSTART.md`, `docs/USAGE.md`, and `prompt.txt` so the main menu lists `0. Salir` before or after options `1..6` consistently.

- [ ] **Step 6: Run full verification**

Run:

```bash
docker compose -f docker/test/docker-compose.yml config
make test
python3 .opencode/meta/hooks/quality_gate.py --workspace . --mode quick --strict
git status --short
```

Expected: Docker config valid, tests pass, quality gate `ok: true`, only intended files changed before commit.
