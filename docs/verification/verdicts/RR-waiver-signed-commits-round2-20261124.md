---
{"schema":"proofbound.requirement-review.v1","review_id":"RR-waiver-signed-commits-round2-20261124","requirement":{"record_kind":"requirement","source":"intent.records","record_id":"BR-waiver-signed-commits-6fioa7","artifact_sha256":"5690adb20e93ff0b0f6afab693ba80c9979e4b9edda09360b569c8070e7e0dde","relation":"reviews"},"declared_reviewer":"agent-requirement-reviewer-round2-20261124","outcomes":[{"obligation_id":"O-2","outcome":"VERIFIABLE","finding":"Testable with fixtures: a policy file at the range base only, in the first parent only, and malformed (which must require signed commits), and an unsigned commit whose check and commit-hook output both say why it was not waived."}],"artifact_path":"docs/verification/verdicts/RR-waiver-signed-commits-round2-20261124.md","artifact_sha256":"173fd328daa870a1e31369078e937588dba7e79bdf14410ea86760151f966a6b"}
---

# Independent requirement review: waivers only for signed commits, round 2

Fresh non-author reviewer context: `agent-requirement-reviewer-round2-20261124`. Reviewed commit: `6637847fd76b61d8299d0d0875e341276b7e99ec`.

This review binds the exact `BR-waiver-signed-commits-6fioa7/0002` revision. The declared reviewer is distinct from the requirement's declared owner, `proofbound-maintainer`.

`VERIFIABLE` means the obligation's words have a conceivable independent test and evidence route. It does not judge business merit or accept any implementation.

- `O-2` VERIFIABLE: Testable with fixtures: a policy file at the range base only, in the first parent only, and malformed (which must require signed commits), and an unsigned commit whose check and commit-hook output both say why it was not waived.
