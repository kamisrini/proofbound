---
schema: vera.verdict.v1
verdict_id: r2-m0-review-round2-20261116
status: NEEDS_WORK
reviewed_commit: 0dc9e087454a58748b82b72d531797ff1966a612
findings:
  - finding_id: r2-m0-review-round2-20261116-F1
    severity: HIGH
  - finding_id: r2-m0-review-round2-20261116-F2
    severity: HIGH
  - finding_id: r2-m0-review-round2-20261116-F3
    severity: HIGH
  - finding_id: r2-m0-review-round2-20261116-F4
    severity: HIGH
  - finding_id: r2-m0-review-round2-20261116-F5
    severity: HIGH
  - finding_id: r2-m0-review-round2-20261116-F6
    severity: HIGH
  - finding_id: r2-m0-review-round2-20261116-F7
    severity: MED
  - finding_id: r2-m0-review-round2-20261116-F8
    severity: MED
  - finding_id: r2-m0-review-round2-20261116-F9
    severity: MED
  - finding_id: r2-m0-review-round2-20261116-F10
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m0-review-round2-20261116.md
artifact_sha: 6ce5fbcf8c21cfb202bebb08ce24389e9aeda8d9d55a9e953e614b2932f08217
---

# Release 2 M0 (SDD Warranty Connector): draft-2 provider SPEC, decisions A2 / B1, and BR revision 0002. Review round 2

## Provenance

- **Under review:** the documentary M0 round-2 deliverables at HEAD `0dc9e08`. They were introduced
  by `f748a6c` (draft 2 and the F10 consistency fixes), `bcfb7ca` (A2 and B1 settled, SDD-INV-21..23)
  and `6a2fee0` (BR revision 0002). Files:
  - `docs/plans/R2-sdd-provider-spec-draft2.md`
  - `docs/intent/records/requirements/BR-sdd-warranty-i9miv7/0002.md`
  - `docs/plans/sdd-warranty-connector-plan.md`
  - `docs/decisions/VD-release-line-r2-sdd-t5j34p.md` (amended)
  - `docs/decisions/VD-sdd-accountable-identity-ypcx7x.md`
  - `docs/verification/r2-m0-mapping-spike.md`
  - the amendment paragraph in `kernel/internal/connector/intent/SPEC.md` §1
- **Follows up:** `docs/verification/verdicts/r2-m0-review-round1-20261116.md` (NEEDS_WORK, F1..F11).
- **Judged against:** CLAUDE.md Build Laws, and the following ground truth:
  - `kernel/internal/connector/intent/SPEC.md` §3-§5, `intent.go` (validation), and
    `records/records.go` (`validateLineages`)
  - `kernel/internal/connector/git/connector.go` (trailer resolution, INV-29)
  - `scripts/intent-delivery-applicability.sh` (read only)
  - `kernel/internal/projections/SPEC.md` and `projections.go` (obligation-verdict.v2,
    requirement-review.v1, the revision-table keys)
  - `kernel/internal/gates/SPEC.md` (intent-delivery-readiness)
  - a shallow clone of `Fission-AI/OpenSpec` (the kit's own repository: 44 main specs, 5 active
    changes, 108 archived changes) and its `src/core/specs-apply.ts`, `src/core/archive.ts`,
    `docs/opsx.md` and `docs/cli.md`
- **HEAD at dispatch:** `0dc9e087454a58748b82b72d531797ff1966a612`. **HEAD at completion:**
  `0dc9e087454a58748b82b72d531797ff1966a612` (unchanged).
- **Driving mechanism:** the `/proofbound-review` procedure, run by a dispatched Agent.
- **Tier:** connector/peripheral. The subject owns record identity and obligation lineage, which the
  procedure names as an escalation trigger. M1 should be reviewed at the foundational tier.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261116-m0-r2`, a **non-author**. It did
  not write any of the work under review, so this verdict counts under Law 9.
- **Probes:** two read-only Python probes over the OpenSpec clone, kept in the session scratch
  directory (`<scratch>/reviewer-m0r2/fold_probe.py` and `tomb_probe.py`):
  - The **fold probe** implements B1's consistency rule: fold the archived deltas in date order, then
    compare the result with the current main specs.
  - The **tombstone probe** asks, for each archived MODIFIED or REMOVED requirement, whether an
    earlier archived delta still carries its predecessor text.
  - The parser is approximate, and its counts are evidence of scale, not exact. Every example quoted
    below was checked by hand (for example with `diff` of the actual blocks).
- **Writes:** no repository file other than this verdict. No cleanroom redactions; the scratch path is
  written as `<scratch>` (1 occurrence).

## Round-1 findings re-verified

- **F1 (delta record status `accepted`): FIXED.** Delta records are always `proposed` (§5.2,
  SDD-INV-17), and `proposed` is in the requirement vocabulary (`intent.go:250`). Change status is a
  single declared `proposed|accepted` value, which the change-intent vocabulary admits
  (`intent.go:265`).
- **F2 (archived REMOVED makes the repo unsyncable; intents re-bind at HEAD): PARTIALLY FIXED.**
  - Fixed: archived changes no longer map, so an archived REMOVED target is no longer resolved
    against HEAD.
  - The same failure class returns through B1's fold rule (new F1), through OpenSpec's early-sync
    workflow (new F7), and through the HEAD-only sync (new F5).
  - Active change intents still bind the main record's digest at HEAD, so an edit to a main spec
    still re-mints them. They now do it under an unchanged `artifact_sha256` (new F6).
- **F3 (path-derived ids fork on archive): FIXED.**
  - Ids derive from native names (§4.1), and SDD-INV-16 tests this across a simulated archive move.
  - Residual: the spike still says "Derive it deterministically from the artifact path" and
    "archived → `accepted`" (spike lines 37 and 71-73; see F9).
- **F4 (one requirement, two homes, no link): PARTIALLY FIXED.**
  - The two homes are now named: the delta record is the proposed home and the main record is the
    accepted home.
  - The delta record is still never superseded, retired, or linked to the main record. It stays
    `proposed` in the ledger forever.
  - Verdicts bound to the delta record never become stale (new F4).
- **F5 (change intent has no defined artifact): PARTIALLY FIXED.**
  - The manifest digest is now defined (§5.1).
  - It does not cover the inputs that determine the change-intent payload (new F6).
- **F6 (reference profile rejects valid changes; REMOVED-only; RENAMED): MOSTLY FIXED.**
  - Fixed: separate `owners` and `sponsors` lists; no delta record for a REMOVED-only or
    RENAMED-only change; RENAMED is mapped.
  - Residual: rename plus modify is mishandled, and a change with no spec deltas is unaddressed
    (new F7).
- **F7 (unordered `by_path`): FIXED.** `by_path` is removed, and change status is one declared
  value.
- **F8 (obligations vanish or change meaning, against intent SPEC §5): NOT FIXED IN EFFECT.** B1 is
  the owner's chosen answer, but as specified it cannot be implemented:
  - The fold fails closed on the kit's own repository (new F1).
  - The tombstones need statements that are absent from the tree, and retired capabilities have no
    record to hold them (new F2).
  - The statement that the fingerprint hashes is undefined (new F3).
- **F9 (id derivation underspecified): MOSTLY FIXED.**
  - `slug`, the byte-wise suffix, the length threshold, and the collision checks are now exact.
  - Residual: the `kind` literal hashed into the suffix is never given.
  - Residual: B1's "cut to 25 bytes" does not say whether the cut is re-trimmed.
  - Residual: the SDD-INV-7 vectors are "computed once and committed", so they pin whatever the
    implementation chooses rather than what the SPEC says (F9 below).
- **F10 (cross-document contradictions): PARTIALLY FIXED.**
  - Fixed: the VD amendment withdraws governance → BD; the plan cites the BR as the one home for
    obligation numbering; the M3 chain states that governance is undeclared.
  - Not fixed: O-5's milestone ownership.
  - New stale statements: the intent SPEC link, the draft's own §0 and §6, and the plan's M0
    status (F9 below).
- **F11 (weak obligations and untestable invariants): PARTIALLY FIXED.**
  - O-6 and O-7 are correctly retired as tombstones, and replaced by O-8 and O-9.
  - O-9 cannot be satisfied for any archived OpenSpec requirement (new F4).
  - The O-2 and O-5 bundling, and SDD-INV-13's untestable half ("kit name read only from the
    profile"), are unchanged (F10 below).

## Findings

### r2-m0-review-round2-20261116-F1 (HIGH): B1's fail-closed fold makes the reference kit's own repository, and every repository that uses supported OpenSpec flows, permanently unsyncable

**Broken property:** SDD-INV-20 ("archiving a change never makes a later sync fail") and SDD-INV-23
contradict each other on real repositories. BR O-1 and the spike's GO verdict both depend on this.

**Evidence:**
- §9.B makes any disagreement between the folded archived deltas and the current main spec
  `intent.ErrUnmappable`. It also makes two same-date changes that touch one requirement
  `ErrUnmappable`.
- `ErrUnmappable` stops the whole batch (`Sync` validates the batch first; SDD-INV-14). Archived
  deltas are never edited again, so once a repository is inconsistent it stays inconsistent.
- The fold probe on `Fission-AI/OpenSpec` (the repository the spike used as GO evidence) gave the
  following results over 358 final requirement states:

  | Result | Count |
  |---|---|
  | Folded text matches the main spec exactly | 217 |
  | Text differs | 77 |
  | Requirement missing from the main spec | 13 |
  | Requirement belongs to a capability whose main spec no longer exists | 46 |
  | Same-date conflicts (one requirement touched by two changes archived on one date) | 16 |

  Hand-checked examples:
  - **Text differs.** `cli-archive` "Archive Validation": the main spec gained a "Proposal warnings
    stay proposal-level" scenario after its last archived delta (`2025-08-19-add-zod-validation`).
  - **Missing from the main spec.** "Workspace Product Language" exists only in
    `archive/2026-05-04-workspace-foundation/specs/openspec-conventions/spec.md`.
  - **Same-date conflict.** `cli-list` "Command Execution" is touched by both
    `2025-08-19-add-change-commands` and `2025-08-19-adopt-verb-noun-cli-structure`.
- The kit itself produces such disagreements on purpose:
  - `openspec archive --skip-specs` archives a change *without* applying its deltas
    (`docs/cli.md:688`, `src/cli/index.ts:615`).
  - `--no-validate` skips validation.
  - Bulk archive can exclude deltas.
- One `--skip-specs` archive therefore ends syncing forever.
- SDD-INV-20's derived test (`TestArchiveNeverBreaksSync`) runs over a fixture in which the author
  controls consistency, so it passes while the property is false for real repositories.

### r2-m0-review-round2-20261116-F2 (HIGH): B1's tombstones need predecessor statements that are not in the tree, and a retired capability has no record to hold them

**Broken property:** SDD-INV-22 versus SDD-INV-11 (read only the named tree), and the "no
fabrication" rule.

**Evidence:**
- §9.B requires a `retired` tombstone "carrying the old id and statement" for every requirement that
  an archived delta REMOVED, MODIFIED, or RENAMED away. Under B1 the old id is itself a function of
  the old statement.
- The archived files do not contain the old statement:
  - An OpenSpec MODIFIED block holds only the *new* text (`specs-apply.ts` replaces the block with
    `mod.raw`).
  - A REMOVED block holds the heading plus `**Reason**` / `**Migration**` (for example
    `archive/2026-02-17-merge-init-experimental/specs/cli-init/spec.md`).
- The old text is therefore recoverable only if an *earlier* archived ADDED or MODIFIED delta
  carried it. For a baseline requirement (one that predates delta use, which §9.B explicitly admits)
  it is not in the tree at all.
- The tombstone probe found that 29 of 113 archived MODIFIED or REMOVED requirements in the
  reference repository have no recoverable predecessor text. One example is `cli-list` "Command
  Execution" at `2025-08-19-add-change-commands`.
- The draft's own fixture hits this case. T0's two capabilities are baseline, and T1 modifies and
  removes their requirements. SDD-INV-22 at T2 can pass only in one of two ways:
  - The provider reads T0, which violates SDD-INV-11.
  - The provider emits a tombstone id and statement that the T2 tree does not contain, which is a
    fabrication.
- OpenSpec also retires whole capabilities: `archive.ts:72` (`isRetirableSpec`) deletes the main
  spec file once every requirement is REMOVED. §4 emits "one requirement record per matching
  main-spec artifact". With no artifact there is no record, so every obligation of that capability
  disappears without a tombstone. This is the intent SPEC §5 harm that B1 was chosen to prevent.

### r2-m0-review-round2-20261116-F3 (HIGH): the statement that B1 fingerprints is undefined, and the only definition in the lineage makes B1 a no-op

**Broken property:** SDD-INV-21 and SDD-INV-5/7 conformance. Two implementations that both conform
to the SPEC would assign different ids.

**Evidence:**
- B1's id is `O-<name-part>-<6 hex of SHA-256(statement)>`, but draft 2 never says what a
  heading-named obligation's `statement` is. The OpenSpec profile declares only
  `"heading": "^### Requirement: (?P<name>.+)$"`.
- Draft 1 §4 line 112 defined it as "Statement = the `statement` group, else the `name` group".
  Draft 2 dropped that line and put nothing in its place.
- If draft 1's rule carries over, the fingerprint is a function of the *name*. A MODIFIED
  requirement keeps its name, so it keeps its id, and SDD-INV-21 is false.
- If the statement is the SHALL line, a MODIFIED delta that changes only scenarios keeps its id.
- If the statement is the whole block's exact bytes, then:
  - any whitespace or reflow edit is a "meaning change";
  - an out-of-workflow edit of that kind is also a fold mismatch (F1);
  - the canonical payload ingests the requirement prose. Intent SPEC §6 lists "no prose-body
    ingestion beyond its source artifact digest" as a non-goal.
- `TestModifiedObligationGetsNewID` can be written to pass under any of these readings, so it does
  not pin the property.

### r2-m0-review-round2-20261116-F4 (HIGH): A2 and O-9 cannot be satisfied under the existing obligation-verdict.v2 rules, and drift is blind for every delta-bound verdict

**Broken property:** §9.A ("no new mechanism"), BR O-8 and O-9, and the plan's claim that M4 is where
the value lands.

**Evidence:**
- `reduceObligationVerdict` (`projections.go:456-515`) accepts a v2 verdict only when three
  conditions hold:
  - `(reviewed_commit, CI@digest)` is an existing row in `commit_intents_view`;
  - every outcome is an exact `intent_targets_view` target of *that* change-intent revision;
  - every evidence event's sequence is at or before the verdict's.
- A2 asks for a verdict "for the main record" after archive. No change intent targets the main
  record's post-archive obligations:
  - Only active changes produce change intents (§5.1).
  - CI-X's targets are the *delta* record's obligations (`implements`) and the main record's *old*
    obligations (`modifies` / `retires`), both at pre-archive digests.
- The archive commit, or any later commit, cannot claim X: an archived-only token is an error (§6,
  §7, SDD-INV-19).
- So no commit claim and change-intent revision exist that a re-attesting verdict could bind to, and
  the projector rejects it as "outcome is not an exact CI target". A2 therefore needs a new
  mechanism, such as a tenant-authored change intent (which `sdd` cannot emit) or a verdict form that
  is not commit-anchored.
- O-9 inherits the same gap: a "new obligation-verdict.v2 bound to the new revision" needs a change
  intent targeting the new revision. Once a requirement is archived, its stale flag can never clear.
- Conversely, every SATISFIED outcome for an ADDED or MODIFIED requirement binds the *delta* record
  (`implements`, §5.3). The delta record is never superseded and never gets a later revision. M4's
  rule ("binds a requirement digest that a later revision has superseded") can therefore never flag
  it, however the main spec changes later. This is round-1 F2/F4's drift blindness, which the draft
  does not resolve.

### r2-m0-review-round2-20261116-F5 (HIGH): "the ledger already holds every revision it had while active" is false, so commits made during a change's life can dangle and stop the projector

**Broken property:** the premise of §5.1, which carries the whole "only active changes map" fix
(round-1 F2), and intent SPEC §4 point-in-time resolution.

**Evidence:**
- `Revisions` reads only `HEAD` (§2), and the plan confirms it ("`sync` reads `HEAD`").
- A change opened and archived between two `sync intent sdd` runs is never ingested. Neither is any
  intermediate revision of a change edited between syncs. In the reference repository, 108 changes
  went through this lifecycle.
- The git connector, however, resolves each tenant commit's `Intent: sdd:X` at *that commit's* tree
  (§6 and §7, git INV-29). The result is a `CI-X@digest` that is not in the ledger.
- `projections.go:216-225` then returns `dangling intent revision` for the commit event. That fails
  the projection transaction (P-INV-7), not just one gate row.
- Prior-art providers share the intermediate-revision gap. `sdd` makes it routine, because the kit's
  normal workflow removes every change from `HEAD`.
- No invariant or test covers binding a commit from a change's active window after that change is
  archived. SDD-INV-20 only asserts that a post-archive sync succeeds.

### r2-m0-review-round2-20261116-F6 (HIGH): `artifact_sha256` no longer determines the payload, so exact-digest bindings (reviews, verdicts, commit claims) can silently change meaning

**Broken property:** intent SPEC §4 ("the provider's immutable revision bytes") and the exact-binding
basis of O-5 and O-8. SDD-INV-18's "and only then" enshrines the defect.

**Evidence:**
- **Change intents.** The manifest covers only "files the mapping reads in that directory" (§5.1). A
  change intent's payload also depends on two inputs outside the manifest:
  - each target's main-record `artifact_sha256` (a different file);
  - the sponsor, taken from `.proofbound/sdd-profile.json`.
- **Requirements.** A requirement's payload depends on inputs outside its artifact bytes:
  - `declared_owner`, from the profile;
  - for the main record, the B1 tombstones, derived from archived delta files.
- In `records` and `specdir`, every payload field comes from the artifact's own bytes.
- The projector keys revisions on `(source, id, artifact_sha256)` and keeps the last write:
  - `requirements_view` replaces `declared_owner` on conflict (`projections.go:335`).
  - `change_intents_view` replaces the payload on conflict (`projections.go:358`).
  - `intent_targets_view` only adds rows (`projections.go:370`).
- Consequences:
  - **A profile edit changes the owner of an already-reviewed revision.** The `declared_owner` of a
    requirement revision that already holds a `requirement-review.v1` changes without changing its
    key, and the reviewer-is-not-owner check (`projections.go:551`) is evaluated against whichever
    owner was written last.
  - **Archiving one change re-points every active change on that capability.** It changes the main
    digest, so each active change intent targeting that capability gets new target rows under its
    *unchanged* key. Commit claims already bound to `CI-Y@manifest` now point at a different target
    set.
  - **Verdicts must cover stale and current targets together.** The verdict rule "cover every
    targeted obligation exactly once" (`projections.go:506`) now counts both the stale and the
    current target rows.
- Git INV-29 / intent INT-INV-8 compare only the artifact digest. When a commit tree's main spec
  differs from HEAD's, the commit binds a digest whose ledger payload was not the payload at its
  tree, and no mismatch is detected.

### r2-m0-review-round2-20261116-F7 (MED): OpenSpec's documented flows break §5.1 and §5.3 while a change is active

**Broken property:** SDD-INV-9 and the §5.1 claim that targets exist "unless someone edited the main
spec outside the workflow".

**Evidence:**
- **Early sync.** `/opsx:sync` (`docs/opsx.md:224`) merges a change's deltas into the main specs
  "without archiving — the change stays active". `specs-apply.ts:461-470` and `:520-530` treat an
  already-absent REMOVED target, or a RENAMED FROM target, as "already synced". After such a sync:
  - a REMOVED or RENAMED target is absent from the main spec of an *active* change, which makes the
    whole repository `ErrUnmappable`;
  - a synced ADDED requirement lives both in the `proposed` delta record and in the main record,
    which reopens round-1 F4.
- **Rename plus modify.** OpenSpec applies RENAMED first and matches MODIFIED "using new names if
  renamed" (`openspec/specs/openspec-conventions/spec.md:184`). §5.3 makes MODIFIED `modifies` "the
  main record's obligation of the same name". At an active tree the main spec still holds the FROM
  name, so the target is missing and the result is `ErrUnmappable` (SDD-INV-9).
- **Changes without deltas.** OpenSpec supports changes with no spec deltas (`skip_specs: true`,
  `docs/cli.md:688`); 8 of the 108 archived changes have no `specs/` directory. While such a change is
  active, "one change intent per active change" yields an intent with zero targets. Canonical
  validation rejects that (`intent.go:265`), so the batch fails. The draft says nothing about this
  case.

### r2-m0-review-round2-20261116-F8 (MED): §6's INV-29 adaptation would turn an independent anti-spoofing check into the provider vouching for itself, and it changes a locked payload that the draft does not mention

**Broken property:** git INV-29 / INV-30 and the soundness of §6. No SDD invariant covers the
connector side.

**Evidence:**
- Today `resolvedCommit` checks `ref.RecordID == id`, where `id` is the trailer token
  (`connector/git/connector.go:208`). That is a check the connector performs *independently* of the
  resolver.
- For `sdd`, the record id is `CI-<slug(name)>-<suffix(kit,kind,name)>`. The connector can derive it
  only by knowing the profile's `kit` at that tree, which only the provider reads.
- So "the token must be the resolved change's native name" can be checked only by the provider, and
  that is the provider certifying its own answer.
- `IntentRef` has exactly `provider`, `record_id` and `artifact_sha256` (`connector.go:42-46`).
  Recording the token as well, as §6 requires, changes the git commit payload, which INV-30 pins.
- §2's locked `Resolve(ctx, tree, token) (intent.Revision, error)` returns no native name for the
  connector to compare against.
- Three existing checks reject `sdd:<name>`, and the draft says no more than that "both parsers
  extend":
  - the connector's provider allowlist (`connector.go:199`);
  - `intentIDRE` (`connector.go:21`);
  - the shell gate's `id_re` (`scripts/intent-delivery-applicability.sh:93`).
- The work is deferred to M3, but §2 locks the `Resolve` signature at M1, so the deferral constrains
  M1's interface now.

### r2-m0-review-round2-20261116-F9 (MED): the round-1 F10 drift class recurs. Stale statements across the SPEC, the plan, the intent SPEC, the spike and the BR

**Broken property:** Law 2 (one home per datum) and the stale-index class.

**Evidence:**
- **Intent SPEC.** The amendment in `kernel/internal/connector/intent/SPEC.md` §1 still names
  `docs/plans/R2-sdd-provider-spec-draft1.md` as the draft contract, which is the superseded
  NEEDS_WORK draft.
- **Draft 2 contradicts itself:**
  - The §0 table maps F8 to "**Open decision B**".
  - The §6 diagram says "OPEN DECISION A".
  - The header says both were settled.
  - §4.1 and §9.B state two rules for one obligation-id datum, joined by a "superseded by" note.
- **Plan.** The M0 status (`sdd-warranty-connector-plan.md:100-103`) still says "open decisions A …
  and B … are the owner's", after `bcfb7ca` settled them.
- **Spike.** It still states "archived → `accepted`" (line 37), path-derived record ids (lines 71-73)
  and "past 31 chars" (line 28). Draft 2 contradicts all three, and the spike carries no amendment
  note.
- **BR 0002.** O-5 is carried over unchanged, but:
  - it says "`Resolve(tree, id)` returns exactly one artifact", while draft 2 has `Resolve(tree,
    token)` over a synthesized manifest that is not "exact committed bytes";
  - it keeps O-5 in M1, while O-5's trailer clause is owned by M3 (draft §11, plan M3). This was
    round-1 F10's last bullet, and it is not fixed.
- **Residual of round-1 F9.**
  - The `kind` string in `suffix(key)` is never given as a literal.
  - The SDD-INV-7 vectors are "computed once and committed" from the implementation, so they cannot
    catch an implementation that diverges from the SPEC.

### r2-m0-review-round2-20261116-F10 (LOW): BR revision 0002 is digest-valid but not yet fit to receive a requirement-review.v1 and go active

**Broken property:** testability and coverage of the BR obligations.

**Evidence:**
- **What holds.** Revision 0002 is well-formed:
  - its self-digest reproduces (`fbf813f6…`);
  - it supersedes revision 0001 at `e03ca43a…`;
  - O-6 and O-7 are kept with `state: retired` and unchanged statements, which matches the records
    rule (`records.go` `validateLineages`);
  - every obligation has a milestone.
- **What blocks review:**
  - O-9 cannot be satisfied as worded under the drafted design (F4).
  - O-5 conflicts with draft 2 (F9).
  - O-2 and O-5 still bundle properties owned by different milestones.
  - SDD-INV-13's "the kit name is read only from the profile" still has no mechanical test form.
  - Change-name reuse (a new active change named like an archived one) merges two distinct changes
    into one record identity. Draft 2 does not mention it.

## What holds

- **Round-1 F1, F3 and F7 are genuinely fixed.**
  - Delta records are always `proposed`.
  - Ids come from native names and survive archive.
  - `by_path` is gone.
- **Ordered `owners` and `sponsors` lists** remove the first-match trap. I again found no path to a
  default, git-author, or placeholder identity (VD-sdd-accountable-identity).
- **Obligation-id length:** B1's `O-` + 25 + `-` + 6 is 32 characters, within `obligationRE`
  (`intent.go:25`). The §4.1 slug and suffix are now byte-exact.
- **Invariant numbering:** SDD-INV-1..14 titles are byte-identical to draft 1, and INV-15..23 are
  appended. This is the append-only numbering discipline.
- **Records and gates:**
  - BR revision 0002's digests and lineage are correct.
  - The three round-2 M0 commits are `applicable=false` under
    `scripts/intent-delivery-applicability.sh`, so they needed no `Intent:` trailer.
  - The VD amendment withdrawing governance → BD is correctly scoped.
- **Build Laws:** Law 6 ordering holds (a SPEC draft and no `intent/sdd` code), and Law 8 holds (JSON
  profiles, stdlib).
- **Section 6 boundaries are sound:** read-only feedback, and "the verdict is the proof" (no test →
  obligation inference).
- **§12's tree-sequence fixture discipline** is the right shape. F1, F2 and F5 show that the
  sequence must include real-kit histories, not only author-consistent trees.

## Fitness for M1

**No.** Draft 2 fixes the snapshot defects of round 1, but the lifecycle it chose does not hold up
against how OpenSpec is actually used:

- **The fold fails on the reference repository.** B1's fail-closed fold rejects the kit's own
  repository and every `--skip-specs` archive (F1).
- **The tombstones cannot be built from the tree.** They need statements the tree does not contain,
  and a capability retired on archive has no record to hold them (F2).
- **The obligation id is undefined.** The statement that B1 fingerprints is not specified (F3).
- **A2 and O-9 cannot be met.** They are unreachable under the existing obligation-verdict.v2 rules,
  and drift cannot see verdicts bound to delta records (F4).
- **Archived changes can break the projector.** The HEAD-only sync never ingests changes archived
  between syncs, so commits from their active window dangle (F5).
- **The artifact digest no longer pins the payload.** That breaks the exact-binding property the
  whole warranty rests on (F6).

These defects shape the M1 interface itself: `Revisions` at HEAD, the `Resolve` return type, the
artifact definition, and the obligation-id function. M1 code written to this contract would encode
them. The contract should be revised, with the lifecycle tests run against at least one real OpenSpec
history, before code begins.
