---
{"schema":"proofbound.requirement-review.v1","review_id":"RR-write-perimeter-round1-20261124","requirement":{"record_kind":"requirement","source":"intent.records","record_id":"BR-write-perimeter-i1xmq8","artifact_sha256":"57fd7f9b9a1a523d2e71b9512cefe3c89fc84765e4e4035df7e46c7be2bfc924","relation":"reviews"},"declared_reviewer":"agent-requirement-reviewer-20261124","outcomes":[{"obligation_id":"O-1","outcome":"AMBIGUOUS","finding":"'anything the hook cannot judge' is open-ended, 'a reason the agent can act on' is subjective, and 'a task is active' is undefined. Proposed rewording: while a task started by proofbound task start is active, the write is refused with a reason naming the file and the declared area; the hook allows it when there is no repository, no active task, no declared area, or unparseable input."}],"artifact_path":"docs/verification/verdicts/RR-write-perimeter-round1-20261124.md","artifact_sha256":"53c9d68e9f454f05540e61d9eb08066e630393f8a3e4f3b156c877e6b8b328ec"}
---

# Independent requirement review: write-time perimeter, round 1

Fresh non-author reviewer context: `agent-requirement-reviewer-20261124`. Reviewed commit: `8d1c786238809ae99776b5f380884d10e39edea3`.

This review binds the exact `BR-write-perimeter-i1xmq8/0001` revision. The declared reviewer is distinct from the requirement's declared owner, `proofbound-maintainer`.

`VERIFIABLE` means the obligation's words have a conceivable independent test and evidence route. It does not judge business merit or accept any implementation.

- `O-1` AMBIGUOUS: 'anything the hook cannot judge' is open-ended, 'a reason the agent can act on' is subjective, and 'a task is active' is undefined. Proposed rewording: while a task started by proofbound task start is active, the write is refused with a reason naming the file and the declared area; the hook allows it when there is no repository, no active task, no declared area, or unparseable input.
