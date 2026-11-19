---
schema: vera.verdict.v1
verdict_id: r2-m3-impl-review-mutation-concurrency-20261119
status: ACCEPTABLE
reviewed_commit: 2d67fc03c1900d9df7de57a841401d9a9d74c6fe
findings:
  - finding_id: r2-m3-impl-review-mutation-concurrency-20261119-F1
    severity: LOW
  - finding_id: r2-m3-impl-review-mutation-concurrency-20261119-F2
    severity: LOW
  - finding_id: r2-m3-impl-review-mutation-concurrency-20261119-F3
    severity: LOW
  - finding_id: r2-m3-impl-review-mutation-concurrency-20261119-F4
    severity: LOW
  - finding_id: r2-m3-impl-review-mutation-concurrency-20261119-F5
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m3-impl-review-mutation-concurrency-20261119.md
artifact_sha: eaea8ac0d78a8296975bb75bef0f2873587f75e89b5c34a263c3b33e7e254be4
---

# R2 M3 implementation review: mutation adequacy and concurrency/ordering lens

## Provenance

- **Reviewer.** An independent non-author agent session (Law 9). I wrote none of the code or tests under review.
- **What was reviewed.** HEAD 2d67fc0 (`2d67fc03c1900d9df7de57a841401d9a9d74c6fe`), covering the M3 range 260b563..2d67fc0 under `kernel/`.
- **HEAD moved during the review.** At the end, main was at `22727923558e5d0639dd92eafe7be194d65be1bb`. Four commits landed while I worked: ed1744d, 9464e06, 6f43108 and 2272792, which hold the other lens's verdict and the author's fixes for it.
  - Every result below was measured on 2d67fc0, in private detached worktrees.
  - From reading the later diff only, 9464e06 appears to add `TestStoreLedgerViewFailsClosedOnUndecodablePayloads` (relevant to F1) and to rewrite `port_home_test.go` as an AST check that includes a "multi-line portless" case (relevant to F4).
  - I did not re-sweep or verify those later commits.
- **Read.**
  - Contract: `docs/plans/R2-sdd-m3-git-binding-spec-draft{5,7,8}.md` (§4.3-§4.6, §6, §7).
  - Mutation tooling: `scripts/mutants.sh`, `tools/mutants/main.go`, `scripts/mutant-apply.sh`, `docs/allowed-survivors.txt`.
  - Code: the M3 diff of `connector.go`, `reviews.go`, `cli.go` and `projections.go`; the store's `ReadEvents`, `BeginSync`/`Append` and lock code.
  - Tests: all M3 tests and `port_home_test.go`.
- **Executed.**
  - `scripts/mutants.sh` full sweeps on `internal/connector/git` and `internal/connector/reviews`, in worktree `<scratch>/wt`.
  - Hand-applied mutants for git, reviews, projections and cli through `scripts/mutant-apply.sh` in a second worktree, `<scratch>/wt2`, using a harness script at `<scratch>/hand.sh`. After each mutant the harness restored the tree with `git checkout -- kernel` and stopped if `git status` was not clean. It never stopped.
  - One full `go test ./... -count=1` on the clean worktree.
  - A guard-evasion probe of `port_home_test.go`, using a temporary test file that was then deleted.
- **Environment.** `$HOME/go/bin` first on PATH (go1.27.4); `GOCACHE` and `TMPDIR` under `<scratch>`.
- **Writes outside scratch.** `tools/mutants` hard-codes `GOCACHE=/tmp/proofbound-mutant-cache` for each mutant run, and I did not override it.
- **Incident.** I stopped my first cli harness run with `pkill`. That left an embedded postgres from `TestM3ClaimBindsOnlyALedgeredRevision` (my scratch DataDir) listening on port 55432 for about 24 minutes, and it may have blocked the other reviewer's port-55432 tests in that window. I killed it once I found it.
  - Mutant C19 ran during that window and was first recorded as "killed" through contention. I struck that result, hardened the harness so that contention reports CONTENDED instead of a kill, and re-ran C19 cleanly.
- **Final state.**
  - Both worktrees were removed with `git worktree remove --force`, and `git worktree list` shows only the main checkout.
  - `git status --short` in the main checkout shows only `?? docs/plans/P5-intent-provenance-plan.md` and `?? kernel/cmd/p6dbprobe/`.
  - No tracked file was rewritten, and `kernel/go.mod` is clean.

## Calibration (stated before any count)

**Repository sweep (`tools/mutants`).** Each sweep reported `calibration neutral=survived invalid=invalid lethal=killed` before any mutant ran:
- neutral: a comment appended, which must survive;
- invalid: an undefined symbol, which must be classed invalid;
- lethal: an injected `t.Fatal` test, which must be killed.

Both sweeps passed calibration.

The tool's operator set is only `==`/`!=`/`&&`/`||` swaps, and a timeout counts as a kill. No sweep mutant reached the 1m timeout: the packages run in seconds, and the cli/sdd harness was not running during the sweeps.

**Hand harness, per stage.** All controls behaved as required:

| Stage | Neutral (`_ = 0` statement) | Invalid (undefined call) | Lethal (`t.Fatal` injected) | Positive (behaviour-changing mutant) |
|---|---|---|---|---|
| git | survived | invalid | killed | killed (`len(skips) > 99`) |
| reviews | survived (see note) | not run | killed | many R-mutants killed |
| projections | survived | not run | not run | P1-P3 killed |
| cli A (TestSDD*) | survived through A, A2 and B | invalid | killed (TestSDDResolverChecksAndCauses) | killed (`Provider: "records"`) |
| cli B (sdd TestM3) | — | — | killed (TestM3InterleavingB, so stage B can kill on its own) | — |

- **Reviews neutral.** My first neutral control was syntactically broken: it was inserted into a one-line function body and was correctly classed invalid. I re-ran a valid one, which survived.
- **Invalid mutants.** These are listed separately, never counted as kills, and every one was re-expressed as a compile-valid mutant and re-run. Their causes:
  - a perl `\x00` that became a NUL byte;
  - a variable left unused;
  - an import left unused.

## Mutation results

### `internal/connector/git` (connector.go), repository sweep

49 candidates: 47 killed, 0 invalid, 2 survived.

- `connector.go:215#39` (`&&` → `||`, second `&&` of `isSDDClaim`). **Real test gap.** `sdx:foo` and `xxd:foo` become sdd claims, so they are reported as "malformed" skips instead of failing the sync as an unknown provider. Part of F3.
- `connector.go:215#40` (third `&&`). **Real test gap.** Any value with `:` at byte 3, for example `abc:CI-…`, or one starting `sdd` without a colon there, for example `sdd-x:…`, becomes a malformed skip instead of a sync failure. Part of F3.

### `internal/connector/git`, hand mutants

13 valid: 8 killed, 5 survived. Also 1 no-op (G10, a perl `$\` artifact, not counted) and 1 invalid (G13, an unused variable, not counted).

- **G1 survived** (`connector.go:212`, `len(raw) < 4` → `< 3`). **Real gap.** No test uses the 3-byte value `sdd`. Under the mutant, `Intent: sdd` panics on `raw[3]`; under the original it fails the sync as a records id. Part of F3.
- **G3 survived** (`connector.go:219`, `b <= 'Z'` → `< 'Z'`). **Equivalent:** only `s` and `d` are ever compared after folding.
- **G7 and G8 survived** (`connector.go:24`, `{1,99}` → `{1,98}` and `{1,100}`). **Real gap:** the D1 grammar's upper bound (99 and 100 characters) is untested. Part of F3.
- **G11 survived** (`continue` after a skip removed). **Equivalent:** the zero ref that leaks through is discarded by the `len(skips) > 0` clear, and the error path is unchanged.
- Killed: G2, G4, G5, G6, G9, G12, G14, and the positive control.

### `internal/connector/reviews`, repository sweep

134 candidates (reviews.go and v2.go): 131 killed, 0 invalid, 3 survived.

- `reviews.go:127#80` and `#81`: the first two `||` of the pre-M3 `c.reader == nil || c.ids == nil` guard. **Pre-existing and unreachable:** the fields are unexported and `New` validates them. The M3-added `|| c.ledger == nil` (#82) was killed.
- `reviews.go:252#85` (`err != nil || !candidate` → `&&` in `duplicateIdentities`). **Equivalent:** when there is no error and the artifact is not a candidate, its first line is neither `---` nor `schema:`, so `artifactSchema` fails and the loop continues anyway.

### `internal/connector/reviews`, hand mutants

19 valid: 17 killed, 2 survived.

- **R17 survived** (the `wireCandidate` check removed from `duplicateIdentities`). **Near-equivalent.** It differs only when a sibling artifact fails `wireCandidate` (invalid UTF-8 or a bad path) yet parses as v2 with a shared id. That artifact fails the sync as malformed either way; only which partner is refused versus appended before the stop changes.
- **R18 survived** (an error from `isRevision`/`Ledgered` ignored). **Real interface-level gap, unreachable in production.**
  - `scriptedLedger.err` fails every method, so the `CommitClaimed` error path fires first (R19 is killed).
  - With `storeLedgerView`, a `Ledgered` error cannot follow a successful `CommitClaimed`/`RequirementLedgered`, because the sets are already loaded.
  - Observation only.
- Killed: R1-R16 and R19, covering:
  - the duplicate-count thresholds and both schemas' duplicate keys;
  - the held/refused swap;
  - the cursor excluding held paths;
  - the provider trim;
  - both requirement loops;
  - final and join-on-malformed errors;
  - the refused-error text;
  - the isRevision ledger use.

### `internal/projections` (one provider clause)

3 valid: 3 killed (P1 rejects sdd, P2 removes the sdd clause, P3 changes `&&` to `||`). No survivors.

### `internal/cli` (cli.go M3 functions), hand mutants

The stages ran in order, and a later stage ran only if the mutant survived the earlier one:
- A: `TestSDD*` (no store);
- A2: `TestStoreLedgerView*|TestSyncReviews*` (ephemeral ports);
- B: sdd package `TestM3*` (default port).

56 valid mutants: 44 killed, 12 survived. There were also 8 invalid, 1 no-op and 1 contended attempt; each was re-expressed and re-run, and those re-runs are inside the 56.

Branches covered: the resolver construction fields; the `case "sdd"` arm; lazy provider and L construction (never and always); each cause mapping; each anti-spoof clause; the segment check, its date offset and equality; L's filter (kind and source) and its keys; claim-skip printing in `sync git` and `sync all`; refusal propagation and print-before-error in `sync reviews` and `sync all`; carrying held/refused through `syncReviewsOnStore`; every `readLedgerView` key; `load` caching; every `storeLedgerView` method forced true or false; `printHolds` words and groups; and the view wiring.

Survivors:
- **C21** (L filter without `Kind`). **Equivalent:** intent.sdd requirement payloads have no `intent_id`, so their keys (`"\x00"+digest`) never match a CI id.
- **C22** (L decode error → `return nil`). **Real gap** in fail-closed behaviour. Part of F1.
- **C26** (`sddDatedRE` unanchored). **Real but defence-in-depth gap.** With `name="2026-01-01-x"` and segment `abcd-ef-gh-2026-01-01-x`, the mutant accepts where the original refuses. The spec puts the guarantee on `Resolve`'s exact-name semantics (SDD-INV-59). Observation.
- **C28** (claim-skip value printed without `printableUnit`). **Real gap.** Part of F2.
- **C29** (write error while printing a claim-skip ignored). **Real, minor gap**, because no failing-writer test covers it. Part of F2.
- **C39** (`probe.Schema == ""` check removed). **Real gap.** Confirmed to survive the full cli and full sdd packages. Part of F1.
- **C42** (git view case for any kind). **Equivalent today:** the git source emits only `commit.recorded` (`core.go:24`).
- **C43** (git payload decode error ignored). **Real gap.** Part of F1.
- **C45** (requirement view case for any source). **Equivalent today:** only `intent.*` sources emit `KindRequirement`.
- **C46** (requirement decode error ignored). **Real gap.** Part of F1.
- **C49** (`load` error ignored, so partial sets are used). **Real gap.** Confirmed to survive the full cli and full sdd packages. Part of F1.
- **C54** (held/refused paths printed without `printableUnit`). **Real gap.** Confirmed to survive the full cli and full sdd packages. Part of F2.

## Concurrency and ordering answers

1. **Can storeLedgerView's snapshot or L go stale in an unsafe direction within one sync?**
   - **Within one process, no.** In `sync all`, L is read lazily on the first sdd claim, after the intent step. The view is read lazily on the first v2/review artifact, after the intent, git, checks and sessions steps, so it sees everything appended earlier in the same run.
   - `sync git` appends only git events, so L cannot miss anything it needs (spec §4.3 holds).
   - The reviews step's own appends are not added to the identity set (draft 8 dropped that rule). Same-identity pairs within one sync are refused up front as `duplicate-identity`, so a stale identity set cannot turn an in-sync revision into a hold.
   - **Across processes.** The store takes an exclusive, non-blocking flock on `<DataDir>.lock` (INV-39) in `Open`, whatever `DATABASE_URL` says. Two CLIs on one root therefore cannot interleave: the second gets `ErrLocked`.
   - The residual case is two different roots sharing one external `DATABASE_URL`, which uses two different lock files. In that case:
     - `ReadEvents` is paged and is not one database snapshot;
     - because the ledger is append-only, the view and L can only under-report;
     - so no artifact is ever appended ahead of its prerequisites, and M3-INV-11/12's safety holds;
     - but a later revision can be classed **held** (exit 0) instead of **refused** (exit 1) for one run, if the other process ledgers its first revision between this view's read and its check. This is F5.
2. **Can a hold or refusal decision depend on map iteration order or path order?**
   - No dependency on map iteration: `duplicateIdentities` reduces its count map to a set, and every view set is order-free.
   - Artifacts are path-sorted and duplicates are refused as a group (killed by R9 and R10, and tested in both path orders).
   - `holdCause` returns `claim-not-ledgered` before any requirement cause, deterministically.
   - Claim skips follow commit order and the `sortedUnique` trailer order.
   - The one path-order effect is pre-existing: a malformed artifact stops the sync, so artifacts sorted after it are not classified. The sync fails in every order.
3. **Is the embedded-port contention guard sound at 2d67fc0?** **No.** I demonstrated this directly.
   - A temporary untagged test in `internal/projections` with a multi-line portless `store.Config{\n Root: …,\n}` passed `TestDefaultPortStoreTestsLiveInOnePackage`, because the regex `store\.Config\{[^}\n]*\}` cannot match across a newline.
   - The same literal on one line tripped the guard (positive control).
   - The guard's own comment admits two other blind spots: `os.Unsetenv("DATABASE_URL")`, and a bare `Run(`/`runCommand(…, "", …)` in package cli without an `openStore` override.
   - No current offender exists. The M3 cli tests override `openStore` with ephemeral ports.
   - This is F4. HEAD 2272792 appears to have replaced the guard with an AST check that has a multi-line case; I did not verify it.
4. **Can verify's two passes classify differently?** No.
   - Pass 2's intent and git syncs append nothing, which `requireNoSecondSyncAppends` enforces.
   - Each `syncGit` builds a fresh resolver and re-reads the same L.
   - Pass 1's review appends only add identities whose prerequisites are present, so pass 2 re-appends them as `existing`.
   - Held artifacts are not ledgered, so they are held again.
   - Refusals and duplicates fail pass 1.
5. **Full suite.** `go test ./... -count=1` at 2d67fc0 on a clean worktree: 19 packages ok, exit 0, 4m50s.

## Findings

All five findings are LOW. None leaves a path where a bad artifact is ledgered or a check passes on an enforcement path that syncs first.

- **F1 (LOW): the fail-closed "View errors" clause and its L analogue are not pinned by any test at 2d67fc0.** Draft 7 §4.6 requires that a git payload that does not decode, or a reviews payload without a `schema`, fail the sync.
  - Evidence: C39 (schema check removed), C43 (git decode error ignored), C46 (requirement decode error ignored), C49 (`load` error ignored) and C22 (L decode error ignored) all survive. C39 and C49 also survive the full cli and full sdd packages.
  - Failure scenario: a regression like C39 lets a schema-less `reviews` event be keyed `"\x00"+id`. A later revision of that verdict then misses the identity set and is reported `held` (exit 0) instead of `refused`, so `sync all` goes green while the old outcome stands.
  - Mitigation: projection decoding would probably fail on the same corrupt ledger elsewhere.
  - HEAD 2272792 appears to add a test for this (unverified).
- **F2 (LOW): the `printableUnit` escaping of claim-skip values and of held/refused paths is unpinned** (C28, C54 confirmed on the full packages; C29 is the related print-error path).
  - Draft 8 explicitly added "refused paths go through printableUnit".
  - Failure scenario: git allows a newline in a file name, and `validPath` checks only prefix and suffix. Under a C54-style regression, a committed file named `docs/verification/verdicts/a.md\nheld claim-not-ledgered docs/verification/verdicts/x.md` could print a forged `held` line into `sync reviews` / `sync all` output.
  - The exit status is unaffected.
- **F3 (LOW): the edges of the sdd claim detector and of the D1 grammar are untested.**
  - Evidence: sweep survivors `connector.go:215#39` and `#40`, plus G1, G7 and G8.
  - Failure scenario: a mutant like #40 turns any trailer with `:` at byte 3 (`abc:CI-x-123456`) into a quiet `claim-skip malformed`. Today that trailer fails the sync as an unknown provider. A loud refusal becomes a reported skip.
  - G1 turns `Intent: sdd` into a panic.
  - The 99/100-character slug bound can move unnoticed.
- **F4 (LOW): the default-port guard is syntactic and was demonstrably evadable at 2d67fc0** (multi-line `store.Config` literal; probe above).
  - Failure scenario: an untagged test outside the sdd package that opens a portless `store.Config` written across several lines passes the guard. It then collides on port 55432 with the sdd package's tests under parallel `go test ./...`, bringing back the flaky `make check` that the LESSON compiled against.
  - No offender exists today.
- **F5 (LOW): held-versus-refused classification is not linearizable when processes share a database across different roots.**
  - Evidence: the lock is per DataDir (`store/lock.go`), and `ReadEvents` pages with `seq > cursor`.
  - Failure scenario: two checkouts sync against one `DATABASE_URL`. Process A appends verdict V rev1. Process B, whose view was read just before that, sees V rev2 with a missing prerequisite and reports `held` with exit 0 instead of `refused` with exit 1. B's next sync refuses it.
  - Nothing is appended ahead of its prerequisites, and claim-skips from a stale L recover by revision on the next `sync git`.

## Other observations

- The sweep tool's four-operator set (no relational-boundary, return-value or statement-deletion operators) would have missed most of the cli and reviews gaps above. Every survivor except the two `isSDDClaim` ones came from hand mutants.
- The tool also counts a timeout as a kill. That is safe here because nothing timed out, but it is an inflation risk under machine contention.
- The tool hard-codes `GOCACHE=/tmp/proofbound-mutant-cache` over any caller setting.
- C26 (unanchored date regex) and R18 (`Ledgered`-only error) are real but defence-in-depth or unreachable; see above.
- The resolver's sdd arm does not compare `ref.RecordID` with the claimed name; it relies on the cli segment check plus `Resolve` semantics. This is by spec (§4.2) and is noted for the spec-truth lens.

## Verdict

**ACCEPTABLE.** The M3 code is well defended where it matters most. Of the mutants that change behaviour, every one that would let an artifact be ledgered ahead of its prerequisites, bind an unledgered or spoofed revision, or swallow a refusal's exit status was killed. That includes the lazy-construction counts, the duplicate-identity rule in both orders, and print-before-error in both commands.

What survives clusters in three unpinned spec clauses:
- fail-closed view errors (F1);
- output escaping (F2);
- the edges of the claim detector and grammar (F3).

There are also two infrastructure or concurrency limits (F4, F5). All five are LOW, and none produces a silent pass on a path that syncs before it enforces.
