# Regeneration-repair — the "is this rebuild safe to trust?" gate

**Status:** Release 3 *preview* — the first, gate-only slice of **witnessed regeneration** (vision
build-order step #2b). This part is built, tested, and part of the build. The automatic *rebuilder*
is deliberately **not** in this slice (see "What this is / isn't" below).

> **Run the live demo (≈1 second):**
> ```bash
> bash docs/releases/r3/regen-repair/demo.sh
> ```
> It shows the gate approving one clean rebuild and refusing three unsafe ones.

---

## The idea in one paragraph

When AI writes software, the usual reflex on a bug is to *patch* the code. On real work that backfired —
patching introduced one-to-three **new** defects per fix, over and over. So Proofbound's plan is the
opposite: when a rule the software must obey gets broken, you **throw the broken code away and rebuild
it** from the rule plus the specific failing example. But a rebuild is only as trustworthy as the rule
you check it against — a weak rule will happily wave through wrong code. **This gate is the safety
check that decides whether a rebuild can be trusted at all.** It says **AUTHORIZE** (safe to rebuild)
or **DENY** (don't trust a rebuild here — fix the problem first), and it fails closed: anything
missing or doubtful is a DENY.

## What the gate checks (in plain terms)

Before it will trust a rebuild, the gate requires *all* of these — miss any one and it refuses:

| Check | In plain English | Why it matters |
|---|---|---|
| **New rule revision** | The failing example was folded into a **new** version of the rule; it can't quietly rewrite the old one. | Keeps an honest, append-only history of how the rule got stronger. |
| **Strong enough rule** | A fresh measurement says the rule catches wrong code with **no known blind spot**. | A rebuild checked against a weak rule proves nothing. This is the whole point. |
| **Real failure, not a strawman** | The failing example was actually shown to break the *old* code. | Stops a fake "failure" from being used to justify a rebuild. |
| **No peeking at the answer key** | Someone other than the rebuilder attests the new code was built from the *rule*, not from the test code. | Stops the rebuild from gaming the tests instead of obeying the rule. |
| **No grading your own work** | The person who wrote the rule, the person who rebuilt the code, and the person who signs off are **three different, named people**. | Proofbound's hardest-won law: no actor accepts its own work. |

## What the demo shows

`demo.sh` runs the real gate four times on real evidence files:

1. **A clean repair →** ✅ AUTHORIZE. Strong rule, real failure, three distinct people.
2. **A weak rule →** 🛑 DENY. The rule has a blind spot, so a rebuild can't be trusted; strengthen the rule first.
3. **A strawman failure →** 🛑 DENY. The "failing example" was never shown to break the old code.
4. **One person grading their own work →** 🛑 DENY. The rule's author is also the rebuilder.

Each refusal names the exact reason. (One input — the rule-strength measurement — is pre-canned in
the demo because the real measurement is a slow mutation sweep; every other check is the real gate on
real files. The demo labels the stubbed line `DEMO-STUB`.)

## What this is / isn't (honest scope)

- **It is** the *judge*: given the evidence, it decides whether a rebuild is trustworthy. A **person
  still does the actual rebuild** by hand.
- **It is not** an automatic code rewriter. Nothing here calls an AI or regenerates code. Building the
  automatic rebuilder is a larger, later step that would reopen a safety-boundary decision; it was
  deliberately left out of this slice.
- **One honest limit:** in a single-person setup the "no peeking at the answer key" check can't be
  fully *mechanically* guaranteed — it rests on a named person's attestation. The gate records this as
  a declared limitation rather than pretending it's airtight.

## Where this fits

This is obligation **O-2** of the regeneration-trust requirement, built on top of the already-accepted
**O-1** "rule-strength" gate. The design is specified (and independently reviewed) in
[`../../../plans/P7-2b-regeneration-loop-spec-draft1.md`](../../../plans/P7-2b-regeneration-loop-spec-draft1.md);
the gate itself is [`../../../../scripts/regen-gate.sh`](../../../../scripts/regen-gate.sh) with a
fail-closed self-test that runs inside `make check`.
