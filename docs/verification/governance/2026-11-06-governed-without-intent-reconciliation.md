# Governance reconciliation — two governed commits that landed without an `Intent:` trailer

**Owner decision 2026-11-06 (recorded by the agent as scribe; owner is the acceptor, Law 9):**
reconcile by **intent-linkage** — each commit below is sanctioned by linking it to the already
owner-authorized work it is a consequence of, so the detector-power work may land on a clean ledger.
This record is the linkage; it creates no new capability and does not weaken any gate.

## Why `make check` stayed green over these commits

`scripts/delivery-enforce.sh` (the change-intent / delivery gate) enforces only **HEAD**, by design —
it asserts about the current commit, which is why a mid-history commit lacking an `Intent:` trailer is
not caught by a later `make check`. Both commits below are governed-path changes that landed without a
trailer; neither is dishonest, but neither was linked, so this record links them.

## The two commits

### `daf9fe4` — `decide(p7): resolve DEF-P7-001 — B1 provenance+licensing+runtime-evidence bar met`
- **Governed paths touched:** `docs/deferrals.tsv` (DEF-P7-001 → RESOLVED), `docs/decisions/INDEX.md`
  (`@generated` regeneration, Law 1), and the new decision record
  `docs/decisions/VD-p7-def001-resolve-2026-11-05.md`.
- **Authorizing work it is a consequence of:** the **owner's go/no-go on DEF-P7-001**, recorded in
  [`VD-p7-def001-resolve-2026-11-05`](../../decisions/VD-p7-def001-resolve-2026-11-05.md). A decision
  record (a VD) IS the authorization; the change-intent reserved this resolution as a separate owner
  decision, which the owner then made. The INDEX.md change is mechanical derived-state (Law 1).
- **Sanction:** accepted as owner-authorized. No `Intent:` trailer is required for a decision record;
  the governed `deferrals.tsv` flip is the direct expression of the owner's decision.

### `6b04f0a` — `fix(gates): restore make check green — hooks-test regressions from today's owner-authorized changes`
- **Governed paths touched:** `scripts/tests/deferrals.test.sh` (repointed from the now-RESOLVED
  DEF-P7-001 to the live OPEN rows), intent-record prose + digests under
  `docs/intent/records/.../BD-witnessed-regeneration-9c4e21`, `BR-regeneration-trust-5f8a03`,
  `CI-p7-detector-power-1a2b3c` (requirement moved `active`→`proposed` until its non-author review),
  `notes/state.md`/journal de-tokenization of the then-unbuilt `scripts/detector-power.sh` path, and a
  `scripts/p6-route-matrix.sh --write` regeneration.
- **Authorizing work it is a consequence of:** the same **DEF-P7-001 resolution** (`daf9fe4`) and the
  **#2b intent chain acceptance** (`830f6b3`). Every edit is a maintenance correction consequent to
  that owner-authorized work — a self-test repointed to match the resolved register, an intent-status
  correction, and `@generated` regenerations — introducing **no new capability**.
- **Sanction:** accepted as owner-authorized maintenance consequent to `daf9fe4`/`830f6b3`.

## Law 7 note (flagged, not resolved here)

"Governed commit without an `Intent:` trailer" has now been observed on these two commits. Law 7 says
a lesson class seen twice must compile to enforcement — but `delivery-enforce` is **HEAD-only by
design** (it cannot assert about non-HEAD commits), so range-level intent-trailer enforcement is a
separate enforcement-design question, not something this reconciliation closes. It is left for the
owner as a follow-on (candidate: a pre-push / range check that every governed-path commit since a
baseline carries a trailer or is listed in a reconciliation record like this one). Flagged in
`notes/state.md`.
