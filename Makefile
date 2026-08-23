SHELL := /usr/bin/env bash
.PHONY: check short hooks-test index index-check link-lint law-citation-lint
check: hooks-test link-lint index-check law-citation-lint kernel-check
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
mutants:
	@scripts/mutants.sh "$${PKG:?set PKG, e.g. PKG=internal/store}"
kernel-check:
	@cd kernel && GOCACHE=$${GOCACHE:-/tmp/vera-go-build} GOLANGCI_LINT_CACHE=$${GOLANGCI_LINT_CACHE:-/tmp/vera-golangci} go build ./... && GOCACHE=$${GOCACHE:-/tmp/vera-go-build} go test ./... -count=1 && GOCACHE=$${GOCACHE:-/tmp/vera-go-build} GOLANGCI_LINT_CACHE=$${GOLANGCI_LINT_CACHE:-/tmp/vera-golangci} golangci-lint run ./...
