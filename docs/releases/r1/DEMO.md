# Proofbound — 5-Minute Demo

This demo builds a tiny sample app, records its delivery history with Proofbound, and reads
back a durable **trust record** — proof IDs bound to exact commits, and a decision bound to the
commit that implements it. No prior Proofbound knowledge required.

For the full walkthrough with explanations, see the [User Guide](USER-GUIDE.md).

---

## 0. Prerequisites (one time)

- Git and Go 1.26+ on your machine.
- Outbound internet on first run **or** your own PostgreSQL (see note at the end).

## 1. Build the Proofbound binary

```bash
git clone <this-repo-url> proofbound
cd proofbound/kernel
CGO_ENABLED=0 go build -o "$HOME/bin/proofbound" ./cmd/proofbound   # ensure ~/bin is on PATH
```

## 2. Create a sample app with some history

```bash
mkdir /tmp/greeter && cd /tmp/greeter
git init -q
git config user.email "dev@example.com" && git config user.name "Sample Dev"

# IMPORTANT: keep Proofbound's local database out of your repo
printf '.proofbound/\n' > .gitignore

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

## 3. Record a decision and the commit that implements it

```bash
mkdir -p docs/decisions
cat > docs/decisions/VD-greeting-style-7a1b2c.md <<'EOF'
# VD-greeting-style-7a1b2c — Greeting uses "Hi" not "Hello"

Status: accepted
Date: 2026-11-12

We shorten the greeting from "Hello, X!" to "Hi, X!" for a friendlier tone.
EOF

sed -i 's/"Hello, "/"Hi, "/'   main.go
sed -i 's/Hello, Ada/Hi, Ada/' main_test.go

git add -A && git commit -q -m "feat: adopt friendly greeting

Implements the decision recorded in VD-greeting-style-7a1b2c."
```

A commit "cites a decision" when its message mentions a `VD-*.md` file that exists under
`docs/decisions/` at that commit. Proofbound picks that up automatically.

## 4. Record the history into Proofbound

Run this **from inside the sample repo**. `DATABASE_URL` is unset, so Proofbound uses its
embedded dev database (first run downloads PostgreSQL into `./.proofbound`).

```bash
proofbound sync git
```

Expected (last line; PostgreSQL log lines above it are normal in dev-mode):

```
listed=3 appended=3 existing=3   # appended=3 on first run; existing=3 if you re-run it
```

## 5. Read the trust record

```bash
proofbound report week
```

Expected (the meaningful block):

```
week 2026-11-05 to 2026-11-12
commits count=3
- 5c8aba35…  feat: greeter prints a greeting   files=3 decisions=                       proof=01M78X6Y1F1V4R59AKN5XQDAGM
- 818744e3…  test: cover Greet                 files=1 decisions=                       proof=01M78X6Y278HFEXFMY21Z4V4A7
- ec34d1cd…  feat: adopt friendly greeting     files=3 decisions=VD-greeting-style-7a1b2c proof=01M78X6Y2SM5F9P5GMDKK5JDFK
checks count=0
sessions count=0
reviews count=0
red_verdict_chains count=0
```

**What you just proved:**

- Each commit now has a permanent **proof ID** (`proof=01M78…`) bound to its exact SHA.
- The third commit shows `decisions=VD-greeting-style-7a1b2c` — the *why* (the decision) is
  bound to the exact commit that implements it. That link survives in the ledger forever, even
  if the commit message is later rewritten or the file moves.

## 6. See that the record is deterministic (this is the point)

```bash
proofbound sync git     # re-run: listed=3 appended=0 existing=3  (idempotent)
proofbound rebuild      # rebuild all projections from the raw ledger
proofbound report week  # identical proof IDs — the record is reproducible, not a snapshot
```

---

## A note on the PostgreSQL log noise

In embedded dev-mode, PostgreSQL prints startup/shutdown lines to stdout around Proofbound's
output. To see only Proofbound's lines:

```bash
proofbound report week | grep -E '^(week|commits|- |checks|sessions|reviews|red_)'
```

Or point Proofbound at your own database, which removes the noise entirely:

```bash
export DATABASE_URL='postgres://user:pass@host:5432/dbname?sslmode=disable'
```

## Cleaning up

```bash
rm -rf /tmp/greeter
```

---

*Next: the [User Guide](USER-GUIDE.md) explains every concept and shows how to run Proofbound on
your real project in a corporate setting.*
