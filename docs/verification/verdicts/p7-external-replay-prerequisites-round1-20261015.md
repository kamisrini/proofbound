# P7 external replay prerequisites — round-1 documentary review

**Reviewed commit:** `01677467b29f4232251d9194e3b47699db27e21f`
**Reviewer identity:** fresh independent non-author context (`Godel`)
**Reviewed artifacts and SHA-256:**

- `docs/decisions/VD-p7-external-replay-prerequisites-2026-10-15.md` — `01b9030403bfb072ce1e641831125b8690a2a066c3907ac2f11c341fb961c772`
- `docs/plans/P7-replay-lock-external-prerequisites-amendment-draft1.md` — `9a1ab9f17b19714402613c8e430d160c8fbc2ec9ad8af45a19a085868f258020`
- `docs/verification/p7-replay-lock-prerequisites/README.md` — `005f012021b8f7fe49320a0df5a8394bb778c9c0511e9e9165cf50d9805ce893`

## VERDICT: AMEND — documentary proposal only

The reviewer inspected the exact committed files; the accepted P7 plan, addendum, matrix,
first-slice proposal and verdicts; the active requirement; relevant P4 `twin`/`store` code
and SPECs; and the pinned embedded-Postgres source. HEAD and all three file hashes matched;
`git diff --check` was clean. The reviewer edited no files. Two pre-existing untracked paths
were present and untouched.

1. **HIGH — higher-level clean-clone criterion not replaced.** The amendment's replacement
   boundary (lines 13–20) does not explicitly amend the accepted P7 plan's replay-from-committed-
   artifacts criterion (lines 152–153). Define the prepared-offline clean-clone control it
   replaces; the original empty-cache/no-network claim cannot receive `PASS`.
2. **HIGH — Postgres provisioning conflicts with accepted first-slice scope.** The amendment
   (lines 52–69) and README (lines 24–35) stage an external tree through `BinariesDir`, while
   the accepted first slice (lines 16–34) requires a companion entry point to receive verified
   archive identity and bytes, stage inside the replay's temporary root, and clean up. State
   whether the companion receives an archive or verified tree, how input binds to reviewed
   digest, and where staging/cleanup occur; explicitly replace conflicting obligations.
3. **MEDIUM — setup-source validation relationships are implicit.** In the amendment (lines
   38–50), bind `go-toolchain` to `tar.gz`/`self`, `postgres` to `jar+txz`/the exact member,
   and `self` to equal download/consumed sizes and hashes. Specify the exact projector-manifest
   field mapping and require `platform` to match its build platform.
4. **MEDIUM — redirects lack a pre-reviewed acceptance rule.** The schema records only
   `source_url`, but the amendment rejects an "unrecorded origin" and records redirects only
   after setup. Reject all redirects or define a reviewed permitted chain/final-URL set and
   check the observed chain against it.
5. **LOW — binary-exclusion wording is broader than its examples.** The owner-direction
   decision (lines 9–12) can be read to exclude module `.zip` source archives. Name the
   excluded Go/Postgres binary archives and built executable as the policy boundary and
   expressly retain reviewed `.mod`/`.zip` source artifacts.

The reviewer found the `.mod`/`.zip` Git-tracked versus external-binary split otherwise clear.
The README correctly states that the setup manifest and runtime corpus do not yet exist and
that the local Postgres cache measurement is unverified. This verdict accepts no contract
change, L12/L17 result, runtime fixture, provenance clearance, full-gate result, requirement
satisfaction, product implementation, or implementation review. Resolve the findings and
obtain a fresh non-author verdict on the exact amended revision before implementation.
