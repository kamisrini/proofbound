---
{"schema":"proofbound.requirement-review.v1","review_id":"RR-signed-waiver-round1-20261124","requirement":{"record_kind":"requirement","source":"intent.records","record_id":"BR-signed-waiver-15j8f9","artifact_sha256":"a85efe1f1307ef49464e043808232acae8910d37478db604d180808dfd758164","relation":"reviews"},"declared_reviewer":"agent-requirement-reviewer-20261124","outcomes":[{"obligation_id":"O-1","outcome":"VERIFIABLE","finding":"Testable: creating a waiver with an expiry more than 90 days ahead must fail, as must creating one over an existing waiver file; the created file must verify against a declared key. Minor: '90 days ahead' is implicitly from the creation date."},{"obligation_id":"O-2","outcome":"AMBIGUOUS","finding":"'had not expired by its date' does not say which commit date: author and committer dates can straddle the expiry. The spec uses the UTC committer date. Proposed rewording: 'had an expiry on or after the UTC date of the commit's committer date'. The rest is testable with fixtures."},{"obligation_id":"O-3","outcome":"VERIFIABLE","finding":"Testable: assert a waived commit appears with signer, waiver, expiry and reason on each named surface (PR check, bundle index, verification page, badge, Audit room), and that a waiver change is listed in the Audit room."}],"artifact_path":"docs/verification/verdicts/RR-signed-waiver-round1-20261124.md","artifact_sha256":"380a970cbffe3e7e9ed0bebe43ba29bf74a66a5ac61903cc30e7f1c41c4f194a"}
---

# Independent requirement review: signed exception waivers, round 1

Fresh non-author reviewer context: `agent-requirement-reviewer-20261124`. Reviewed commit: `8d1c786238809ae99776b5f380884d10e39edea3`.

This review binds the exact `BR-signed-waiver-15j8f9/0001` revision. The declared reviewer is distinct from the requirement's declared owner, `proofbound-maintainer`.

`VERIFIABLE` means the obligation's words have a conceivable independent test and evidence route. It does not judge business merit or accept any implementation.

- `O-1` VERIFIABLE: Testable: creating a waiver with an expiry more than 90 days ahead must fail, as must creating one over an existing waiver file; the created file must verify against a declared key. Minor: '90 days ahead' is implicitly from the creation date.
- `O-2` AMBIGUOUS: 'had not expired by its date' does not say which commit date: author and committer dates can straddle the expiry. The spec uses the UTC committer date. Proposed rewording: 'had an expiry on or after the UTC date of the commit's committer date'. The rest is testable with fixtures.
- `O-3` VERIFIABLE: Testable: assert a waived commit appears with signer, waiver, expiry and reason on each named surface (PR check, bundle index, verification page, badge, Audit room), and that a waiver change is listed in the Audit room.
