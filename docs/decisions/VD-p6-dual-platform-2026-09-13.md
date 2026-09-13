# VD-p6-dual-platform-2026-09-13: retain Linux and native Windows execution

**Status:** Accepted

**Date:** 2026-09-13

## Authority and provenance

This decision supplements the accepted P6 consolidation authority
[`VD-p6-consolidation-2026-09-12`](VD-p6-consolidation-2026-09-12.md), which adopts the ratified
P6 draft-2 plan and cites the non-author plan adjudication
[`p6-consolidation-plan-round1.md`](../verification/verdicts/p6-consolidation-plan-round1.md).
The founder's declared-authority platform receipt is
[`p6-dual-platform-ratification.md`](../verification/verdicts/p6-dual-platform-ratification.md).

The receipt is not cryptographically authenticated and does not itself constitute a Windows test
result. Acceptance remains subject to the mechanical C8-002 evidence contract.

## Decision

P6 retains both Linux and native Windows as supported execution paths for the existing Proofbound
tooling. Windows uses PowerShell for setup and entry, with Git Bash or MSYS2 supplying Bash and GNU
Make compatibility for the repository's existing scripts. WSL is not native Windows evidence.

This is a platform-validation decision, not a product-capability expansion: no new provider, event
kind, product behavior, or P7+ vision capability is authorized. C8-002 remains open until a real
native Windows run passes the repository gate and its result is committed as evidence.

## Consequence

The Task 3 reproducibility SPEC now names the dual-platform prerequisites and acceptance commands.
The prior Windows `make`-missing result remains red historical evidence. The Linux clean-clone
historical-evidence portability decision remains separate and unresolved.
