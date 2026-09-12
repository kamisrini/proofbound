# Proofbound P5 mutation evidence

This record binds the package-acceptance mutation sweep for the P5 implementation. The
runner calibration passed for every invocation (`neutral=survived`, `invalid=invalid`,
`lethal=killed`). Integration ranges used disposable PostgreSQL services and
`MUTANT_TEST_TAGS=integration`; each range was run with `-p=1` and the repository-root
handoff environment required by the gate and CLI fixtures.

Connector packages were fully killed: intent 93/93, intent/records 52/52,
intent/specdir 47/47, git 34/34, and reviews 114/114.

The projections package had 405 candidates in the initial segmented sweep. All candidates
were killed; one unobservable dead assignment in `report.go` was then removed, leaving
404 production candidates, all killed after the final focused reruns. The completed ranges
were 1–126, 127–240, 241–323, and 324–405, with focused reruns for the prior survivors
and the final envelope-validation tests.

The gates package completed 109/109 killed (segmented ranges with focused reruns for the
validation and delivery-proof selectors). The CLI package completed 246/246 killed
(segmented ranges with focused reruns for sync-all ordering, provider-error propagation,
gate enforcement, verification-stage errors, and timeout labeling).

No invocation reported an invalid or surviving mutant. The focused integration suites were
also run in isolation after the final test additions:

```
go test -tags=integration ./internal/projections -count=1
go test -tags=integration ./internal/gates -count=1
go test -tags=integration ./internal/cli -count=1
```

All three suites passed.
