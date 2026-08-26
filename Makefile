SHELL := /usr/bin/env bash
.PHONY: check check-witnessed index-check-witnessed law-citation-witnessed spec-numbering-witnessed verify gates-canary gates-enforce short hooks-test index index-check invariants-lock invariant-table-lint spec-numbering-lint link-lint law-citation-lint
check: hooks-test link-lint index-check law-citation-lint invariant-table-lint spec-numbering-lint kernel-check
check-witnessed:
	@bash kernel/scripts/check-witness.sh
index-check-witnessed:
	@VERA_CHECK_TARGET=index-check bash kernel/scripts/check-witness.sh
law-citation-witnessed:
	@VERA_CHECK_TARGET=law-citation-lint bash kernel/scripts/check-witness.sh
spec-numbering-witnessed:
	@VERA_CHECK_TARGET=spec-numbering-lint bash kernel/scripts/check-witness.sh
verify:
	@cd kernel && go run ./cmd/vera verify
gates-canary:
	@cd kernel && go run ./cmd/vera gates canary
gates-enforce:
	@cd kernel && go run ./cmd/vera gates enforce
short: hooks-test
hooks-test:
	@for f in scripts/tests/*.test.sh; do bash "$$f"; done
index:
	@scripts/gen-index.sh
index-check:
	@scripts/index-check.sh
invariants-lock:
	@scripts/gen-invariants-lock.sh
invariant-table-lint:
	@scripts/invariant-table-lint.sh
spec-numbering-lint:
	@scripts/spec-numbering-lint.sh
link-lint:
	@scripts/link-lint.sh
law-citation-lint:
	@scripts/law-citation-lint.sh
mutants:
	@scripts/mutants.sh "$${PKG:?set PKG, e.g. PKG=internal/store}"
kernel-check:
	@cd kernel && GOCACHE=$${GOCACHE:-/tmp/vera-go-build} GOLANGCI_LINT_CACHE=$${GOLANGCI_LINT_CACHE:-/tmp/vera-golangci} go build ./... && GOCACHE=$${GOCACHE:-/tmp/vera-go-build} go test ./... -count=1 && GOCACHE=$${GOCACHE:-/tmp/vera-go-build} GOLANGCI_LINT_CACHE=$${GOLANGCI_LINT_CACHE:-/tmp/vera-golangci} golangci-lint run ./...
