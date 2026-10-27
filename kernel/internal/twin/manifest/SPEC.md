# SPEC — `internal/twin/manifest`

**Status:** P7-NONRUNTIME-PROJECTOR-MANIFEST-DECLARATION-1 authorized non-runtime preparation slice · 2026-10-27
**Authority:** [`CI-p7-projector-manifest-declaration-1a2b3c`](../../../../docs/intent/records/change-intents/CI-p7-projector-manifest-declaration-1a2b3c/0001.md)

## Purpose and boundary

This package validates the documentary `proofbound.p7.projector-manifest.v1` contract and its
`proofbound.p7.source-toolchain-inventory.v1` evidence inventory. It binds
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
| DECL-INV-3 | Declaration serialization is canonical and invalid, stale, duplicate, unsafe, missing, or contradictory inputs fail closed | manifest_test.go::TestManifestDeclarationCanonicalSerializationAndValidation; manifest_test.go::TestDeclareCommittedManifestRejectsInvalidAndUnavailableInputs |
