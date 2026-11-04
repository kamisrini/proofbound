# VD-p7-nofetch-witness-2026-11-04: no-fetch witness is network-namespace isolation with a labeled declaration fallback

**Status:** Accepted
**Date:** 2026-11-04
**Context:** The P7 B1 unit-3 runtime-fixture spec
([`docs/plans/P7-postgres-b1-unit3-runtime-fixture-spec.md`](../plans/P7-postgres-b1-unit3-runtime-fixture-spec.md))
defines TWIN-INV-19, the no-fetch witness: proof that the real from-source PostgreSQL run performs no
network fetch for binaries. The round-1 adversarial review (finding F4) established that a code-level
assertion — "the `fergusstrange/embedded-postgres` remote-fetch strategy is never constructed" — is a
property of the harness wiring, not an observation that no egress occurred, and left the *mechanism*
as the one genuinely-open owner decision (spec §8.4). The choices put to the owner were: (A) a
no-egress network namespace as the primary witness with the labeled declaration as a fallback where
namespaces are unavailable; (B) a network namespace required with no fallback (the fixture cannot run,
case stays `UNVERIFIED`, where userns/`unshare` is unavailable); (C) the labeled configuration
declaration only, no namespace. See [[VD-p7-postgres-source-route-2026-11-01]] (the B1 route this
unit serves).

**Decision:** Adopt **Option A**: the no-fetch witness is a **no-egress network namespace**
(`unshare -rn` / equivalent) as the **primary witness** wherever the platform supports it; where it
does not, the fixture falls back to the **labeled configuration declaration** (the remote-fetch
strategy is never constructed *and* `BinariesPath` points at the verified local dir), **recorded
honestly as a declaration, not an observation**. Each run records which of the two it used, so a
reader can tell a real no-egress observation from a configuration assertion.

**Consequences:** Makes easy — the strongest witness actually available in a given environment is
used, and the fixture stays runnable in CI/WSL environments that cannot create user namespaces
(portability preserved), so the no-fetch outcome is never silently blocked. Makes hard / cost
accepted — the witness is not uniform: a run under the declaration fallback carries a weaker claim
than one under the namespace, and the evidence must surface that difference rather than paper over it
(the spec requires the labeled distinction). We reject (B) because an all-or-nothing namespace
requirement would make the unit unrunnable on environments this project actually uses, converting a
provenance step into an environment gate; we reject (C) because a configuration assertion alone would
repeat the F4 over-claim (presenting harness wiring as an egress observation). This decision settles
spec §8.4 only; it does not accept any produced fixture, witness, or `case_status` — those remain
gated behind the round-4+ non-author spec re-verify, an owner-approved change intent, and the
package-acceptance bar (Law 9).
