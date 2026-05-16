.PHONY: help run run-all run-all-replace run-gemini test quality clean test-build test-up test-down test-shell opencode.init opencode.link opencode.verify opencode.open opencode.start

IMAGE ?=
TEST_COMPOSE ?= docker/test/docker-compose.yml
TEST_SERVICE ?= go-scripts
TEST_RUN ?= docker compose -f $(TEST_COMPOSE) exec $(TEST_SERVICE)
GO ?= go
APP ?= ./cmd/vinyl-quoter

help:
	@printf "VinylQuoter commands:\n"
	@printf "  make run IMAGE=data/src/DSC01.jpg  Process one image inside the test container\n"
	@printf "  make run-all                 Update data/report/album_catalog.csv inside the test container\n"
	@printf "  make run-all-replace         Regenerate data/report/album_catalog.csv inside the test container\n"
	@printf "  make run-gemini              Update report using Gemini inside the test container\n"
	@printf "  make test-build              Build the Go scripts test image\n"
	@printf "  make test-up                 Start the Go scripts test container\n"
	@printf "  make test-down               Stop the Go scripts test container\n"
	@printf "  make test-shell              Open a shell in the Go scripts test container\n"
	@printf "  make test                    Run Go unit tests inside the test container\n"
	@printf "  make quality                 Run tests and strict quick quality gate\n"
	@printf "  make clean                   Remove local Python cache files\n"

run:
	@if [ -z "$(IMAGE)" ]; then \
		$(TEST_RUN) $(GO) run $(APP); \
	else \
		$(TEST_RUN) $(GO) run $(APP) --image "$(IMAGE)"; \
	fi

run-all:
	$(TEST_RUN) $(GO) run $(APP) --all

run-all-replace:
	$(TEST_RUN) $(GO) run $(APP) --all --replace

run-gemini:
	$(TEST_RUN) $(GO) run $(APP) --all --provider gemini

test:
	$(TEST_RUN) $(GO) test ./...

quality: test
	python3 .opencode/meta/hooks/quality_gate.py --workspace . --mode quick --strict

clean:
	rm -rf __pycache__ .DS_Store .cache/go-build

test-build:
	docker compose -f $(TEST_COMPOSE) build

test-up:
	docker compose -f $(TEST_COMPOSE) up -d

test-down:
	docker compose -f $(TEST_COMPOSE) down --remove-orphans

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
