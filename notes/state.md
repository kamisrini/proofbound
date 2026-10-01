# Proofbound — Current State

> THE resume note. Overwrite in place; never append and never create a second state file.
> A fresh session reads CLAUDE.md, then this file, before acting.

As of: 2026-10-01

## Resume — 2026-10-01

P7 Gate 1 is resolved: retain the P6 snapshot-provider skip. The exact next action is to draft and
independently review a bounded behavior-locks/regeneration planning slice using existing feeds.
This is planning only. Do not implement a snapshot provider or introduce a feed, vendor, schema,
event kind, or other P7+ capability.

## Branch and repository status

- Branch: main, tracking origin/main; local main remains ahead because prior push publication was
  blocked by policy.
- Today's changes are documentation-only and the preserved untracked artifacts remain excluded:
  docs/plans/P5-intent-provenance-plan.md (SHA-256
  eacb706918adf23cb90ae74547e1d519b76c613beb042c26425503b9a0358439) and
  kernel/cmd/p6dbprobe/main.go (SHA-256
  f2a747ed588141723ce8f0932aa07814fcca63f9901b7dca0d71119b93ee15a0).

## Completed

- P6 final consolidation remains complete: round-C ACCEPTABLE, 327/327 census rows closed, and
  P1–P5 deep-complete.
- Recorded the founder decision in
  docs/decisions/VD-p7-snapshot-provider-skip-2026-10-01.md.
- Updated the P7 plan and roadmap to show Gate 1 resolved in favor of the skip.
- Regenerated docs/decisions/INDEX.md. No implementation or frozen wire identity changed.

## Open, blocked, and unverified

- Open: draft and independently review the next bounded behavior-locks/regeneration planning
  slice within existing feeds.
- Blocked: snapshot-provider work and all external-feed expansion until a new decision names a
  concrete lawful feed and grants a narrow exception.
- Bare make check passed all repository, generated-freshness, census, package-acceptance, and
  platform stages, then stopped at kernel-check because golangci-lint was unavailable (exit 127).
  The expected index stale; run make index output was an intentional negative-control message.
  No source lint failure was reported.
- Unverified today: kernel build/test/lint and make verify, because the compatible linter binary is
  unavailable. The prior compatible-toolchain result remains historical evidence.
- Native Windows acceptance remains an evidence requirement and is not assumed green.

## Institutionalized improvement

Keep capability boundaries decision-first: record the concrete feed and authorization before
minting a provider SPEC or writing provider code. When no real feed is needed, retain the skip and
continue only within existing feeds. No new automation is justified.

## Standing cautions

- Preserve frozen vera.witness.v1, vera.verdict.v1, vera.replay.v1, pinned vectors, and
  historical artifacts.
- No provider, feed, vendor, schema, event kind, or P7+ capability is authorized by this
  decision.
- P7 planning is allowed only within existing feeds and separately authorized slices.

## Exact next action

Draft and independently review a bounded behavior-locks/regeneration planning slice using existing
feeds. Do not implement a snapshot provider or widen the provider boundary.
