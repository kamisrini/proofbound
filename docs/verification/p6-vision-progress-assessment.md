# Proofbound vision progress — 2026-09-12

The clearest verdict is: Proofbound has built the evidence spine of the vision, but not yet the
autonomous “verified delivery organism.”

The completed P0–P5 work is substantial—it can connect intent, an exact commit, evidence, an
independent verdict, and deployment observation. But most of the 2028/100x promise—automatic test
generation, production-shaped simulation, cryptographic warranties, autonomous repair, and
governance—is still ahead.

Legend: 🟢 accepted foundation · 🟡 partial precursor · 🔵 bounded spike · ⚪ not built

## Vision versus reality

| Final goal | What Proofbound has now | Status | What is still missing |
|---|---|---:|---|
| One trustworthy record of reality | Append-only PostgreSQL ledger, immutable events, content identity, replayable projections | 🟢 | Complete capture of all builds, deploys, credentials, runtime behavior, and external systems |
| Dashboards derived from evidence | Weekly, GitHub delivery, intent, and requirement reports with event proof | 🟢 | End-user UI, cross-organization reporting, operational scale |
| Gates that cannot be silently skipped | Data-defined gates and a proven `make delivery-enforce` boundary | 🟢 within Proofbound | Integration into real customer deployment boundaries and broader applicability policy |
| Intent becomes the durable asset | BD → BR obligations → CI → commit chain with exact revisions | 🟡 | Behavior locks, domain calculus, controlled language, examples, precedents, risk envelopes |
| Multiple intent sources | Native records and `specdir` providers under one canonical contract | 🟢 for file sources | Real API-backed systems, mutable-upstream snapshots, customer tool integrations |
| Independent verification | Requirement reviews, per-obligation verdicts, mutation sweeps, non-author acceptance | 🟡 | Cryptographic verifier identities, disjoint builder/verifier keys, scalable verifier orchestration |
| Proof-carrying delivery | Exact chain from intent through commit and evidence to observed deployment | 🟡 | Portable signed warranty consumable by another organization |
| Living simulated twin | Isolated replay, disposable projections, deterministic replay proof, calibration primitives | 🔵 | Production-shaped replicas, realistic users, traffic replay, chaos, surrogates, confidence-scored futures |
| Testing becomes redundant for humans | Tests and checks are witnessed, independently reviewed, mutation-calibrated, and enforced | 🟡 early foundation | Tests generated from intent, continuous twin execution, automatic verdict interpretation, anomaly-only human review |
| Code becomes disposable | Exact intent revisions now survive independently of commits | 🟡 conceptual precursor | Witnessed code regeneration, behavior-preserving rewrites, generated migrations, regeneration-as-repair |
| Production continuously validates claims | GitHub workflow/deployment observations can join exact commits | 🟡 | Runtime invariant probes, production telemetry, trust decay, sim-to-real calibration |
| Humans handle only meaning | Requirement authors and independent reviewers have distinct semantic roles | 🟡 | Ambiguity alternatives, consequence briefs, decision inbox, precedent ledger |
| Autonomous governed operation | Proofbound has an executable build constitution and scoped delivery gates | 🟡 precursor | Governor, policy envelopes, credential control, autonomy ratchet, portfolio optimizer |
| Portable global trust fabric | Provider-neutral schemas and exact proof references establish the grammar | ⚪ | Signing, trust anchors, cross-company admission, warranties, insurance, verifier marketplace |
| Proofbound builds Proofbound | P5 was self-hosted through its own intent and acceptance chain | 🟢 first milestone | Regenerating and operating itself under a complete warranty |

## The journey as a ladder

```text
FINAL: Software is a continuously kept, verified promise
  │
  ├─ 5. Portable warranties and verifier marketplace       ⚪
  ├─ 4. Governor and autonomous decision routing           ⚪ / early governance precursor
  ├─ 3. Production-shaped living twin                      🔵 replay/calibration spike
  ├─ 2. Behavior locks and witnessed regeneration          🟡 intent provenance precursor
  └─ 1. Flight recorder, evidence, and admission gates      🟢 substantially built
       └─ Current position: P0–P5 accepted and self-hosted
```

## QA-redundancy ladder

The vision carefully says testing becomes redundant as a human activity—not that evidence or
checking disappears.

| Vision level | Current position |
|---|---|
| L1: Tests generated automatically from intent | Not yet. Tests are still explicitly authored, though derived from invariant tables |
| L2: Tests run continuously without scheduling | Partial. Delivery enforcement automates witnessed checks, but there is no continuous twin |
| L3: Humans stop reading results | Not yet. Independent verdict production and review remain explicit activities |
| L4: Formal proofs replace tests for critical kernels | Not built |
| L5: Testing and operating become one continuous evidence stream | Not built |

Proofbound currently makes QA more rigorous, traceable, and difficult to fake. It does not yet make
QA obsolete.

## What P5 changed strategically

Before P5, Proofbound mostly knew:

```text
commit → checks → deployment
```

After P5, it knows:

```text
declared business choice
→ exact requirement obligations
→ exact change intent
→ exact commit
→ witnessed evidence
→ independent per-obligation conclusion
→ observed deployment
```

That is the first credible slice of “software as a verified promise.” It establishes the vocabulary
and evidence chain on which the twin, automatic verification, regeneration, and portable warranty
can eventually operate.

One important correction P5 made to the original plain-English vision: Proofbound cannot eliminate
claims entirely. Business intent and authority begin as authored claims. P5 therefore labels them
honestly as declared authority, keeps observations distinct, and gives verdicts their own
independent category.

## Overall position

- Against the committed roadmap: P0–P5 are complete.
- Against the 2028 product: the foundational kernel is real; the Intent Fabric and Trust Engine have
  credible early slices; the Twin is only a spike; the Governor and portable warranty economy are
  mostly unbuilt.
- Against the 100x north star: Proofbound now preserves pieces of intent and memory, but
  implementation is not yet fungible and behavior is not “born proven.”
- Release posture: credible internal alpha/research prototype for proof-bound delivery—not yet a
  product that can claim to replace QA, testing, audits, or the software-delivery stack.

## Canonical sources

- `vision-100x.md`
- `vision-2028.md`
- `ROADMAP.md`
- `docs/plans/P5-PROOFBOUND-INTENT-PROVENANCE-v3.md`
- `notes/state.md`
