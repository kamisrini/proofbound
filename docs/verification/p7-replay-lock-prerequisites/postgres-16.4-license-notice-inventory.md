# P7 B1 — PostgreSQL 16.4 (minimal-permissive build) component license/notice inventory

**Status:** component license/notice inventory for B1 unit 2. Reviewable non-binary evidence; not an
accepted unit until a non-author verdict says so (Law 9).
**Build it describes:** [`postgres-source-16.4-build-20261103.md`](postgres-source-16.4-build-20261103.md)
(`./configure --without-icu --without-readline`, zlib default-on).
**Decision:** [`VD-p7-postgres-source-route-2026-11-01`](../../decisions/VD-p7-postgres-source-route-2026-11-01.md)
(B1). **Deferral:** [`DEF-P7-001`](../../deferrals.tsv).

## Method (grounded in the actual link set, not configure flags)

The components below are exactly what `ldd bin/postgres` reported for the produced server binary,
cross-checked against the opt-in library list — **no** `libicu / libreadline / libedit / libssl /
libcrypto / libxml / liblz4 / libzstd` is linked. The inventory is therefore the complete set of
third-party licenses that attach to the produced server; it is not a superset copied from a default
build. System glibc runtime (`libc`, `libm`, dynamic loader) is the platform C library, not a
bundled/redistributed component, and is listed for completeness only.

## Inventory

| component | role in the build | version | license | obligation if redistributed |
|---|---|---|---|---|
| **PostgreSQL** | the server/client we built | 16.4 | **PostgreSQL License** (BSD/MIT-style, permissive) | reproduce the copyright + permission notice |
| **zlib** | dynamically linked (`libz.so.1`); compression | system `1:1.2.11.dfsg-2ubuntu9.2` (`zlib1g`) | **zlib license** (permissive) | reproduce the zlib notice; do not misrepresent origin |
| glibc runtime (`libc`, `libm`, loader) | platform C library (not bundled) | GLIBC 2.35 | LGPL-2.1+ (system component) | n/a — platform library, not redistributed with the fixture |

**Net license posture:** the produced binary's third-party license set is **all-permissive**
(PostgreSQL License + zlib license). There is **no copyleft** component linked (readline/GPLv3 was
deliberately excluded; it links only into the unused `psql` client, which is not part of the server
the fixture runs). No OpenSSL, ICU/Unicode, libxml, lz4, or zstd notices apply.

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

## Scope / honest limits

- This inventory covers the **produced server** under the minimal-permissive config. If a future
  build re-enables ICU, OpenSSL, readline, libxml, lz4 or zstd, this inventory does **not** apply and
  must be regenerated from that build's `ldd` output.
- zlib here is the **system** `zlib1g` resolved at link time, not a copy vendored into this repo. A
  redistributable package would either depend on the platform zlib or bundle a pinned zlib, and must
  then carry the zlib notice accordingly.
- This slice does **not** redistribute the binary (slice non-goal). The inventory is the evidence
  that *would* make redistribution lawful; the redistribution go/no-go remains the owner's and is out
  of scope here.
