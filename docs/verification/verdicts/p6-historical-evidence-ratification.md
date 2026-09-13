# Founder ratification — P6 historical-evidence portability

**Date:** 2026-09-13

**Founder statement on receipt:** “ratify option 1”

**Applies to:** P6 Task 3 C8-001/C8-002 historical evidence portability

This is a declared-authority record. The founder statement ratifies the immediately preceding
recommended option: preserve the exact historical event envelopes cited by the accepted P5
obligation verdict in a strict migration-only archive so a fresh clone can reconstitute the
evidence without relaxing projection integrity.

## Ratified decision

1. The archive is limited to the exact envelopes for
   `01M28TPW9C8R7ND19MNDCJ9GDG` and `01M29HMPE5V977AR3VMW47DVDE`, including their original ledger
   sequence numbers and event fields.
2. The archive is migration evidence, not a connector source, ordinary sync input, new event kind,
   or replacement for the accepted P5 artifact.
3. Import is explicit, opt-in, exact-ID-bound, append-only, and available only for a fresh local
   migration before verification. Normal sync, projection, and report paths remain unchanged and
   dangling references continue to fail closed.
4. The archive must be validated before import and must reject additions, omissions, altered
   envelopes, duplicate IDs, sequence changes, and non-empty target ledgers.

No new provider, platform promise, product capability, or P7+ vision capability is authorized.
