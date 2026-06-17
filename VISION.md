<!-- version: 1.0.0 -->
# cARL — Project Vision

---

## Vision Statement

**cARL makes AI-assisted software development consistent, governed, and durable — across sessions, across teams, and across models.**

Coding agents are increasingly capable. What they lack is discipline by default. cARL provides that discipline as committed infrastructure: version-controlled, auditable, and always in scope.

---

## The Problem We Are Solving

AI coding agents lose context. Every new session begins from zero. Engineering standards, architectural decisions, security requirements, and scope boundaries must be re-communicated on every task — or they are ignored.

This creates a gap between what engineering teams expect and what agents deliver:

- Agents make decisions without knowing the established patterns
- Agents introduce dependencies without knowing the dependency policy
- Agents drift out of scope because no PR contract constrained them
- Agents repeat the same mistakes because field-test lessons were never persisted
- Agents produce correct-looking but architecturally wrong outputs

The result is inconsistent code quality, security regressions, and a false sense of acceleration that masks accumulating technical debt.

---

## What cARL Is

cARL is a **Cognitive Agent Runtime Layer** — a set of committed files inside your repository that govern how an AI coding agent behaves before, during, and after each task.

It is:

- **A governance layer**, not a framework
- **A memory layer**, not a session log
- **A contract layer**, not a prompt template
- **A discipline layer**, not a style guide
- **Version-controlled infrastructure**, not ad-hoc instructions

cARL does not replace the agent. It shapes the agent — providing the operating context that makes agent behaviour predictable, safe, and aligned with engineering intent.

---

## Core Beliefs

**1. Agents should be accountable, not ambient.**
Delegated cognition is a governed resource. Agent tasks should have explicit contracts, defined scope, and clear stop conditions — just like human pull requests.

**2. Memory should outlast sessions.**
Architectural decisions, invariants, and field-test lessons should persist in the repository, not in someone's head or a stale prompt file.

**3. Governance should be version-controlled.**
If it matters, it should be in git. Engineering conventions, security rules, and planning contracts should be diffable, reviewable, and auditable.

**4. Packs should be modular and composable.**
No single monolithic instruction file. Focused packs per concern, composed at the root operating model, reusable across repositories.

**5. The agent should do less, not more, when uncertain.**
Minimum sufficient reasoning depth. Escalate when risk warrants. Default to plan-first. Pause before destructive actions.

---

## Who cARL Is For

| Audience | How cARL helps |
|---|---|
| **Individual developers** | Consistent agent behaviour across tasks without re-prompting |
| **Engineering teams** | Shared, version-controlled conventions that apply to every team member's agent sessions |
| **Security-conscious teams** | Codified security baselines that cannot be forgotten or bypassed |
| **Platform / DevEx teams** | A reusable governance layer to standardise agent behaviour across multiple repositories |
| **Tech leads and architects** | Durable memory for architectural decisions that agents must respect |

---

## What cARL Is Not

- cARL is **not an agent framework** — it does not orchestrate agents or define tool schemas
- cARL is **not a prompt library** — it is committed infrastructure, not copy-paste prompts
- cARL is **not a linter** — it governs reasoning and behaviour, not syntax
- cARL is **not a replacement for code review** — it reduces drift, it does not eliminate human judgment
- cARL is **not a solution to model limitations** — it reduces their impact, it does not eliminate them

---

## The Desired Future State

A developer opens a coding agent session. Without writing a single instruction, the agent:

1. Knows the repository's architecture and constraints from the memory cache
2. Plans before acting, and pauses for approval before implementing
3. Stays within the scope of the PR contract
4. Applies the correct security, dependency, and identity rules for the language and platform
5. Records durable lessons and decision rationale for the next session

This is the future cARL is building toward — one repository at a time.
