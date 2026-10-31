---
schema: vera.verdict.v1
verdict_id: p7-twin-manifest-acceptance-round1-20261031
status: ACCEPTABLE
reviewed_commit: 87577ce431a9db0bad91e73f807ba8fb57a6280b
findings: []
artifact_path: docs/verification/verdicts/p7-twin-manifest-acceptance-round1-20261031.md
artifact_sha: 7e9b9fbfdf197483c085c118c6721eb83bea9d8197ab0fdcdaba874bae812cad
---

# P7 `internal/twin/manifest` — package acceptance (Task-8 bar), non-author verdict

**ACCEPTABLE.** `internal/twin/manifest` is accepted as the 17th production package under the P6
Task-8 acceptance bar (calibrated mutation evidence + a non-author verdict). It landed after the
original P6 freeze as authorized P7 non-runtime work (`CI-p7-projector-manifest-declaration-1a2b3c`)
and had not yet been bound into the frozen package set; `DEF-P7-003` tracked that gap.

## Binding

- Reviewed commit: `87577ce431a9db0bad91e73f807ba8fb57a6280b`.
- Production package tree object: `git rev-parse 87577ce:kernel/internal/twin/manifest` =
  `ac36cf84b258abed173fc5d102fe4cda604789a6`. The production `.go` files are byte-identical to the
  tree the non-author code reviewer examined; the acceptance commit that carries this verdict adds
  only mutation-kill `_test.go` coverage and documentation, no production-source change.

## Non-author code review

An independent reviewer that did **not** author the package reviewed every non-test file against the
package SPEC and the Build Laws, ran `go test` (18 tests PASS) and `go vet` (clean), and audited
imports (stdlib + `internal/core` only — no `os`, `os/exec`, `net`, `database/sql`, `pgx`, or any
store/ledger/lock reference). Verdict: **ACCEPTABLE**, no open defect, no broken SPEC invariant, no
Build-Law violation. It confirmed the package stays within its SPEC non-goals: no filesystem access,
git invocation, network, binary retrieval/build/execution, writable truth, lock evaluator, PostgreSQL
route, provider, or runtime path — all committed evidence is obtained only through the caller-injected
`CommittedEvidenceReader`. Two non-blocking notes (an invariant→test citation that points at an
adjacent test, and two guarantees covered mechanically but not named by a case) carried no reproducible
harm. The package is imported by nothing else in `kernel/` — a standalone non-runtime contract
validator, correctly the connector/peripheral tier (one adversarial review).

## Calibrated mutation evidence

`scripts/mutants.sh internal/twin/manifest` (stdlib-only deterministic operators). **Calibration
controls behaved: neutral = survived, invalid = invalid, lethal = killed** — so the counts are
believed.

| Candidates | Killed | Invalid | Survived (declared) |
|---|---|---|---|
| 190 | 185 | 0 | 5 |

The sweep first exposed 25 survivors — real per-branch coverage gaps in compound `||`/`&&`
conditions. They were closed with targeted kill-tests (`mutation_kill_test.go`) that make exactly one
sub-condition true per case, plus a positive test pinning the executable `100755` source mode. The
remaining **5 survivors are all EQUIVALENT mutants, declared in `docs/allowed-survivors.txt`**: four
are input-layer validation checks in `validateManifestDeclarationInput` that `DeclareCommittedManifest`
re-runs via `Bind→Validate` on the built manifest (so no input changes the accept/reject outcome by
mutating the earlier copy), and one is a defensive `exactObject` nil-guard masked by the downstream
required-field loop. Zero undeclared survivors.

## Law 9 note

The non-author code-review verdict (ACCEPTABLE) is the independent judgment of the production code.
The mutation-kill tests were authored by the acceptance session, but their efficacy is established
objectively by the calibrated sweep (185/190 killed with declared equivalents), not by author
self-assessment. The owner (Srinivas Kamisetty), as a non-author, authorized and accepts this package
into the frozen set under the standing authorization (`VD-agent-standing-authorization-5ouz7q`).
