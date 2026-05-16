.PHONY: help build run run-cli run-all run-all-replace run-gemini test quality clean docker-build docker-up docker-down docker-shell test-build test-up test-down test-shell opencode.init opencode.link opencode.verify opencode.open opencode.start

IMAGE ?=
MODEL ?=
ARGS ?=
TEST_COMPOSE ?= docker/test/docker-compose.yml
TEST_SERVICE ?= go-scripts
DOCKER_RUN ?= docker compose -f $(TEST_COMPOSE) run --rm $(TEST_SERVICE)
DOCKER_EXEC ?= docker compose -f $(TEST_COMPOSE) exec $(TEST_SERVICE)
GO ?= go
APP ?= ./cmd/vinyl-quoter
BIN ?= bin/vinyl-quoter
LM_STUDIO_BASE_URL ?= http://host.docker.internal:1234/v1
MODEL_ARG := $(if $(MODEL),--model "$(MODEL)",)

help:
	@printf "VinylQuoter commands:\n"
	@printf "  make docker-build                 Build the Docker image used to run the app\n"
	@printf "  make build                        Compile bin/vinyl-quoter inside Docker\n"
	@printf "  make run                          Run the interactive menu inside Docker\n"
	@printf "  make run-cli ARGS='--all'         Run raw CLI flags inside Docker\n"
	@printf "  make run IMAGE=DSC01.jpg          Process one image inside Docker\n"
	@printf "  make run-all                      Update data/report/album_catalog.csv inside Docker\n"
	@printf "  make run-all MODEL=gemma-3-4b-it  Use a specific LM Studio model inside Docker\n"
	@printf "  make run-all-replace              Regenerate data/report/album_catalog.csv inside Docker\n"
	@printf "  make run-gemini                   Update report using Gemini inside Docker\n"
	@printf "  make docker-shell                 Open a shell in the Docker runtime\n"
	@printf "  make test                         Run Go unit tests inside Docker\n"
	@printf "  make quality                 Run tests and strict quick quality gate\n"
	@printf "  make clean                   Remove local Python cache files\n"

build:
	$(DOCKER_RUN) $(GO) build -o $(BIN) $(APP)

run:
	@if [ -z "$(IMAGE)" ]; then \
		$(DOCKER_RUN) $(GO) run $(APP) --lm-studio-base-url "$(LM_STUDIO_BASE_URL)"; \
	else \
		$(DOCKER_RUN) $(GO) run $(APP) --lm-studio-base-url "$(LM_STUDIO_BASE_URL)" $(MODEL_ARG) --image "$(IMAGE)"; \
	fi

run-cli:
	$(DOCKER_RUN) $(GO) run $(APP) --lm-studio-base-url "$(LM_STUDIO_BASE_URL)" $(ARGS)

run-all:
	$(DOCKER_RUN) $(GO) run $(APP) --lm-studio-base-url "$(LM_STUDIO_BASE_URL)" $(MODEL_ARG) --all

run-all-replace:
	$(DOCKER_RUN) $(GO) run $(APP) --lm-studio-base-url "$(LM_STUDIO_BASE_URL)" $(MODEL_ARG) --all --replace

run-gemini:
	$(DOCKER_RUN) $(GO) run $(APP) --all --provider gemini

test:
	$(DOCKER_RUN) $(GO) test ./...

quality: test
	python3 .opencode/meta/hooks/quality_gate.py --workspace . --mode quick --strict

clean:
	rm -rf __pycache__ .DS_Store .cache/go-build bin

docker-build:
	docker compose -f $(TEST_COMPOSE) build

docker-up:
	docker compose -f $(TEST_COMPOSE) up -d

docker-down:
	docker compose -f $(TEST_COMPOSE) down --remove-orphans

docker-shell:
	$(DOCKER_RUN) sh

test-build: docker-build

test-up: docker-up

test-down: docker-down

test-shell:
	docker compose -f $(TEST_COMPOSE) exec $(TEST_SERVICE) sh

opencode.init:
	$(MAKE) -C opencode-bundle bundle-init-all

opencode.link:
	$(MAKE) -C opencode-bundle link-parent

opencode.verify:
	$(MAKE) -C opencode-bundle bundle-verify-all

opencode.open:
	$(MAKE) -C opencode-bundle opencode-all

opencode.start: opencode.init opencode.link opencode.verify opencode.open
