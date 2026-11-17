"""Change-scoped, history-native mapping probe over a real OpenSpec repository.

Walks every commit that touches openspec/changes, reads the change folders AT THAT COMMIT,
and measures what would break a change-scoped warranty:
  - changes never visible as active (archived without ever being observed active)
  - changes with no requirement entries at all (a change intent needs >= 1 target)
  - MODIFIED / REMOVED targets missing from the main spec at the same commit
  - requirements without a statement paragraph
  - within an active change, between consecutive observations: requirements that disappear,
    and requirements whose statement changes (both are intent SPEC section 5 events)
  - change names reused after archive
Usage: python3 -I history_probe.py <repo>
"""
import collections, re, subprocess, sys

repo = sys.argv[1]
def git(*a):
    return subprocess.run(["git", "-C", repo, *a], capture_output=True, check=True).stdout

commits = git("rev-list", "--reverse", "--topo-order", "HEAD", "--", "openspec/changes").decode().split()
cat = subprocess.Popen(["git", "-C", repo, "cat-file", "--batch"], stdin=subprocess.PIPE, stdout=subprocess.PIPE)
def blob(rev, path):
    cat.stdin.write(f"{rev}:{path}\n".encode()); cat.stdin.flush()
    head = cat.stdout.readline().decode().split()
    if len(head) < 3 or head[1] != "blob":
        return None
    data = cat.stdout.read(int(head[2])); cat.stdout.read(1)
    return data.decode("utf-8", "replace")

SECTION = re.compile(r"^## (ADDED|MODIFIED|REMOVED|RENAMED) Requirements\s*$")
REQ = re.compile(r"^### Requirement: (.+?)\s*$")
def parse_delta(text):
    out = []  # (section, name, statement)
    section = None; lines = text.split("\n")
    for i, l in enumerate(lines):
        m = SECTION.match(l)
        if m: section = m.group(1); continue
        if l.startswith("## "): section = None; continue
        r = REQ.match(l)
        if r and section:
            stmt = ""
            for l2 in lines[i + 1:]:
                if l2.startswith("#"): break
                if l2.strip(): stmt = l2.strip(); break
            out.append((section, r.group(1), stmt))
    return out
def main_reqs(text):
    return {m.group(1) for l in text.split("\n") if (m := REQ.match(l))}

stats = collections.Counter()
reuse_commits = collections.defaultdict(set)  # name -> commits where it is both active and archived
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
            for section, rname, stmt in parse_delta(text):
                entries += 1
                if section in ("ADDED", "MODIFIED"):
                    reqs[(m.group(1), rname)] = stmt
                    if not stmt: no_stmt.add((name, rname))
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
print(f"commits walked:                     {stats['commits']}")
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
