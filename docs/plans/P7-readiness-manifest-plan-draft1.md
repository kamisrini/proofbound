# P7 readiness manifest — planning draft 1

**Status:** OWNER-APPROVED FOR PLANNING — non-author review required before implementation;
no runtime implementation, fixture corpus, binary validation, or acceptance claim is authorized.

**Date:** 2026-10-17

**Purpose:** define the evidence index needed before a separately authorized P7 implementation
slice can be reviewed. The planning shape is settled by
[`VD-p7-readiness-manifest-2026-10-17`](../decisions/VD-p7-readiness-manifest-2026-10-17.md).
This document organizes existing P7 requirements; it does not replace the
[ReplayIsolated lock contract](P7-replay-lock-spec-addendum-draft1.md), the
[fixture matrix](P7-replay-lock-fixture-matrix-draft1.md), or the
[external-prerequisites amendment](P7-replay-lock-external-prerequisites-amendment-draft1.md).

## Boundary

The case manifest is an evidence index for one bounded P7 case. It is not:

- the behavior-lock artifact;
- the frozen `vera.replay.v1` proof;
- a new ledger event, wire identity, or writable truth;
- a replacement for the setup-source manifest;
- approval of the pinned PostgreSQL input; or
- authorization to implement `ReplayIsolated`, regeneration, or any provider.

The manifest may be designed now because its shape is independent of the deferred third-party
binary provenance/licensing go-no-go. Any case that needs the PostgreSQL input remains
`UNVERIFIED` until that prerequisite is explicitly activated and cleared.

## Settled planning shape

The following is the owner-approved planning shape. It still requires fresh non-author review
before implementation. The emitted manifest uses UTF-8 RFC 8785 canonical JSON with exactly the
listed top-level fields, no unknown/missing/null members, and explicit pending status objects.

```json
{
  "schema": "proofbound.p7.case-manifest.v1",
  "case_id": "<stable-case-id>",
  "subject": "ReplayIsolated",
  "case_status": "UNVERIFIED",
  "evidence_revision": "<full-git-commit-id>",
  "contract_refs": {
    "lock_addendum": {"path": "<path>", "sha256": "<digest>"},
    "fixture_matrix": {"path": "<path>", "sha256": "<digest>"},
    "implementation_scope": {"path": "<path>", "sha256": "<digest>"}
  },
  "inputs": {
    "lock": {"path": "<path>", "sha256": "<digest>"},
    "source": {"path": "<path>", "sha256": "<digest>"},
    "candidate": {"path": "<path>", "sha256": "<digest>"},
    "setup_sources": {"path": "<path>", "sha256": "<digest>"},
    "projector_manifest": {"path": "<path>", "sha256": "<digest>"}
  },
  "commands": [
    {
      "id": "<stable-command-id>",
      "argv": ["<exact-command>"],
      "cwd": "<repository-relative-directory>",
      "network_policy": "<exact-policy>",
      "environment_digest": "<digest>",
      "exit_code": 0,
      "stdout": {"path": "<path>", "sha256": "<digest>"},
      "stderr": {"path": "<path>", "sha256": "<digest>"}
    }
  ],
  "witnesses": [
    {"kind": "<ledger-before>", "path": "<path>", "sha256": "<digest>"},
    {"kind": "<ledger-after>", "path": "<path>", "sha256": "<digest>"},
    {"kind": "<network-observation>", "path": "<path>", "sha256": "<digest>"},
    {"kind": "<cleanup-observation>", "path": "<path>", "sha256": "<digest>"}
  ],
  "outcomes": {
    "source_binding": {"status": "UNVERIFIED", "evidence": []},
    "baseline_provenance": {"status": "UNVERIFIED", "evidence": []},
    "candidate_binding": {"status": "UNVERIFIED", "evidence": []},
    "ledger_preserved": {"status": "UNVERIFIED", "evidence": []},
    "snapshot_comparison": {"status": "UNVERIFIED", "evidence": []},
    "proof_binding": {"status": "UNVERIFIED", "evidence": []},
    "no_fetch": {"status": "UNVERIFIED", "evidence": []},
    "cleanup": {"status": "UNVERIFIED", "evidence": []},
    "intent_binding": {"status": "UNVERIFIED", "evidence": []}
  },
  "artifacts": [
    {"role": "<evidence-role>", "path": "<path>", "size_bytes": 1, "sha256": "<digest>"}
  ],
  "review": {
    "author_commit": "<full-git-commit-id>",
    "independent_verdict": "PENDING",
    "independent_verdict": {"status": "PENDING"}
  }
}
```

## Field rules before implementation

| Area | Required decision or evidence | Current state |
|---|---|---|
| Identity | Stable case ID, exact subject, evidence revision, and canonical bytes | Settled for planning; non-author review pending |
| Contract binding | Exact digests for the addendum, fixture matrix, and authorized implementation scope | Shape settled; no case binding exists |
| Lock/source/candidate | Content-addressed paths and bytes for the exact case inputs | No runtime corpus exists |
| Setup inputs | Setup-source manifest and projector/build-manifest relationship | Documentary setup accepted; provenance/licensing remains deferred |
| Commands | Exact argv, working directory, relevant environment, network policy, exit code, and output digests | No case run exists |
| Ledger witness | Complete before/after ledger bytes, including records beyond `ThroughSeq` | Required by the accepted matrix; not run |
| Network witness | Observable zero-fetch evidence, not merely a blocked request | Required; not run |
| Cleanup witness | Temporary store and staging removal on success and every failure path | Required; not run |
| Outcome vocabulary | `PASS`, `FAIL`, `UNVERIFIED`, `UNDECLARED`, or `PENDING` with key-specific rules | Settled for planning; no outcomes exist |
| Review | Non-author verdict against immutable manifest and artifact digests; pending uses an explicit status object | No implementation verdict exists |

## Readiness checklist

This checklist is a planning gate, not a claim that the items are complete.

1. Obtain fresh non-author review of the canonical manifest encoding and exact member set.
2. Bind the case to one authorized implementation scope and one evidence revision.
3. Resolve the deferred binary provenance/licensing item before any case requiring that input is
   run or labeled beyond `UNVERIFIED`.
4. Separately review the live `twin` API/SPEC amendment and implementation route.
5. Create the fixture corpus and content-addressed evidence files.
6. Run the case under its declared network and cleanup witnesses.
7. Record all outcomes, including negative cases and expected failures.
8. Obtain a fresh non-author implementation verdict.

Until those gates are separately authorized and satisfied, this draft remains planning material.
The owner may leave any prerequisite dormant; the deferral register governs that decision.

## Explicit non-goals

This draft does not select or validate a PostgreSQL binary, resolve licensing, download runtime
inputs, add a case-manifest parser, alter the live twin API, add fixtures, implement behavior
locks, implement regeneration, or reopen any dormant deferral.
