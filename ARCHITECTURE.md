<!-- version: 1.5.0 -->
# cARL — Architecture Overview

---

## Conceptual Architecture

cARL is a three-layer system:

```
┌─────────────────────────────────────────────────────────────┐
│                     Agent Session                           │
│  (Copilot, Claude Code, or Codex)                           │
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

This is the root governance file and shared adapter loader. GitHub Copilot
reads it directly; the production Claude Code and Codex adapters route their
agent sessions through it. It defines:

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

### Pack Metadata Model (schema version 1)

Packs are addressable units of policy, not just files. Each pack has a
versioned metadata record (`schemaVersion: 1`) derived from its canonical
file plus runtime state:

| Field | Meaning |
|---|---|
| `id` | Stable identity `<category>/<name>`, derived from the canonical path `.github/instructions/<category>/<name>.instructions.md` |
| `version` | Semantic version from the pack's `<!-- version: X.Y.Z -->` header |
| `title` / `description` | First `#` heading and first paragraph of the pack file |
| `category` | `core`, `languages`, `platform`, or `cloud` — must match the ID prefix |
| `source` | Where the pack was discovered: `bundled`, `repository-local`, both, or `registry:<id>` |
| `state` | `bundled` / `installed` / `selected` / `active` flags |
| `ownedArtifacts` | Repository paths the pack owns (currently its own instruction file) |
| `dependencies` | Pack IDs this pack requires, parsed from the `<!-- requires: ... -->` header (validated for existence and cycles) |
| `compatibility` | Optional minimum CLI / runtime version constraints |
| `precedence` | Optional priority + mode (`additive`, `overridable`, `restrictable-only`, `immutable`) + explicit overrides, parsed from `<!-- priority: N -->`, `<!-- precedence-mode: ... -->`, and `<!-- overrides: ... -->` headers |
| `provenance` | For registry-managed packs: explicit registry ID/location, relative artifact, and verified SHA-256 |

Pack state is a chain of distinct facts:

- **bundled** — shipped inside the `carl` binary,
- **installed** — present as a file in the repository,
- **selected** — recorded in `.github/carl/packs.json` (written by
  `carl pack select` / `unselect`; falls back to the legacy
  `.github/carl/runtime.json` derivation when `packs.json` is absent),
- **active** — explicitly activated by organisation/repository defaults plus
  the current named profile and its role/task overlays in
  `.github/carl/profiles.json`. When that artefact is absent, active falls
  back to selected for compatibility.

Discovery merges bundled, repository-local, and manifest sources
deterministically (sorted by pack ID); repository-local metadata takes
precedence over bundled metadata for the same ID. Filesystem enumeration
order is never policy order. `carl pack list` / `carl pack show` expose this
model (see [CLI.md](CLI.md)).

**Selection vs priority vs override authority.** These are deliberately
separate concepts: *selection* decides which packs are in play; *priority*
decides ordering among selected packs; *override authority* decides whether
one pack may relax another's rules. All three are modelled as of Pack
Phase 2: selection is an explicit committed artefact
(`.github/carl/packs.json`), while priority and override authority come only
from explicit pack metadata headers — never from load order.

**Effective pack set.** `carl pack effective` computes the composed policy
surface: explicitly selected packs plus transitively expanded required
dependencies, each entry carrying explicit reasons. Composition is
conservative — packs add constraints; an override is honoured only when
declared in explicit metadata *and* the target pack declares itself
`overridable`; overridden packs remain in the set flagged with
`overriddenBy`, so no pack silently disables another. Conflicts (missing
dependencies, overriding a non-overridable pack, mutual overrides) fail with
a non-zero exit. Ordering is priority descending with pack-ID tie-breaks. Any
future policy intermediate representation compiled from pack metadata builds
on this effective set and must never be inferred from load order.

### Profiles and Agent Roles (schema version 1)

`.github/carl/profiles.json` is a user-owned committed policy artefact. It
contains:

- additive organisation and repository default pack sets;
- named profiles with base pack sets;
- per-profile role and task overlays;
- one explicit active profile/role/task context.

Every referenced pack must already be selected. Profile resolution is
deterministic and additive: defaults, profile packs, the active role overlay,
and the active task overlay form the active seeds; Phase 2 dependency,
precedence, override, and conflict rules then produce the effective set.
`carl pack profile activate` and `clear` write only `profiles.json`;
`runtime.json` remains init-only. Invalid profile IDs, duplicate profiles,
unknown contexts, and unselected pack references are explicit errors.

### Registry and Installation (schema version 1)

Registry support is an optional boundary around the existing repository-local
pack model. There is no default registry and ordinary pack discovery,
selection, profile, and composition commands remain network-free.

`.github/carl/registries.json` records explicit HTTPS or repository-local
index locations. Each strict schema-versioned index advertises immutable pack
releases by ID, semantic version, relative artifact location, and SHA-256.
Resolution selects the highest semantic version or an exact requested version.
Equal winning versions from multiple registries are rejected as ambiguous
unless the caller names a registry.

`carl pack install` downloads only relative, same-origin artifacts, checks
bounded response sizes, verifies SHA-256 and pack-declared metadata, resolves
unavailable required dependencies from the same registry, and validates the
whole planned pack set before mutation. Writes are confined to canonical
instruction-pack paths plus `.github/carl/installed-packs.json` and roll back
if a transaction fails. Installation does not alter selection or activation.

The installed-pack manifest is committed provenance rather than runtime
authority. It records the configured registry, artifact, digest, version, and
installed path. Discovery exposes that provenance and rejects digest drift.
`carl pack update` uses the recorded source, rejects local drift and
same-version registry mutation, and never downgrades.

SHA-256 establishes artifact integrity relative to the configured index. It
does not authenticate a publisher, establish a signing-key trust root, or
provide a signing trust chain.

### Policy Provenance and Explanation (schema version 1)

`carl explain <pack-id>` and `carl trace` are read-only diagnostic views over
the existing pack evaluator. They do not introduce a second policy authority
or evaluation algorithm:

1. strict discovery validates bundled, repository-local, registry-managed,
   selected, and profile-driven state;
2. `ComputeEffectiveSet` remains the source of dependency, precedence,
   override, and conflict truth;
3. the explanation layer attaches repository-relative canonical definitions
   and structured source artefacts to those observable results.

The explained policy unit is an instruction pack. The runtime does not parse
natural-language pack prose into individual rules and has no compiled policy
intermediate representation. An effective pack therefore reports that it adds
constraints at pack level; permitted override relationships name their source
and target packs. Overridden packs remain visible, while invalid and mutual
overrides remain unresolved conflicts with non-zero exits.

The trace order is identical to `carl pack effective`: priority descending,
then pack ID. Activation steps distinguish legacy selection, organisation and
repository defaults, named profiles, role/task overlays, and dependency
expansion. Registry-managed definitions retain their verified artifact
provenance, but registry integrity does not become policy authority.

Explanation output is derived diagnostic evidence, not canonical governance.
Both commands are local-only, network-free, and make no repository writes.
They explicitly do not expose prompts, hidden model reasoning, or
chain-of-thought.

### Cognitive Repository Graph (schema version 1)

`carl map` preserves its original repository inventory fields and adds a
schema-versioned `graph` to `.github/carl/repo-map.json`.

The graph uses stable repository-relative IDs for the repository root,
components, Go packages, entry points, workflows, governance artefacts,
documentation, and instruction-pack policy definitions. `contains` edges
represent structural containment. `depends_on` edges come only from statically
parsed repository-local Go imports and carry the source files as evidence.
Direct reverse dependency edges populate each target node's `change_impact`.

Nodes provide deterministic orientation metadata:

- purpose and agent context;
- `high`, `medium`, or `low` criticality as a navigation heuristic;
- repository, governance, policy, or automation trust-boundary classification;
- explicit component/package policy attachment points;
- direct dependency-based change impact where available.

The graph records evidence coverage separately for ownership, dependencies,
data flows, trust boundaries, criticality, policy attachments, and change
impact. Static imports are not runtime data-flow proof, ownership is never
guessed, and policy attachment points do not claim policy activation. Active
policy provenance remains owned by `carl trace`.

Graph output is derived orientation evidence, not canonical governance,
ownership, risk, or runtime evidence. Existing `carl reconcile` consumers
continue to read the unchanged inventory fields and ignore the additive graph.

---

## Layer 3: Governance Artefacts

**Directory:** `.github/carl/`

Templates and data artefacts used by cARLv2 packs. These are not instruction-pack logic — they are the runtime state that makes governance durable across sessions.

| Artefact | Purpose |
|---|---|
| `memory.md` | Durable architectural truth cache: purpose, invariants, trust boundaries, field findings, open questions |
| `current-pr-contract.md` | Active PR scope contract: goal, approved scope, forbidden scope, constraints, stop conditions |
| `current-pr-contract.template.md` | Blank template — copy to `current-pr-contract.md` at the start of each new PR |
| `invariants.yml` | Machine-readable invariant set: secrets policy, scope discipline, security baseline, plan-first, dependency approval |
| `trust-boundaries.md` | Trust boundary classification table and crossing rules |
| `tool-policy.yml` | Tier 0/1/2 tool permission policy |
| `profiles.json` | User-owned named profiles, defaults, role/task overlays, and active context (created only when configured) |
| `registries.json` | User-owned explicit HTTPS or repository-local pack registry configuration (created only when configured) |
| `installed-packs.json` | User-owned deterministic provenance for SHA-256-verified registry-managed packs (created on first install) |
| `repo-map.example.json` | Schema-versioned cognitive repository inventory and graph for fast agent orientation |
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
