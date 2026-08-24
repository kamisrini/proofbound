# internal/connector/git/gitcmd — SPEC

Contract for the standard-library Git process adapter. This package implements the parent
connector's `Repo` interface; it owns repository/ref semantics and output parsing, but never event
creation, persistence, cursors, or database access.

## 1. Interface

```go
var ErrShallow = errors.New("gitcmd: repository history is incomplete")

type Repo struct { /* unexported */ }

func New(root string) (*Repo, error)
func (r *Repo) Commits(ctx context.Context) ([]git.Commit, error)
func (r *Repo) Tips(ctx context.Context) (map[string]string, error)
```

`New` requires a non-empty path to a non-bare work tree. It refuses shallow repositories and
repositories with legacy `.git/info/grafts`. `Commits` repeats both refusal checks so a repository
cannot become partial after construction.

## 2. Git protocol

- Commands run with `--no-replace-objects`, `-c core.quotepath=true`, and
  `-c diff.renames=false`; repository and user configuration cannot alter payload semantics.
- History starts from `--all`, excluding exactly `refs/stash`, `refs/notes/*`, and
  `refs/replace/*`. HEAD remains included, including detached HEAD.
- Commit scalar fields are emitted by distinct placeholders with NUL separators. A root commit's
  file block is its full tree; every other commit, including a merge, uses the tree delta from its
  first parent. Filenames are NUL-delimited raw bytes and need no trimming or C-quote decoding;
  non-UTF-8 paths are refused because JSON cannot represent their byte identity reversibly.
  Scalar fields are likewise refused unless they are valid UTF-8; JSON replacement characters
  never stand in for observed Git bytes.
- Decision citations are resolved against exact direct files in the commit's own
  `docs/decisions/VD-*.md` directory; nested paths do not qualify. A token
  that is only a prefix of a real id, or merely looks id-shaped, is never emitted.
- Tips are peeled commit ids for every ref/HEAD route admitted by `Commits`; excluded refs never
  appear. A failure from Git is always an error except the documented unborn-HEAD/empty-repository
  cases.

## 3. Invariants

1. **G-INV-1 — No commit content can break the framing.** Scalar fields are NUL-framed and paths
   come from a `-z` stream.
2. **G-INV-2 — A misframed stream is a loud error.** Missing or extra scalar fields cannot be
   silently assigned to a neighbouring commit.
3. **G-INV-3 — Stash and notes refs are not history.** Replace refs are excluded too.
4. **G-INV-4 — A commit on a non-checked-out branch, reachable only from a tag, or sitting on a DETACHED HEAD, IS history.**
5. **G-INV-5 — A shallow repository is refused, at construction AND at every listing.**
6. **G-INV-6 — A broken repository is an error, never an empty one.** Every loose ref, packed ref,
   and detached HEAD target object must exist; a syntactically valid missing-object id is broken.
7. **G-INV-7 — An empty repository is not an error.**
8. **G-INV-8 — Every scalar comes from its own git placeholder.** Committer date, author fields,
   committer fields, subject, and body are not substituted for one another.
9. **G-INV-9 — A c-quoted path is decoded.** The implementation obtains raw `-z` paths, which is
   the stronger form: quoted spellings never enter the payload.
10. **G-INV-10 — Decision ids are sorted and de-duplicated.**
11. **G-INV-11 — Ref tips are peeled, and cover EVERYTHING `Commits` ingests.**
12. **G-INV-12 — Path bytes are significant; nothing is trimmed.** Leading/trailing whitespace
    and newlines in a legal UTF-8 filename survive exactly. Non-UTF-8 paths are refused rather than
    collapsed through Unicode replacement characters.
13. **G-INV-13 — A citation is never fabricated.** Citations must resolve to exact decision files
    in the observed commit tree.
14. **G-INV-14 — Object replacement never rewrites what is recorded.** Every object-reading
    command uses `--no-replace-objects`; replace refs are not ingested as history.
15. **G-INV-15 — Merge paths are the first-parent tree delta.** Root commits report the full tree;
    every other commit compares its tree to its first parent, including merges.
16. **G-INV-16 — Rename reporting is configuration-independent.** Rename detection is forced off,
    so a pure rename records both the removed and added path for the same SHA under every local or
    user `diff.renames` setting.
17. **G-INV-17 — Invalid UTF-8 scalar bytes are refused.** SHA, author/committer names and emails,
    date, subject, and body cannot reach JSON through Unicode replacement.
18. **G-INV-18 — Ref validation follows annotated-tag referents.** A tag object whose referent is
    absent is a broken repository for both commit listing and tips; valid non-commit refs remain
    outside history without becoming errors.

Legacy graft refusal is a route to G-INV-5's harm: grafts make a boundary commit appear to change
its whole tree, producing contradictory payloads for one SHA. Both `New` and `Commits` therefore
refuse a non-empty `.git/info/grafts` with `ErrShallow`.

## 4. Non-goals

- No date, cursor, count, or previous-tip restriction.
- No Git library dependency.
- No database access or event creation.
- No attempt to repair partial, corrupt, shallow, grafted, or replacement-rewritten history.

## 5. Invariant table

| Invariant | Statement | Proving test |
|---|---|---|
| G-INV-1 | No commit content can break framing | gitcmd_test.go::TestCommits_ContentCannotBreakFraming |
| G-INV-2 | Misframing is loud | gitcmd_test.go::TestParseScalars_RejectsMisframing |
| G-INV-3 | Stash, notes and replace refs are excluded | gitcmd_test.go::TestCommits_ExcludesNonHistoryRefs |
| G-INV-4 | Branch, tag and detached HEAD commits are included | gitcmd_test.go::TestCommits_IncludesEveryHistoryRoute |
| G-INV-5 | Shallow repositories and grafts are refused twice | gitcmd_test.go::TestRepo_RefusesPartialHistory |
| G-INV-6 | Broken repositories and missing ref/HEAD objects are errors | gitcmd_test.go::TestRepo_MissingObjectRoutesAreErrors |
| G-INV-7 | Empty repositories are accepted | gitcmd_test.go::TestCommits_EmptyRepositoryIsNotAnError |
| G-INV-8 | Scalar placeholders are mapped exactly | gitcmd_test.go::TestCommits_MapsEveryScalar |
| G-INV-9 | Quoted paths decode to raw names | gitcmd_test.go::TestCommits_PreservesHostilePaths |
| G-INV-10 | Citations are sorted and unique | gitcmd_test.go::TestCommits_ResolvesCitationsAgainstTheCommitTree |
| G-INV-11 | Tips cover the admitted history scope | gitcmd_test.go::TestTips_CoverCommitsAndPeelTags |
| G-INV-12 | Valid UTF-8 paths survive exactly and invalid UTF-8 paths are refused through the adapter | gitcmd_test.go::TestCommits_PathIdentityIsPreservedOrRefused |
| G-INV-13 | Prefixes and id-shaped fiction are not citations | gitcmd_test.go::TestCommits_ResolvesCitationsAgainstTheCommitTree |
| G-INV-14 | Replacement objects never alter payloads | gitcmd_test.go::TestCommits_IgnoresReplacementObjects |
| G-INV-15 | Root and first-parent merge file semantics are deterministic | gitcmd_test.go::TestCommits_MergeFilesAreFirstParentDelta |
| G-INV-16 | Local rename configuration cannot change one SHA's paths | gitcmd_test.go::TestCommits_RenamePayloadIgnoresLocalConfig |
| G-INV-17 | Invalid UTF-8 scalar bytes are refused before payload construction | gitcmd_test.go::TestCommits_RefusesInvalidUTF8Scalars |
| G-INV-18 | Missing annotated-tag referents are errors while valid non-commit refs are skipped | gitcmd_test.go::TestRepo_MissingObjectRoutesAreErrors |

## 6. Dependencies

Only the standard library and the parent `internal/connector/git` package. `os/exec` is confined
here. No database driver, Git library, or persistence package is permitted.
