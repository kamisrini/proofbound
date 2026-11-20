# Proofbound v0.3.0 release notes

**Release 3, first public cut: install it, gate pull requests, and hand over signed warranties.**
Tag `v0.3.0` · Decisions [`VD-r3-warranty-first-s59ktv`](../../decisions/VD-r3-warranty-first-s59ktv.md)
and [`VD-vertical-mvp-first-ai3dhn`](../../decisions/VD-vertical-mvp-first-ai3dhn.md) · Scope in
[`plans/RELEASE-PLAN.md`](../../plans/RELEASE-PLAN.md).

**Status:** R3 slice 1 accepted by the owner 2026-11-20. MVP steps 1 to 4 built 2026-11-20, each with
its spec and code reviewed by a non-author (Law 11 budget). Published by the owner's decision of
2026-11-20.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/kamisrini/proofbound/main/install.sh | sh
proofbound version
```

Linux and macOS (amd64, arm64); Windows users download `proofbound_windows_amd64.tar.gz`. The script
checks the archive against `SHA256SUMS` and never uses `sudo`. Quickstart:
[`docs/releases/mvp/QUICKSTART.md`](../mvp/QUICKSTART.md).

## What's new since v0.2.0

| What | Command | Where |
|---|---|---|
| Portable signed warranty: a file anyone can verify against a clone, no database | `keys init`, `warranty export`, `warranty verify` | [R3 demo](DEMO.md) |
| Reviewer proof: VALID only when each named reviewer's own declared key signed (at seal, or by cosigning) | `review seal` (signs), `warranty cosign`; verify exit 3 = UNREVIEWED | [`mvp-3` spec](../../plans/mvp-3-reviewer-signed-spec.md) |
| Release archives and a checksum-verified installer | `install.sh`, `proofbound version` | [`mvp-1` spec](../../plans/mvp-1-install-spec.md) |
| Pull-request check and GitHub Action; `init` writes the workflow | `proofbound check <base>`, `uses: kamisrini/proofbound@main` | [`mvp-2` spec](../../plans/mvp-2-pr-check-spec.md) |
| Release bundle: signed files, a verification page, an index and a badge | `proofbound warranty bundle <base>` | [`mvp-4` spec](../../plans/mvp-4-shareable-proof-spec.md) |

## Changes in behaviour

- `warranty verify` now returns **UNREVIEWED (exit 3)** for a warranty whose reviewer has not signed;
  warranties made before reviewer signatures existed verify UNREVIEWED until the reviewer cosigns.
- Reviewer and owner names are compared by key name: `Alice` and `alice` are one person.
- `proofbound check` lists a mistyped claim (for example `intent: SDD:x`) as UNBOUND instead of ignoring it.

## Known issue

- **The GitHub Action fails on reviewed commits in CI.** A sealed verdict cites its evidence by an
  event id that exists only in the reviewer's local database, so `proofbound sync all` on a fresh CI
  runner stops with `v2 verdict evidence … is dangling` (`DEF-VERDICT-EVIDENCE-PORTABILITY-001`, found
  by the first live Action run, 2026-11-20). `proofbound check` and warranties work locally, and
  `warranty verify` is unaffected (it needs no database). A fix is planned for the next release.

## Caveats

- The repository is the trust anchor: anyone who can commit to it can declare a key. Clone from a
  source you trust and pin the exporter's fingerprint, obtained out of band.
- A reviewer signature proves the key declared for that name signed, not who holds it.
- The embedded database downloads a PostgreSQL build from Maven Central on first use; its published
  signature did not verify when we checked (`DEF-P7-001`). Set `DATABASE_URL` to use your own.
- Release archives are checksummed, not signed. Pin the action to a tag once you rely on it.
- Only OpenSpec and GitHub Spec Kit are supported.
