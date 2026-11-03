# P7 B1 — PostgreSQL 16.4 (minimal-permissive build) component license/notice inventory

**Status:** component license/notice inventory for B1 unit 2. Reviewable non-binary evidence; not an
accepted unit until a non-author verdict says so (Law 9).
**Build it describes:** [`postgres-source-16.4-build-20261103.md`](postgres-source-16.4-build-20261103.md)
(`./configure --without-icu --without-readline`, zlib default-on).
**Decision:** [`VD-p7-postgres-source-route-2026-11-01`](../../decisions/VD-p7-postgres-source-route-2026-11-01.md)
(B1). **Deferral:** [`DEF-P7-001`](../../deferrals.tsv).

> **Round-1 correction (2026-11-03).** A first non-author review (NEEDS_CHANGES,
> `p7-postgres-b1-unit2-build-round1-20261103`) found the original inventory under-counted: it
> derived the component set from `ldd` alone, which sees only **dynamically-linked** shared objects
> and is blind to third-party source **vendored into the PostgreSQL tree and statically compiled**
> into the binary (e.g. Ryu, hyperloglog). The reviewer's all-permissive / no-copyleft conclusion
> held, but the completeness claim did not. This revision fixes the methodology and enumerates the
> bundled components. See the round-1 verdict.

## Method (two linkage dimensions, not just `ldd`)

A produced binary carries third-party licenses by **two** paths, and a complete inventory must cover
both:

1. **Dynamically linked** shared libraries — enumerated from `ldd bin/postgres` / `objdump -p`
   (`DT_NEEDED`). For this minimal-permissive build that is **zlib** only, plus the platform glibc
   runtime (`libc`, `libm`, loader — platform C library, not a bundled/redistributed component).
   Confirmed absent: `libicu / libreadline / libedit / libssl / libcrypto / libxml / liblz4 /
   libzstd / libintl(NLS)`.
2. **Statically compiled vendored source** — third-party code that lives inside the PostgreSQL
   source tree and is compiled directly into the binary; `ldd` cannot see it. Enumerated by scanning
   per-file copyright/license headers in the compiled source dirs
   (`grep -rhiE 'Copyright' src/backend src/common src/port src/include`, filtering out PGDG/Regents)
   and confirming the object is linked (`ar t libpgcommon_srv.a`, compiled `.o`s). The authoritative
   home of each notice is its source-file header; this inventory cites the path and reproduces the
   notices that carry explicit preservation obligations.

PostgreSQL publishes the aggregate distribution under a **single** `COPYRIGHT` file (the PostgreSQL
License). The bundled third-party files below are distributed under that umbrella **and** each
retains its own permissive notice, which a redistributor must preserve.

## Inventory

### Dynamically linked

| component | role | version | license | obligation if redistributed |
|---|---|---|---|---|
| **PostgreSQL** | the server/client we built | 16.4 | **PostgreSQL License** (BSD/MIT-style, permissive) | reproduce the copyright + permission notice |
| **zlib** | `libz.so.1`, compression | system `1:1.2.11.dfsg-2ubuntu9.2` (`zlib1g`) | **zlib license** (permissive) | reproduce the zlib notice; do not misrepresent origin |
| glibc runtime (`libc`, `libm`, loader) | platform C library (not bundled) | GLIBC 2.35 | LGPL-2.1+ (system component) | n/a — platform library, not redistributed with the fixture |

### Statically compiled vendored source (in the PostgreSQL tree, compiled into `bin/postgres`)

All permissive; all carry a notice-preservation obligation. Confirmed compiled into the server
(object present in a linked archive / backend).

| component | source path | copyright | license | reproduced below |
|---|---|---|---|---|
| **Ryu** float output | `src/common/d2s.c`, `f2s.c`, `ryu_common.h`, `d2s_intrinsics.h`, `d2s_full_table.h` | 2018 Ulf Adams | **Boost-1.0** (PG's chosen option; upstream offers Apache-2.0 OR Boost) | ✓ |
| **HyperLogLog** | `src/backend/lib/hyperloglog.c` | 2013 Hideaki Ohno | **MIT** | ✓ |
| **Spencer regex** (via Tcl) | `src/backend/regex/*` | 1998,1999 Henry Spencer; 1998 Scriptics Corp. | BSD-style / Tcl-style (permissive) | cite header |
| **strlcpy/strlcat** | `src/port/strlcpy.c` etc. | 1998 Todd C. Miller | ISC/OpenBSD (permissive) | cite header |
| **inet_net_pton** / network | `src/backend/utils/adt/inet_net_pton.c` | 1996,1999 Internet Software Consortium; 2004 ISC | ISC (permissive) | cite header |
| **SHA-2 fallback** | `src/common/sha2.c` (used because `--without-openssl`) | 2000-2001 Aaron D. Gifford | BSD-3-clause (permissive) | cite header |
| **WIDE Project** net code | `src/backend/utils/adt/*` | 1995-1998 WIDE Project | BSD-style (permissive) | cite header |
| misc BSD utility | e.g. `src/backend/...` (assert/qsort/snprintf lineage) | 1983,1995,1996 Eric P. Allman; Regents | BSD-style (permissive) | cite header |
| CJK encoding tables | `src/backend/utils/mb/**` | Project X0213; earthian@tama.or.jp; I'O; NARITA Tomio; SKC/SRA | permissive (per-file) | cite header |

**Completeness note:** the vendored list is the set of distinct non-PGDG/Regents copyright holders
surfaced by a header scan of the compiled dirs; the per-file headers are the authoritative,
reproducible home (re-run the `grep` in **Method**). PostgreSQL's documentation *Acknowledgements*
and the source-file headers are the upstream references. No copyleft (GPL/LGPL) third-party source
was found compiled into the server (glibc, LGPL, is the platform library only).

## Net license posture

The produced binary's third-party license set is **all-permissive** — PostgreSQL License + zlib
(dynamic) + a set of BSD/MIT/ISC/Boost-licensed vendored sources (static). There is **no copyleft**
component: readline (GPLv3) was deliberately excluded and is absent from both `postgres` and `psql`
(`ldd bin/psql` shows no readline/libedit); no OpenSSL, ICU/Unicode, libxml, lz4, zstd, or NLS. Every
obligation is notice-preservation (retain the copyright + permission text); none is reciprocal.

## Required notices (verbatim)

### PostgreSQL License (source `COPYRIGHT`)

```
PostgreSQL Database Management System
(formerly known as Postgres, then as Postgres95)

Portions Copyright (c) 1996-2024, PostgreSQL Global Development Group
Portions Copyright (c) 1994, The Regents of the University of California

Permission to use, copy, modify, and distribute this software and its
documentation for any purpose, without fee, and without a written agreement
is hereby granted, provided that the above copyright notice and this
paragraph and the following two paragraphs appear in all copies.

IN NO EVENT SHALL THE UNIVERSITY OF CALIFORNIA BE LIABLE TO ANY PARTY FOR
DIRECT, INDIRECT, SPECIAL, INCIDENTAL, OR CONSEQUENTIAL DAMAGES, INCLUDING
LOST PROFITS, ARISING OUT OF THE USE OF THIS SOFTWARE AND ITS
DOCUMENTATION, EVEN IF THE UNIVERSITY OF CALIFORNIA HAS BEEN ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.

THE UNIVERSITY OF CALIFORNIA SPECIFICALLY DISCLAIMS ANY WARRANTIES,
INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY
AND FITNESS FOR A PARTICULAR PURPOSE.  THE SOFTWARE PROVIDED HEREUNDER IS
ON AN "AS IS" BASIS, AND THE UNIVERSITY OF CALIFORNIA HAS NO OBLIGATIONS TO
PROVIDE MAINTENANCE, SUPPORT, UPDATES, ENHANCEMENTS, OR MODIFICATIONS.
```

### zlib license (from the linked `zlib1g` 1.2.11)

```
Copyright (C) 1995-2017 Jean-loup Gailly and Mark Adler

This software is provided 'as-is', without any express or implied
warranty.  In no event will the authors be held liable for any damages
arising from the use of this software.

Permission is granted to anyone to use this software for any purpose,
including commercial applications, and to alter it and redistribute it
freely, subject to the following restrictions:

1. The origin of this software must not be misrepresented; you must not
   claim that you wrote the original software. If you use this software
   in a product, an acknowledgment in the product documentation would be
   appreciated but is not required.
2. Altered source versions must be plainly marked as such, and must not be
   misrepresented as being the original software.
3. This notice may not be removed or altered from any source distribution.

Jean-loup Gailly        Mark Adler
jloup@gzip.org          madler@alumni.caltech.edu
```

### Ryu — `src/common/d2s.c` (PostgreSQL uses the Boost-1.0 option)

```
This is a modification of code taken from github.com/ulfjack/ryu under the
terms of the Boost license (not the Apache license). The original copyright
notice follows:

Copyright 2018 Ulf Adams

The contents of this file may be used under the terms of the Apache
License, Version 2.0.
    (See accompanying file LICENSE-Apache or copy at
     http://www.apache.org/licenses/LICENSE-2.0)
Alternatively, the contents of this file may be used under the terms of the
Boost Software License, Version 1.0.
    (See accompanying file LICENSE-Boost or copy at
     https://www.boost.org/LICENSE_1_0.txt)

Unless required by applicable law or agreed to in writing, this software is
distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.
```

### HyperLogLog — `src/backend/lib/hyperloglog.c` (MIT; copyright terms of Ohno's original)

```
Copyright (c) 2013 Hideaki Ohno <hide.o.j55{at}gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the 'Software'), to
deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER …
```

The remaining vendored components (Spencer/Tcl regex, Todd C. Miller strlcpy, ISC inet code,
Aaron D. Gifford SHA-2, WIDE Project, Eric P. Allman, CJK encoding tables) are BSD/ISC-style
notice-preservation licenses; their authoritative text is the per-file header at the paths above and
must be preserved verbatim in a redistribution.

## Scope / honest limits

- This inventory covers the **produced server** under the minimal-permissive config. If a future
  build re-enables ICU, OpenSSL, readline, libxml, lz4 or zstd, this inventory does **not** apply and
  must be regenerated.
- The dynamic set is grounded in `ldd`/`objdump -p`; the vendored set is grounded in a per-file
  header scan of the compiled dirs (reproducible via the **Method** `grep`). The per-file headers —
  not this table — are the authoritative home of each notice (Law 2: cite, don't become a second
  source of truth that can drift).
- zlib here is the **system** `zlib1g` resolved at link time, not vendored into this repo. A
  redistributable package would either depend on the platform zlib or bundle a pinned zlib, and
  carry the zlib notice accordingly.
- This slice does **not** redistribute the binary (slice non-goal). The inventory is the evidence
  that *would* make redistribution lawful (preserve the notices above); the redistribution go/no-go
  remains the owner's and is out of scope here.
