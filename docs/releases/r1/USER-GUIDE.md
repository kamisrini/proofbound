# Proofbound User Guide (Release 1 — the Trust Record)

**Who this is for:** a developer who has never used Proofbound and wants to understand what it
does, see the value on a real example, and start using it on their own project — including in a
corporate setting as a single developer.

**What you'll do:** build the tool, then build a small sample app from scratch and watch
Proofbound turn its ordinary git history into a durable, machine-readable *trust record*. By the
end you'll know which commands to run on your own repository and what each one proves.

- Just want to copy-paste? Use the [5-minute demo](DEMO.md).
- Want the scope and caveats of this release? See the [Release Notes](RELEASE-NOTES.md).

---

## 1. What problem Proofbound solves

Modern software — increasingly written and changed by automated systems — leaves behind source
code and a green CI badge. But a green badge is temporary and forgetful. A month later you cannot
reliably answer:

- What was *this* change supposed to do?
- Which decision authorized it, and who owns that decision?
- Which **exact** commit was checked, and what proof is bound to it?
- What is still *unverified*?

Proofbound records the answers as they happen, into an append-only ledger, and binds them to the
exact commit they describe. The result is **proof that travels with the software**, not a log line
that scrolls away.

Proofbound does not replace your CI or test runner. Those *execute* builds and checks. Proofbound
*records* what was claimed, what was checked, what the evidence supports, and where responsibility
lies — and makes proven, failed, and unverified work distinguishable forever.

## 2. The core ideas (2-minute mental model)

| Term | What it means |
|---|---|
| **Ledger** | An append-only event store. Nothing is edited or deleted; new facts are appended. This is the source of truth. |
| **Connector** | A read-only importer that turns an external source into ledger events: `git` (commits), `checks` (check runs), `sessions` (agent sessions), `reviews` (verdicts), `github` (Actions/deployments). |
| **Projection** | A read model built *from* the ledger (e.g. the weekly report). Projections are derived, so they can always be rebuilt from scratch — that reproducibility is what makes the record trustworthy. |
| **Proof ID** | A permanent identifier Proofbound assigns to each recorded fact (e.g. a commit), bound to its exact SHA. |
| **Decision (`VD-*`)** | A decision record under `docs/decisions/`. When a commit mentions one, Proofbound binds the *why* to that commit. |
| **Intent chain (BD → BR → CI)** | An optional, deeper layer: a business-decision → requirement → change-intent chain that states what a change must do and ties it to the commit that satisfies it. |
| **Verdict** | An independent accept/needs-work judgement bound to a reviewed commit. |
| **Gate** | A rule evaluated against the ledger (canary = observe, enforce = block). |

You only need the first five rows to get real value today. The rest scale up as your project
adopts more of Proofbound.

## 3. Prerequisites

- **Git** and **Go 1.26+** (the maintained build uses Go 1.27).
- For the embedded database: **outbound internet on first run** (it downloads a local PostgreSQL),
  *or* your **own PostgreSQL** reachable via a connection string.
- No Docker, no C toolchain, no Java. (`golangci-lint` and GNU Make are only needed to run
  Proofbound's *own* `make check`, not to build or use the binary.)

## 4. Install the binary

```bash
git clone <this-repo-url> proofbound
cd proofbound/kernel
CGO_ENABLED=0 go build -o "$HOME/bin/proofbound" ./cmd/proofbound
```

Make sure `~/bin` (or wherever you put it) is on your `PATH`:

```bash
proofbound                       # prints usage
```

**How Proofbound finds your project:** it walks up from your current directory until it finds a
`.git` folder, and stores its ledger in `.proofbound/` at that repository's root. So the rule is
simple: **`cd` into the repository you want to record, then run `proofbound`.**

## 5. Choose a database

Proofbound needs a PostgreSQL to hold the ledger. Two options:

**Option A — your own PostgreSQL (recommended for real/corporate use):**

```bash
export DATABASE_URL='postgres://user:pass@host:5432/dbname?sslmode=disable'
```

**Option B — embedded dev-mode (zero setup, great for trying it out):**

Leave `DATABASE_URL` unset. On first run Proofbound downloads an embedded PostgreSQL into
`./.proofbound/pgbin` and keeps all its data under `./.proofbound`. Dev-mode is for local use,
**not** for redistribution.

> **Always add `.proofbound/` to your `.gitignore`.** It is a local database, not part of your
> source. (If you forget, a commit can accidentally swallow tens of thousands of database files.)

---

## 6. Tutorial: record a sample app from scratch

We'll build a tiny Go "greeter" app with a few commits, then record it. Every command below is
real and verified; the output shown is what you should see.

### Step 1 — create the repo

```bash
mkdir /tmp/greeter && cd /tmp/greeter
git init -q
git config user.email "dev@example.com" && git config user.name "Sample Dev"
printf '.proofbound/\n' > .gitignore      # keep the local DB out of the repo
```

### Step 2 — first feature commit

```bash
cat > go.mod <<'EOF'
module example.com/greeter

go 1.21
EOF

cat > main.go <<'EOF'
package main

import "fmt"

// Greet returns a greeting for name.
func Greet(name string) string { return "Hello, " + name + "!" }

func main() { fmt.Println(Greet("world")) }
EOF
git add -A && git commit -q -m "feat: greeter prints a greeting"
```

### Step 3 — add a test

```bash
cat > main_test.go <<'EOF'
package main

import "testing"

func TestGreet(t *testing.T) {
	if got := Greet("Ada"); got != "Hello, Ada!" {
		t.Fatalf("got %q", got)
	}
}
EOF
git add -A && git commit -q -m "test: cover Greet"
```

### Step 4 — record a decision, and a commit that implements it

This is where Proofbound starts earning its keep. First write down the decision:

```bash
mkdir -p docs/decisions
cat > docs/decisions/VD-greeting-style-7a1b2c.md <<'EOF'
# VD-greeting-style-7a1b2c — Greeting uses "Hi" not "Hello"

Status: accepted
Date: 2026-11-12

We shorten the greeting from "Hello, X!" to "Hi, X!" for a friendlier tone.
EOF
```

Then make the change, and **mention the decision in the commit message**:

```bash
sed -i 's/"Hello, "/"Hi, "/'   main.go
sed -i 's/Hello, Ada/Hi, Ada/' main_test.go
git add -A && git commit -q -m "feat: adopt friendly greeting

Implements the decision recorded in VD-greeting-style-7a1b2c."
```

Proofbound binds a commit to a decision when the commit message mentions a `VD-*.md` file that
exists under `docs/decisions/` at that commit. No special syntax, no IDs to compute.

### Step 5 — record the history into the ledger

Run this **from inside `/tmp/greeter`**:

```bash
proofbound sync git
```

The meaningful line of output:

```
listed=3 appended=3 existing=0
```

`listed` = commits seen, `appended` = newly recorded, `existing` = already in the ledger. (In
dev-mode you'll also see PostgreSQL startup/shutdown logs — that's normal; see
[the note below](#a-note-on-the-postgresql-log-noise).)

### Step 6 — read the trust record

```bash
proofbound report week
```

```
week 2026-11-05 to 2026-11-12
commits count=3
- 5c8aba35…  feat: greeter prints a greeting   files=3 decisions=                        proof=01M78X6Y1F1V4R59AKN5XQDAGM
- 818744e3…  test: cover Greet                 files=1 decisions=                        proof=01M78X6Y278HFEXFMY21Z4V4A7
- ec34d1cd…  feat: adopt friendly greeting     files=3 decisions=VD-greeting-style-7a1b2c proof=01M78X6Y2SM5F9P5GMDKK5JDFK
checks count=0  sessions count=0  reviews count=0  red_verdict_chains count=0
```

Read what this gives you:

- **`proof=01M78…`** — every commit now has a permanent proof ID bound to its exact SHA. You can
  cite it anywhere; it will always resolve to this exact state.
- **`decisions=VD-greeting-style-7a1b2c`** — the third commit is now *provably* the one that
  implemented that decision. The link lives in the ledger; it survives a reworded commit message
  or a moved file.
- **`red_verdict_chains count=0`** — nothing recorded here is marked failed-but-unaddressed.

> `report week` shows a **trailing 7-day window**. Commits older than a week are still in the
> ledger but won't appear in the weekly view.

### Step 7 — prove the record is reproducible

Trust comes from determinism, so confirm it:

```bash
proofbound sync git        # listed=3 appended=0 existing=3   (idempotent — re-running is safe)
proofbound rebuild         # discard and rebuild every projection from the raw ledger
proofbound report week     # identical proof IDs — this is a reproducible record, not a snapshot
```

That's the whole core loop: **record (`sync`) → read (`report`) → prove reproducible
(`rebuild`).** You now have a durable trust record for an app that didn't know Proofbound existed
an hour ago.

---

## 7. Optional: ingest GitHub Actions and deployments

Release 1 includes a read-only GitHub connector. Configure it with environment variables, then
sync:

```bash
export PROOFBOUND_GITHUB_OWNER='your-org-or-user'
export PROOFBOUND_GITHUB_REPOS='repo-a,repo-b'                   # comma-separated
export GITHUB_TOKEN='ghp_...'                                    # a read-only token
export PROOFBOUND_GITHUB_API_BASE_URL='https://api.github.com'   # set to your host for GitHub Enterprise
proofbound sync github
proofbound report github
```

Without `PROOFBOUND_GITHUB_OWNER`/`PROOFBOUND_GITHUB_REPOS`, `sync github` exits with
`github connector: API, Owner, IDs, and Logger are required`.

## 8. Command reference

```
proofbound sync git                 # record commits from the current repository
proofbound sync checks              # record check runs (from the local spool)
proofbound sync sessions            # record agent sessions
proofbound sync reviews             # record verdicts
proofbound sync github              # record GitHub Actions runs + deployments (needs env vars)
proofbound sync all                 # intent + git + checks + sessions + reviews in one pass
proofbound sync intent {records|specdir|all}   # record the BD/BR/CI intent chain
proofbound rebuild                  # rebuild all projections from the ledger
proofbound verify                   # Proofbound's own delivery boundary (see §10)
proofbound report week              # the weekly trust record (trailing 7 days)
proofbound report github            # GitHub activity view
proofbound report intent <CI-id>    # render one change-intent chain and its verdicts
proofbound report requirement <BR-id>   # render one requirement and its authorization
proofbound intent check --commit <sha>  # check a commit against its declared intent
proofbound gates {canary|enforce}   # evaluate data gates (observe | block)
proofbound twin replay [--through <seq>] [--candidate <file>]
```

`sync` is always idempotent: re-running appends only new events. On an empty repo, connectors for
sources you don't have (checks, sessions, reviews, intents) simply append `0` — that's fine.

## 9. Using Proofbound on your real project (corporate, single developer)

1. **Point it at your own PostgreSQL.** Set `DATABASE_URL` to a database your team controls. This
   avoids the embedded download, keeps the ledger on managed infrastructure, and removes the log
   noise. The embedded dev-mode is for local experimentation, not production.

2. **`cd` into your repo and record it.** Start with the parts you have:

   ```bash
   cd /path/to/your/project
   proofbound sync git
   proofbound report week
   ```

3. **Adopt the decision convention.** Put decision records under `docs/decisions/` named
   `VD-<slug>-<id>.md`, and mention the `VD-…` id in the commit message of the change that
   implements it. From then on, your weekly record shows *why* each change happened, bound to the
   commit — at essentially zero extra effort.

4. **Wire `sync` into your workflow.** Run `proofbound sync git` (or `sync all`) after merges — a
   post-merge hook, a nightly job, or a CI step against your shared `DATABASE_URL`. Because `sync`
   is idempotent, running it often is safe and cheap.

5. **Grow into the deeper layers when you need them.** Check runs, agent sessions, verdicts, the
   BD/BR/CI intent chain, and gates are all available in Release 1. The intent chain (which states
   what a change *must* do and binds the evidence to it) is the strongest guarantee Proofbound
   offers today. Proofbound's **own** repository is the fully worked example — browse
   `docs/intent/records/` and run `proofbound report requirement <BR-id>` or
   `proofbound report intent <CI-id>` inside this checkout to see a complete chain rendered.

## 10. What `verify` is and isn't

`proofbound verify` is **Proofbound's own delivery boundary**, not a general command for your app.
It rebuilds projections and checks snapshot equality against a *witnessed* `make check` run — a
witness only Proofbound's self-hosting workflow produces. On a plain application repo you'll see:

```
proofbound: verify: witness read: verify: no make check witness found in spool
```

That is expected. For your own project, your trust surface is **`sync` + `report` + `rebuild`**;
reproducibility is demonstrated by `rebuild` producing the same projections. `verify` becomes
relevant only if you adopt Proofbound's full witnessed-delivery workflow.

## 11. Limitations (Release 1)

Release 1 **records and verifies**. It does not regenerate failing code, predict outcomes, govern
actions with policy, issue portable signed warranties, or redistribute a database. Those are
**Release 2** and are not built yet — see [`plans/RELEASE-PLAN.md`](../../plans/RELEASE-PLAN.md) for the
full boundary.

Other caveats: embedded dev-mode needs internet on first run and isn't for redistribution; embedded
PostgreSQL logs to stdout in dev-mode; `report week` is a trailing 7-day window; the GitHub
connector is read-only.

---

## Appendix

### A note on the PostgreSQL log noise

In embedded dev-mode, PostgreSQL prints startup/shutdown lines to stdout around Proofbound's own
output. To isolate Proofbound's lines:

```bash
proofbound report week | grep -E '^(week|commits|- |checks|sessions|reviews|red_)'
```

Using your own `DATABASE_URL` removes the noise entirely.

### Resetting the embedded dev database

The ledger is append-only. If a bad import pollutes your *dev* database and you want a clean slate,
move the data directory aside (don't `rm -rf` it mid-run):

```bash
mv .proofbound/db .proofbound/db.aside-$(date +%s)
```

The next command re-initializes a fresh embedded database.

---

*See also: [Release Notes](RELEASE-NOTES.md) · [5-minute demo](DEMO.md) · [README](../../../README.md)*
