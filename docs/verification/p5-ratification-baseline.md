# P5 ratification baseline — local-machine attachment

**Captured:** 2026-09-10  
**Purpose:** Attach the repository state and bare blocking gate requested by the round-1
adjudication, whose verifier could not access this checkout. This is mechanical evidence, not an
independent acceptance verdict.

The received adjudication, reviewed exhibit, and founder ratification were first preserved verbatim
in commit `578898b`, in accordance with `VD-verdicts-are-artifacts-rl0rab`. The exhibit's SHA-256 was
verified before that commit as
`9d203c96243a73cad2cc119662a9b58ea04dc7e9b17d8e5ed3b1a418118d5ffa`, exactly matching the
adjudication provenance header.

## Receipt baseline

Command:

```text
git log -1 --oneline
```

Output (exit 0):

```text
578898b docs: preserve P5 ratification artifacts on receipt
```

Command, run bare and unpiped:

```text
make check
```

Output (exit 2):

```text
index stale; run make index
snap-confine is packaged without necessary permissions and cannot continue
required permitted capability cap_dac_override not found in current capabilities:
  =: Function not implemented
snap-confine is packaged without necessary permissions and cannot continue
required permitted capability cap_dac_override not found in current capabilities:
  =: Function not implemented
make: *** [Makefile:46: kernel-check] Error 1
```

The first failure is the expected derived-index consequence of committing new verdict artifacts.
The second is an execution-environment restriction on the Snap-packaged Go toolchain. Neither result
is suppressed; the remediated rerun is appended below when complete.

## Post-ratification gate

The host has `golangci-lint` installed at `/home/thamm/go/bin/golangci-lint`, but that directory was
not present in the invoking session's `PATH`. After adding the existing tool directory to the
process environment, the following `make check` command itself was again run bare and unpiped.

Command:

```text
make check
```

Output (exit 0):

```text
index stale; run make index
ok  	github.com/kamisrini/proofbound/tools/mutants	0.010s
ok  	github.com/kamisrini/proofbound/kernel/cmd/vera	0.015s
ok  	github.com/kamisrini/proofbound/kernel/internal/connector/checks	5.615s
ok  	github.com/kamisrini/proofbound/kernel/internal/connector/git	1.760s
ok  	github.com/kamisrini/proofbound/kernel/internal/connector/git/gitcmd	3.593s
ok  	github.com/kamisrini/proofbound/kernel/internal/connector/github	0.012s
ok  	github.com/kamisrini/proofbound/kernel/internal/connector/reviews	0.006s
ok  	github.com/kamisrini/proofbound/kernel/internal/connector/sessions	0.023s
ok  	github.com/kamisrini/proofbound/kernel/internal/core	0.007s
ok  	github.com/kamisrini/proofbound/kernel/internal/gates	0.007s
ok  	github.com/kamisrini/proofbound/kernel/internal/projections	0.008s
ok  	github.com/kamisrini/proofbound/kernel/internal/specfirst	0.021s
ok  	github.com/kamisrini/proofbound/kernel/internal/store	0.053s
ok  	github.com/kamisrini/proofbound/kernel/internal/twin	45.189s
0 issues.
```

The `index stale; run make index` line is emitted by the index hook's negative self-test and is
expected within a successful aggregate run; the generated index itself matched when checked
directly. Exit status 0 is the blocking-gate result.
