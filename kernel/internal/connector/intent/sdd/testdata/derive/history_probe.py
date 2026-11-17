"""Change-scoped, history-native mapping probe over a real OpenSpec repository.

Walks every commit that touches openspec/changes, reads the change folders AT THAT COMMIT,
and measures what would break a change-scoped warranty:
  - changes never visible as active (archived without ever being observed active)
  - changes with no requirement entries at all (a change intent needs >= 1 target)
  - MODIFIED / REMOVED targets missing from the main spec at the same commit
  - requirements without a statement paragraph
  - within an active change, between consecutive observations: requirements that disappear,
    and requirements whose statement changes (both are intent SPEC section 5 events)
  - change names reused after archive (active and archived at once, or archived twice)
  - draft 4 section 0 walk counts: the git connector's --all walk, merges, roots, distinct SDD
    states, path-limited vs full-history coverage, tags on dual-presence commits, RENAMED entries
Usage: python3 -I history_probe.py <repo>
"""
import collections, re, subprocess, sys

repo = sys.argv[1]
def git(*a):
    return subprocess.run(["git", "-C", repo, *a], capture_output=True, check=True).stdout

commits = git("rev-list", "--reverse", "--date-order", "--exclude=refs/stash", "--exclude=refs/notes/*",
              "--exclude=refs/replace/*", "--all").decode().split()  # the provider's walk (draft 6 section 3)
cat = subprocess.Popen(["git", "-C", repo, "cat-file", "--batch"], stdin=subprocess.PIPE, stdout=subprocess.PIPE)
def blob(rev, path):
    cat.stdin.write(f"{rev}:{path}\n".encode()); cat.stdin.flush()
    head = cat.stdout.readline().decode().split()
    if len(head) < 3 or head[1] != "blob":
        return None
    data = cat.stdout.read(int(head[2])); cat.stdout.read(1)
    return data.decode("utf-8", "replace")

# Draft 6 section 4.2: OpenSpec's delta READER (src/core/parsers/requirement-blocks.ts parseDeltaSpec,
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
def name_of(raw): return js_trim(re.sub(r"[ \t]+#+[ \t]*$", "", raw))
def split_lines(text):
    if text.startswith("﻿"): text = text[1:]
    return re.sub(r"\r\n?", "\n", text).split("\n")
def title_kind(raw):
    t = name_of(raw)
    return TITLES.get(t.lower()) if t.isascii() else None
def parse_delta(text, masked=True, unpaired=None):
    """[(SEC, name, statement)] for one delta file; `unpaired` collects unpaired RENAMED lines."""
    lines = split_lines(text)
    mask = fence_mask(lines) if masked else [False] * len(lines)
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
                close(); entry = None; out.append((kind, name_of(bm.group(1)), name_of(bm.group(1)))); continue
            if entry and entry[2] and entry[1] is None:
                if l.startswith("#"): entry[2] = False
                elif not blank(l): entry[1] = l
        close()
    return out
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
