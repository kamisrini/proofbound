---
{"schema":"proofbound.requirement-review.v1","review_id":"RR-warranty-head-reviews-round1-20261124","requirement":{"record_kind":"requirement","source":"intent.records","record_id":"BR-warranty-head-reviews-gdjrrv","artifact_sha256":"475a89a9e32a57759baadb18c14f8cb04b90ca8e0da27f649b02ef7587de7cf0","relation":"reviews"},"declared_reviewer":"agent-requirement-reviewer-20261124","outcomes":[{"obligation_id":"O-1","outcome":"VERIFIABLE","finding":"Testable with a git fixture: seal a review on a branch, abandon it, export a later commit on main: the export must succeed and not cite that review; removing the HEAD review of a claimed requirement must make the export fail."}],"artifact_path":"docs/verification/verdicts/RR-warranty-head-reviews-round1-20261124.md","artifact_sha256":"3200190186a6fc9fad6b48b44c1aa2cbac9fc5bfd5e91f799932692b6f6f2d43"}
---

# Independent requirement review: a warranty cites only reviews committed at head, round 1

Fresh non-author reviewer context: `agent-requirement-reviewer-20261124`. Reviewed commit: `8d1c786238809ae99776b5f380884d10e39edea3`.

This review binds the exact `BR-warranty-head-reviews-gdjrrv/0001` revision. The declared reviewer is distinct from the requirement's declared owner, `proofbound-maintainer`.

`VERIFIABLE` means the obligation's words have a conceivable independent test and evidence route. It does not judge business merit or accept any implementation.

- `O-1` VERIFIABLE: Testable with a git fixture: seal a review on a branch, abandon it, export a later commit on main: the export must succeed and not cite that review; removing the HEAD review of a claimed requirement must make the export fail.
