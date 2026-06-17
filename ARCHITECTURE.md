<!-- version: 1.0.0 -->
# cARL — Architecture Overview

---

## Conceptual Architecture

cARL is a three-layer system:

```
┌─────────────────────────────────────────────────────────────┐
│                     Agent Session                           │
│  (GitHub Copilot coding agent)                              │
└──────────────────────┬──────────────────────────────────────┘
                       │ reads at session start
┌──────────────────────▼──────────────────────────────────────┐
│                  Operating Model Layer                      │
│  .github/copilot-instructions.md                            │
│  (root constitution: modes, discipline, security baseline)  │
└──────────────────────┬──────────────────────────────────────┘
                       │ references
       ┌───────────────┴──────────────────┐
       │                                  │
┌──────▼──────────┐            ┌──────────▼──────────────────┐
│  Instruction     │            │  Governance Artefacts       │
│  Pack Layer      │            │  Layer                      │
│                  │            │                             │
│  .github/        │            │  .github/carl/              │
│  instructions/   │            │  ├── memory.md              │
│  ├── core/       │            │  ├── current-pr-contract.md │
│  ├── languages/  │            │  ├── invariants.yml         │
│  ├── platform/   │            │  ├── trust-boundaries.md    │
│  └── cloud/      │            │  ├── tool-policy.yml        │
│                  │            │  ├── repo-map.example.json  │
│  Single-concern  │            │  └── plans/                 │
│  instruction     │            │      ├── README.md          │
│  packs           │            │      └── plan-template.md   │
└──────────────────┘            └─────────────────────────────┘
```

---

## Layer 1: Operating Model (Root Constitution)

**File:** `.github/copilot-instructions.md`

This is the root governance file. GitHub Copilot reads it automatically at the start of every agent session. It defines:

- The agent's default operating mode (plan-first)
- Mode selection logic (plan-only / assisted / automatic)
- Core engineering principles (spec before code, small changes, tests are mandatory)
- Security baseline
- Dependency discipline
- Cognition governance overview (cARLv2)
- Required final response headings

This file acts as the repository constitution. It should be stable and modified only via deliberate governance change.

---

## Layer 2: Instruction Packs

**Directory:** `.github/instructions/`

Modular, single-concern instruction files that provide focused guidance per language, platform, or cloud provider. Organized into four categories:

### Core Packs (`.github/instructions/core/`)

Foundational rules for every task:

| Pack | Content |
|---|---|
| `baseline` | Engineering operating model, plan-first, test discipline |
| `security` | Secret hygiene, input validation, SSRF, auth enforcement |
| `dependency` | CVE thresholds, native-first preference, justification rules |
| `identity` | Token validation, confused deputy prevention, trust planes |
| `carl` | cARLv2 phase model (shaping → planning → execution → validation → reset) |
| `cognition-governance` | Minimum sufficient depth, correction budget, model fallback |
| `tool-permission-tiers` | Tier 0 (read), Tier 1 (scoped write), Tier 2 (destructive) |
| `memory-cache` | Durable truth cache update discipline |
| `pr-contract` | Scope enforcement, assertion mapping, escalation triggers |

### Language Packs (`.github/instructions/languages/`)

Language-specific conventions and guardrails. Current packs:
- Python, TypeScript, JavaScript, Terraform, PowerShell, HTML

### Platform Packs (`.github/instructions/platform/`)

Deployment and infrastructure conventions:
- CI/CD, Docker, Kubernetes

### Cloud Packs (`.github/instructions/cloud/`)

Provider-specific security and operational guidance:
- Azure, Microsoft Entra ID, Microsoft Graph, Google Cloud Platform, Netlify

---

## Layer 3: Governance Artefacts

**Directory:** `.github/carl/`

Templates and data artefacts used by cARLv2 packs. These are not instruction-pack logic — they are the runtime state that makes governance durable across sessions.

| Artefact | Purpose |
|---|---|
| `memory.md` | Durable architectural truth cache: purpose, invariants, trust boundaries, field findings, open questions |
| `current-pr-contract.md` | Active PR scope contract: goal, approved scope, forbidden scope, constraints, stop conditions |
| `invariants.yml` | Machine-readable invariant set: secrets policy, scope discipline, security baseline, plan-first, dependency approval |
| `trust-boundaries.md` | Trust boundary classification table and crossing rules |
| `tool-policy.yml` | Tier 0/1/2 tool permission policy |
| `repo-map.example.json` | Cognitive repository map for fast agent orientation |
| `plans/README.md` | Prompt-as-code guidance and when to use it |
| `plans/plan-template.md` | Reusable planning contract template |

---

## cARLv2 Phase Model

The cARLv2 cognition governance model separates agent work into five phases:

```
┌──────────┐   ┌──────────┐   ┌───────────┐   ┌────────────┐   ┌───────────┐
│ Shaping  │──▶│ Planning │──▶│ Execution │──▶│ Validation │──▶│  Reset    │
│          │   │          │   │           │   │            │   │           │
│ Clarify  │   │ PR       │   │ Implement │   │ Compare    │   │ Archive   │
│ scope    │   │ contract │   │ inside    │   │ against    │   │ contract  │
│ Reduce   │   │ Contract │   │ approved  │   │ contract   │   │ Update    │
│ ambiguity│   │ assertions│  │ scope     │   │ not just   │   │ memory    │
│          │   │          │   │           │   │ tests      │   │           │
└──────────┘   └──────────┘   └───────────┘   └────────────┘   └───────────┘
```

**Key properties:**
- Phases are distinct; do not blend planning and execution
- PR contract constrains execution scope
- Validation checks contract compliance, not just test passage
- Context reset archives the contract and updates the memory cache

---

## Tool Permission Tier Model

```
Tier 0 ──── Read-only ──── No escalation required
            (read files, search code, inspect artefacts)

Tier 1 ──── Scoped write ── Declare intent, confirm scope
            (edit approved files, create approved paths)

Tier 2 ──── Destructive ─── Require explicit user approval
            (delete files, bulk changes, CI workflow edits)
```

---

## Prompt-as-Code Pattern

For substantial, long, or boundary-sensitive tasks, prefer a committed plan file over a UI prompt:

```
.github/carl/plans/prN-short-description.md
```

Benefits:
- Version-controlled and diffable
- Survives session resets and model switches
- Line-addressable for targeted corrections
- Immune to UI prompt truncation
- Shared with the team via git

---

## Composition Model

The root operating model (`.github/copilot-instructions.md`) is the composition point. It:

1. Defines the overarching operating model in full
2. References the cARLv2 packs for detailed cognition governance
3. Defers to individual packs for language, platform, and cloud specifics
4. Points to `.github/carl/` for durable artefacts

Individual packs are self-contained and can be used independently by copying them to another repository.

---

## Versioning

Each file carries a `<!-- version: X.Y.Z -->` header comment:

- **MAJOR** — breaking change to conventions (removes previously required behaviour)
- **MINOR** — new guidance added backwards-compatibly
- **PATCH** — clarifications, corrections, wording improvements
