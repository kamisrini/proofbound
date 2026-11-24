---
{"schema":"proofbound.requirement-review.v1","review_id":"RR-tests-can-fail-round4-20261124","requirement":{"record_kind":"requirement","source":"intent.records","record_id":"BR-tests-can-fail-k2h06j","artifact_sha256":"3bb0d48b597d69830551c0cb23903be5b359c7cd266f3bb13f1aee62d6447e21","relation":"reviews"},"declared_reviewer":"agent-requirement-reviewer-round4-20261124","outcomes":[{"obligation_id":"O-4","outcome":"AMBIGUOUS","finding":"'as git diff -U0 against the commit's first parent lists it (the same line ranges C10 lists)' gives two definitions that disagree: plain git diff detects renames and prints no hunk for a binary or mode-only change, while C10 (area.go) and tests.go use --no-renames and list a binary, empty or mode-only change as one whole-file entry. A renamed code file with a one-line edit, or a binary change, would give testers different hunk counts. Proposed rewording: undo each changed hunk, as git diff -U0 --no-renames against the commit's first parent lists it, a binary, empty or mode-only file change counting as one whole-file hunk (the same entries C10 lists), of a claiming commit outside ..."}],"artifact_path":"docs/verification/verdicts/RR-tests-can-fail-round4-20261124.md","artifact_sha256":"a4055fc22b979c2c2403090bd541d405abddb5f6ee4ad15cd6750dc2c96ad306"}
---

# Independent requirement review: tests that can fail, round 4

Fresh non-author reviewer context: `agent-requirement-reviewer-round4-20261124`. Reviewed commit: `59d54f52e206fb56008cc4d81a63f2bb5a9318aa`.

This review binds the exact `BR-tests-can-fail-k2h06j/0004` revision. The declared reviewer is distinct from the requirement's declared owner, `proofbound-maintainer`.

`VERIFIABLE` means the obligation's words have a conceivable independent test and evidence route. It does not judge business merit or accept any implementation.

- `O-4` AMBIGUOUS: 'as git diff -U0 against the commit's first parent lists it (the same line ranges C10 lists)' gives two definitions that disagree: plain git diff detects renames and prints no hunk for a binary or mode-only change, while C10 (area.go) and tests.go use --no-renames and list a binary, empty or mode-only change as one whole-file entry. A renamed code file with a one-line edit, or a binary change, would give testers different hunk counts. Proposed rewording: undo each changed hunk, as git diff -U0 --no-renames against the commit's first parent lists it, a binary, empty or mode-only file change counting as one whole-file hunk (the same entries C10 lists), of a claiming commit outside ...
