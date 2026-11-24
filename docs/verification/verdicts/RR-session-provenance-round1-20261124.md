---
{"schema":"proofbound.requirement-review.v1","review_id":"RR-session-provenance-round1-20261124","requirement":{"record_kind":"requirement","source":"intent.records","record_id":"BR-session-provenance-p3v8nc","artifact_sha256":"e1e00c458cf0e025bef85f367e3ab37d15b8dc9350d1ca35987e90cb7a85f220","relation":"reviews"},"declared_reviewer":"agent-requirement-reviewer-20261124","outcomes":[{"obligation_id":"O-1","outcome":"AMBIGUOUS","finding":"'too common to count' has no threshold in the statement. Proposed rewording: 'or common (three or fewer non-space characters, or repeated within the commit)'. Hashing into a local session log, per-line matching, the reviewer-signed PROV record and C12 never blocking are testable."}],"artifact_path":"docs/verification/verdicts/RR-session-provenance-round1-20261124.md","artifact_sha256":"b09d6a30f8c1069000cf6ab3d6c5cf7910a90bc26d2a77a90ac6587f8ee63e51"}
---

# Independent requirement review: session-to-line provenance, round 1

Fresh non-author reviewer context: `agent-requirement-reviewer-20261124`. Reviewed commit: `8d1c786238809ae99776b5f380884d10e39edea3`.

This review binds the exact `BR-session-provenance-p3v8nc/0001` revision. The declared reviewer is distinct from the requirement's declared owner, `proofbound-maintainer`.

`VERIFIABLE` means the obligation's words have a conceivable independent test and evidence route. It does not judge business merit or accept any implementation.

- `O-1` AMBIGUOUS: 'too common to count' has no threshold in the statement. Proposed rewording: 'or common (three or fewer non-space characters, or repeated within the commit)'. Hashing into a local session log, per-line matching, the reviewer-signed PROV record and C12 never blocking are testable.
