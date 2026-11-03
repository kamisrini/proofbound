---
schema: vera.verdict.v1
verdict_id: p7-postgres-b1-unit2-build-round1-20261103
status: NEEDS_CHANGES
reviewed_commit: caf94c8f91e1a2b8edff10684e7ab33ee58bf949
findings:
  - id: 1
    severity: MED
    law: 2
    summary: license inventory under-counts — ldd-only methodology is blind to statically-compiled vendored third-party source (Ryu Boost-1.0, hyperloglog MIT, Spencer/Tcl regex, ISC, Gifford SHA2, WIDE, Allman, encoding tables); all-permissive/no-copyleft conclusion still holds
artifact_path: docs/verification/verdicts/p7-postgres-b1-unit2-build-round1-20261103.md
artifact_sha: 17b4d331e553a4354cf8432c37a30ff3e90ee96f52554792f965de12c11bc472
---

# P7 B1 unit-2 build + license inventory (round 1) — non-author verdict

## Provenance

- **Under review:** `a79d53d` (B1 unit-2 build evidence + license/notice inventory), at HEAD
  `caf94c8` (a notes-only commit atop `a79d53d`; the two evidence docs are the committed `a79d53d`
  state, unchanged by `caf94c8`).
- **Driving mechanism:** adversarial non-author review per CLAUDE.md Law 9 and the standing
  authorization (adversarial rounds via dispatched non-author reviewers).
- **Reviewer identity:** an independent reviewer dispatched via the Agent tool by the authoring
  session, which did NOT write the evidence or run the authoring build. The reviewer reproduced the
  central claims from its own from-source build in an isolated scratch dir. Law 9 satisfied
  (non-author).
- **Code/evidence author:** `VERA Maintainer` (authoring session).
- **Cleanroom redaction:** the reviewer worked in its own `mktemp` scratch; the report below
  contains no absolute paths or usernames.

## Verdict (verbatim reviewer report)

status: NEEDS_CHANGES
reviewed_commit: caf94c8f91e1a2b8edff10684e7ab33ee58bf949 (HEAD; a notes-only commit atop the work under review, a79d53d, which the two evidence docs reflect unchanged)

what I ran (own mktemp scratch, never in the clone; toolchain gcc 11.4.0 / bison 3.8.2 / flex 2.6.4 / make 4.3 / glibc 2.35 — matches the evidence):
- Independent HTTPS re-fetch of tarball + .sha256 from ftp.postgresql.org/pub/source/v16.4 with `curl --proto '=https' --max-redirs 0`: both http_code=200, effective==requested (no redirect). sha256 = 971766d6…3386d6f, size = 24765786 — matches the pin; published .sha256 matches.
- Repo verifier `scripts/p7-postgres-source-verify.sh …` → OK, exit 0.
- Extracted the verified tarball; `./configure --without-icu --without-readline` (zlib detected "inflate in -lz... yes"), `make -j28`, `make install` — all exit 0, 32 bin/ tools.
- `ldd bin/postgres` → ONLY linux-vdso, libz.so.1, libm.so.6, libc.so.6, ld-linux; `objdump -p` DT_NEEDED = libz/libm/libc. None of libicu/libreadline/libedit/libssl/libcrypto/libxml/liblz4/libzstd; no libintl/gettext. Central link-set claim reproduced on my build.
- `ldd bin/psql` → libpq, libm, libc, loader — NO libreadline/libedit. --without-readline leaves no GPLv3 anywhere.
- My bin/postgres sha256 = d2bb954fa908… vs author's 11d734c405f1… — differ, prefix path embedded; the evidence's "pin THIS build, not a cross-env anchor" caveat is accurate and present.
- Functional (short socket dir): initdb --locale-provider=libc OK; start OK; `select version()` = "PostgreSQL 16.4 … gcc 11.4.0, 64-bit"; datlocprovider = c; clean stop + rm; nothing left running.
- Validated the unit-3 socket note: install prefix alone is 110 chars (>107), so a deep-path socket overruns PG's 107-byte limit — note accurate.
- License text accuracy: source COPYRIGHT matches the inventory's PostgreSQL License block verbatim; /usr/include/zlib.h matches the zlib block verbatim; dpkg zlib1g = 1:1.2.11.dfsg-2ubuntu9.2 matches.
- ADVERSARIAL license hunt: grepped compiled dirs (src/backend, src/common, src/port) for non-PGDG/Regents copyright holders. Found distinct-license bundled third-party source compiled into the server: Ryu (Ulf Adams 2018; Apache-2.0 OR Boost-1.0) — d2s_srv.o/f2s_srv.o in libpgcommon_srv.a → postgres; hyperloglog (Hideaki Ohno 2013; MIT) — hyperloglog.o in backend; plus Henry Spencer regex (1998,1999), WIDE Project, Eric Allman, Todd C. Miller, ISC, Aaron Gifford SHA2, Scriptics/Tcl.

findings:
1. MED / Law 2 (+ slice-plan scope) — LICENSE INVENTORY UNDER-COUNTS. The inventory states its component set is "exactly what `ldd bin/postgres` reported … the complete set of third-party licenses that attach to the produced server" and concludes the third-party set is "PostgreSQL License + zlib" only. But `ldd` shows only dynamically-linked shared objects; it is blind to third-party source VENDORED INTO the tree and statically compiled in. Confirmed Ryu (Apache-2.0/Boost) and hyperloglog (MIT) compiled into the server, plus other distinct-copyright permissive notices. The slice plan scoped the "bundled/linked component set … inventory completely" — bundled, not just linked — so the inventory's own completeness claim is false. Fix: add the bundled-source components (and their verbatim notices) OR narrow the claim to "PG License governs the aggregate; dynamically-linked third parties beyond it = zlib only" and state bundled permissive sources are under that umbrella rather than asserting "PG License + zlib" is the exhaustive, fact-grounded set.

NON-blocking note: the substantive conclusions "all-permissive" and "no copyleft" remain TRUE — every bundled component found is permissive (Boost/Apache-2.0, MIT, BSD/ISC); readline/GPLv3 absent from postgres AND psql. The defect is completeness/accuracy of the enumeration, not the copyleft-free verdict, so the owner's go/no-go input is not misled.

scenarios I tried and could not break: source provenance (re-fetched, digest+size+no-redirect match, verifier OK); link set (only zlib+glibc, by ldd and DT_NEEDED); no-copyleft (readline absent from postgres and psql; no GPL/LGPL compiled in beyond platform glibc); runs on libc (datlocprovider=c, exact version string); binary-digest honesty (my digests differ exactly as caveated); socket-path note, toolchain versions, PG COPYRIGHT text, zlib license text, zlib version — all verified correct.

build-laws check:
- Law 1: PASS — neither new doc carries an `@generated` marker; both correctly hand-authored.
- Law 2: MED concern via the finding (the inventory asserts a "complete third-party set" that is incomplete). Separately, no drift seed from re-pinning the source digest: the full 64-hex digest appears 0 times in both docs; the build doc uses only the truncated form and links the manifest/unit-1 as the home (cite, don't copy).
- Law 9: PASS — independent non-author reproduced the central claims from source.

reviewer summary: I independently reproduced every central CLAIM of the build evidence — source verifies, minimal-permissive build links ONLY zlib+glibc, runs on libc, no copyleft in postgres or psql — on my own from-source build. The build evidence doc is solid and I would accept it alone. The license inventory under-counts: `ldd`-only methodology cannot see vendored statically-compiled source, so "complete third-party set = PG License + zlib" is false (Ryu/Boost, hyperloglog/MIT and others are compiled in). The all-permissive/no-copyleft bottom line survives, but the completeness claim and verbatim-notices list do not — NEEDS_CHANGES. High confidence.

## Disposition (author note)

Finding 1 is a real, correct undercount and is **accepted**. Fixed in the same revision of the
inventory (author may fix; Law 9 requires only that a non-author ACCEPT): the inventory now covers
**two linkage dimensions** — dynamic (`ldd`: zlib only) and statically-compiled vendored source,
enumerated from a per-file license-header scan of the compiled dirs, with Ryu (Boost-1.0, PG's
chosen option) and hyperloglog (MIT) reproduced verbatim and the remaining BSD/ISC-style components
cited by source path. The methodology sentence in the build-evidence doc was corrected to scope
`ldd` to dynamic linkage and point at the companion inventory for the static set. The
all-permissive / no-copyleft conclusion is unchanged and now better supported. A round-2 non-author
re-verify of the corrected inventory is owed before unit 2 is ACCEPTED (Law 9); `DEF-P7-001` stays
OPEN/BLOCKED.
