#!/usr/bin/env python3
"""Inventory committed P7 source inputs without fetching, building, or executing them."""

import argparse
import hashlib
import json
import pathlib
import re
import subprocess
import sys
import zipfile


ROOT = pathlib.Path(__file__).resolve().parents[1]
CORPUS = "docs/verification/p7-replay-lock-fixtures/bytes/sha256/"


def git(*args):
    return subprocess.run(
        ["git", "-C", str(ROOT), *args], check=True, capture_output=True
    ).stdout


def entries(revision):
    result = {}
    for entry in git("ls-tree", "-r", "-z", revision, "kernel", CORPUS).split(b"\0"):
        if not entry:
            continue
        metadata, name = entry.split(b"\t", 1)
        mode, kind, _oid = metadata.decode("ascii").split(" ")
        if kind != "blob" or mode not in ("100644", "100755"):
            raise ValueError("unsupported committed entry: " + name.decode())
        result[name.decode("utf-8")] = mode
    return result


def source_path(path):
    if path in ("kernel/go.mod", "kernel/go.sum", "kernel/internal/store/migrations/001_ledger.sql"):
        return True
    return path.startswith("kernel/") and path.endswith(".go") and not (
        path.endswith("_test.go") or path.endswith("_windows.go")
    )


def module_name(raw):
    match = re.search(rb"(?m)^module\s+([^\s]+)", raw)
    if not match:
        return None
    return match.group(1).decode("utf-8").strip('"')


def zip_identity(path):
    with zipfile.ZipFile(path) as archive:
        names = archive.namelist()
    matches = [re.match(r"^(.+?@[^/]+)/", name) for name in names]
    if not names or any(match is None for match in matches):
        return None
    roots = {match.group(1) for match in matches}
    if len(roots) != 1:
        return None
    prefix = next(iter(roots))
    if "@" not in prefix:
        return None
    module, version = prefix.rsplit("@", 1)
    return re.sub(r"!([a-z])", lambda m: m.group(1).upper(), module), version


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--revision", required=True)
    args = parser.parse_args()
    revision = git("rev-parse", "--verify", args.revision + "^{commit}").decode().strip()
    if not re.fullmatch(r"[0-9a-f]{40}", revision):
        raise ValueError("revision must resolve to a full Git commit")
    tracked = entries(revision)
    relevant = [p for p in tracked if source_path(p) or p.startswith(CORPUS)]
    # Prevent working-tree substitutions for the named Git revision.
    subprocess.run(
        ["git", "-C", str(ROOT), "diff", "--quiet", revision, "--", *relevant],
        check=True,
    )
    sources, mod_by_name, zip_by_name, issues = [], {}, {}, []
    for name in sorted(relevant):
        path = ROOT / name
        if not path.is_file() or path.is_symlink():
            raise ValueError("missing or nonregular committed path: " + name)
        raw = path.read_bytes()
        digest = hashlib.sha256(raw).hexdigest()
        item = {"path": name, "mode": tracked[name], "sha256": digest}
        if source_path(name):
            sources.append(item)
            continue
        if name != CORPUS + digest:
            raise ValueError("digest-named corpus mismatch: " + name)
        if zipfile.is_zipfile(path):
            identity = zip_identity(path)
            if identity is None:
                issues.append("zip identity ambiguous: " + name)
            else:
                zip_by_name.setdefault(identity[0], []).append((identity[1], item))
        else:
            identity = module_name(raw)
            if identity is None:
                issues.append("module directive missing: " + name)
            else:
                mod_by_name.setdefault(identity, []).append(item)
    modules = []
    for name in sorted(set(mod_by_name) | set(zip_by_name)):
        mods, zips = mod_by_name.get(name, []), zip_by_name.get(name, [])
        if len(mods) == len(zips) == 1:
            version, archive = zips[0]
            modules.append({"module": name, "version": version, "mod": mods[0], "zip": archive})
        else:
            issues.append(f"ambiguous module pairing: {name}: {len(mods)} .mod, {len(zips)} .zip")
    result = {
        "schema": "proofbound.p7.source-build-input-inventory.v1",
        "evidence_revision": revision,
        "selection": "committed production Go candidates for linux/amd64, go.mod, go.sum, embedded migration; closure unverified",
        "source_files": sources,
        "modules": modules,
        "issues": sorted(issues),
        "status": "PENDING" if issues else "UNVERIFIED",
    }
    print(json.dumps(result, sort_keys=True, separators=(",", ":")))


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, subprocess.CalledProcessError, zipfile.BadZipFile) as exc:
        print(f"p7 source build input binding: {exc}", file=sys.stderr)
        sys.exit(1)
