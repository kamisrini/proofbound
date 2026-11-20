# MVP step 2: the pull-request check — spec

**Status:** built 2026-11-20 (as amended in §6), for one non-author combined review (Law 11). **Decision:** `VD-vertical-mvp-first-ai3dhn`
step 2. **Intent:** `CI-mvp-pr-check-8hq2rm` → `BR-vertical-mvp-xc31gp` O-2.

## 1. Payoff

A team adds one workflow file to their repository and every pull request gets a Proofbound result in
the Checks tab and the job summary:

```yaml
# .github/workflows/proofbound.yml  (written by `proofbound init`, or copied by hand)
on: pull_request
jobs:
  proofbound:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: kamisrini/proofbound@v0     # the composite action at this repository's root
```

```
Proofbound: NOT WARRANTED (1 of 2 claiming commits verified)
  a1b2c3d4e5f6 sdd:add-csv-export   PASS  2 of 2 obligations verified SATISFIED; 2 of 2 reviews VERIFIABLE
  0f9e8d7c6b5a sdd:add-pdf-export   NO REVIEW  no readiness gate gates/ready-0f9e8d7c6b5a.yaml; run `proofbound review draft 0f9e8d7c6b5a`
Drift: add-csv-export CURRENT
```

## 2. Interface

### 2.1 `proofbound check <base>` (kernel, `internal/cli`)

Reads the ledger after the caller's `proofbound sync all`. For every commit in `<base>..HEAD`
(git rev-list order, oldest first) whose message carries an `Intent: sdd:<name>` claim, it reports
one line `<commit12> sdd:<name> <STATE> <reason>`:

| STATE | When |
|---|---|
| `PASS` | `gates/ready-<commit12>.yaml` exists at HEAD with rule `intent-delivery-readiness` scoped to the commit, and evaluates PASS (the same evaluation `warranty export` uses) |
| `BLOCKED` | that gate exists and evaluates to anything else; reason = the gate's reason |
| `NO REVIEW` | no such gate; reason names the `review draft` command |
| `UNBOUND` | the commit is not in the ledger with a bound sdd ref (claim skipped); reason = "run proofbound sync all" or the skip cause |

The first line is `Proofbound: WARRANTED (n of n claiming commits verified)` when every claiming
commit is `PASS` and there is at least one, `Proofbound: NOT WARRANTED (k of n …)` when some is not,
and `Proofbound: NO CLAIMS (no commit in <base>..HEAD claims a spec)` when none claims. Then the
`report drift` output for the changes those commits claim. Exit codes: 0 for WARRANTED and NO CLAIMS,
1 for NOT WARRANTED, 2 for usage. `--markdown` renders the same content as a Markdown table for
`$GITHUB_STEP_SUMMARY`. It writes nothing and never evaluates gates in enforce mode.

### 2.2 `action.yml` (repository root, composite action)

Inputs: `version` (default `latest`), `base` (default `origin/${{ github.base_ref }}`),
`require-claims` (default `false`: NO CLAIMS passes). Steps: run `install.sh` with
`PROOFBOUND_VERSION`; `proofbound sync all`; `proofbound check <base> --markdown >>
$GITHUB_STEP_SUMMARY`, then the plain form for the log; the step's exit code is the check's.
No token is needed (no GitHub API call); the result is the job's pass/fail.

### 2.3 `proofbound init` writes the workflow

`init {openspec|speckit}` also writes `proofbound.yml` in the tenant's `.github/workflows/` (the block in §1), under the
starter kit's never-overwrite rule (CLI-INV-18).

## 3. Invariants and proving tests

| ID | Statement | Proving test |
|---|---|---|
| CHK-1 | Each claiming commit in range gets exactly one state from §2.1's table; non-claiming commits are not listed | `cli/check_test.go` + an sdd end-to-end test (claim, no review → NO REVIEW; seal → PASS; seal with a NEEDS_WORK outcome → BLOCKED) |
| CHK-2 | WARRANTED requires ≥1 claim and all PASS; exit codes as §2.1 | same |
| CHK-3 | `check` writes no file and appends no ledger event | ledger length and `git status` unchanged |
| CHK-4 | `init` writes the workflow once and never overwrites it | `starter_test.go` (CLI-INV-18 extended) |
| ACT-1 | `action.yml` runs install, sync, check, and fails the step iff check exits non-zero | `scripts/tests/action.test.sh` runs the action's steps as a script with a fake `GITHUB_STEP_SUMMARY` |

## 4. Non-goals

- Posting PR comments or check-run annotations through the GitHub API (needs a token; later).
- Running the tenant's tests: the check reads verdicts people recorded; it does not re-run anything.
- Gating merges: branch protection is the team's choice; the action only reports.
- Publishing the action: pushes are held; the action is proven locally (ACT-1) and on a real PR once
  the owner publishes.

## 5. Open question for the owner (default chosen)

Should a PR with **no** spec-claiming commits pass? Default **yes** (`require-claims: false`), so
adopting Proofbound never blocks unrelated PRs; a team can flip it.

## 6. As built (amendments to draft 1)

- `check` takes `--require-claims` itself (the action passes it through), so the CLI alone gives the
  same answer as the action.
- Drift: `check` prints the whole `report drift` output under `Drift:` (Markdown: a collapsed block),
  not a per-claim filter; it is skipped when nothing in the range claims a spec.
- UNBOUND is decided first: a claim the ledger does not hold bound gets no gate lookup.
- Invariants: CHK-1..3 are CLI-INV-26; CHK-4 is CLI-INV-18 (amended); ACT-1 is
  `scripts/tests/action.test.sh`, which drives `scripts/action-check.sh` (the action's body) with a
  stub `proofbound`.
- The workflow `init` writes uses `kamisrini/proofbound@main` until a release tag exists.
