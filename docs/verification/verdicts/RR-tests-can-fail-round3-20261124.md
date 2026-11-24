---
{"schema":"proofbound.requirement-review.v1","review_id":"RR-tests-can-fail-round3-20261124","requirement":{"record_kind":"requirement","source":"intent.records","record_id":"BR-tests-can-fail-k2h06j","artifact_sha256":"cf24d3089ba83a06414c4561d07ba50b92641a5cc9bd5ad1c64e3419dc661820","relation":"reviews"},"declared_reviewer":"agent-requirement-reviewer-round3-20261124","outcomes":[{"obligation_id":"O-3","outcome":"AMBIGUOUS","finding":"The exclusion is fixed, but each changed hunk has two readings: git's default -U3 merges nearby changes into one hunk, while tests.go and area.go use -U0. Testers would disagree on the hunk count and the decisions seal requires. Proposed rewording: undo each changed hunk, as git diff -U0 against the commit's first parent lists it (the same line ranges C10 lists), of a claiming commit outside ..."}],"artifact_path":"docs/verification/verdicts/RR-tests-can-fail-round3-20261124.md","artifact_sha256":"d31f23cdd33f123654a2810a69d23216aa7c679b51ce9ec1354580804eec0aac"}
---

# Independent requirement review: tests that can fail, round 3

Fresh non-author reviewer context: `agent-requirement-reviewer-round3-20261124`. Reviewed commit: `74e354de34805154a28cc60598ef14ca10ec1b4c`.

This review binds the exact `BR-tests-can-fail-k2h06j/0003` revision. The declared reviewer is distinct from the requirement's declared owner, `proofbound-maintainer`.

`VERIFIABLE` means the obligation's words have a conceivable independent test and evidence route. It does not judge business merit or accept any implementation.

- `O-3` AMBIGUOUS: The exclusion is fixed, but each changed hunk has two readings: git's default -U3 merges nearby changes into one hunk, while tests.go and area.go use -U0. Testers would disagree on the hunk count and the decisions seal requires. Proposed rewording: undo each changed hunk, as git diff -U0 against the commit's first parent lists it (the same line ranges C10 lists), of a claiming commit outside ...
