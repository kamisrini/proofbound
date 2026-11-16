---
schema: vera.verdict.v1
verdict_id: r2-m0-review-round1-20261116
status: NEEDS_WORK
reviewed_commit: 1de1b21ac2b04a04474fcb33d006e95356754f7f
findings:
  - finding_id: r2-m0-review-round1-20261116-F1
    severity: HIGH
  - finding_id: r2-m0-review-round1-20261116-F2
    severity: HIGH
  - finding_id: r2-m0-review-round1-20261116-F3
    severity: HIGH
  - finding_id: r2-m0-review-round1-20261116-F4
    severity: MED
  - finding_id: r2-m0-review-round1-20261116-F5
    severity: MED
  - finding_id: r2-m0-review-round1-20261116-F6
    severity: MED
  - finding_id: r2-m0-review-round1-20261116-F7
    severity: MED
  - finding_id: r2-m0-review-round1-20261116-F8
    severity: MED
  - finding_id: r2-m0-review-round1-20261116-F9
    severity: MED
  - finding_id: r2-m0-review-round1-20261116-F10
    severity: MED
  - finding_id: r2-m0-review-round1-20261116-F11
    severity: LOW
artifact_path: docs/verification/verdicts/r2-m0-review-round1-20261116.md
artifact_sha: 79e39f2935465ea94de5e29006a04ef09a3739009cbc0021732008d35bb37f81
---

# Release 2 M0 (SDD Warranty Connector): plan, spike, intent home, decisions, and draft `sdd` provider SPEC. Review round 1

## Provenance

- **Under review:** the documentary M0 deliverables of Release 2 at commits `a028984` (release line),
  `9dc7729` (mapping spike), `23a6fcc` (BD/BR/CI intent home + `VD-sdd-accountable-identity-ypcx7x`),
  and `b0f5e5d` (draft provider SPEC + intent SPEC amendment). Files:
  `docs/plans/sdd-warranty-connector-plan.md`, `docs/verification/r2-m0-mapping-spike.md`,
  `docs/intent/records/{business-decisions/BD-sdd-warranty-jdyjhh,requirements/BR-sdd-warranty-i9miv7,change-intents/CI-sdd-warranty-m0-a49f5b}/0001.md`,
  `docs/decisions/VD-release-line-r2-sdd-t5j34p.md`, `docs/decisions/VD-sdd-accountable-identity-ypcx7x.md`,
  `docs/plans/R2-sdd-provider-spec-draft1.md`, and the amendment paragraph in
  `kernel/internal/connector/intent/SPEC.md` §1.
- **Judged against:** CLAUDE.md Build Laws 1-10; `kernel/internal/connector/intent/SPEC.md`; the actual
  validation in `kernel/internal/connector/intent/intent.go`; prior art `specdir/` and `records/`.
- **HEAD at dispatch:** `1de1b21ac2b04a04474fcb33d006e95356754f7f`. **HEAD at completion:**
  `1de1b21ac2b04a04474fcb33d006e95356754f7f` (unchanged).
- **Driving mechanism:** `/proofbound-review` procedure, executed by a dispatched Agent.
- **Tier:** connector/peripheral (one adversarial review). Several findings below touch record
  identity and lineage, which the review procedure names as an escalation trigger for M1.
- **Reviewer identity:** `dispatched-non-author-reviewer-20261116-r2m0`, a **non-author**. It did
  not write any of the work under review, so this verdict counts as acceptance evidence under Law 9.
- **Writes:** no repository files other than this verdict and its sibling
  `intent-trailer-gate-review-round1-20261116.md` were written. Probes ran only in a scratch directory.
- No cleanroom redactions were needed.

## Findings

### r2-m0-review-round1-20261116-F1 (HIGH): the delta requirement record gets a status that the requirement vocabulary rejects, so any OpenSpec repo with an archived change cannot sync

**Broken property:** SDD-INV-2/INV-12/family conformance ("every emitted revision passes
`intent.ValidateRevision`", draft §9) and BR O-1/O-4.

**Evidence:** Draft §5.1 (lines 126-129) says that a delta requirement record takes its "status from
`changes.status`". The reference profile maps `changes.status.by_path` to `"accepted"` /
`"proposed"` (line 71). However, `intent.go:250` admits a requirement only with status in
{`proposed`, `active`, `superseded`, `retired`}, and `accepted` is not in that set. So every archived
OpenSpec change that has an ADDED or MODIFIED section emits an invalid requirement revision.
`Sync` validates the whole batch before appending anything (`intent.go:182`, `validateBatch`), so
one such revision makes the entire tenant sync fail. The spike's own reference repo has 108 archived
changes. The draft's own test fixture (§9, line 210: "one archived change with ADDED, MODIFIED, and
REMOVED sections") cannot be the neutral control the draft says must map. No rule maps change-intent
statuses onto requirement statuses.

### r2-m0-review-round1-20261116-F2 (HIGH): MODIFIED/REMOVED targets are resolved against the main spec at HEAD, so archiving a REMOVED change makes the repo permanently unmappable, and archived intents can never go stale

**Broken property:** SDD-INV-9 / SDD-INV-11, and BR O-5 and O-7.

**Evidence:**
- `Revisions` reads at `"HEAD"` (draft §2). §5.1 (lines 130-134) resolves `modifies` and `retires`
  against "the capability's **main** requirement record at the same tree", and a target obligation
  that is absent there is `intent.ErrUnmappable`.
- Archiving an OpenSpec change applies its deltas to `openspec/specs/...`. After a REMOVED change is
  archived, the main spec no longer contains the removed requirement, while the archived delta (which
  is never edited again) still lists it under `## REMOVED Requirements`. That mapping is
  `ErrUnmappable` on every later sync, so the tenant can never sync again.
- The same thing happens to any archived MODIFIED change whose requirement a later change removes or
  renames.
- For archived intents that still map, the `modifies` target binds the main record's digest *at
  HEAD*, not the digest the change was made against. Two harms follow:
  - Every edit to a main spec mints a new revision of every archived change intent that ever touched
    that capability: the payload's target digest changes, so the `content_sha` changes.
  - Such an intent always binds the latest digest, so the O-7 drift projection ("a reference bound to
    a digest a later revision supersedes") can never flag it.
- The plan names M4 drift as "where the differentiated value lands". For archived OpenSpec changes,
  this mapping makes that value structurally unreachable.

### r2-m0-review-round1-20261116-F3 (HIGH): record identity is derived from the path, so archiving an OpenSpec change silently forks its identity and the proposed→accepted transition never happens

**Broken property:** SDD-INV-7 is true but harmful. The intent SPEC §5 lifecycle ("a successor may
supersede exact earlier revisions") is unreachable, and the spike's status claim fails.

**Evidence:**
- §4 (lines 102-104) derives the id suffix from `"<kit>:<path>"`. §5.1 (line 135) derives the CI slug
  and suffix from the change directory path.
- `openspec archive` moves `openspec/changes/<name>/` to `openspec/changes/archive/YYYY-MM-DD-<name>/`.
  That changes the CI's slug and suffix, and the suffix of every delta requirement record.
- §4 line 116 and §5 always emit `supersedes: []`. As a result:
  - The pre-archive `proposed` CI and its delta BR stay in the ledger as `proposed` forever. They are
    never superseded or withdrawn.
  - The archived CI is a new, unrelated record with no lineage to the proposed one.
  - Tenant commits whose `Intent:` trailer named the pre-archive id are bound to the orphan.
- The spike (`r2-m0-mapping-spike.md:37`, "Active change → `proposed`; archived → `accepted`")
  describes a status *transition* of one change. Under the derivation, no single record ever makes
  that transition.
- SDD-INV-7's proving test (`TestRecordIDsAreDeterministic`) passes while this property is false,
  because determinism of a function of the path is exactly what causes the fork.

### r2-m0-review-round1-20261116-F4 (MED): an ADDED requirement ends up with two homes and no supersession link (Law 2)

**Broken property:** Law 2 (one home per datum) and BR O-7.

**Evidence:**
- §5.1 puts ADDED requirements, and also MODIFIED ones (line 128), as obligations of a *delta*
  requirement record built from the change's delta file.
- After archive, the same requirement also lives in the capability's main spec record, with the same
  name-derived obligation id.
- Nothing links the two: `supersedes` is always `[]`, and the main record never cites the delta
  record. A change intent `implements` the delta record's obligation, and the archived delta file is
  never edited again.
- Later edits to that requirement in the main spec therefore never make that intent's reference stale
  (O-7).
- The MODIFIED requirements copied into the delta record are not targeted by anything, so they are an
  unreferenced second copy of the main record's obligations.

### r2-m0-review-round1-20261116-F5 (MED): a change intent has no defined artifact

**Broken property:** the intent SPEC §3/§4 envelope (`artifact_path` + `artifact_sha256` of "the
provider's immutable revision bytes"), SDD-INV-8, and SDD-INV-3 for change intents.

**Evidence:**
- §5.1 emits "one change intent per change directory" (line 135). It also says each delta file
  `<change>/specs/<capability>/spec.md` is "one change's view of one capability" (line 123), so a
  change that touches N capabilities has N files.
- The canonical change-intent payload carries exactly one `artifact_path` and one `artifact_sha256`
  (`intent.go:265`, and the envelope binding check after it).
- The draft does not say which bytes the CI's digest covers, or which path an `identity` glob is
  matched against for a CI (the directory? each delta file?).
- SDD-INV-8 ("artifact_sha256 is the exact committed bytes") cannot be stated, or tested, for a
  multi-file change intent.

### r2-m0-review-round1-20261116-F6 (MED): the draft's own reference profile and rules reject valid OpenSpec changes

**Broken property:** SDD-INV-3 / SDD-INV-6 as applied to §5.1, and the neutral calibration control
in §9.

**Evidence:**
- **Identity.** The delta requirement record's path is `openspec/changes/<c>/specs/<cap>/spec.md`.
  Under "first match wins" (line 95), its first, and only, matching identity entry is
  `{"pattern": "openspec/changes/**", "sponsor": ...}` (line 76). That entry has no `owner`, and §3
  makes "an entry covering it without the needed field" `ErrUnmappable`. So the reference profile
  rejects every change with an ADDED or MODIFIED section.
- **Zero obligations.** A delta file with only `## REMOVED Requirements` yields a delta record with
  zero obligations, which §4 makes `ErrUnmappable`. A REMOVED-only change is ordinary OpenSpec usage,
  and the batch-wide validation turns it into a whole-repo sync failure.
- **Unknown sections.** OpenSpec's delta format also has a `## RENAMED Requirements` operation (from
  the reviewer's knowledge of the format; not re-verified against the spike's clone). It is absent
  from `sections`, and the draft does not specify what happens to requirement headings under an
  unlisted section.

### r2-m0-review-round1-20261116-F7 (MED): `changes.status.by_path` is order-dependent but unordered and outside SDD-INV-4

**Broken property:** SDD-INV-4 ("status comes only from the explicit map or a declared absent value")
and SDD-INV-1.

**Evidence:**
- Line 71 declares `by_path` as a JSON *object*. Its two keys overlap: `openspec/changes/**` matches
  every archive path too.
- Precedence is the only thing that gives archived changes `accepted`, but JSON object member order
  is not semantically preserved, and a stdlib `map` decode loses it.
- §3's field rules define status only via `line` / `map` / `absent`. `by_path` is a third mechanism
  with no stated precedence and no rule in SDD-INV-4.
- An implementation that iterates a Go map assigns `proposed` or `accepted` to archived changes
  non-deterministically. `TestStatusMapFailsClosed`, as derived, would not detect this.
- `identity` correctly uses an ordered array. `by_path`, and to a lesser degree `sections`, do not.

### r2-m0-review-round1-20261116-F8 (MED): obligations can silently disappear or be re-identified, against intent SPEC §5

**Broken property:** intent SPEC §5 ("Active obligation IDs cannot disappear … retirement leaves a
tombstone"), and SDD-INV-5 read in its spirit.

**Evidence:**
- The `records` provider enforces §5 (`records/records.go:231,237`: "obligation %s disappeared" /
  "changed meaning").
- The draft has no equivalent rule, invariant, or test. Its requirements carry `supersedes: []` and
  status `active` via `absent`.
- After a REMOVED change is archived, the main record's next revision simply lacks an obligation
  that was active, with no tombstone and no `retired` state.
- OpenSpec obligation ids are derived from the heading *name* (§4 lines 106-109), so renaming a
  heading changes the id. One obligation vanishes and an unrelated one appears; neither the drift
  projection nor any lineage connects them.
- "Never positional" is achieved, but name-derived identity reproduces the same class of harm on
  rename. The draft states no position on this.

### r2-m0-review-round1-20261116-F9 (MED): ID derivation is underspecified enough that conforming implementations disagree or emit invalid ids

**Broken property:** SDD-INV-5 / SDD-INV-7 and family conformance.

**Evidence:**
- **The `[a-z0-9-]` mapping is not defined.** §4 line 103 says the slug is the directory name
  "lower-cased to `[a-z0-9-]`" without saying whether other characters are dropped or replaced.
- **Short or empty slugs fail the record-id regex.** `recordIDRE` (`intent.go:24`) requires
  `[a-z0-9][a-z0-9-]{1,62}`, that is ≥ 2 characters starting alphanumeric. A capability or feature
  directory named `a`, `_internal`, `-x`, or a purely non-ASCII name yields `BR-a-…`, `BR--…`, or
  `BR--…`, all of which fail the regex. The batch-wide validation turns that into a whole-repo
  failure, and the draft does not classify it as `ErrUnmappable`.
- **The suffix rule is ambiguous.** "the first 6 characters of the SHA-256 … mapped to `[a-z0-9]` by
  `byte % 36`" can mean raw digest bytes or hex characters, and the two give different ids.
- **Obligation names can produce invalid or colliding ids.**
  - A name made only of non-ASCII characters collapses to the empty string, giving `O-`, which is
    invalid.
  - Names such as `User Login` and `user-login` collapse to the same id, which is `ErrUnmappable` for
    the whole repo.
- **No uniqueness check within a tree.** Two different artifacts whose ids collide pass
  `validateBatch`, because its duplicate key includes the digest (`intent.go:372-374`). Downstream they
  would be treated as revisions of one record. The probability is low, but the draft states no
  detection rule.
- **The threshold disagrees with the spike.** The spike says "hash-suffixed past 31 chars"
  (`r2-m0-mapping-spike.md:28`); the draft says "longer than 32" (line 108).

### r2-m0-review-round1-20261116-F10 (MED): plan, VD, BD, BR, and SPEC contradict each other, including a stale obligation numbering

**Broken property:** Law 2, and the stale-index class the Law-numbering preamble warns about.

**Evidence:**
- **Governance mapping.**
  - Accepted `VD-release-line-r2-sdd-t5j34p.md:20` still decides "governance/ADRs → business
    decisions".
  - Plan M2 (`sdd-warranty-connector-plan.md:108`) still delivers "ADR → business decisions" and
    "proposals/tasks → change intents".
  - The draft (SDD-INV-2) forbids business decisions, and its OpenSpec mapping ignores proposals and
    tasks.
  - The spike (`:86`) says governance→BD was "moved to non-goals for R2", but the plan's Non-goals
    list does not contain it.
  - The BD outcome and BR O-1 still describe emitting "business-decision" revisions.
- **Obligation numbering.**
  - Plan `:75` lists the BR obligations as `O-1` provider, `O-2` profiles, `O-3` binding/enforce,
    `O-4` drift.
  - The ratified BR has `O-3` = accountable identity, `O-4` = profiles, `O-6` = gate, and `O-7` =
    drift.
  - Anyone citing the plan's O-3 or O-4 gets the wrong obligation. Both still resolve to real
    obligations, so the error is silent.
- **Plan M3 gate chain.** It specifies "change intent → requirement obligation → governance"
  (`:116`), but every `sdd` requirement has `authorization: undeclared`. The plan neither resolves
  this nor states it as a gap.
- **Ownership of O-5.**
  - The draft header says this package discharges O-1..O-5.
  - The BR body assigns O-5 to M1.
  - O-5 includes "tenant Intent: trailers bind through it", which draft §8 assigns to the git
    connector and the trailer gate in M3.

### r2-m0-review-round1-20261116-F11 (LOW): several BR obligations and invariant tests cannot discriminate as worded

**Broken property:** testability of BR O-2/O-5/O-6/O-7 and of SDD-INV-13, which matters for
promoting BR `proposed`→`active`.

**Evidence:**
- **O-6** ("refuses an applicable tenant change whose intent targets no requirement obligation") is
  nearly vacuous for `sdd`. Canonical validation already rejects a change intent with zero targets
  (`intent.go:265`) and a requirement target without obligations (`intent.go:327-331`). A gate that
  adds no new check satisfies O-6 for every sdd-sourced intent. The real spec→code gaps (a missing
  trailer, or an unresolvable `sdd:` trailer) are not stated.
- **O-7**'s "until re-verified" is undefined, and F2-F4 show it is structurally unreachable for
  archived and ADDED requirements.
- **O-2 and O-5** each bundle 3-4 separately-failing properties that are owned by different
  milestones.
- **SDD-INV-13**'s derivation ("fails if the kit name is consulted anywhere except the profile's own
  field") has no mechanical test form.

On F1-F11 together, the BR is not fit to move from `proposed` to `active` at this HEAD.

## What holds

- **Accountable identity (SDD-INV-3, VD-sdd-accountable-identity).** Ordered first-match, a required
  field per kind, and fail-closed on no match: I found no path to a default, git-author, or
  placeholder identity. The defect in F6 is a too-strict rejection, not a fabrication.
- **Status (SDD-INV-4) for `line` / `map` / `absent`.** It needs an explicit table or a declared
  value and fails closed; I found no path to an invented requirement status. F7 is confined to
  `by_path`.
- **Business decisions.** No BD is emitted, requirements carry `authorized_by: []`, and the intent
  SPEC §3 sanctions empty `authorized_by` for a provider without an authorization segment.
- **Digests.** Exact committed bytes, consistent with intent SPEC §4. JSON profiles with RE2 named
  groups and a hand-written `**` glob need no new dependency (Law 8 holds).
- **Spec Kit.** Obligation ids `O-FR-001` etc. fit `obligationRE`, and feature-level-only change
  intents avoid an invented task→requirement link.
- **Intent SPEC amendment.** Correctly scoped as conditional ("production only once its package
  exists and its proofs pass `make check`"). Law 6 ordering (spec draft before any code) holds; no
  `intent/sdd` code exists.
- **Records.** The BD/BR/CI records' self-digests, id formats, and cross-reference digests are
  consistent (BR → BD `b451adb8…`; CI → BR `e03ca43a…`). The CI scope matches the M0 commits, and
  `23a6fcc` passes the Intent-trailer gate.
- **Truncation length.** The "first 25 + `-` + 6 hex" obligation-id truncation stays within
  `obligationRE` (32 characters after `O-`).
