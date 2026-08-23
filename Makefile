SHELL := /usr/bin/env bash
.PHONY: check short hooks-test index index-check link-lint law-citation-lint
check: hooks-test link-lint index-check law-citation-lint
short: hooks-test
hooks-test:
	@for f in scripts/tests/*.test.sh; do bash "$$f"; done
index:
	@scripts/gen-index.sh
index-check:
	@scripts/index-check.sh
link-lint:
	@scripts/link-lint.sh
law-citation-lint:
	@scripts/law-citation-lint.sh
