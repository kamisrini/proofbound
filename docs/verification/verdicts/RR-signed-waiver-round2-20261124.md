---
{"schema":"proofbound.requirement-review.v1","review_id":"RR-signed-waiver-round2-20261124","requirement":{"record_kind":"requirement","source":"intent.records","record_id":"BR-signed-waiver-15j8f9","artifact_sha256":"43d14bf9bfa8d60d167125e1c4d51a4fb1a95f5327915a9253232a8e6af48df0","relation":"reviews"},"declared_reviewer":"agent-requirement-reviewer-round2-20261124","outcomes":[{"obligation_id":"O-1","outcome":"VERIFIABLE","finding":"Confirmed from round 1: an expiry more than 90 days ahead and an overwrite of an existing waiver must be refused, and the created file must verify against a declared key."},{"obligation_id":"O-3","outcome":"VERIFIABLE","finding":"Confirmed from round 1: signer, waiver, expiry and reason must appear on each named surface, and a waiver change must be listed in the Audit room."},{"obligation_id":"O-4","outcome":"VERIFIABLE","finding":"The round 1 date ambiguity is resolved: the expiry must be on or after the latest of the commit's committer day, the range base's committer day and the waiver's signing day (UTC), with keys declared at both the base and the parent; each condition, including independence from author, agent and signer, has a pass and a fail fixture."}],"artifact_path":"docs/verification/verdicts/RR-signed-waiver-round2-20261124.md","artifact_sha256":"a2f382045ea480a10f66fdcd495e680f63b8f0aa522066348e352370abaedebc"}
---

# Independent requirement review: signed exception waivers, round 2

Fresh non-author reviewer context: `agent-requirement-reviewer-round2-20261124`. Reviewed commit: `6637847fd76b61d8299d0d0875e341276b7e99ec`.

This review binds the exact `BR-signed-waiver-15j8f9/0002` revision. The declared reviewer is distinct from the requirement's declared owner, `proofbound-maintainer`.

`VERIFIABLE` means the obligation's words have a conceivable independent test and evidence route. It does not judge business merit or accept any implementation.

- `O-1` VERIFIABLE: Confirmed from round 1: an expiry more than 90 days ahead and an overwrite of an existing waiver must be refused, and the created file must verify against a declared key.
- `O-3` VERIFIABLE: Confirmed from round 1: signer, waiver, expiry and reason must appear on each named surface, and a waiver change must be listed in the Audit room.
- `O-4` VERIFIABLE: The round 1 date ambiguity is resolved: the expiry must be on or after the latest of the commit's committer day, the range base's committer day and the waiver's signing day (UTC), with keys declared at both the base and the parent; each condition, including independence from author, agent and signer, has a pass and a fail fixture.
