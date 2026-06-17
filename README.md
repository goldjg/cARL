<!-- version: 1.0.0 -->
# cARL — Cognitive Agent Runtime Layer

> **"cARL remembers why you made that decision three months ago, because neither you nor your coding agent will."**

---

## What is cARL?

cARL (Cognitive Agent Runtime Layer) is a version-controlled governance and instruction layer that sits inside your repository and shapes how AI coding agents behave — every session, every task, without re-prompting.

It is not a framework. It is not a library. It is a set of committed files that give your agent persistent memory, engineering discipline, and bounded execution contracts.

---

## The Problem cARL Solves

AI coding agents are powerful but amnesiac. Every session starts from scratch. Each new task re-discovers the same architectural decisions, security constraints, and engineering conventions — or ignores them entirely.

The result is inconsistent behaviour, security regressions, dependency sprawl, and agents that confidently do the wrong thing because nobody told them otherwise — again.

cARL solves this by providing:

| Problem | cARL solution |
|---|---|
| Agent forgets architecture decisions | Durable truth cache in `.github/carl/memory.md` |
| Agent ignores engineering standards | Modular instruction packs loaded every session |
| Agent drifts out of scope mid-task | PR contracts constrain execution to approved scope |
| Agent uses wrong tools recklessly | Tiered tool-permission governance |
| Agent over-reasons or under-reasons | Cognition governance: minimum sufficient depth |
| Session is lost, work is lost | Prompt-as-code plans in `.github/carl/plans/` |

---

## How cARL Differs from ADRs

Architecture Decision Records (ADRs) are documents written for humans. They record decisions after the fact and are read by developers during onboarding or code review.

cARL is agent-readable, partly structured governance written for agents. It is loaded before every task, not consulted after the fact. Where an ADR records why a decision was made, cARL enforces what the agent should and should not do as a result.

| | ADR | cARL |
|---|---|---|
| **Audience** | Human developers | AI coding agents |
| **When used** | After decision, during review | Before every task |
| **Format** | Prose narrative | Structured instruction packs + YAML |
| **Enforcement** | None (human discipline) | Loaded into agent context every session |
| **Scope** | Single decision | Entire operating model |

cARL complements ADRs. Use ADRs for human record-keeping. Use cARL to make those decisions machine-enforceable.

---

## How cARL Differs from Prompt Engineering

Prompt engineering is per-task. You write a prompt, get a result, and the instruction vanishes when the session ends.

cARL is persistent. Instructions live in committed files. They are version-controlled, diffable, auditable, and always in scope. They survive model upgrades, session resets, and team member turnover.

| | Prompt Engineering | cARL |
|---|---|---|
| **Persistence** | Per-session | Version-controlled |
| **Discoverability** | Lost when session ends | Always in the repository |
| **Consistency** | Re-typed or forgotten | Loaded automatically every session |
| **Auditability** | None | Full git history |
| **Team sharing** | Copy-paste | Committed and shared via version control |

---

## How cARL Differs from Agent Frameworks

Agent frameworks (LangChain, AutoGen, CrewAI, etc.) are code. They orchestrate agents programmatically, define tool schemas, and wire up models in code.

cARL is behavioural governance. It does not execute code. It shapes how an agent reasons, plans, and acts inside an existing tool (GitHub Copilot). No new runtime dependencies. No code to run.

| | Agent Frameworks | cARL |
|---|---|---|
| **Nature** | Code libraries | Committed governance files |
| **Deployment** | Runtime dependency | Repository files |
| **Target** | New agent applications | Existing coding agents (Copilot) |
| **Concern** | Orchestration and tooling | Behavioural governance and discipline |
| **Language coupling** | Yes | Language-agnostic |

---

## How cARL Evolved from AADLC

cARL is the productised form of AADLC (Autonomous Agent Development Lifecycle), an internal governance model developed through practitioner experience with GitHub Copilot coding agents.

AADLC established the core concepts:
- Phase separation (shaping → planning → execution → validation → reset)
- Durable memory caches
- PR contracts
- Tool permission tiers
- Prompt-as-code

cARL takes these concepts, renames them for clarity, organises them into a coherent product, and adds the documentation and positioning needed for broader adoption.

The runtime semantics are unchanged. This is a rename and productisation, not a redesign. The `aadlc.instructions.md` file is now `carl.instructions.md`. The `.github/aadlc/` directory is now `.github/carl/`. All governance behaviour is preserved.

---

## Repository Structure

```
.github/
├── copilot-instructions.md          # Root operating model — loaded by Copilot automatically
├── carl/
│   ├── memory.md                    # Durable architectural truth cache
│   ├── current-pr-contract.md          # Active PR contract (populate before each PR)
│   ├── current-pr-contract.template.md # Blank template — copy to current-pr-contract.md
│   ├── invariants.yml               # Machine-readable governance invariants
│   ├── trust-boundaries.md          # Trust boundary definitions and crossing rules
│   ├── tool-policy.yml              # Tool permission tier policy (Tier 0/1/2)
│   ├── plans/
│   │   ├── README.md                # Prompt-as-code guidance for substantial tasks
│   │   └── plan-template.md         # Reusable cARL planning contract template
│   └── repo-map.example.json        # Example cognitive repo map for fast orientation
└── instructions/
    ├── core/
    │   ├── baseline.instructions.md
    │   ├── security.instructions.md
    │   ├── dependency.instructions.md
    │   ├── identity.instructions.md
    │   ├── carl.instructions.md              # cARLv2 cognition governance phase model
    │   ├── cognition-governance.instructions.md
    │   ├── tool-permission-tiers.instructions.md
    │   ├── memory-cache.instructions.md
    │   └── pr-contract.instructions.md
    ├── languages/
    │   ├── python.instructions.md
    │   ├── typescript.instructions.md
    │   ├── javascript.instructions.md
    │   ├── terraform.instructions.md
    │   ├── powershell.instructions.md
    │   └── html.instructions.md
    ├── platform/
    │   ├── cicd.instructions.md
    │   ├── docker.instructions.md
    │   └── kubernetes.instructions.md
    └── cloud/
        ├── azure.instructions.md
        ├── entra.instructions.md
        ├── microsoft-graph.instructions.md
        ├── gcp.instructions.md
        └── netlify.instructions.md

README.md
VISION.md
ARCHITECTURE.md
ROADMAP.md
GLOSSARY.md
```

---

## How It Works

1. GitHub Copilot automatically reads `.github/copilot-instructions.md` at the start of every agent session.
2. That file is the root operating model — plan-first discipline, security constraints, cognition governance.
3. Individual packs under `.github/instructions/` provide focused guidance per language, platform, or cloud provider.
4. `.github/carl/` contains durable governance artefacts: memory cache, PR contract, invariants, trust boundaries, and plans.

---

## Usage

### Using this repository directly

Fork or copy into your GitHub account or organisation. Copilot picks up the instructions automatically.

### Copying packs into another repository

Copy any individual pack from `.github/instructions/` into the same path in your target repository. For full cARLv2 usage, also copy `.github/carl/` and populate the artefacts for your project.

### Versioning

Each file carries a `<!-- version: X.Y.Z -->` comment. Increment using Semantic Versioning:
- **MAJOR** — breaking change to conventions
- **MINOR** — new guidance added compatibly
- **PATCH** — clarifications or corrections

---

## Pack Summary

### Core Packs
| Pack | Purpose |
|---|---|
| `baseline` | Engineering operating model, plan-first, test discipline |
| `security` | No secrets, input validation, SSRF prevention |
| `dependency` | CVE management, native-first, justification required |
| `identity` | Token validation, trust boundary discipline |
| `carl` | cARLv2 phase model and cognition governance |
| `cognition-governance` | Minimum sufficient reasoning depth |
| `tool-permission-tiers` | Tier 0/1/2 tool classification and escalation |
| `memory-cache` | Durable truth cache governance |
| `pr-contract` | Scoped implementation contract lifecycle |

### Language Packs
Python · TypeScript · JavaScript · Terraform · PowerShell · HTML

### Platform Packs
CI/CD · Docker · Kubernetes

### Cloud Packs
Azure · Microsoft Entra ID · Microsoft Graph · Google Cloud Platform · Netlify

---

## Contributing

1. Open an issue describing the gap or new pack.
2. Follow naming: `<name>.instructions.md` in the appropriate subdirectory.
3. Keep each pack focused on a single concern.
4. Add `<!-- version: 1.0.0 -->` at the top of any new file.
5. Update this README.

---

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned evolution.

---

## License

See repository root for licence information.
