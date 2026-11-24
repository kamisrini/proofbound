---
{"schema":"proofbound.requirement-review.v1","review_id":"RR-ui-bundle-round1-20261124","requirement":{"record_kind":"requirement","source":"intent.records","record_id":"BR-ui-bundle-da6ue0","artifact_sha256":"7769aa5334dea4c738fc9666a68f17f185814d5e2dc5f1ccf503ce78202a92d7","relation":"reviews"},"declared_reviewer":"agent-requirement-reviewer-20261124","outcomes":[{"obligation_id":"O-1","outcome":"VERIFIABLE","finding":"Testable: copy only the bundle folder to a machine with no clone, database or DATABASE_URL and open the page; writes must fail; the Audit room's guard tests (cross-origin request, foreign Host header) must pass unchanged against the bundle page."},{"obligation_id":"O-2","outcome":"VERIFIABLE","finding":"Testable by comparison: per warranty, the page's result must match warranty verify on the same file without and with a clone; deleting or editing a warranty must show missing or altered against the index, and the index must be labelled unsigned."},{"obligation_id":"O-3","outcome":"VERIFIABLE","finding":"Testable: what is shown must come only from the signed warranties (editing the unsigned index must not change it); waived commits must appear marked unsigned; a warranty with one byte altered must fail the browser's own signature check."}],"artifact_path":"docs/verification/verdicts/RR-ui-bundle-round1-20261124.md","artifact_sha256":"64c6f4848d6adf63a849944bf161ef4fe4d0fbbbccdc6aab76f9515bb08d50e6"}
---

# Independent requirement review: release evidence in a browser, round 1

Fresh non-author reviewer context: `agent-requirement-reviewer-20261124`. Reviewed commit: `8d1c786238809ae99776b5f380884d10e39edea3`.

This review binds the exact `BR-ui-bundle-da6ue0/0001` revision. The declared reviewer is distinct from the requirement's declared owner, `proofbound-maintainer`.

`VERIFIABLE` means the obligation's words have a conceivable independent test and evidence route. It does not judge business merit or accept any implementation.

- `O-1` VERIFIABLE: Testable: copy only the bundle folder to a machine with no clone, database or DATABASE_URL and open the page; writes must fail; the Audit room's guard tests (cross-origin request, foreign Host header) must pass unchanged against the bundle page.
- `O-2` VERIFIABLE: Testable by comparison: per warranty, the page's result must match warranty verify on the same file without and with a clone; deleting or editing a warranty must show missing or altered against the index, and the index must be labelled unsigned.
- `O-3` VERIFIABLE: Testable: what is shown must come only from the signed warranties (editing the unsigned index must not change it); waived commits must appear marked unsigned; a warranty with one byte altered must fail the browser's own signature check.
