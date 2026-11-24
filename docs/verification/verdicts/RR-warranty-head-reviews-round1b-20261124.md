---
{"schema":"proofbound.requirement-review.v1","review_id":"RR-warranty-head-reviews-round1b-20261124","requirement":{"record_kind":"requirement","source":"intent.records","record_id":"BR-warranty-head-reviews-gdjrrv","artifact_sha256":"002b84ec84216e50c44eb60ef17b3996bbb5ba25582c75851e091e1d8081c35b","relation":"reviews"},"declared_reviewer":"agent-requirement-reviewer-20261124","outcomes":[{"obligation_id":"O-1","outcome":"VERIFIABLE","finding":"statement unchanged from 0001; Testable with a git fixture: seal a review on a branch, abandon it, export a later commit on main: the export must succeed and not cite that review; removing the HEAD review of a claimed requirement must make the export fail."}],"artifact_path":"docs/verification/verdicts/RR-warranty-head-reviews-round1b-20261124.md","artifact_sha256":"35bbc73c776dd858c13f0c03c45fcb5d423987a4252606e6c386b7f94920ac09"}
---

# Independent requirement review: a warranty cites only reviews committed at head, round 1b

Reviewed commit: `8d1c786238809ae99776b5f380884d10e39edea3`.

This review binds the active revision `BR-warranty-head-reviews-gdjrrv/0002`, whose obligations are identical (ids and statements) to `0001`, reviewed in `RR-warranty-head-reviews-round1-20261124`. Declared reviewer: `agent-requirement-reviewer-20261124`.

- `O-1` VERIFIABLE: statement unchanged from 0001; Testable with a git fixture: seal a review on a branch, abandon it, export a later commit on main: the export must succeed and not cite that review; removing the HEAD review of a claimed requirement must make the export fail.
