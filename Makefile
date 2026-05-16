.PHONY: help run run-all run-all-replace run-gemini test quality clean opencode.init opencode.link opencode.verify opencode.open opencode.start

PYTHON ?= python3
IMAGE ?=

help:
	@printf "VinylQuoter commands:\n"
	@printf "  make run IMAGE=src/DSC01.jpg  Process one image with the interactive/default provider\n"
	@printf "  make run-all                 Update report/album_catalog.csv from all src images\n"
	@printf "  make run-all-replace         Regenerate report/album_catalog.csv from all src images\n"
	@printf "  make run-gemini              Update report using Gemini instead of LM Studio\n"
	@printf "  make test                    Run unit tests\n"
	@printf "  make quality                 Run tests and strict quick quality gate\n"
	@printf "  make clean                   Remove local Python cache files\n"

run:
	@if [ -z "$(IMAGE)" ]; then \
		$(PYTHON) vinyl_quoter.py; \
	else \
		$(PYTHON) vinyl_quoter.py --image "$(IMAGE)"; \
	fi

run-all:
	$(PYTHON) vinyl_quoter.py --all

run-all-replace:
	$(PYTHON) vinyl_quoter.py --all --replace

run-gemini:
	$(PYTHON) vinyl_quoter.py --all --provider gemini

test:
	PYTHONDONTWRITEBYTECODE=1 $(PYTHON) -m unittest discover -s tests

quality: test
	$(PYTHON) .opencode/meta/hooks/quality_gate.py --workspace . --mode quick --strict

clean:
	rm -rf __pycache__ tests/__pycache__ .DS_Store

opencode.init:
	$(MAKE) -C opencode-bundle bundle-init-all

opencode.link:
	$(MAKE) -C opencode-bundle link-parent

opencode.verify:
	$(MAKE) -C opencode-bundle bundle-verify-all

opencode.open:
	$(MAKE) -C opencode-bundle opencode-all

opencode.start: opencode.init opencode.link opencode.verify opencode.open
