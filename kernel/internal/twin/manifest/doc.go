// Package manifest validates the non-runtime P7 documentary contracts for the twin
// replay-lock line: the projector/build manifest and its source/toolchain evidence
// inventory, the ReplayIsolated case evidence index (case-manifest.v1), and the
// authored ReplayIsolated behavior-lock (behavior-lock.v1). Every operation validates
// declarations and caller-supplied committed bytes; none retrieves, builds, executes,
// or approves a runtime input.
package manifest
