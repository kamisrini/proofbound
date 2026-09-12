SHELL := /usr/bin/env bash
.PHONY: check check-witnessed index-check-witnessed law-citation-witnessed spec-numbering-witnessed invariant-table-witnessed link-witnessed kernel-check-witnessed delivery-enforce verify gates-canary gates-enforce short hooks-test index index-check invariants-lock laws-lock invariant-table-lint invariant-lint spec-numbering-lint link-lint law-citation-lint identity-inventory commit-cadence state-freshness cleanroom-lint lesson-recurrence figure-provenance prescription-lint meta-tax backup state wrap-verify
check: hooks-test commit-cadence state-freshness cleanroom-lint lesson-recurrence figure-provenance prescription-lint link-lint index-check law-citation-lint invariant-table-lint invariant-lint spec-numbering-lint identity-inventory kernel-check
check-witnessed:
	@bash kernel/scripts/check-witness.sh
index-check-witnessed:
	@PROOFBOUND_CHECK_TARGET=index-check bash kernel/scripts/check-witness.sh
law-citation-witnessed:
	@PROOFBOUND_CHECK_TARGET=law-citation-lint bash kernel/scripts/check-witness.sh
spec-numbering-witnessed:
	@PROOFBOUND_CHECK_TARGET=spec-numbering-lint bash kernel/scripts/check-witness.sh
invariant-table-witnessed:
	@PROOFBOUND_CHECK_TARGET=invariant-table-lint bash kernel/scripts/check-witness.sh
link-witnessed:
	@PROOFBOUND_CHECK_TARGET=link-lint bash kernel/scripts/check-witness.sh
kernel-check-witnessed:
	@PROOFBOUND_CHECK_TARGET=kernel-check bash kernel/scripts/check-witness.sh
delivery-enforce:
	@bash scripts/delivery-enforce.sh
verify:
	@cd kernel && go run ./cmd/proofbound verify
gates-canary:
	@cd kernel && go run ./cmd/proofbound gates canary
gates-enforce:
	@cd kernel && go run ./cmd/proofbound gates enforce
short: hooks-test
	@cd kernel && GOCACHE=$${GOCACHE:-/tmp/proofbound-go-build} go test ./... -short -count=1
hooks-test:
	@for f in scripts/tests/*.test.sh; do bash "$$f"; done
index:
	@scripts/gen-index.sh
index-check:
	@scripts/index-check.sh
invariants-lock:
	@scripts/gen-invariants-lock.sh
laws-lock:
	@scripts/gen-laws-lock.sh
invariant-table-lint:
	@scripts/invariant-table-lint.sh
invariant-lint:
	@scripts/invariant-lint.sh
spec-numbering-lint:
	@scripts/spec-numbering-lint.sh
link-lint:
	@scripts/link-lint.sh
law-citation-lint:
	@scripts/law-citation-lint.sh
identity-inventory:
	@bash scripts/identity-inventory.sh >/dev/null
commit-cadence:
	@scripts/commit-cadence.sh
state-freshness:
	@scripts/state-freshness.sh
cleanroom-lint:
	@scripts/cleanroom-lint.sh
lesson-recurrence:
	@scripts/lesson-recurrence.sh
figure-provenance:
	@scripts/figure-provenance.sh
prescription-lint:
	@scripts/prescription-lint.sh
meta-tax:
	@scripts/meta-tax.sh
backup:
	@scripts/backup.sh
state:
	@scripts/gen-state.sh
wrap-verify:
	@scripts/wrap-verify.sh
mutants:
	@scripts/mutants.sh "$${PKG:?set PKG, e.g. PKG=internal/store}"
kernel-check:
	@scripts/kernel-check.sh
