# VD-audit-room-ui-q2v6x3: A local UI, Audit room first, signed acts without a terminal

**Status:** Accepted
**Date:** 2026-11-22

**Context:** Every Proofbound answer is a terminal command or a pull-request comment. The owner named
compliance and auditors as the primary customer ("who want to keep track of agents doing what they are
supposed to, nothing outside the authorized perimeter") and asked for a UI that people who are not
developers can use without the CLI. The plan is [ui-local-plan.md](../plans/ui-local-plan.md): five
workspaces over one ledger, local first (option 2), hosted later (option 3). The forcing questions:
which workspace first, what stack, and may a UI write?

**Decision:** The owner approved on 2026-11-22 ("go ahead lets do these", "yes", "lets build it, so we
can ship it to clients as part of proofbound"):
1. **Audit room first** (U1): `proofbound ui` serves it on `127.0.0.1`; the other workspaces follow.
2. **Stack:** Go standard library only (`net/http`, `html/template`, `embed`) and plain embedded
   JavaScript and CSS: no framework, no build step, no CDN, works offline. No new dependency (Law 8).
3. **Lock:** the UI never holds the ledger between requests. Each refresh runs `sync all`, reads the
   audit into memory, and closes the store, so other commands are blocked only for seconds.
4. **Signed acts:** exploring writes nothing. An act (an exception outcome, a transcript
   countersignature) is previewed exactly, confirmed, signed with the person's own declared key on their
   machine, and committed as a file any clone can verify. Writes need the loopback `Host`, a same-origin
   `Origin`, and a per-run token, so no other web page can trigger one.
5. **Release:** ships as v0.3.3; v0.3.2 stays as published. *Amended 2026-11-22 by the owner:* v0.3.2 had
   reached no one outside (public launch 2026-11-30), so the UI is folded into v0.3.2 and its release is
   rebuilt and republished at the new commit, rather than adding v0.3.3.

**Consequences:** The primary customer can audit a release in a browser and close each exception with a
signed record. Given up for U1: in-browser signature checks (the page says the binary checked them),
the time rail, the sample draw, the evidence-bundle viewer, and the other four workspaces. The local
server becomes a signing service for its user, which is the main new risk; the spec states the threat
model and the code review tests it.

**Revisit when:** a user asks for hosted access, or U1 shows the Audit room answers the wrong
questions.
