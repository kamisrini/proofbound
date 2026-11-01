# SPEC — `internal/twin/manifest`

**Status:** P7-NONRUNTIME-PROJECTOR-MANIFEST-DECLARATION-1 authorized non-runtime preparation slice · 2026-10-27;
extended with the `proofbound.p7.case-manifest.v1` evidence index and the
`proofbound.p7.behavior-lock.v1` `ReplayIsolated` lock contract · 2026-11-01
**Authority:** [`CI-p7-projector-manifest-declaration-1a2b3c`](../../../../docs/intent/records/change-intents/CI-p7-projector-manifest-declaration-1a2b3c/0001.md) ·
[`CI-p7-case-manifest-1a2b3c`](../../../../docs/intent/records/change-intents/CI-p7-case-manifest-1a2b3c/0001.md) ·
[`CI-p7-behavior-lock-1a2b3c`](../../../../docs/intent/records/change-intents/CI-p7-behavior-lock-1a2b3c/0001.md)

## Purpose and boundary

This package validates the documentary `proofbound.p7.projector-manifest.v1` contract and its
`proofbound.p7.source-toolchain-inventory.v1` evidence inventory. It also validates the
`proofbound.p7.case-manifest.v1` case evidence index and the `proofbound.p7.behavior-lock.v1`
authored `ReplayIsolated` behavior-lock contract. It binds
exact source and projector revisions, declared Go/toolchain identity, selected module byte
identities, and build settings. It validates supplied declarations and verifies bytes returned by
an injected committed-tree reader; it does not itself access the filesystem, invoke Git, retrieve,
build, execute, or validate a binary or runtime input.

The package does not add a provider, event kind, writable truth, case-manifest parser, lock
evaluator, regeneration path, runtime fixture, PostgreSQL route, or external feed. Incomplete
evidence remains absent from this contract and cannot be rendered as a runtime `PASS`.

## Canonical contract

The input is strict RFC 8785 canonical JSON with exactly these top-level fields:

```json
{
  "schema": "proofbound.p7.projector-manifest.v1",
  "evidence_revision": "<40 lowercase hex characters>",
  "projector_revision": "<40 lowercase hex characters>",
  "source": {
    "files": [
      {"path": "<safe relative path>", "mode": "100644", "sha256": "<64 lowercase hex>"}
    ]
  },
  "toolchain": {
    "go_version": "go1.26.7",
    "goos": "linux",
    "goarch": "amd64",
    "archive_format": "tar.gz",
    "archive_sha256": "<64 lowercase hex>"
  },
  "modules": [
    {"module": "<module path>", "version": "<version>", "mod_sha256": "<64 lowercase hex>", "zip_sha256": "<64 lowercase hex>"}
  ],
  "build": {
    "command": "<exact non-empty command>",
    "cgo_enabled": false,
    "gcflags": "",
    "ldflags": "",
    "tags": []
  }
}
```

Source files are non-empty, safe, sorted, and unique. Modes are `100644` or `100755`. Modules
are non-empty, sorted by module then version, and unique. Tags are sorted and unique. Every
declared digest is lowercase SHA-256. `Parse` rejects unknown, missing, null, duplicate, or
non-canonical fields. `Digest` hashes the validated canonical bytes only.

The toolchain archive digest is an identity declaration, not proof that the archive was fetched,
licensed, extracted, or executed. Those activities remain outside this authorized slice.

## Source/toolchain evidence inventory

The inventory is strict RFC 8785 canonical JSON with exactly `schema`, `evidence_revision`,
`manifest_digest`, and `fields`. `fields` contains the exact sorted set of manifest paths: schema,
revision fields, source files, toolchain fields, modules, and build fields. Each field is either
`BOUND`, with a safe relative committed evidence path, matching evidence revision, and lowercase
SHA-256 declaration, or `UNVERIFIED` with all evidence fields empty. `manifest_digest` must equal
the SHA-256 of the validated canonical projector manifest supplied by the caller.

Inventory validation validates declarations only. `ResolveInventory` reads only bytes returned by
the caller’s `CommittedEvidenceReader` at each declared revision/path. It does not itself read the
filesystem, resolve a Git tree, retrieve a toolchain or module, build an executable, or establish
completeness beyond the declared field coverage.

## Committed evidence resolution

`ResolveInventory` first validates the inventory and manifest. For each `BOUND` field it asks the
injected reader for the declared path at the declared evidence revision and requires the returned
bytes to hash to the declared SHA-256. A missing path, unavailable revision, unsafe path, or digest
mismatch fails closed. `UNVERIFIED` fields are returned unchanged and are never passed to the
reader.

## Evidence closure evaluator

`EvaluateEvidenceClosure` is a pure evaluator over an accepted inventory and resolution list. It
requires exact sorted coverage of every required field and classifies each field as `BOUND`,
`UNVERIFIED`, or explicit `PENDING`. Missing, extra, duplicate, contradictory, stale, malformed,
or evidence-bearing non-`BOUND` inputs fail closed. Overall status precedence is `PENDING`, then
`UNVERIFIED`, then `BOUND`; an incomplete result can never become `BOUND` by absence.

This result is documentary completeness evidence only. It does not establish a reproducible build,
binary validity, runtime behavior, PostgreSQL approval, or any behavior-lock acceptance.

## Evidence-manifest skeleton

`BuildEvidenceManifest` deterministically derives a strict documentary skeleton from the accepted
projector manifest, source/toolchain inventory, committed-evidence resolutions, and closure result.
It records the manifest and inventory digests, closure overall state, and one exact status row per
required field. It preserves `UNVERIFIED` and `PENDING` states and rejects any supplied closure or
digest that does not exactly match the derivation inputs.

The skeleton is an evidence index only. It does not claim complete projector/build evidence,
reproducibility, binary validity, runtime behavior, PostgreSQL approval, or behavior-lock
acceptance.

## Projector/build manifest declaration

`DeclareCommittedManifest` reads only bytes returned by the caller-supplied committed-tree reader
at one exact evidence revision. It binds sorted source paths and accepted module artifact paths to
the manifest’s SHA-256 identities, carries the exact projector revision and build settings, and
emits a canonical declaration with explicit evidence states. Toolchain/archive bytes remain
`UNVERIFIED` when unavailable in the committed slice; declaration fields whose completeness is not
established remain `PENDING`.

This declaration is a non-runtime identity record. It does not retrieve, build, execute, validate,
or approve a toolchain, executable, PostgreSQL artifact, complete source corpus, or reproducible
build.

## Case evidence manifest

`proofbound.p7.case-manifest.v1` is the documentary evidence index for one bounded P7
`ReplayIsolated` case. It binds the contract references, exact case inputs, recorded commands and
witnesses, per-outcome evidence states, produced artifacts, and the independent-review pointer for
one evidence revision. It is an index of declarations and content digests; it runs nothing,
retrieves nothing, and approves no runtime input.

The input is strict RFC 8785 canonical JSON with exactly these top-level fields: `schema`,
`case_id`, `subject`, `case_status`, `evidence_revision`, `contract_refs`, `inputs`, `commands`,
`witnesses`, `outcomes`, `artifacts`, and `review`.

- `case_id` is a lowercase stable slug (`[a-z0-9][a-z0-9-]{0,63}`); `subject` is exactly
  `ReplayIsolated`; `evidence_revision` and `review.author_commit` are lowercase 40-hex Git
  revisions.
- `contract_refs` has exactly `lock_addendum`, `fixture_matrix`, and `implementation_scope`, each a
  `{path, sha256}` reference with a safe relative path and lowercase SHA-256.
- `inputs` has exactly `lock`, `source`, `candidate`, `setup_sources`, and `projector_manifest`,
  each a `{path, sha256}` reference.
- `commands` is an ordered, possibly empty list; each entry has exactly `id`, `argv`, `cwd`,
  `network_policy`, `environment_digest`, `exit_code`, `stdout`, and `stderr`. `argv` is a non-empty
  list of non-empty strings; `cwd` is `.` or a safe relative path; `network_policy` is one of
  `none`, `loopback-only`, or `offline`; `environment_digest` is a lowercase SHA-256;
  `stdout`/`stderr` are `{path, sha256}` references; command ids are unique safe slugs.
- `witnesses` is an ordered, possibly empty list of `{kind, path, sha256}` with unique kinds.
- `artifacts` is an ordered, possibly empty list of `{role, path, size_bytes, sha256}` with unique
  roles and `size_bytes >= 0`.
- `review` has exactly `author_commit` and `independent_verdict`; the verdict is `{status}` with a
  status in `PENDING`, `ACCEPTED`, or `REJECTED`.

`outcomes` has exactly the nine keys `source_binding`, `baseline_provenance`, `candidate_binding`,
`ledger_preserved`, `snapshot_comparison`, `proof_binding`, `no_fetch`, `cleanup`, and
`intent_binding`. Each is `{status, evidence}` with a status in `PASS`, `FAIL`, `UNVERIFIED`,
`UNDECLARED`, or `PENDING`. `PASS` and `FAIL` require a non-empty `evidence` list whose entries each
name an existing artifact `role`; `UNVERIFIED`, `UNDECLARED`, and `PENDING` require an empty
`evidence` list. An outcome can therefore never be `PASS` by absence of evidence.

`case_status` is the derived worst-wins fold over the nine outcomes and must equal the declared
value: `PENDING` if any outcome is `PENDING`; else `FAIL` if any is `FAIL`; else `UNVERIFIED` if any
is `UNVERIFIED` or `UNDECLARED`; else `PASS`. A case with no produced artifacts therefore cannot be
`PASS`. `DeriveCaseStatus` returns this fold; `ValidateCaseManifest` rejects a declared status that
does not match it.

This manifest is a documentary evidence index. It does not run a case, fetch an input, approve the
pinned PostgreSQL route, resolve `DEF-P7-001`, or establish any runtime `PASS`.

## ReplayIsolated behavior lock

`proofbound.p7.behavior-lock.v1` is the authored, content-addressed behavior-lock contract for the
bounded subject `ReplayIsolated`
([`BR-replay-evidence-6e3d2b`](../../../../docs/intent/records/requirements/BR-replay-evidence-6e3d2b/0001.md)
`O-1`). The lock declares the protected observable behavior so a later regeneration can be judged
against the lock rather than the implementation that happened to exist. Materializing the lock and
cross-binding it to supplied, already-committed `vera.replay.v1` proof bytes are non-runtime
operations: the package validates declarations and caller-supplied bytes and never runs a replay,
writes the ledger, or fetches an input.

The input is strict RFC 8785 canonical JSON with exactly these top-level fields: `subject`,
`lock_revision`, `source_prefix`, `candidates`, `baseline`, `candidate_snapshot`, `verdict`,
`sequence_map`, `replay_proof`, and `intent_binding`.

- `subject` is exactly `ReplayIsolated`.
- `lock_revision` has exactly `subject`, `number`, and `parent`; `number` is a positive integer and
  `parent` is either a lowercase 64-hex prior lock digest or the literal `genesis`.
- `source_prefix` has exactly `through_seq`, `events`, and `source_digest`; `events` is a non-empty
  list ordered by ascending `seq` of `{event_id, seq}` with unique ids and seqs, each seq `>= 1` and
  `<= through_seq`; `source_digest` is a lowercase SHA-256.
- `candidates` is a non-empty list ordered by ascending `seq` of `{event_id, seq, bytes_sha256}`
  with unique ids and seqs.
- `baseline` carries the full provenance tuple: exactly `through_seq`, `source_digest`,
  `projector_revision`, `projector_revision_digest`, and `baseline_snapshot_digest`. Its
  `through_seq` and `source_digest` must equal `source_prefix`.
- `candidate_snapshot` has exactly `snapshot_digest`.
- `verdict` is exactly `preserved` or `changed` (the existing twin verdict vocabulary).
- `sequence_map` is a list of `{event_id, requested, effective}` covering exactly the candidate ids,
  ordered by ascending `requested`.
- `replay_proof` has exactly `schema`, `through_seq`, `source_digest`, `candidate_digest`,
  `baseline_snapshot_digest`, and `candidate_snapshot_digest`; `schema` is exactly `vera.replay.v1`.
- `intent_binding` is either the explicit object `{state: "UNDECLARED"}` or the exact P5
  provider-scoped requirement tuple `{record_kind: "requirement", source: "intent.<provider>",
  record_id, artifact_sha256, obligation_id}`.

`EvaluateLock` is a pure evaluator over a validated lock, caller-supplied committed proof bytes, and
an injected `RequirementResolver`. It recomputes the proof digest over the supplied bytes, requires
the parsed proof `schema=vera.replay.v1`, and requires every `through_seq`, `source_digest`,
`candidate_digest`, `baseline_snapshot_digest`, and `candidate_snapshot_digest` to equal the
corresponding lock members, where the candidate digest is the SHA-256 of the newline-joined ordered
candidate `bytes_sha256` values. For a declared `intent_binding`, the resolver must resolve the exact
`record_id`, `artifact_sha256`, and `obligation_id`; an `UNDECLARED` binding is never passed to the
resolver. The overall result is `LOCKED` only when identity, source/candidate, baseline, and proof
all verify and a declared intent resolves exactly; any missing or mismatched member, a baseline
provenance tuple that does not verify, a digest-only proof with no retrievable bytes, an unresolved
declared intent, or an `UNDECLARED` intent yields derived `UNVERIFIED`. A derived `UNVERIFIED` fails
the lock gate and is never converted to `preserved` or `changed`.

A lock revision digest covers only its own canonical bytes; a later revision is a new artifact and
cannot reinterpret an earlier proof or verdict. Synthetic candidates and fixtures may exercise the
contract but are never presented as production observations. This contract is a non-runtime authored
artifact; it does not run a replay, approve the PostgreSQL route, resolve `DEF-P7-001`, or establish
runtime acceptance.

## Interface

```go
const Schema = "proofbound.p7.projector-manifest.v1"
func Parse([]byte) (Manifest, error)
func Validate(Manifest) error
func CanonicalBytes(Manifest) ([]byte, error)
func Digest([]byte) (string, error)
func Bind(BindingInput) (Manifest, error)
const InventorySchema = "proofbound.p7.source-toolchain-inventory.v1"
func ParseInventory([]byte, Manifest) (EvidenceInventory, error)
func ValidateInventory(EvidenceInventory, Manifest) error
func CanonicalInventoryBytes(EvidenceInventory, Manifest) ([]byte, error)
type CommittedEvidenceReader interface {
    ReadCommittedFile(revision, path string) ([]byte, error)
}
func ResolveInventory(EvidenceInventory, Manifest, CommittedEvidenceReader) ([]EvidenceResolution, error)
const EvidenceClosureSchema = "proofbound.p7.evidence-closure.v1"
func EvaluateEvidenceClosure(EvidenceInventory, Manifest, []EvidenceResolution) (EvidenceClosure, error)
const EvidenceManifestSchema = "proofbound.p7.evidence-manifest.v1"
func BuildEvidenceManifest(Manifest, EvidenceInventory, []EvidenceResolution, EvidenceClosure) (EvidenceManifest, error)
func ValidateEvidenceManifest(EvidenceManifest, Manifest, EvidenceInventory, []EvidenceResolution, EvidenceClosure) error
func ParseEvidenceManifest([]byte, Manifest, EvidenceInventory, []EvidenceResolution, EvidenceClosure) (EvidenceManifest, error)
func CanonicalEvidenceManifestBytes(EvidenceManifest, Manifest, EvidenceInventory, []EvidenceResolution, EvidenceClosure) ([]byte, error)
const ManifestDeclarationSchema = "proofbound.p7.projector-manifest-declaration.v1"
func DeclareCommittedManifest(ManifestDeclarationInput, CommittedEvidenceReader) (ManifestDeclaration, error)
func ValidateManifestDeclaration(ManifestDeclaration, ManifestDeclarationInput) error
func ParseManifestDeclaration([]byte, ManifestDeclarationInput) (ManifestDeclaration, error)
func CanonicalManifestDeclarationBytes(ManifestDeclaration, ManifestDeclarationInput) ([]byte, error)
const CaseManifestSchema = "proofbound.p7.case-manifest.v1"
func ParseCaseManifest([]byte) (CaseManifest, error)
func ValidateCaseManifest(CaseManifest) error
func CanonicalCaseManifestBytes(CaseManifest) ([]byte, error)
func CaseManifestDigest([]byte) (string, error)
func DeriveCaseStatus(CaseManifest) (string, error)
const BehaviorLockSchema = "proofbound.p7.behavior-lock.v1"
func ParseBehaviorLock([]byte) (BehaviorLock, error)
func ValidateBehaviorLock(BehaviorLock) error
func CanonicalBehaviorLockBytes(BehaviorLock) ([]byte, error)
func BehaviorLockDigest([]byte) (string, error)
type RequirementResolver interface {
    ResolveRequirement(recordID, artifactSHA256, obligationID string) error
}
func EvaluateLock(BehaviorLock, []byte, RequirementResolver) (LockEvaluation, error)
```

`Bind` accepts caller-supplied bytes only. It sorts source/module identities, computes their
SHA-256 values, and validates the resulting manifest. It performs no filesystem, network,
external-artifact, process, toolchain, executable, or runtime operation.

## Invariants and proving tests

| Invariant | Statement | Proving test |
|---|---|---|
| MAN-INV-1 | Only the exact schema and field set are accepted | manifest_test.go::TestParseRejectsShapeAndCanonicality |
| MAN-INV-2 | Evidence and projector revisions are lowercase 40-hex identities | manifest_test.go::TestDigestAndRevisionBinding |
| MAN-INV-3 | Source files and modules are content-addressed, safe, sorted, and unique | manifest_test.go::TestValidateRejectsBindingOrderAndDigest |
| MAN-INV-4 | Toolchain and module digests are lowercase SHA-256 declarations | manifest_test.go::TestValidateRejectsBindingOrderAndDigest |
| MAN-INV-5 | Equal typed inputs produce equal canonical bytes and digest | manifest_test.go::TestDigestAndRevisionBinding |
| MAN-INV-6 | Non-canonical, missing, unknown, null, and malformed inputs fail closed | manifest_test.go::TestParseRejectsShapeAndCanonicality |
| MAN-INV-7 | Caller-supplied bytes bind deterministically without external operations | manifest_test.go::TestBindComputesByteIdentitiesAndSorts |
| MAN-INV-8 | Duplicate or invalid source/module bindings fail closed | manifest_test.go::TestBindRejectsDuplicateAndInvalidInputs |
| INV-INV-1 | Inventory has exact canonical shape and complete sorted field coverage | manifest_test.go::TestParseInventoryRejectsUnknownAndNonCanonicalShape |
| INV-INV-2 | Every field is BOUND with a committed path and matching revision or explicit UNVERIFIED | manifest_test.go::TestInventoryRejectsMissingExtraDuplicateStaleAndDigestInputs |
| INV-INV-3 | Inventory manifest digest binds the supplied canonical manifest exactly | manifest_test.go::TestInventoryRejectsMissingExtraDuplicateStaleAndDigestInputs |
| RES-INV-1 | BOUND bytes resolve at the declared revision and UNVERIFIED fields are preserved without reader calls | manifest_test.go::TestResolveInventoryVerifiesCommittedBytesAndPreservesUnverified |
| RES-INV-2 | Missing paths, unavailable revisions, unsafe paths, and digest mismatches fail closed | manifest_test.go::TestResolveInventoryRejectsMissingRevisionUnsafeAndDigestEvidence |
| CLOSURE-INV-1 | Exact resolution coverage classifies complete, unverified, and pending fields deterministically | manifest_test.go::TestEvaluateEvidenceClosureClassifiesWithoutClaimingAcceptance |
| CLOSURE-INV-2 | Missing, extra, duplicate, contradictory, stale, and malformed closure inputs fail closed | manifest_test.go::TestEvaluateEvidenceClosureRejectsMissingExtraContradictoryAndStaleInputs |
| SKELETON-INV-1 | Skeleton serialization binds exact manifest/inventory digests and closure-derived fields | manifest_test.go::TestEvidenceManifestBuildsCanonicalSkeleton |
| SKELETON-INV-2 | UNVERIFIED and PENDING states survive deterministic assembly without false completeness | manifest_test.go::TestEvidenceManifestPreservesUnverifiedAndPending |
| SKELETON-INV-3 | Missing, extra, stale, mismatched, contradictory, unknown, and non-canonical inputs fail closed | manifest_test.go::TestEvidenceManifestRejectsMismatchedAndNonCanonicalInputs |
| DECL-INV-1 | Committed source/module bytes bind exact sorted paths and SHA-256 identities at one revision | manifest_test.go::TestDeclareCommittedManifestBindsBytesAndPreservesIncompleteEvidence |
| DECL-INV-2 | Toolchain/archive unavailability and incomplete declaration coverage remain explicit `UNVERIFIED` or `PENDING` | manifest_test.go::TestDeclareCommittedManifestBindsBytesAndPreservesIncompleteEvidence |
| DECL-INV-3 | Declaration serialization is canonical and contradictory declaration evidence fails closed | manifest_test.go::TestManifestDeclarationCanonicalSerializationAndValidation |
| DECL-INV-4 | Invalid, stale, duplicate, unsafe, or missing committed inputs fail closed | manifest_test.go::TestDeclareCommittedManifestRejectsInvalidAndUnavailableInputs |
| CASE-INV-1 | Only the exact case-manifest schema, field set, and `ReplayIsolated` subject are accepted | casemanifest_test.go::TestParseCaseManifestRejectsShapeAndCanonicality |
| CASE-INV-2 | Case id, revisions, paths, and every digest are safe, well-formed, and content-addressed | casemanifest_test.go::TestParseCaseManifestRejectsShapeAndCanonicality |
| CASE-INV-3 | Contract refs, inputs, and outcomes have exactly their required keys; commands/witnesses/artifacts are unique | casemanifest_test.go::TestCaseManifestRejectsKeyAndUniquenessViolations |
| CASE-INV-4 | `PASS`/`FAIL` outcomes require evidence naming an existing artifact role; other states require empty evidence | casemanifest_test.go::TestCaseManifestOutcomeEvidenceRules |
| CASE-INV-5 | `case_status` equals the derived worst-wins fold; absence can never produce `PASS` | casemanifest_test.go::TestDeriveCaseStatusPrecedenceAndNoPassByAbsence |
| CASE-INV-6 | Equal typed inputs produce equal canonical bytes and digest; non-canonical input fails closed | casemanifest_test.go::TestCaseManifestCanonicalRoundTripAndDigest |
| LOCK-INV-1 | Only the exact behavior-lock schema, field set, and `ReplayIsolated` subject are accepted | behaviorlock_test.go::TestParseBehaviorLockRejectsShapeAndCanonicality |
| LOCK-INV-2 | Revision ancestry, ordered unique source events and candidates, and all digests are well-formed | behaviorlock_test.go::TestBehaviorLockRejectsOrderAncestryAndDigestViolations |
| LOCK-INV-3 | Baseline through_seq/source_digest equal source_prefix; sequence map covers exactly the candidate ids | behaviorlock_test.go::TestBehaviorLockRejectsOrderAncestryAndDigestViolations |
| LOCK-INV-4 | `replay_proof` carries `vera.replay.v1` and intent is UNDECLARED or an exact requirement tuple | behaviorlock_test.go::TestParseBehaviorLockRejectsShapeAndCanonicality |
| LOCK-INV-5 | EvaluateLock is LOCKED only when all bindings verify over supplied proof bytes and a declared intent resolves | behaviorlock_test.go::TestEvaluateLockLockedOnlyWhenEveryBindingVerifies |
| LOCK-INV-6 | Missing/mismatched proof members, unverifiable baseline, digest-only proof, unresolved or UNDECLARED intent yield UNVERIFIED | behaviorlock_test.go::TestEvaluateLockDerivesUnverifiedAndFailsClosed |
