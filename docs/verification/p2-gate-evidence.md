# P2 gate acceptance evidence

The P2 canary-to-enforce requirement is proven by
`kernel/internal/gates/gates_integration_test.go::TestCanaryThenEnforceRejectsBadWitness`.

The test creates a temporary Git repository containing a committed `bad.go` file and a real
`Makefile` target that fails when that bad file is present. It runs the actual witness wrapper,
ingests the resulting `check.run` event, loads the committed `gates/kernel-check-success.yaml`,
evaluates it in `canary` mode, and verifies `BLOCKED` plus event ID and sequence. It then promotes
the same loaded definition to `enforce` and verifies that the proof is rejected.

Reproduction command:

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:55433/vera?sslmode=disable \
  go test -tags integration ./internal/gates -run TestCanaryThenEnforceRejectsBadWitness -count=1
```
