# P6 intent applicability canary

**Anchor:** immediately after `f426ca8`

**Head:** `13503c7c96bad63ce1b95405ee8a83f09ccdab5a`

**Policy:** founder-ratified path-only predicate from `docs/decisions/VD-p6-consolidation-2026-09-12.md`.

This is a pre-enforcement observation. Applicability is content-independent; no commit is retroactively rejected.

| Commit | Subject | Changed paths | Applicable | Valid exact Intent claim |
|---|---|---|---|---|
| `1d1112df2622` | docs: record P5 close-out verification | notes/journal/2026-09-12.md, notes/state.md | false | no |
| `1e5e96cd25c7` | chore: ignore local status reports | .gitignore | false | no |
| `ceeae16e1280` | docs: adjudicate P6 consolidation draft | docs/plans/P6-CONSOLIDATION-PLAN-draft1.md, docs/verification/verdicts/p6-consolidation-plan-round1.md | false | no |
| `67a3259db786` | docs: fold P6 consolidation findings | docs/plans/P6-CONSOLIDATION-PLAN-draft2.md, notes/journal/2026-09-12.md, notes/state.md | false | no |
| `28d87167f972` | docs: record P6 founder ratification | docs/verification/verdicts/p6-founder-ratification.md | false | no |
| `708ad123912f` | docs: authorize P6 consolidation | ROADMAP.md, docs/decisions/INDEX.md, docs/decisions/VD-p6-consolidation-2026-09-12.md | true | no |
| `1ba2336e310d` | docs: specify P6 debt census | docs/plans/p6-census-SPEC.md | false | no |
| `7046b425d6de` | build: add closed P6 census mechanism | scripts/p6-census.sh, scripts/tests/p6-census.test.sh | true | no |
| `7bb9a66df00c` | build: close P6 census discovery universe | scripts/p6-census.sh | true | no |
| `4c54a707b531` | docs: publish first P6 debt census | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, notes/journal/2026-09-12.md, notes/state.md, scripts/identity-inventory.sh, scripts/tests/identity-inventory.test.sh | true | no |
| `cce0a5d46eba` | docs: specify P6 mechanism restoration | docs/plans/p6-task1-mechanisms-SPEC.md | false | no |
| `dd6154e7824c` | build: restore P6 operational gates | Makefile, docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, scripts/backup.sh, scripts/commit-cadence.sh, scripts/gen-laws-lock.sh, scripts/gen-state.sh, scripts/identity-inventory.sh, scripts/meta-tax.sh, scripts/state-freshness.sh, scripts/tests/commit-cadence.test.sh, scripts/tests/identity-inventory.test.sh, scripts/tests/make-contract.test.sh, scripts/tests/operational-tools.test.sh, scripts/tests/state-freshness.test.sh, scripts/wrap-verify.sh | true | no |
| `908d03a37ef2` | docs: checkpoint P6 Task 1 restoration | notes/journal/2026-09-12.md, notes/state.md | false | no |
| `7fd328d005aa` | build: restore P6 lexical gates | Makefile, docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, scripts/cleanroom-lint.sh, scripts/figure-provenance.sh, scripts/invariant-lint.sh, scripts/kernel-check.sh, scripts/lesson-recurrence.sh, scripts/prescription-lint.sh, scripts/skip-lint.sh, scripts/tests/cleanroom-lint.test.sh, scripts/tests/figure-provenance.test.sh, scripts/tests/invariant-lint.test.sh, scripts/tests/lesson-recurrence.test.sh, scripts/tests/prescription-lint.test.sh, scripts/tests/skip-lint.test.sh | true | no |
| `7d840ba33b27` | build: restore repository hook controls | .claude/hooks/block-generated-edit.sh, .claude/hooks/block-secrets.sh, .claude/hooks/lint-on-write.sh, .claude/hooks/stop-check.sh, .claude/settings.json, docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, scripts/tests/claude-hooks.test.sh | true | no |
| `4de1439bd8d9` | build: enforce exact SPEC citations | Makefile, docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, kernel/internal/core/SPEC.md, kernel/internal/gates/SPEC.md, kernel/internal/store/SPEC.md | true | no |
| `2d35faa06102` | docs: record P6 legacy-storage gate | notes/journal/2026-09-12.md, notes/state.md | false | no |
| `d4b240cbe470` | docs: ratify private legacy storage exception | docs/verification/verdicts/p6-legacy-storage-ratification.md | false | no |
| `a128ac098723` | docs: freeze migrated private storage identity | docs/decisions/INDEX.md, docs/decisions/VD-p6-private-storage-compat-2026-09-12.md | true | no |
| `8e287acbf3c0` | refactor: retire external legacy identity aliases | .gitignore, Makefile, docs/gates.md, docs/proofbound-identity-migration.md, kernel/cmd/vera/main.go, kernel/internal/cli/SPEC.md, kernel/internal/cli/cli.go, kernel/internal/cli/cli_test.go, kernel/scripts/check-witness.sh, scripts/identity-inventory.sh, scripts/tests/identity-inventory.test.sh, scripts/tests/index-check.test.sh | true | no |
| `0729da5fbb4b` | build: close P6 documentation and gate truth | Makefile, README.md, docs/gates.md, docs/verification/p6-c1-c2-closure.md, scripts/p6-task1-close.sh, scripts/tests/p6-task1-close.test.sh | true | no |
| `287e1e7c4ce2` | docs: checkpoint P6 Task 1 closure | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, docs/verification/p6-c1-c2-closure.md, notes/journal/2026-09-12.md, notes/state.md | false | no |
| `5e17c9524ef1` | docs: specify P6 connector reality proof | docs/plans/p6-task2-connector-SPEC.md | false | no |
| `8ccf593eaf79` | test: reprove connector and delivery boundaries | kernel/internal/cli/SPEC.md, kernel/internal/cli/cli.go, kernel/internal/cli/cli_test.go, kernel/internal/connector/github/SPEC.md, kernel/internal/connector/github/github_test.go, kernel/internal/connector/sessions/live_acceptance_test.go, scripts/tests/delivery-enforce.test.sh, scripts/tests/make-contract.test.sh | true | no |
| `8c1381e0fdd7` | docs: record P6 connector reality evidence | docs/verification/p6-connector-reality.md, scripts/p6-task2-close.sh, scripts/tests/p6-task2-close.test.sh | true | no |
| `e169b63569be` | docs: checkpoint P6 Task 2 closure | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, notes/journal/2026-09-12.md, notes/state.md | false | no |
| `d3a78ac60671` | docs: specify P6 event and reproducibility proof | docs/plans/p6-task3-reproducibility-SPEC.md | false | no |
| `c4c9ebaaa755` | test: prove complete event route coverage | kernel/internal/gates/SPEC.md, kernel/internal/gates/gates_integration_test.go, kernel/internal/twin/SPEC.md, kernel/internal/twin/replay_test.go | true | no |
| `ddbccbc5dd29` | docs: record complete P6 event route matrix | docs/invariants.lock, docs/verification/p6-event-route-matrix.md, scripts/p6-route-matrix.sh, scripts/tests/p6-route-matrix.test.sh | true | no |
| `a5383b947ecc` | docs: audit P6 evidence integrity | docs/verification/p6-artifact-integrity.md, docs/verification/p6-vision-progress-assessment.md, notes/journal/2026-09-12.md, notes/state.md, scripts/p6-artifact-integrity.sh, scripts/tests/p6-artifact-integrity.test.sh | true | no |
| `e1bb67df7ed2` | docs: record P6 reproducibility blockers | docs/verification/p6-fresh-clone-linux.md, docs/verification/p6-windows-platform.md, notes/journal/2026-09-12.md, notes/state.md | false | no |
| `3810a851f7e1` | build: bind P6 census evidence freshness | docs/gates.md, docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, notes/journal/2026-09-12.md, notes/state.md, scripts/tests/p6-census-current.test.sh | true | no |
| `81380334d068` | docs: record P6 end-of-day handoff | notes/journal/2026-09-12.md, notes/state.md | false | no |
| `10fa22c6b124` | docs: keep dated resume handoff at top | CLAUDE.md, docs/eod-prompt.md, notes/journal/2026-09-12.md, notes/state.md | true | no |
| `fd8fc2b29a9c` | docs: ratify dual-platform P6 execution | README.md, ROADMAP.md, docs/decisions/INDEX.md, docs/decisions/VD-p6-dual-platform-2026-09-13.md, docs/plans/p6-task3-reproducibility-SPEC.md, docs/verification/p6-windows-platform.md, docs/verification/verdicts/p6-dual-platform-ratification.md, notes/journal/2026-09-13.md, notes/state.md | true | no |
| `33bc32bca4a9` | docs: refresh P6 artifact integrity census | docs/verification/p6-artifact-integrity.md | false | no |
| `5ddaed05b5ef` | build: define native Windows Bash toolchain | check-windows.ps1, notes/journal/2026-09-13.md, notes/state.md, scripts/tests/windows-contract.test.sh, setup-windows.ps1 | true | no |
| `dc7d587e2b2d` | build: make Windows gate portable | check-windows.ps1, scripts/tests/windows-contract.test.sh, setup-windows.ps1 | true | no |
| `423e74a50a3f` | build: normalize Windows repository checks | scripts/gen-invariants-lock.sh, scripts/identity-inventory.sh | true | no |
| `f46606d9fecb` | build: make identity inventory path-stable | scripts/identity-inventory.sh | true | no |
| `a26f3e56c33f` | build: parse Windows identity paths safely | scripts/identity-inventory.sh | true | no |
| `33c8632e650a` | docs: record Windows acceptance handoff | notes/journal/2026-09-13.md, notes/state.md | false | no |
| `879f04d3f303` | fix: support store locking on Windows | kernel/internal/store/SPEC.md, kernel/internal/store/lock.go, kernel/internal/store/lock_unix.go, kernel/internal/store/lock_windows.go | true | no |
| `1a3dbaa44fac` | docs: refresh invariant lock wording | docs/invariants.lock | true | no |
| `f6c2030ea462` | test: bound Windows filesystem fixtures | kernel/internal/connector/checks/SPEC.md, kernel/internal/connector/checks/emitter_test.go, kernel/internal/connector/checks/process_group_test_unix.go, kernel/internal/connector/checks/process_group_test_windows.go, kernel/internal/connector/git/gitcmd/SPEC.md, kernel/internal/connector/git/gitcmd/gitcmd_test.go | true | no |
| `212cd18a570f` | docs: refresh connector closure evidence | docs/verification/p6-connector-reality.md, scripts/p6-task2-close.sh | true | no |
| `3842bf0e244d` | docs: refresh native acceptance resume state | notes/journal/2026-09-13.md, notes/state.md | false | no |
| `98f5e05defd5` | test: normalize native Windows fixture paths | kernel/internal/connector/checks/emitter_test.go, kernel/internal/connector/git/gitcmd/SPEC.md, kernel/internal/connector/git/gitcmd/gitcmd_test.go | true | no |
| `06105f067ef4` | docs: bind connector evidence to current tests | docs/verification/p6-connector-reality.md | false | no |
| `59814ad7a488` | test: close native Windows filesystem gaps | kernel/internal/connector/checks/SPEC.md, kernel/internal/connector/checks/emitter_test.go, kernel/internal/connector/sessions/SPEC.md, kernel/internal/connector/sessions/sessions.go, kernel/internal/connector/sessions/sessions_test.go, kernel/internal/store/SPEC.md, kernel/internal/store/lock_test.go, kernel/internal/store/surface_test.go | true | no |
| `9bd8b351490f` | docs: bind native Windows gate prerequisites | docs/allowed-skips.txt, docs/verification/p6-connector-reality.md, notes/journal/2026-09-13.md, notes/state.md | true | no |
| `01e0dc99787a` | fix: accept valid detached Git checkouts | docs/verification/p6-connector-reality.md, kernel/internal/connector/git/gitcmd/SPEC.md, kernel/internal/connector/git/gitcmd/gitcmd.go, kernel/internal/connector/git/gitcmd/gitcmd_test.go | true | no |
| `b0e3d36e7d70` | docs: record detached checkout acceptance blocker | notes/journal/2026-09-13.md, notes/state.md | false | no |
| `a2e3aedddfd3` | docs: refresh connector closure binding | docs/verification/p6-connector-reality.md | false | no |
| `1efd821a2c4e` | docs: bind native Windows acceptance result | docs/verification/p6-windows-platform.md, notes/journal/2026-09-13.md, notes/state.md | false | no |
| `5e21cc5a2b4b` | docs: finalize P6 resume state | notes/journal/2026-09-13.md, notes/state.md | false | no |
| `9c57ade6902c` | docs: ratify historical evidence portability | docs/decisions/INDEX.md, docs/decisions/VD-p6-historical-evidence-portability-2026-09-13.md, docs/verification/verdicts/p6-historical-evidence-ratification.md | true | no |
| `80fb20cbd542` | test: specify historical evidence portability migration | docs/plans/p6-historical-evidence-portability-SPEC.md, docs/verification/p6-historical-evidence.jsonl, kernel/internal/cli/cli_test.go, kernel/internal/migration/migration_test.go | true | no |
| `727331144151` | fix: restore cited historical evidence on migration | README.md, docs/verification/p6-artifact-integrity.md, kernel/internal/cli/cli.go, kernel/internal/migration/SPEC.md, kernel/internal/migration/migration.go, kernel/internal/migration/migration_test.go, kernel/internal/store/SPEC.md, kernel/internal/store/lock.go, kernel/internal/store/store.go, scripts/identity-inventory.sh | true | no |
| `d89f423b2526` | docs: bind historical migration to task three | docs/plans/p6-task3-reproducibility-SPEC.md, docs/verification/p6-event-route-matrix.md, scripts/p6-route-matrix.sh | true | no |
| `277554ed5fec` | docs: bind migration census and resume state | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, docs/verification/p6-connector-reality.md, notes/state.md | false | no |
| `b3303100f649` | fix: include historical intent chain prerequisites | docs/decisions/VD-p6-historical-evidence-portability-2026-09-13.md, docs/plans/p6-historical-evidence-portability-SPEC.md, docs/verification/p6-historical-evidence.jsonl, kernel/internal/migration/SPEC.md, kernel/internal/migration/migration.go, kernel/internal/migration/migration_test.go, notes/state.md | true | no |
| `a887e46d757b` | docs: close P6 Task 3 reproducibility evidence | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, docs/verification/p6-fresh-clone-linux.md, docs/verification/p6-windows-platform.md, notes/journal/2026-09-13.md, notes/state.md | false | no |
| `3b8db1cf729e` | docs: refresh P6 resume state | notes/state.md | false | no |
| `ee56d9abcca3` | feat: specify and match P6 intent applicability paths | docs/plans/p6-task4-applicability-SPEC.md, scripts/intent-applicability.sh, scripts/tests/intent-applicability.test.sh | true | no |
| `c1d8da6e226d` | docs: record Task 4 applicability progress | notes/journal/2026-09-13.md, notes/state.md | false | no |
| `aadd6601a68d` | feat: record P6 intent applicability canary | docs/plans/p6-task4-applicability-SPEC.md, notes/journal/2026-09-13.md, notes/state.md, scripts/p6-intent-canary.sh | true | no |
| `bbd9288241b7` | docs: bind complete P6 applicability canary | docs/verification/p6-intent-applicability-canary.md, notes/journal/2026-09-13.md, notes/state.md | false | no |
| `d6abc5b3c777` | feat: gate delivery on applicable commit intent | docs/plans/p6-task4-applicability-SPEC.md, scripts/delivery-enforce.sh, scripts/intent-delivery-applicability.sh, scripts/tests/intent-delivery-applicability.test.sh | true | no |
| `3d3fdfbad99b` | docs: bind Task 4 delivery boundary evidence | docs/verification/p6-intent-applicability-canary.md, notes/journal/2026-09-13.md, notes/state.md | false | no |
| `8fa7fded7023` | docs: close P6 intent coverage census | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, docs/verification/p6-intent-coverage.md, notes/journal/2026-09-13.md, notes/state.md | false | no |
| `1df968ffc401` | docs: mark P6 Task 4 complete | notes/journal/2026-09-13.md, notes/state.md | false | no |
| `7d72898ef0e4` | build: close P6 Task 5 requirement reviews | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, docs/plans/p6-task5-requirement-review-SPEC.md, docs/verification/p6-c1-c2-closure.md, docs/verification/p6-requirement-review.md, notes/journal/2026-09-14.md, notes/state.md, scripts/p6-task1-close.sh, scripts/p6-task5-review.sh, scripts/tests/p6-task5-review.test.sh | true | no |
| `44eedc4b9f47` | docs: refresh P6 connector evidence binding | docs/verification/p6-connector-reality.md, notes/journal/2026-09-14.md, notes/state.md | false | no |
| `329bceaa4176` | docs: record P6 Task 5 acceptance | notes/journal/2026-09-14.md, notes/state.md | false | no |
| `ea7c499f7ef0` | docs: bind P6 measurement and falsifier results | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, docs/plans/p6-task6-measurements-SPEC.md, docs/plans/p6-task6-results.tsv, docs/verification/p6-measurements-falsifiers.md, notes/journal/2026-09-14.md, notes/state.md, scripts/p6-task6-results.sh, scripts/tests/p6-task5-review.test.sh, scripts/tests/p6-task6-results.test.sh | true | no |
| `f20376740368` | docs: record P6 Task 7 skip decision | notes/journal/2026-09-14.md, notes/state.md | false | no |
| `b47fc2cf8888` | docs: freeze P6 Task 8 package universe | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, docs/plans/p6-task8-package-acceptance-SPEC.md, notes/journal/2026-09-14.md, notes/state.md, scripts/tests/p6-task8-package-universe.test.sh | true | no |
| `5b75fe690495` | test: close P6 core mutation survivors | kernel/internal/core/core.go, kernel/internal/core/core_test.go, notes/journal/2026-09-14.md, notes/state.md | true | no |
| `d9055fc0f089` | docs: record CLI mutation audit | notes/journal/2026-09-14.md, notes/state.md | false | no |
| `7df0c502e432` | docs: refine CLI mutation resume state | notes/journal/2026-09-14.md, notes/state.md | false | no |
| `1e92aaa1b683` | test: harden CLI error-path contracts | kernel/internal/cli/cli.go, kernel/internal/cli/cli_test.go, notes/journal/2026-09-14.md, notes/state.md | true | no |
| `3f115eadafb4` | test: harden CLI mutation contracts | kernel/internal/cli/cli.go, kernel/internal/cli/cli_test.go, notes/journal/2026-09-14.md, notes/state.md | true | no |
| `9a4760183c0b` | docs: bind CLI package acceptance evidence | docs/verification/p6-task8-cli-mutation-3f115ea.md, notes/journal/2026-09-14.md, notes/state.md | false | no |
| `27c2bdfc5662` | docs: bind checks mutation evidence | docs/verification/p6-task8-checks-mutation-3f115ea.md, notes/journal/2026-09-14.md, notes/state.md | false | no |
| `343fa5d331b6` | docs: bind git mutation evidence | docs/verification/p6-task8-checks-mutation-3f115ea.md, docs/verification/p6-task8-git-mutation-3f115ea.md, notes/journal/2026-09-14.md, notes/state.md | false | no |
| `a5ef80329441` | test: close gitcmd mutation survivors | kernel/internal/connector/git/gitcmd/gitcmd.go, kernel/internal/connector/git/gitcmd/gitcmd_test.go | true | no |
| `f84af261c8b5` | docs: bind frozen package mutation results | docs/verification/p6-task8-package-results-a5ef803.md, notes/journal/2026-09-14.md, notes/state.md | false | no |
| `a5ce3b618100` | test: close GitHub mutation survivors | kernel/internal/connector/github/github_test.go | true | no |
| `ad69db938455` | test: bind gates mutation fixture root | kernel/internal/gates/gates_integration_test.go | true | no |
| `95498bbf9a00` | test: close gate validation mutation survivors | kernel/internal/gates/gates_test.go | true | no |
| `86683f63ba4d` | test: bind migration fixture root | kernel/internal/migration/migration_test.go | true | no |
| `65dabcad95b3` | test: close migration mutation survivors | kernel/internal/migration/migration_test.go | true | no |
| `6dfb4d19d866` | test: close remaining migration survivors | kernel/internal/migration/migration.go, kernel/internal/migration/migration_test.go | true | no |
| `4499fa5f0217` | test: assert migration trailing error contract | kernel/internal/migration/migration_test.go | true | no |
| `c6677e4038f4` | docs: bind migration mutation evidence | docs/verification/p6-task8-migration-mutation-4499fa5.md, notes/journal/2026-09-14.md, notes/state.md | false | no |
| `3b53b6b18698` | docs: refresh generated verification artifacts | docs/verification/p6-connector-reality.md, docs/verification/p6-event-route-matrix.md | false | no |
| `9ef6eec2952e` | docs: record migration resume state | notes/journal/2026-09-14.md, notes/state.md | false | no |
| `45b7a35123b0` | docs: finalize resume checkpoint | notes/journal/2026-09-14.md, notes/state.md | false | no |
| `7cb2f6cdc7b7` | docs: record projections execution checkpoint | notes/journal/2026-09-14.md, notes/state.md | false | no |
| `cc74bb529325` | test: support focused mutation execution | tools/mutants/main.go, tools/mutants/main_test.go | true | no |
| `07f04c4b16b2` | test: close requirement review digest survivors | kernel/internal/projections/intent_review_test.go | true | no |
| `e134bbd216c9` | test: enforce exact obligation verdict coverage | kernel/internal/projections/intent_review_test.go | true | no |
| `b04013dbe57f` | test: cover requirement review envelope invariants | kernel/internal/projections/intent_review_test.go | true | no |
| `99fa0ca7d955` | test: cover commit intent reference validation | kernel/internal/projections/projection_test.go | true | no |
| `3ae97a3aee79` | test: isolate intent reference validation cases | kernel/internal/projections/projection_test.go | true | no |
| `5b1b30789e39` | test: cover empty requirement reviewer | kernel/internal/projections/intent_review_test.go | true | no |
| `2f2520b32fcf` | test: cover obligation outcome domain | kernel/internal/projections/intent_review_test.go | true | no |
| `17cb4b1669a2` | test: cover workflow payload validation branches | kernel/internal/projections/projection_test.go | true | no |
| `e71f669aa73c` | test: propagate finding projection errors | kernel/internal/projections/intent_review_test.go | true | no |
| `894b6854267b` | test: propagate report scan errors | kernel/internal/projections/report_integration_test.go | true | no |
| `d69392c394af` | test: propagate session report scan errors | kernel/internal/projections/report_integration_test.go | true | no |
| `69794c086484` | test: propagate review report scan errors | kernel/internal/projections/report_integration_test.go | true | no |
| `4240cd6baf91` | test: parallelize isolated projection integration tests | kernel/internal/projections/projection_test.go | true | no |
| `2e674ed199a8` | test: close projection report proof survivors | kernel/internal/projections/report_integration_test.go | true | no |
| `b3e9e1aaeab5` | test: target report session and review writer failures | kernel/internal/projections/report_integration_test.go | true | no |
| `d44b540dd337` | test: distinguish projection insert error propagation | kernel/internal/projections/intent_review_test.go | true | no |
| `80a60d566bd2` | test: preserve requirement review insert errors | kernel/internal/projections/intent_review_test.go | true | no |
| `172f4c46d925` | test: cover GitHub deployment validation fields | kernel/internal/projections/projection_test.go | true | no |
| `21b8bb5be33c` | test: isolate intent reference validation cases | kernel/internal/projections/projection_test.go | true | no |
| `12f0d74b948f` | test: cover report review chain output failures | kernel/internal/projections/report_integration_test.go | true | no |
| `be3c5c2b3c28` | test: cover report reachability and chain filtering | kernel/internal/projections/report_integration_test.go, kernel/internal/projections/review_integration_test.go | true | no |
| `be58cd7dcb17` | test: enforce projection report interval boundaries | kernel/internal/projections/projection_test.go, kernel/internal/projections/review_integration_test.go | true | no |
| `ab78e827b4aa` | docs: record projections mutation acceptance | docs/verification/p6-task8-projections-mutation-be58cd7.md, notes/journal/2026-09-15.md, notes/state.md | false | no |
| `22eaff4ef120` | docs: refresh route matrix binding | docs/verification/p6-event-route-matrix.md | false | no |
| `ab5060a2e324` | docs: record gate toolchain limitation | notes/journal/2026-09-15.md, notes/state.md | false | no |
| `eb40978ad885` | test: close store mutation survivors | kernel/internal/store/integration_test.go, kernel/internal/store/store.go, kernel/internal/store/surface_test.go, tools/mutants/main.go, tools/mutants/main_test.go | true | no |
| `e4c8e7740769` | fix: preserve isolated replay sequence updates | kernel/internal/store/store.go | true | no |
| `f1b73f7559a2` | docs: record store and twin mutation acceptance | docs/verification/p6-task8-store-twin-mutation-e4c8e77.md, notes/journal/2026-09-15.md, notes/state.md | false | no |
| `a199229351f9` | docs: record frozen package reruns | docs/verification/p6-task8-frozen-reruns-e4c8e77.md, notes/state.md | false | no |
| `e1e35a84848e` | docs: checkpoint 2026-09-16 workday (gate pending) | notes/journal/2026-09-16.md, notes/state.md | false | no |
| `83777166b8fe` | docs: record end-of-day verification outcome | notes/journal/2026-09-16.md, notes/state.md | false | no |
| `0067dc61755e` | docs: finalize durable EOD state | notes/journal/2026-09-16.md, notes/state.md | false | no |
| `8bef29020d07` | feat: add frozen package acceptance gate | docs/verification/p6-artifact-integrity.md, docs/verification/p6-task8-cli-projections-mutation-e4c8e77.md, docs/verification/p6-task8-package-results-e4c8e77.tsv, docs/verification/verdicts/p6-task8-current-round1-e4c8e77.md, notes/journal/2026-09-18.md, notes/state.md, scripts/p6-census.sh, scripts/p6-task8-package-acceptance.sh, scripts/tests/p6-task8-package-acceptance.test.sh | true | no |
| `e02dbb1c0569` | fix: compare non-author identity consistently | scripts/p6-task8-package-acceptance.sh, scripts/tests/p6-task8-package-acceptance.test.sh | true | no |
| `c10fb8d2b66e` | docs: refresh P6 census intent coverage | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, docs/verification/p6-intent-applicability-canary-task8-20260918.md, docs/verification/p6-intent-applicability-canary.md, docs/verification/p6-intent-coverage.md, notes/journal/2026-09-18.md, notes/state.md | false | no |
| `624e7040b3e5` | docs: close P6 Task 8 package acceptance rows | docs/plans/p6-census-rows.tsv, docs/plans/p6-census.md, notes/journal/2026-09-18.md, notes/state.md | false | no |
| `054da9db9921` | docs: record P6 Task 8 durability checkpoint | notes/journal/2026-09-18.md, notes/state.md | false | no |
| `4254ac7110bc` | docs: record 2026-09-19 durability checkpoint | notes/journal/2026-09-19.md, notes/state.md | false | no |
| `7af4e683d732` | docs: record final 2026-09-19 gate result | notes/journal/2026-09-19.md, notes/state.md | false | no |
| `63a60fc4c41d` | docs: finalize 2026-09-19 handoff | notes/journal/2026-09-19.md, notes/state.md | false | no |
| `04d60c12c550` | docs: align final backup handoff state | notes/journal/2026-09-19.md, notes/state.md | false | no |
| `7b1dda2ebc4a` | docs: record final EOD bundle identity | notes/journal/2026-09-19.md, notes/state.md | false | no |
| `a20d7ea079a9` | docs: finalize handoff state wording | notes/state.md | false | no |
| `13503c7c96ba` | fix: execute P6 result census probes | docs/plans/p6-census-rows.tsv, scripts/p6-census.sh, scripts/tests/p6-census.test.sh | true | no |

## Summary

- Commits examined: `145`.
- Applicable commits: `82`.
- Applicable commits with valid exact Intent claims: `0`.
- Coverage: `0.0%`.
- Known false negatives: `0`.
- False-positive rate under the ratified path definition: `0.0%`; the >10% redesign threshold did not fire.
- Enforcement status: `canary`; only the existing explicit `make delivery-enforce` boundary may later enforce this rule.
