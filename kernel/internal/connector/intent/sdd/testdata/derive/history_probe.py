"""Change-scoped, history-native mapping probe over a real OpenSpec repository (draft 9 §0, §4.2).

Walks the provider's own walk — every commit of `rev-list --all` (the git connector's arguments),
unsimplified, date order — reads the change folders AT EACH COMMIT with draft 9 §4.2's parser
(OpenSpec's parseDeltaSpec rules, ECMAScript regex semantics made explicit), and prints every count
draft 9 §0 cites.

Usage:
  python3 -I history_probe.py <openspec-clone>      # print the §0 counts
  python3 -I history_probe.py --pins <testdata-dir>     # regenerate <testdata-dir>/expected-obligations.json
  python3 -I history_probe.py --vectors <testdata-dir>  # regenerate <testdata-dir>/id-vectors.json
The --pins / --vectors modes need no git repository.
"""
import collections, json, os, re, subprocess, sys

MODE = sys.argv[1] if sys.argv[1] in ("--pins", "--vectors") else None
repo = sys.argv[2] if MODE else sys.argv[1]
REGULAR = ("100644", "100755")  # regular-file modes; 120000 symlink and 160000 gitlink are not
def git(*a):
    return subprocess.run(["git", "-C", repo, *a], capture_output=True, check=True).stdout

# Draft 8 section 4.2: OpenSpec's delta READER (src/core/parsers/requirement-blocks.ts parseDeltaSpec,
# code-fence.ts) at f085c850, with ECMAScript regex semantics made explicit: \s is the JS set below,
# `.` excludes JS line terminators, and the only case-insensitive match (the requirement header's /i,
# non-unicode mode) folds ASCII letters only.
WS = "\t\n\v\f\r    -     　﻿"
W, DOT = f"[{WS}]", "[^\n\r  ]"
def ci(word): return "".join(f"[{c.upper()}{c.lower()}]" if c.isalpha() else re.escape(c) for c in word)
SECTION_LINE = re.compile(rf"^##{W}+({DOT}+)$")
TOP = re.compile(rf"^##{W}+")
REQ = re.compile(rf"^###{W}*{ci('Requirement:')}{W}*({DOT}+){W}*$")
REMOVED_BULLET = re.compile(rf"^{W}*[-*+]{W}*`?###{W}*Requirement:{W}*({DOT}+?)`?{W}*$")
FROM = re.compile(rf"^{W}*[-*+]?{W}*FROM:{W}*`?###{W}*Requirement:{W}*({DOT}+?)`?{W}*$")
TO = re.compile(rf"^{W}*[-*+]?{W}*TO:{W}*`?###{W}*Requirement:{W}*({DOT}+?)`?{W}*$")
OPEN = re.compile(rf"^{W}*(`{{3,}}|~{{3,}})({DOT}*)")
CLOSE = re.compile(rf"^{W}*(`{{3,}}|~{{3,}}){W}*$")
TITLES = {"added requirements": "ADDED", "modified requirements": "MODIFIED",
          "removed requirements": "REMOVED", "renamed requirements": "RENAMED"}
JS_WS = set("\t\n\v\f\r       　﻿") | {chr(c) for c in range(0x2000, 0x200b)}
GO_SPACE = {"\t", "\n", "\v", "\f", "\r", " ", "\x85", "\xa0", " ", " ", " ", " ", " ", "　"} | {chr(c) for c in range(0x2000, 0x200b)}
def js_trim(t): return t.strip("".join(JS_WS))
def blank(t): return all(ch in JS_WS or ch in GO_SPACE for ch in t)
def fence_mask(lines):
    mask, active = [False] * len(lines), None
    for i, l in enumerate(lines):
        if active is None:
            m = OPEN.match(l)
            if m and not (m.group(1)[0] == "`" and "`" in m.group(2)):
                active = (m.group(1)[0], len(m.group(1))); mask[i] = True
            continue
        mask[i] = True
        m = CLOSE.match(l)
        if m and m.group(1)[0] == active[0] and len(m.group(1)) >= active[1]:
            active = None
    return mask
def naive_fence_mask(lines):
    """The NAIVE TOGGLE (draft 8 §13): the mutant SDD-INV-39 must kill. Every line matching
    ^[ \t]*(```|~~~) flips an in-fence flag and is itself masked; marker, length and info string
    are ignored."""
    mask, inside = [], False
    for l in lines:
        if re.match(r"^[ \t]*(```|~~~)", l):
            inside = not inside; mask.append(True); continue
        mask.append(inside)
    return mask
def name_of(raw): return js_trim(re.sub(r"[ \t]+#+[ \t]*$", "", raw))
def split_lines(text):
    if text.startswith("﻿"): text = text[1:]
    return re.sub(r"\r\n?", "\n", text).split("\n")
def title_kind(raw):
    t = name_of(raw)
    return TITLES.get(t.lower()) if t.isascii() else None
def parse_delta(text, masked=True, unpaired=None):
    """[(SEC, name, statement)] for one delta file. masked: True (OpenSpec mask), False (no mask),
    or "naive" (the naive toggle). `unpaired` collects unpaired RENAMED lines."""
    lines = split_lines(text)
    mask = naive_fence_mask(lines) if masked == "naive" else fence_mask(lines) if masked else [False] * len(lines)
    heads = [(i, title_kind(m.group(1))) for i, l in enumerate(lines) if not mask[i] and (m := SECTION_LINE.match(l))]
    out = []
    for n, (start, kind) in enumerate(heads):
        if kind is None: continue
        end = heads[n + 1][0] if n + 1 < len(heads) else len(lines)
        body = [(i, lines[i]) for i in range(start + 1, end) if not mask[i]]
        if kind == "RENAMED":
            pending = None
            for _, l in body:
                f, t = FROM.match(l), TO.match(l)
                if f:
                    if pending is not None and unpaired is not None: unpaired.append(pending)
                    pending = name_of(f.group(1))
                elif t:
                    if pending is None:
                        if unpaired is not None: unpaired.append(name_of(t.group(1)))
                        continue
                    out.append((kind, pending, l)); pending = None
            if pending is not None and unpaired is not None: unpaired.append(pending)
            continue
        entry = None  # [name, statement-or-None, open]
        def close():
            if entry: out.append((kind, entry[0], entry[1] if entry[1] is not None else entry[0]))
        for _, l in body:
            r = REQ.match(l)
            if r:
                close(); entry = [name_of(r.group(1)), None, True]; continue
            if TOP.match(l):
                if entry: entry[2] = False
                continue
            if kind == "REMOVED" and (bm := REMOVED_BULLET.match(l)):
                close(); entry = None  # a REMOVED bullet is its own entry and ends the open block (§4.2)
                out.append((kind, name_of(bm.group(1)), name_of(bm.group(1)))); continue
            if entry and entry[2] and entry[1] is None:
                if l.startswith("#"): entry[2] = False
                elif not blank(l): entry[1] = l
        close()
    return out
if MODE == "--pins":
    base = os.path.join(repo, "fenced", "openspec", "changes")
    cases = {}
    for ch in sorted(os.listdir(base)):
        specs = os.path.join(base, ch, "specs")
        for cap in sorted(os.listdir(specs)):
            rel = os.path.relpath(os.path.join(specs, cap, "spec.md"), repo)
            text = open(os.path.join(repo, rel), encoding="utf-8").read()
            lists = {k: [list(x[:2]) for x in parse_delta(text, m)] for k, m in
                     (("openspec_mask", True), ("naive_toggle", "naive"), ("unmasked", False))}
            lists["discriminates"] = lists["openspec_mask"] != lists["naive_toggle"] or lists["openspec_mask"] != lists["unmasked"]
            cases[rel] = lists
    out = {"@generated": "@generated by python3 -I kernel/internal/connector/intent/sdd/testdata/derive/history_probe.py --pins kernel/internal/connector/intent/sdd/testdata (do not hand-edit)",
           "note": "Draft 8 SDD-INV-39: obligations (section, name) per fenced case under OpenSpec's mask (the contract), the naive toggle (draft 8 §13; the mutant), and no mask. Regenerate; never hand-edit.",
           "schema": "proofbound.sdd-expected-obligations.v1",
           "cases": cases}
    with open(os.path.join(repo, "expected-obligations.json"), "w") as f:
        json.dump(out, f, indent=2); f.write("\n")
    sys.exit(0)


# Draft 8 section 4.2 / 4.3 identity functions (byte-wise ASCII; no Unicode case mapping).
import hashlib
ALNUM_UP, ALNUM_LO = b"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", b"abcdefghijklmnopqrstuvwxyz0123456789"
def _map(raw, keep, lower):
    out = bytearray()
    for b in raw:
        if lower and 65 <= b <= 90: b += 32
        elif not lower and 97 <= b <= 122: b -= 32
        out += bytes([b]) if b in keep else b"-"
    return re.sub(rb"-+", b"-", bytes(out)).strip(b"-")
def namepart(name):
    np = _map(name.encode(), ALNUM_UP, False)
    if not np: raise ValueError("empty-name-part")
    if len(np) > 28: np = np[:21].strip(b"-") + b"-" + hashlib.sha256(name.encode()).hexdigest()[:6].upper().encode()
    return np.decode()
def slug(s):
    out = _map(s.encode(), ALNUM_LO, True)[:40].strip(b"-")
    if len(out) < 2: raise ValueError("slug-too-short")
    return out.decode()
def suffix(kit, kind, key):
    h = hashlib.sha256(kit.encode() + b"\0" + kind.encode() + b"\0" + key.encode()).digest()[:6]
    return "".join("abcdefghijklmnopqrstuvwxyz0123456789"[b % 36] for b in h)
def change_intent_id(name, kit="openspec"): return f"CI-{slug(name)}-{suffix(kit, 'change-intent', name)}"
def requirement_id(key, digest, kit="openspec"): return f"BR-{slug(key)}-{suffix(kit, 'requirement', key + chr(0) + digest)}"
if MODE == "--vectors":
    vectors = []
    for sec, name in (("ADD", "Export X"), ("MOD", "expoſe"), ("ADD", "ıd check"), ("REM", "Kelvin limit"),
                      ("ADD", "A requirement whose name is far longer than twenty-eight bytes"), ("ADD", "!!!"), ("REN", "Café menu")):
        try: vectors.append({"kind": "obligation", "section": sec, "name": name, "id": f"O-{sec}-{namepart(name)}"})
        except ValueError as e: vectors.append({"kind": "obligation", "section": sec, "name": name, "error": str(e)})
    for name in ("add-login", "expoſe-api", "Kelvin", "a", "ünïcode-only", "x" * 60):
        try: vectors.append({"kind": "change-intent", "name": name, "id": change_intent_id(name)})
        except ValueError as e: vectors.append({"kind": "change-intent", "name": name, "error": str(e)})
    digest = hashlib.sha256(b"manifest").hexdigest()
    for key in ("add-login/auth", "expoſe-api/core"):
        vectors.append({"kind": "requirement", "key": key, "artifact_sha256": digest, "id": requirement_id(key, digest)})
    out = {"@generated": "@generated by python3 -I kernel/internal/connector/intent/sdd/testdata/derive/history_probe.py --vectors kernel/internal/connector/intent/sdd/testdata (do not hand-edit)",
           "note": "Draft 8 SDD-INV-7: id vectors for slug / namepart / suffix (kit openspec). Non-ASCII inputs pin byte-wise ASCII mapping: no Unicode case folding.",
           "schema": "proofbound.sdd-id-vectors.v1", "vectors": vectors}
    with open(os.path.join(repo, "id-vectors.json"), "w") as f:
        json.dump(out, f, indent=2, ensure_ascii=False); f.write("\n")
    sys.exit(0)

commits = git("rev-list", "--reverse", "--date-order", "--exclude=refs/stash", "--exclude=refs/notes/*",
              "--exclude=refs/replace/*", "--all").decode().split()  # the provider's walk (draft 8 section 3)
cat = subprocess.Popen(["git", "-C", repo, "cat-file", "--batch"], stdin=subprocess.PIPE, stdout=subprocess.PIPE)
def blob(rev, path):
    cat.stdin.write(f"{rev}:{path}\n".encode()); cat.stdin.flush()
    head = cat.stdout.readline().decode().split()
    if len(head) < 3 or head[1] != "blob":
        return None
    data = cat.stdout.read(int(head[2])); cat.stdout.read(1)
    return data.decode("utf-8", "replace")

def main_reqs(text):
    return {m.group(1) for l in text.split("\n") if (m := REQ.match(l))}

stats = collections.Counter()
reuse_commits = collections.defaultdict(set)  # name -> commits where it is both active and archived
twice_commits = collections.defaultdict(set)  # name -> commits where it is archived twice
renamed_entries = 0
observed_active = set(); archived_names = collections.Counter(); prev = {}  # name -> {req: stmt}
zero_target = set(); missing_target = []; no_stmt = set()
disappear = []; meaning = []
for c in commits:
    paths = git("ls-tree", "-r", "--name-only", c, "--", "openspec").decode().split("\n")
    active = collections.defaultdict(list); archived_here = set(); mains = {}
    for p in paths:
        m = re.match(r"^openspec/changes/archive/(\d{4}-\d{2}-\d{2})-([^/]+)/", p)
        if m: archived_here.add(m.group(2)); continue
        m = re.match(r"^openspec/changes/([^/]+)/(.*)$", p)
        if m and m.group(1) != "archive": active[m.group(1)].append(m.group(2))
        m = re.match(r"^openspec/specs/([^/]+)/spec\.md$", p)
        if m: mains[m.group(1)] = p
    for name in set(active) & archived_here:
        reuse_commits[name].add(c[:8])
    dirs = {p.split("/")[3] for p in paths if p.startswith("openspec/changes/archive/") and len(p.split("/")) > 4}
    by_name = collections.Counter(re.sub(r"^\d{4}-\d{2}-\d{2}-", "", d) for d in dirs if re.match(r"^\d{4}-\d{2}-\d{2}-", d))
    for name, k in by_name.items():
        if k > 1: twice_commits[name].add(c[:8])
    for name in archived_here:
        archived_names[name] = max(archived_names[name], 1)
    current = {}
    for name, files in active.items():
        observed_active.add(name)
        reqs = {}; entries = 0
        for f in files:
            m = re.match(r"^specs/([^/]+)/spec\.md$", f)
            if not m: continue
            text = blob(c, f"openspec/changes/{name}/{f}") or ""
            renamed_entries += sum(1 for sec, _, _ in parse_delta(text) if sec == "RENAMED")
            for section, rname, stmt in parse_delta(text):
                entries += 1
                if section in ("ADDED", "MODIFIED"):
                    reqs[(m.group(1), rname)] = stmt
                    if stmt == rname: no_stmt.add((name, rname))
                if section in ("MODIFIED", "REMOVED"):
                    mp = mains.get(m.group(1))
                    mt = main_reqs(blob(c, mp) or "") if mp else set()
                    if rname not in mt: missing_target.append((c[:8], name, section, m.group(1), rname))
        if entries == 0: zero_target.add(name)
        if name in prev:
            for k, s in prev[name].items():
                if k not in reqs: disappear.append((c[:8], name, k))
                elif reqs[k] != s: meaning.append((c[:8], name, k))
        current[name] = reqs
    prev = {**prev, **current}
    stats["commits"] += 1

all_archived = set(archived_names)
never_active = all_archived - observed_active
print(f"commits walked (the provider's unsimplified --all walk): {stats['commits']}")
print(f"changes ever observed active:       {len(observed_active)}")
print(f"changes ever archived:              {len(all_archived)}")
print(f"archived but NEVER observed active: {len(never_active)}  e.g. {sorted(never_active)[:3]}")
print(f"names both active and archived at a commit: {len(reuse_commits)} names, {sum(len(v) for v in reuse_commits.values())} commits  e.g. {sorted(reuse_commits)[:2]}")
print(f"changes with zero requirement entries (at some commit): {len(zero_target)}")
print(f"MODIFIED/REMOVED target missing in main at same commit: {len(missing_target)} (distinct changes: {len({x[1] for x in missing_target})})")
print(f"ADDED/MODIFIED requirements with no statement paragraph: {len(no_stmt)}")
print(f"within-change: requirement disappeared between observations: {len(disappear)} (changes: {len({x[1] for x in disappear})})")
print(f"within-change: statement changed between observations:      {len(meaning)} (changes: {len({x[1] for x in meaning})})")
for label, rows in (("missing-target sample", missing_target), ("disappear sample", disappear), ("meaning sample", meaning)):
    print(f"  {label}: {rows[:2]}")
print(f"names archived twice at a commit: {len(twice_commits)} names, {sum(len(v) for v in twice_commits.values())} commits")
print(f"RENAMED FROM entries in active deltas, fences masked (summed over walked commits): {renamed_entries}")

# Draft 4 section 0: the git connector's walk (gitcmd.go), unsimplified, deduplicated by SDD state.
walk = git("rev-list", "--reverse", "--date-order", "--exclude=refs/stash", "--exclude=refs/notes/*",
           "--exclude=refs/replace/*", "--all").decode().split()
merges = len(git("rev-list", "--merges", "--all").decode().split())
roots = len(git("rev-list", "--max-parents=0", "--all").decode().split())
def state_key(c):
    r = subprocess.run(["git", "-C", repo, "rev-parse", "-q", "--verify", f"{c}:openspec/changes"],
                       capture_output=True)
    return r.stdout.decode().strip() or "none"
states = {state_key(c) for c in walk}
limited = len(git("rev-list", "HEAD", "--", "openspec/changes").decode().split())
full = len(git("rev-list", "--all", "--full-history", "--", "openspec/changes").decode().split())
heads = {}
for line in git("for-each-ref", "--format=%(objectname) %(refname:short)", "refs/heads", "refs/remotes", "refs/tags").decode().splitlines():
    sha, ref = line.split(" ", 1); heads.setdefault(sha, []).append(ref)
dual, delta_blobs = [], {}
for c in walk:  # the provider's own walk: unsimplified --all
    names = git("ls-tree", "--name-only", c, "openspec/changes/").decode().split("\n")
    act = {n.split("/")[-1] for n in names if n and n.split("/")[-1] != "archive"}
    arch = git("ls-tree", "--name-only", c, "openspec/changes/archive/").decode().split("\n")
    archn = [m.group(1) for a in arch if a and (m := re.match(r"^\d{4}-\d{2}-\d{2}-(.+)$", a.split("/")[-1]))]
    if act & set(archn) or len(archn) != len(set(archn)):
        dual.append(c)
for line in git("rev-list", "--all", "--objects", "--", "openspec/changes").decode().splitlines():
    parts = line.split(" ", 1)
    if len(parts) == 2 and re.match(r"^openspec/changes/.*/specs/[^/]+/spec\.md$", parts[1]):
        delta_blobs[parts[0]] = parts[1]
differ = []
for sha, path in delta_blobs.items():
    text = git("cat-file", "-p", sha).decode("utf-8", "replace")
    if parse_delta(text, True) != parse_delta(text, False):
        differ.append(path)
tips = sorted(r for c in dual for r in heads.get(c, []))
print(f"git-connector walk (--all): {len(walk)} commits, {merges} merges, {roots} roots")
print(f"distinct SDD states over that walk (openspec/changes tree id): {len(states)}")
print(f"commits touching openspec/changes: HEAD path-limited {limited} / --all --full-history {full}")
print(f"dual-presence or archived-twice commits on the unsimplified --all walk: {len(dual)}; refs pointing at them: {len(tips)} {tips}")
print(f"distinct delta blobs: {len(delta_blobs)}; whose obligations change under fence masking: {len(differ)} {sorted(set(differ))}")
fp = git("rev-list", "--first-parent", "--reverse", "HEAD").decode().split()
dual_set, run, best, best_nonmerge = set(dual), 0, 0, 0
nonmerge = set(git("rev-list", "--no-merges", "HEAD").decode().split())
cur_nm = 0
for c in fp:
    if c in dual_set:
        run += 1; cur_nm += c in nonmerge
        if run > best: best, best_nonmerge = run, cur_nm
    else:
        run = cur_nm = 0
print(f"longest run of HEAD first-parent commits on a dual-presence state: {best} ({best_nonmerge} non-merge)")
root_blob_commits, root_blob_names = 0, set()
for c in walk:
    blobs = [l.split("\t", 1)[1] for l in git("ls-tree", c, "openspec/changes/").decode().splitlines() if l.split()[0] in REGULAR]
    if blobs:
        root_blob_commits += 1; root_blob_names.update(b.rsplit("/", 1)[-1] for b in blobs)
print(f"commits with a regular file directly under openspec/changes (never a change, draft 9 section 5): {root_blob_commits} {sorted(root_blob_names)}")
# Type by MODE, not by ls-tree's object type: git lists a symlink (120000) as type "blob" too, and
# section 4.1 treats a symlink at these positions differently from a regular file (M1 re-check F4).
spec_notes, specs_files, specs_links = set(), set(), set()
for c in walk:
    for line in git("ls-tree", "-r", c, "--", "openspec/changes").decode().splitlines():
        meta, p = line.split("\t", 1)
        mode = meta.split()[0]
        if re.match(r"^openspec/changes/(archive/)?[^/]+/specs/[^/]+$", p):
            (spec_notes if mode in REGULAR else specs_links).add(p)
        elif re.match(r"^openspec/changes/(archive/)?[^/]+/specs$", p):
            (specs_files if mode in REGULAR else specs_links).add(p)
print(f"regular files directly under a change's specs/ (ignored, draft 10 section 4.1): {len(spec_notes)} {sorted(spec_notes)}")
print(f"regular files AT a change's specs (ignored, draft 10 section 4.1): {len(specs_files)} {sorted(specs_files)}")
print(f"symlinks or gitlinks at or directly under a change's specs (non-regular-file units): {len(specs_links)} {sorted(specs_links)}")
