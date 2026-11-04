---
schema: vera.verdict.v1
verdict_id: p7-b1-unit3-spec-review-round1-20261104
status: NEEDS_CHANGES
reviewed_commit: d468d9cfeb9921e2cb74be026f6e6e0263af0f27
findings:
  - finding_id: p7-b1-unit3-spec-review-round1-F1
    severity: HIGH
    summary: LOCKED is structurally independent of the runtime; TWIN-INV-22 over-claims
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round1-F2
    severity: HIGH
    summary: TWIN-INV-18 launched-binary witness re-labels the exact code the deferral calls insufficient; permitted mechanisms do not witness the process
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round1-F3
    severity: MED
    summary: Running process maps un-inventoried host shared libraries (zlib/libc); binary witness cannot cover full behavior
    verdict: PLAUSIBLE
  - finding_id: p7-b1-unit3-spec-review-round1-F4
    severity: MED
    summary: TWIN-INV-19 zero-fetch witness is a harness-wiring assertion, not an observation that no network fetch occurred
    verdict: PLAUSIBLE
  - finding_id: p7-b1-unit3-spec-review-round1-F5
    severity: MED
    summary: §8.1 "RESOLVED" proven on full install tree, but unit ships the bin/+share/ subset — shipping artifact unproven
    verdict: CONFIRMED
  - finding_id: p7-b1-unit3-spec-review-round1-F6
    severity: MED
    summary: §8.2 drop-vs-dereference undecidable from spec; dereference at prep time is unconstrained attack surface outside the staging guard
    verdict: PLAUSIBLE
  - finding_id: p7-b1-unit3-spec-review-round1-F7
    severity: MED
    summary: Law 1 (@generated) status of committed derived artifacts unresolved; cannot be regenerated in make check
    verdict: PLAUSIBLE
  - finding_id: p7-b1-unit3-spec-review-round1-F8
    severity: MED
    summary: Law 2 drift seed — toolchain identity copied into the new manifest instead of citing the unit-2 evidence home
    verdict: PLAUSIBLE
  - finding_id: p7-b1-unit3-spec-review-round1-F9
    severity: MED
    summary: TWIN-INV-20/-21 restate existing invariants without a discriminating positive control
    verdict: PLAUSIBLE
  - finding_id: p7-b1-unit3-spec-review-round1-F10
    severity: MED
    summary: t.Skip-on-absence plus DEF-KERNELCHECK-FLAKE-001 port contention create a silent-green acceptance hole
    verdict: PLAUSIBLE
  - finding_id: p7-b1-unit3-spec-review-round1-F11
    severity: LOW
    summary: DeriveCaseStatus PASS reachable from synthetic fixtures + injected stubs with no real postgres; needs a negative control
    verdict: PLAUSIBLE
artifact_path: docs/verification/verdicts/p7-b1-unit3-spec-review-round1-20261104.md
artifact_sha: d2b9b4c385348bfd4052dd70a0bcd830473de3dfed50e64cee2daba8f3ba5878
---

# P7 B1 unit-3 runtime-fixture spec — non-author adversarial review

## Provenance

- **Under review:** the DRAFT design spec
  `docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md` (authored by prior-session commits
  `0e9c93e` + `0ca1ec0`; no code under review — spec-first per Build Law 6). Focus per the request:
  the §8 open decisions (symlink dereference, launched-binary witness, zero-fetch witness) and the
  proposed TWIN-INV-18..22.
- **HEAD at dispatch and at receipt:** `d468d9cfeb9921e2cb74be026f6e6e0263af0f27` — unchanged
  between dispatch and receipt of this verdict (verification-only; the tree was not modified during
  the review).
- **Driving mechanism:** `/proofbound-review` (CLAUDE.md pre-merge adversarial review). Tier:
  connector/peripheral — a single adversarial review of a design document (no mutation evidence, as
  there is no code yet).
- **Reviewer identity:** `dispatched-non-author-reviewer-20261104` — an independent reviewer
  dispatched via the Agent tool; it did NOT author the spec. Adjudication by this session, which
  also did NOT author the spec (authored by prior sessions `0e9c93e`/`0ca1ec0`). Law 9 is satisfied:
  this is a genuine non-author verdict, returning NEEDS_CHANGES rather than acceptance.
- **Adjudicator independent verification:** the two HIGH findings were confirmed against the code
  before acceptance into this verdict (the adjudicator discards findings lacking a reproducible
  proof):
  - **F1 CONFIRMED** — `kernel/internal/twin/replay.go:195-199`: the `vera.replay.v1` `Proof`
    carries only `Schema`, `ThroughSeq`, `SourceDigest`, `CandidateDigest`,
    `BaselineSnapshotDigest`, `CandidateSnapshotDigest`. No postgres-binary digest, binaries dir,
    launched-process, or no-fetch fact is bound. `EvaluateLock→LOCKED` over these bytes is therefore
    pure digest arithmetic, producible with no PostgreSQL running; TWIN-INV-22's "produced by this
    real run" over-claims what LOCKED witnesses.
  - **F2 CONFIRMED** — `kernel/internal/twin/verified_archive.go:267-273` already re-hashes on-disk
    `binariesDir/bin/postgres` (error string "launched postgres executable digest mismatch"), while
    `kernel/internal/twin/SPEC.md:84-85` (TWIN-INV-14) states that seam "does not establish that
    runtime fact." §8.3's first listed mechanism is exactly that pre-existing insufficient code.
  - **F5 CONFIRMED** — the spec itself concedes (§8.1) the subset "must still be confirmed" while
    labeling the risk RESOLVED; the probe used the full install tree, not the shipping
    `bin/`+`share/` subset.
  - F3, F4, F6–F11 are recorded at the reviewer's severities as PLAUSIBLE design/spec-truth
    findings (not reproduced line-by-line, but internally consistent with the cited invariants and
    the authored/observed contract). None was discarded.
- **Cleanroom redaction:** none required; the report below contains no absolute paths or usernames.

## Findings asymmetry notice

Per `VD-verification-asymmetry-2dyjnd`, a verified finding transfers zero trust to any remedy it
suggests. The remedy sketches in the reviewer's "Verdict recommendation" are the reviewer's
proposals, not accepted fixes. The spec author decides the remedies; a non-author must re-verify any
revised spec. Findings below are stated as broken invariants / over-claims, not as instructions.

## Verdict (verbatim reviewer report)

The following is the dispatched reviewer's report, reproduced verbatim.

---

### F1 — LOCKED is structurally independent of the runtime. (HIGH) — over-claim, defeats the unit's purpose
The whole point of unit 3 is to produce a *runtime* fact and bind it so `EvaluateLock→LOCKED`. But I confirmed in `kernel/internal/twin/replay.go:196-198` that the `vera.replay.v1` proof binds only `SourceDigest`, `CandidateDigest`, `BaselineSnapshotDigest`, `CandidateSnapshotDigest` — all pure functions of event/snapshot bytes. The proof does **not** reference `PostgresBinarySHA256`, `BinariesDir`, the launched process, or any no-fetch fact. TWIN-INV-9 further guarantees the proof is deterministic from those inputs. Consequence: the exact proof bytes that make `EvaluateLock→LOCKED` can be produced by the fast `Replay` path (injected projector), or computed from digests, with **no PostgreSQL running at all**. TWIN-INV-22 ("committed proof bytes are produced by this real run and make EvaluateLock==LOCKED") therefore over-claims — LOCKED witnesses digest arithmetic, not a real from-source run. A proving test that commits synthetic-fixture proof bytes and asserts LOCKED passes identically whether or not any postgres ever launched; the key variable (did the real binary run?) is held constant and can never flip the result. The runtime evidence lives only in the case-manifest *outcomes* (self-declared states + evidence pointers), not in the proof the lock gates on. This is the load-bearing gap: the chain "LOCKED + PASS ⟹ a real from-source postgres ran" is not sound as specified.

### F2 — TWIN-INV-18 "launched-binary witness" does not witness the launched process; its weakest listed mechanism is ALREADY the code the deferral says is insufficient. (HIGH)
`verified_archive.go:267-273` already re-hashes on-disk `binariesDir/bin/postgres` after staging and even labels the error "launched postgres executable digest mismatch" — yet the twin SPEC explicitly states "the current seam does not observe the daemonized process's launched executable" and TWIN-INV-14 says "the current seam does not establish that runtime fact." Of §8.3's three options:
- **re-hash staged `bin/postgres`** (listed first): this is the existing code. It proves the file at that path is unchanged; it does **not** prove the running process mapped those bytes (process image is already in memory; file could be restored; pg_ctl could exec a different image). Adopting this as TWIN-INV-18 closes DEF-P7-001's runtime deferral by *renaming the exact code the deferral declares insufficient*.
- **`version()` string match**: trivially forgeable — any server (or a proxy) can print `PostgreSQL 16.4 … gcc 11.4.0`. Witnesses zero bytes.
- **resolve launched PID's executable** (e.g. hash `/proc/<pid>/exe`): the only option that actually ties the running process to an inode, and even it needs the "(deleted)" check. The spec leaves the choice open ("pick the strongest practical") and `and/or`s them, so TWIN-INV-18 can be "satisfied" by the two non-witnessing options. As stated ("the launched postgres executable's bytes equal the declared digest") the invariant claims more than its permitted mechanisms deliver. No negative control is specified — there is no required test where a swapped/other binary makes the witness FAIL. Accepting this would also propagate the over-claim into the accepted twin SPEC (amending TWIN-INV-14 to say the runtime fact is established when it isn't).

### F3 — The running process maps un-inventoried host shared libraries. (MED/HIGH)
Unit-2 configured `--without-icu --without-readline` with zlib default-on; the from-source `postgres` dynamically links host `libc`, `libz`, etc. (zlib default-on links the *system* `libz.so`, not a bundled copy). The archive pins only `bin/postgres` bytes; the actual process = declared bytes **plus arbitrary host `.so` files that are neither inventoried nor witnessed**. So even a perfect launched-PID witness proves the *text segment of `bin/postgres`* only, not the full behavior of the process. "Witness that the launched process ran the declared bytes" is unachievable as long as dynamic deps are ambient. The spec never acknowledges this.

### F4 — TWIN-INV-19 "zero-fetch witness" is hermetic theater as specified. (MED/HIGH)
"A test seam asserts the download path is never invoked" witnesses a property of the *harness wiring* (the Go fetch-strategy constructor wasn't called), not an observation that no network fetch occurred. It covers only `fergusstrange/embedded-postgres`'s own download path (skipped when `BinariesPath/bin/pg_ctl` exists, per §8.1). It does not cover DNS, a cache-dir fallback, or any network I/O by the `initdb`/`postgres` subprocesses themselves. The seam is true-by-construction whenever `BinariesPath` is set, so the test passes even on a host where the binary is absent (it would just fail later) — the witness does not discriminate the thing it claims. `network_policy=loopback-only` in the command logs is a *declaration*, not an enforced sandbox. A genuine no-fetch witness needs network-namespace isolation (`unshare -n`) or an egress monitor, not a code seam. Compounding: the witness also *trusts the very library it audits* — if embedded-postgres fetched anyway, the seam wouldn't know.

### F5 — §8.1 "RESOLVED" is proven on the full install tree, but the spec ships the subset. (MED) — internal inconsistency / over-claim
§8.1 and commit `0ca1ec0` say the highest risk is RESOLVED — but the probe "pointed at the full install tree," while unit 3 ships a symlink-free `bin/`+`share/` **subset** with libs dropped/dereferenced. The spec itself concedes "the staged `.txz` subset … must still be confirmed" and "confirm when packaging the subset." So the shipping artifact's initdb/start behavior is **not** resolved; only a different artifact is. Labeling it RESOLVED (and the commit subject "highest risk RESOLVED") overstates what was proven.

### F6 — §8.2 symlink "drop vs dereference" is undecidable from the spec, and dereference is unconstrained attack surface. (MED)
(a) The rule "drop if not in subset, dereference if retained" is circular: which symlinks are *retained* is itself the open `lib/` subset question (§8.1, "settle empirically"). So the decision is not settled in the spec. (b) All 8 symlinks are client libs the server doesn't link (§3.2/§8.1), so the correct answer is simply **drop all** — dereference is unjustified. (c) Worse, dereference happens at *prep time, outside the staging guard*. TWIN-INV-13's safe-path/no-link guard inspects archive *entry paths*, not what a symlink *target* pointed at during prep. A symlink with an absolute or `../` target could be dereferenced into a regular file carrying arbitrary host bytes under an innocuous in-tree path; the staging guard would see only a safe-path regular file and pass. The honest digest pins whatever got packaged, but a reviewer reading "bin/+share/, symlink-free" would not catch smuggled host bytes. Keeping dereference as an option adds this surface for no runtime need.

### F7 — Law 1 (@generated) not addressed for committed derived artifacts, and they cannot be regenerated in `make check`. (MED)
§6 commits the prepared-archive manifest, proof bytes, behavior-lock, case-manifest, and witness records — all *generated*. Law 1 requires every generated artifact to carry an `@generated` marker + regeneration command, and `make check` fails if regeneration diffs. But these are generated by a run that requires an **offline archive** not present in `make check` (and the fixture `t.Skip`s without it). The spec is silent on whether these are `@generated` (regenerable — impossible in CI) or hand-committed one-time evidence (which Law 1 would classify as "dead state" needing an explicit exemption). This dichotomy must be resolved before implementation, or `make check` either breaks or silently exempts load-bearing artifacts.

### F8 — Law 2: toolchain identity copied, not linked. (MED)
§3 says the prepared-archive manifest includes "toolchain identity from the unit-2 build evidence." The home for that datum is the unit-2 build-evidence record; copying its fields into the new manifest is a drift seed. It should cite the unit-2 evidence revision/path, not duplicate the values. (Digest duplication across proof/lock/case-manifest is pre-existing cross-binding and acceptable; this toolchain copy is new.)

### F9 — TWIN-INV-20 / -21 restate existing invariants without a discriminating control. (MED)
- **TWIN-INV-20 (ledger preservation):** byte-identity before/after is already structurally guaranteed by TWIN-INV-1 + TWIN-INV-8 (replay never writes source; projects in a temp store). A before/after-hash test can therefore *never fail* — the mutating variable is held constant by construction. The spec honestly calls it "extends TWIN-INV-1 with a recorded witness," but as a new *invariant* it adds no force without a positive control (an operation that *does* mutate must make the witness FAIL, proving it can detect mutation).
- **TWIN-INV-21 (cleanup):** TWIN-INV-17 already says "removed on success and every failure path." INV-21 = INV-17 + "witnessed," but the spec doesn't require the witness to *stat the path and assert absence* — a witness that merely records "we called RemoveAll" passes while cleanup silently failed. Restatement risk unless the witness observes actual absence on each failure path.

### F10 — `t.Skip`-on-absence + existing embedded-postgres flake create a silent-green acceptance hole. (MED)
§4.1's skip-when-env-unset is correct for the *committed* case (stays UNVERIFIED, honors CASE-INV-5 "no PASS by absence"). But the spec does not require the **non-author verifier to run WITH the archive present and witness LOCKED+PASS live**. A verifier lacking the archive sees a green SKIP and could grant ACCEPTABLE over a test that never executed — the acceptance itself becomes pass-by-absence. Separately, `DEF-KERNELCHECK-FLAKE-001` already records embedded-postgres **port contention (55461-55463)** making `make check` non-deterministic; adding a second real-server fixture to the same suite worsens this, and §4.1 addresses only the 107-byte socket-path limit, not port contention/parallelism. A flaky gate around a real-run acceptance is itself a Law 7 determinism defect.

### F11 — Reaching PASS can be done with stub injections. (LOW/MED, partly pre-existing)
`intent_binding`→PASS requires the lock's declared P5 requirement to resolve, but `RequirementResolver` is injected; a test stub that resolves anything makes it pass. Combined with F1 (synthesizable proof bytes) and F2 (non-witnessing launched-binary option), a full `DeriveCaseStatus→PASS` is reachable from synthetic fixtures + stubs with no real postgres. This is the Law 9 corollary writ large across the unit: the proving tests must hold real-run identity as the *varied* key and demonstrate PASS collapses to UNVERIFIED/FAIL when the real binary is absent or swapped (a negative control), or the acceptance proves nothing.

### Law 8 note (not a blocker)
`fergusstrange/embedded-postgres v1.31.0` is already in `go.mod` (indirect, via `internal/store`) and covered by `VD-stack-go` / `VD-p7-postgres-source-route`, so adding it is not a new-dependency gap. But the spec promotes it to a **direct, load-bearing kernel-acceptance launcher** and makes the no-fetch witness depend on auditing that same library (F4). That architectural promotion — using the provenance-*failed* tool (the root of DEF-P7-001) as the launcher of the provenance-*good* binary — deserves a one-line decision note, even if defensible (it's a process launcher, not the trusted binary).

### Consistency checks that PASSED
- Invariant numbering appends 18-22 correctly (TWIN-INV-5 gap is a pre-existing tombstone, not introduced here) — Law "append-only" respected.
- Case-manifest nine-outcome mapping (§5) is structurally consistent with CASE-INV-3/4/5 and the worst-wins fold.
- Spec-first (Law 6) is honored: DRAFT, no code, no DEF-P7-001 movement, change-intent flagged (§8.5).
- Honest-digest caveat (§3) and §7 non-goals correctly decline cross-host reproducibility — no over-claim there.

### Verdict recommendation
**NOT sound enough to proceed to implementation as written — needs changes first.** The two HIGH findings are structural, not cosmetic: (F1) the proof the lock evaluates is independent of the runtime, so LOCKED does not witness a real run; and (F2) the launched-binary witness as specified re-labels the exact code the deferral declares insufficient and its permitted mechanisms don't witness the process. Before implementation the spec should: bind some runtime fact into the gated artifact (or explicitly scope the claim down so DEF-P7-001 closure does not rest on LOCKED alone); fix TWIN-INV-18 to a single mechanism that provably observes the launched process (PID-exe resolution) with a required negative control; replace the TWIN-INV-19 code-seam with real network isolation or restate the invariant honestly as a declaration; resolve the subset-vs-full-tree gap (F5); settle drop-vs-dereference as drop-all (F6); and decide the Law 1 status of the committed derived artifacts (F7). F9/F10/F11 require negative/positive controls so the proving tests can actually fail. The §8 "RESOLVED" label on the highest risk should be downgraded to "resolved for the full tree; subset unproven."

---

## Adjudication summary

- **Status: NEEDS_CHANGES.** Two independently-confirmed HIGH structural findings (F1, F2) mean the
  spec's central claim — that unit 3 makes `LOCKED + PASS` witness a real from-source PostgreSQL run
  — is not sound as written. F5 (also confirmed) means the highest-risk "RESOLVED" label overstates
  what the probe proved. The remaining MED/LOW findings are design gaps and missing negative
  controls that must be settled before implementation.
- **This is not an acceptance.** The spec author (a future session, under the owner's standing
  authorization) addresses the findings; a non-author must re-verify the revised spec (round 2)
  before implementation proceeds.
- **Bearing on the owner's §8 decisions:** the review strengthens, not weakens, the case for taking
  these to the owner — decisions §8.3 and §8.4 are where the HIGH/MED findings concentrate, and the
  review supplies concrete direction (PID-exe witness with a negative control; real network
  isolation vs. an honest down-scoped declaration; drop-all for symlinks).
