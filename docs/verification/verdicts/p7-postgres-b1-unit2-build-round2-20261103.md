---
schema: vera.verdict.v1
verdict_id: p7-postgres-b1-unit2-build-round2-20261103
status: ACCEPTABLE
reviewed_commit: a4642fe9ca4c3aa4aaffbf6555306ae3e6f172f8
findings:
  - id: 1
    severity: LOW
    law: 0
    summary: HyperLogLog MIT verbatim block is truncated with an ellipsis on the final warranty/liability sentence; authoritative per-file header is cited, so non-blocking
  - id: 2
    severity: LOW
    law: 0
    summary: CJK-encoding row lists holders that live in the UCS_to_*.pl generator scripts (not compiled objects); errs toward over-counting (safe direction), hedged
  - id: 3
    severity: LOW
    law: 0
    summary: IANA timezone code is statically compiled into bin/postgres and unlisted; public-domain so no obligation, but a one-line note would make the static set airtight
artifact_path: docs/verification/verdicts/p7-postgres-b1-unit2-build-round2-20261103.md
artifact_sha: 285decc1059ddbd30117dff4ccc6568e66e3c561afe7eff742fd0b6b5b2ac15e
---

# P7 B1 unit-2 build + license inventory (round 2) — non-author verdict

## Provenance

- **Under review:** `a4642fe` (the round-1 undercount fix: the two-dimension license/notice
  inventory + the build-doc methodology edit), together with the round-1-accepted build evidence.
  This closes the NEEDS_CHANGES finding of
  [`p7-postgres-b1-unit2-build-round1-20261103`](p7-postgres-b1-unit2-build-round1-20261103.md).
- **Driving mechanism:** adversarial non-author re-verify per Law 9 + the standing authorization.
- **Reviewer identity:** the same independent reviewer dispatched for round 1 (via the Agent tool by
  the authoring session), resumed with its context. It did NOT author the fix; it re-verified from
  its own fresh from-source build of the verified 16.4 source. Law 9 satisfied (non-author).
- **Code/evidence author:** `VERA Maintainer` (authoring session).
- **Cleanroom redaction:** reviewer worked in its own `mktemp` scratch; report contains no absolute
  paths or usernames.

## Verdict (verbatim reviewer report)

status: ACCEPTABLE
reviewed_commit: a4642fe9ca4c3aa4aaffbf6555306ae3e6f172f8 (HEAD; the undercount fix)

what I ran (fresh from-source rebuild in my own mktemp scratch, never in the clone; toolchain gcc 11.4.0 / bison 3.8.2 / flex 2.6.4 / make 4.3 / glibc 2.35):
- Re-fetched the 16.4 tarball + .sha256 over HTTPS (--max-redirs 0) and re-ran `scripts/p7-postgres-source-verify.sh` → OK sha256=971766d6…3386d6f size=24765786. Re-extracted, `./configure --without-icu --without-readline`, `make -j`, `make install` — all clean (rebuilt because round-1 scratch was already cleaned).
- Re-read both changed docs. Build-doc edit: the ldd sentence now correctly scopes itself to "dynamic-linkage set" and points at the companion inventory's Method for the static set — accurate and minimal.
- Verified the inventory's TWO-dimension method against my build:
  * Dynamic set: `ldd bin/postgres` / `objdump -p` DT_NEEDED = libz + libm + libc + loader only; no icu/readline/edit/ssl/crypto/xml/lz4/zstd/intl. Matches the "zlib only + platform glibc" claim.
  * Static vendored set: ran the inventory's own Method grep. The complete set of distinct non-PGDG/Regents copyright holders it surfaces is EXACTLY what the table enumerates: Ulf Adams (Ryu), Hideaki Ohno (HLL), Henry Spencer + Scriptics (regex), Todd C. Miller (strlcpy), ISC (inet), Aaron D. Gifford (SHA-2), WIDE Project, Eric P. Allman + BSD (misc), SKC/SRA (CJK). No distinct-copyright holder in the compiled static dirs is missing from the table.
- Confirmed the listed components ARE compiled into bin/postgres: `ar t libpgcommon_srv.a` lists d2s_srv.o/f2s_srv.o (Ryu); hyperloglog.o compiled; `nm bin/postgres` shows pg_sha256_* (in-tree sha2.c used because --without-openssl); regex/inet/strlcpy paths carry the claimed copyrights.
- Verbatim-notice accuracy: Ryu block matches src/common/d2s.c exactly incl. the "Boost license (not the Apache license)" lead + dual grant + AS-IS; Boost-vs-Apache nuance stated correctly. HLL MIT block matches src/backend/lib/hyperloglog.c.
- HUNTED for a still-missing obligation-bearing component:
  * Snowball stemmer (BSD) — a loose grep "hit" (14 symbols) proved via `nm --defined-only` to be ALL false positives ("sy-stem_"); snowball lives entirely in loadable dict_snowball.so, NOT statically in bin/postgres, and libstemmer carries no distinct header. Correctly excluded.
  * IANA timezone — localtime.o IS statically in bin/postgres; files are public-domain → no notice-preservation obligation, so omission doesn't break the "every obligation is notice-preservation" claim.
  * Unicode data — no Unicode, Inc. notice in the compiled source (PG's tables are PGDG). Nothing missing.
- Cleanup: no postgres process left running; rebuilt scratch removed; repo working tree unchanged (only the two pre-existing untracked items, not mine); inventory carries no @generated marker.

findings:
No blocking defects found. The round-1 MED undercount is RESOLVED. LOW/cosmetic only:
1. LOW / completeness — the HyperLogLog MIT verbatim block is truncated with "…" on the final warranty/liability sentence; a redistributor copying from this doc alone would get incomplete MIT text. Mitigated: the inventory names the per-file header as authoritative home. Non-blocking.
2. LOW / minor over-inclusion — the CJK-encoding row lists holders (Project X0213, earthian, I'O, NARITA Tomio) that live in the UCS_to_*.pl generator scripts (generate .map data, not compiled into bin/postgres); my compiled-dir sweep surfaced only SKC/SRA. Errs toward OVER-counting (safe direction for a license inventory), hedged. Does not undercount or affect the conclusion.
3. LOW / completeness — IANA timezone code is statically compiled into bin/postgres and not listed; public-domain so no obligation, but a one-line "public-domain, no preservation obligation" mention would make the static-set enumeration airtight. Non-blocking.

scenarios I tried and could not break:
- Enumeration completeness: my independent Method-grep produced the same distinct-copyright set the table lists; no obligation-bearing component statically in bin/postgres is missing.
- Snowball red herring: provably NOT in the server binary (nm --defined-only), so excluding it is correct.
- Verbatim fidelity: Ryu and HLL blocks match the real 16.4 source headers; Boost-not-Apache nuance correct.
- Compiled-in proof: Ryu (ar membership), HLL (.o), SHA-2 (nm symbols), regex/inet/strlcpy (paths + headers) all confirmed present.
- All-permissive / no-copyleft: every component permissive (Boost-1.0, MIT, BSD-3, ISC, Tcl/BSD, zlib, PG License); no GPL/LGPL compiled in; readline absent from postgres AND psql; glibc flagged platform-only. Conclusion holds, now correctly supported by the two-dimension method.

build-laws check:
- Law 1: PASS — rewritten inventory carries no `@generated` marker; correctly hand-authored; nothing claims to be derived that isn't.
- Law 2: PASS — the fix makes the per-file source headers (not the table) the authoritative home ("cite, don't become a second source of truth that can drift"), gives the reproducible Method grep, and reproduces verbatim ONLY the two obligation-bearing notices (Ryu, HLL) citing paths for the rest. Both verbatim blocks match the digest-locked source, so no material drift seed. Source digest still cited, not re-pinned (full 64-hex absent). Correct Law 2 posture.
- Law 9: PASS — independent non-author round-2 re-verification reproduced from my own from-source build; satisfies the gate for unit 2 acceptance.

reviewer summary: The round-1 undercount is genuinely fixed. The inventory now covers both linkage dimensions, and its enumeration of statically-compiled vendored source matches, component-for-component, an independent header sweep of my own build of the verified 16.4 source; the two obligation-bearing notices (Ryu Boost-1.0, HyperLogLog MIT) are reproduced faithfully with the Boost-vs-Apache nuance correct, and the remainder are cited to their authoritative per-file headers (correct Law 2 posture). I specifically hunted the most likely still-missing component (snowball) and proved it is NOT in the server binary. The all-permissive / no-copyleft conclusion holds and is now properly supported. Only cosmetic LOW items remain — none blocking. High confidence: ACCEPTABLE; unit 2's build evidence (accepted round 1) and the corrected license/notice inventory are both sound.

## Disposition (author note)

Accepted at the reviewed commit `a4642fe`. The round-1 MED undercount is closed: the reviewer
independently reproduced the inventory's Method grep and got a component-for-component match on its
own from-source build, confirmed each listed component is compiled into `bin/postgres` (`ar`/`nm`),
verified the Ryu/HLL verbatim blocks against the real source, and proved the most-likely missing
candidate (snowball) is in a loadable `.so` rather than the server binary. **Unit 2 (build +
license/notice inventory) is non-author ACCEPTED.**

The three LOW findings are non-blocking and, following the standing precedent for non-false-accept
cosmetic LOWs (not worth moving the tree off an accepted commit and re-opening the non-author cycle),
are recorded here as **optional future cleanup** rather than fixed under this acceptance:
- (1) complete the HyperLogLog MIT verbatim tail;
- (2) narrow the CJK row to the compiled holders (SKC/SRA) or keep the hedged over-inclusion;
- (3) add a one-line public-domain note for the statically-compiled IANA timezone code.

`DEF-P7-001` stays OPEN/BLOCKED pending unit 3 (prepared-offline runtime fixture binding a real
`vera.replay.v1` proof over this server tree, honoring the short-Unix-socket-dir constraint).
