# VERA — The Vision in Plain English

*Companion to [vision-2028.md](vision-2028.md) — same ideas, everyday words.*

**★ The north star, in one breath:** one day software won't be a thing you own and maintain — it will be a promise the world provably keeps. What we're building now (the proof machine below) is the first floor of that. *(Ratified: VD-north-star-6io56h · full dream: [vision-100x.md](vision-100x.md))*

## The big idea

- In 2 years, AI agents will write code so fast and cheap that **the code itself is no longer the valuable thing**.
- The valuable thing becomes: **"Can you PROVE this software works, is safe, and who checked it?"**
- VERA is the tool that gives that proof. Like a diamond comes with a certificate, **software comes with a warranty** — machine-checkable, signed, portable.

## The problem it solves

- Today's tools run on **people claiming things**: "I tested it ✓", "it's deployed ✓", "status: done ✓".
- Claims can be wrong, forgotten, or faked. Tracking systems drift away from reality, so the numbers on a dashboard and the state of the code can disagree — with no mechanical way to tell which is right.
- **VERA's rule: nobody claims anything. The system only records what actually happened, with proof.** A security camera instead of a witness statement.

## The five parts

**1. The Flight Recorder** *(an airplane black box)*
- Every action — every test run, every deploy, every agent's work — is automatically recorded with a tamper-proof receipt.
- Dashboards and statuses are just replays of the recording — they can never disagree with reality.

**2. The Rulebook of Intent** *(humans say WHAT, machines figure out HOW)*
- Humans write the business rules: "a support ticket unanswered for 24 hours must escalate to a duty manager."
- Agents write and rewrite the code to satisfy those rules — code becomes disposable; the rules are what's precious.
- When a rule is unclear, the system doesn't guess — it **shows two working versions side by side** ("A closes at 60 calendar days, B at 60 business days — here are 14 real invoices where it matters") and you pick. Your answer is remembered forever.

**3. The Crystal Ball** *(a practice copy of your whole system)*
- Before anything goes live, it runs in a **realistic simulation** — with fake-but-realistic users clicking through real journeys.
- **Testing becomes 100% redundant as a human job.** Nobody writes test cases, nobody runs test plans, nobody reads test reports — checking happens automatically on every change, the way compiling does. "Did we test it?" becomes a question nobody asks anymore.
- One honest limit: the *evidence* never disappears — reality is still the final check (a proof can't tell you the rule itself was what the business meant, and the outside world follows no spec). So the system keeps watching production continuously; the checking just becomes invisible and free instead of being someone's job.
- **UAT becomes a walkthrough**: the business user tries *next week's system today*, on their own accounts — the walkthrough IS the sign-off.
- It predicts problems early: *"This queue will overflow in 9 days — here's the fix, already tested."*

**4. The Proof Checker** *(the trust part)*
- The agent that builds something is **never allowed to grade its own work** — independent checkers do that and sign their results like a notary.
- Other AI agents are literally **paid to try to break things** before customers find them.
- An audit that takes 3 months today becomes **a one-hour automated report**, because every change already carries its proof.

**5. The Autopilot with a Human Steering Wheel**
- The system works 24/7 — fixing, upgrading, testing — but only inside limits humans set (like a spending limit on a credit card).
- Humans get a **short daily inbox — 10–15 minutes** — with only the decisions that truly need human judgment, each shown with "here's what happens if you pick A vs B."
- Every automated action traces back to a named human who approved the rule that allowed it — **"the AI decided" is never an acceptable answer**.
- Status meetings, sprint planning, and status decks disappear — the system already knows the status, provably.

## What a normal day looks like

- Morning: read a short report — "here's what got done overnight, here's the proof, here are 4 decisions only you can make."
- No chasing status. No writing test cases. No preparing audit evidence. No 3am pages — problems are caught in simulation a week before they'd happen.

## What to build first

- **Start with the black box** (the Flight Recorder). Even alone, it answers the question every company using AI agents is about to panic about: *"What did our AI agents actually do — and can you prove it?"*
- Everything else builds on top of it, one piece at a time — and each piece is useful on its own.

---

**The one-liner:** today's tools track what people *say* is happening; VERA records what *actually* happened, proves it, and lets AI do the work while humans only make the judgment calls.
