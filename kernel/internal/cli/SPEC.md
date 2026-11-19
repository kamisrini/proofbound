# CLI composition contract

## 1. Boundary

This package is Proofbound's reusable command composition root. It parses commands, opens the
ledger, constructs connectors and projections, and renders command results. Executable packages
under `cmd/` are deliberately tiny; `cmd/proofbound` is the only live entry point.

The package owns no semantic event contract. Frozen wire identities remain owned by their connector
or replay SPECs.

P5 adds `proofbound sync intent {records|specdir|all}`. Its committed-tree reader uses only
`git ls-tree` and `git show` at an explicit tree. `sync all` ingests both intent providers before
Git so an explicit commit claim cannot outrun the exact record revision it names.
It also routes `proofbound report intent <id>`, `proofbound report requirement <id>`, and
`proofbound intent check --commit <sha>` to the proof-bearing projection API without interpreting
chain semantics in the CLI.

The R2 starter kit (`docs/plans/R2-starter-kit-spec.md`) adds `init openspec`, which writes a tenant's
profile, agent instructions, reviewer subagent and commit hook, and `review {draft|seal}`, which turn a
reviewer's decisions on one commit into review and verdict artifacts plus that commit's readiness gate
(CLI-INV-18..20).

R2 M3 (`CI-sdd-warranty-m3-xaqocp`, `VD-sdd-m3-git-binding-2026-11-19`) adds the `sdd` arm of the
commit-claim resolver used by `sync git`, `sync all` and `verify`. The `sdd` provider and the ledger
set `L` (every ledgered `intent.sdd` change intent's id and digest) are built lazily, on the first
`sdd` claim. An `sdd` revision is accepted only if its source, kind, CI-grammar id, 64-hex digest and
artifact-path last segment (`<name>` or `YYYY-MM-DD-<name>`) all check, else the sync fails. It binds
only if it is in `L`. Every tenant outcome is a `*ClaimUnboundError` (`unknown-change`,
`archived-earlier`, `no-change-intent`, `unmappable:<cause>`, `not-ledgered`). `sync git` and
`sync all` print `claim-skip <cause> <value> at <commit>` after their counts line, the value through
`printableUnit`; `verify` prints none. The reviews connector receives a store-backed `LedgerView`
(one empty-filter ledger pass, lazily, decoding git payloads as `connectorgit.Commit`); `sync reviews`
and `sync all` print `held <cause> <path>` and `refused <cause> <path>` lines after their counts line,
before returning any reviews error, and a refusal fails the command. The contract is the M3 SPEC
(`docs/plans/R2-sdd-m3-git-binding-spec-draft5.md`, as amended by drafts 6-8).

`proofbound twin replay [--through <seq>] [--candidate <file>]` runs `twin.ReplayIsolated` in a
throwaway embedded store, never mutating the source ledger. The baseline is the isolated projection
of the committed prefix through the same `--through` seq (default: the latest committed seq), so the
verdict reflects only the candidate's effect — never the events a lower `--through` drops. With no
candidate it re-derives and proves that prefix state (verdict `preserved` plus a `vera.replay.v1`
proof). `--candidate <file>` injects exactly one counterfactual event, supplied as canonical JSON
`{"seq": <int>, "event": {<core.Event>}}` with no trailing data, and the command reports `preserved`
or `changed`. The CLI interprets no replay or proof semantics itself — the `twin` SPEC owns them.

## 2. Interface lock

```go
func Run(ctx context.Context, program string, args []string, stdout, stderr io.Writer) int
```

`program` is retained for composition compatibility but does not select an identity. Unknown command forms return status 2 and
print usage for the live name. Runtime failures return status 1 and are prefixed with `proofbound:`.

## 3. Identity and compatibility

- Local state is rooted only at `<repository>/.proofbound`. The one-time migration is an explicit,
  fail-safe directory move documented in `docs/proofbound-identity-migration.md`; commands never
  merge two state directories or silently fall back to the old path.
- Live environment names use only `PROOFBOUND_*`; legacy names are ignored after their one-phase
  P5 window.
- `cmd/proofbound` is the only executable wrapper.
- `DATABASE_URL`, `GITHUB_TOKEN`, and standard tool environment names are not product-namespaced and
  are unchanged.
- Removal is governed by the dated advisory in `docs/gates.md`, not by an undocumented future edit.

## 4. Non-goals

- No interpretation or rewriting of frozen `vera.witness.v1`, `vera.verdict.v1`, or
  `vera.replay.v1` bytes.
- No automatic merge or deletion of local state directories.
- No compatibility promise beyond the single ratified phase.

## 5. Invariants and proving tests

| Invariant | Statement | Proving test |
|---|---|---|
| CLI-INV-1 | The live executable and all usage/runtime diagnostics name Proofbound | cli_test.go::TestRunUsesProofboundIdentity |
| CLI-INV-2 | Every command opens state and witness spool paths below `.proofbound` | cli_test.go::TestOpenStoreUsesProofboundState |
| CLI-INV-3 | Live environment values are read only from the Proofbound namespace | cli_test.go::TestProductEnvPrefersProofbound |
| CLI-INV-4 | Legacy environment names are ignored after P5 | cli_test.go::TestLegacyEnvironmentIsIgnored |
| CLI-INV-5 | The legacy executable wrapper is absent | cli_test.go::TestLegacyCommandWrapperIsAbsent |
| CLI-INV-6 | Frozen witness bytes remain accepted without reinterpretation | sync_integration_test.go::TestSyncChecksRebuildAndVerify |
| CLI-INV-7 | Intent commands select only records, specdir, or both and read committed trees | cli_test.go::TestParseIntentCommandsAndCommittedReader |
| CLI-INV-8 | Full sync orders all intent providers before commit ingestion | cli_test.go::TestSyncAllOrdersIntentBeforeGit |
| CLI-INV-9 | Intent report and check identifiers are passed unchanged to projections | cli_test.go::TestParseIntentReportAndCheckCommands |
| CLI-INV-10 | Commands outside a repository fail before opening product state | cli_test.go::TestRunStopsWhenRepositoryRootIsMissing |
| CLI-INV-11 | `twin replay` evaluates ReplayIsolated in a throwaway store against the same-prefix projection and never mutates the source ledger | twin_replay_test.go::TestTwinReplayIsolatesAndProves |
| CLI-INV-12 | A `--candidate` file injects exactly one counterfactual event and yields a verdict with a vera.replay.v1 proof whose candidate snapshot reflects the change | twin_replay_test.go::TestTwinReplayCounterfactualVerdict |
| CLI-INV-13 | An `sdd` claim binds only an accepted revision that is in `L`; the resolver's anti-spoofing checks and skip-cause mapping are exact. Also proven end to end (claim-skip line, rebuild, binding) by the sdd package's `TestM3ClaimBindsOnlyALedgeredRevision` | sdd_claims_test.go::TestSDDResolverChecksAndCauses |
| CLI-INV-14 | A history without an `sdd` value neither constructs the `sdd` provider nor reads `L`; one or more `sdd` values do each once per sync | sdd_claims_test.go::TestSDDClaimWorkIsLazy |
| CLI-INV-15 | The store-backed reviews `LedgerView` sees every provider's commit claims and intent requirements (imported historical events included) and every reviews identity by payload schema, and reads the ledger lazily, at most once per sync and never for a sync without a v2 or requirement-review artifact. Also proven by TestStoreLedgerViewReadsLazilyOnce | ledger_view_test.go::TestStoreLedgerViewSeesEveryProvider |
| CLI-INV-16 | `sync reviews` and `sync all` print their counts line, then held and refused lines, before returning any reviews error; holds alone exit 0 and a refusal fails the command | ledger_view_test.go::TestSyncReviewsPrintsHoldsAndFailsOnRefusal |
| CLI-INV-17 | `report intent sdd:<change-name>` reports the change intent that name resolves to at HEAD through the sdd provider's `Resolve` (the same resolution commit claims bind through); a name that does not resolve fails the command with the resolver's error and reports nothing; any other argument is an intent id, unchanged | report_intent_test.go::TestReportIntentResolvesAnSDDChangeName |
| CLI-INV-18 | `init openspec [--owner <name>]` writes the starter kit (`.proofbound/sdd-profile.json`, two `.gitignore` lines, the AGENTS.md block, `.claude/agents/proofbound-reviewer.md`, the `commit-msg` hook) without overwriting any existing file or content, and a second run changes nothing | starter_test.go::TestInitOpenSpecNeverOverwritesAndIsIdempotent |
| CLI-INV-19 | `review seal <draft-dir>` writes requirement-review and obligation-verdict artifacts that the reviews connector's own parsers accept, with status ACCEPTABLE iff every outcome is SATISFIED, plus the reviewed commit's own readiness gate `gates/ready-<commit12>.yaml`. Also proven end to end (draft, decide, seal, commit, sync: PASS when all SATISFIED/VERIFIABLE by a non-owner, BLOCKED otherwise) by the sdd package's `TestStarterKitReviewFlow` | starter_test.go::TestReviewSealWritesAcceptedArtifacts |
| CLI-INV-20 | `review seal` takes only the reviewer's decisions (name, outcomes, findings) from the editable draft and rebuilds every reference, digest, owner, evidence id and obligation set from the ledger; it writes nothing when any check fails: a draft that does not match the ledger (an edited digest, a missing or extra obligation), an undecided or unknown outcome, a blank reviewer, the ledger's declared owner as reviewer, an artifact the parser rejects, an existing target file, or a write error (written files are removed) | starter_test.go::TestReviewSealWritesNothingOnAnyFailure |
