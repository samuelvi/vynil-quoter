# Menu Confirm Env Defaults Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Confirm exit from the interactive menu and load default menu/config values from `.env` plus process environment variables.

**Architecture:** Keep hardcoded defaults as the base, then overlay `.env`, process environment, and finally CLI flags. Use Go standard library `os.LookupEnv` for process environment and a small local dotenv parser for `KEY=VALUE` files; do not add dependencies.

**Tech Stack:** Go 1.23, Docker Compose test runtime, Makefile, standard library environment handling.

---

### Task 1: Exit Confirmation

**Files:**
- Modify: `tests/internal/ui/menu_test.go`
- Modify: `tests/internal/app/app_test.go`
- Modify: `internal/ui/menu.go`

- [ ] **Step 1: Write failing tests**
- [ ] **Step 2: Run focused UI/app tests and verify RED**
- [ ] **Step 3: Implement confirmation prompt**
- [ ] **Step 4: Run focused UI/app tests and verify GREEN**

### Task 2: Environment Defaults

**Files:**
- Create: `internal/config/env.go`
- Create: `tests/internal/config/env_test.go`
- Modify: `internal/app/app.go`
- Modify: `.gitignore`
- Create: `.env.example`

- [ ] **Step 1: Write failing config/app tests**
- [ ] **Step 2: Run focused config/app tests and verify RED**
- [ ] **Step 3: Implement `.env` parsing and env overlay**
- [ ] **Step 4: Wire `app.ParseArgs` to `config.LoadDefaults()`**
- [ ] **Step 5: Run focused config/app tests and verify GREEN**

### Task 3: Documentation and Final Verification

**Files:**
- Modify: `docs/QUICKSTART.md`
- Modify: `docs/USAGE.md`
- Modify: `docs/DEVELOPMENT.md`
- Modify: `docs/index.md`
- Modify: `README.md`
- Modify: `prompt.txt`

- [ ] **Step 1: Document exit confirmation and env defaults**
- [ ] **Step 2: Run `docker compose -f docker/test/docker-compose.yml config`**
- [ ] **Step 3: Run `git diff --check`**
- [ ] **Step 4: Run `make test`**
- [ ] **Step 5: Run strict quick quality gate**
