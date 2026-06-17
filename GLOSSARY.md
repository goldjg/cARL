<!-- version: 1.0.0 -->
# cARL — Glossary

This glossary defines the terms used in cARL documentation, instruction packs, and governance artefacts.

---

## A

### Agent
An AI model operating with tool access and autonomous execution capability within a software development context. In cARL, the primary agent is GitHub Copilot's coding agent.

### Agent session
A single continuous interaction between a user and an agent, from invocation to task completion. Agent sessions are stateless by default — cARL provides the persistent context.

### Artefact
A file in `.github/carl/` that carries durable governance state across sessions. Artefacts include the memory cache, PR contract, invariants, trust boundaries, tool policy, and plan files.

### Assisted implementation mode
An operating mode in which the agent implements changes after receiving explicit user approval of a plan. See also: Plan-only mode, Automatic mode.

### Automatic mode
An operating mode in which the agent implements changes end-to-end without seeking confirmation between steps. Activated only when the user explicitly requests it.

---

## C

### cARL
**Cognitive Agent Runtime Layer.** A version-controlled governance and instruction layer that shapes how AI coding agents behave across sessions. The productised form of AADLC.

### cARLv2
The second major version of the cARL cognition governance model. Defines the five-phase discipline (shaping, planning, execution, validation, context reset) and the supporting artefacts.

### Cognition governance
The discipline of managing how an agent reasons, at what depth, and with what constraints. cARLv2 cognition governance covers minimum sufficient reasoning depth, correction budgets, and model fallback.

### Contract assertion
An explicit, verifiable statement of expected behaviour that a test must directly prove. Contract assertions are identified before implementation (not derived from it) and are listed in the PR contract or plan file.

### Correction budget
The maximum number of corrective prompts allowed before resetting a session. In cARLv2: one corrective prompt is acceptable; two means reset; three means abandon and restart with a clearer plan.

### Current PR contract
The active PR contract file at `.github/carl/current-pr-contract.md`. Defines the goal, approved scope, forbidden scope, constraints, and stop conditions for the current task.

---

## D

### Delegated cognition
Agent reasoning and execution performed on behalf of a human. cARL treats delegated cognition as a governed resource: accountable, scoped, and subject to contract constraints.

### Durable truth cache
See: Memory cache.

---

## E

### Escalation trigger
A condition defined in the PR contract or plan that requires the agent to stop and seek user confirmation before proceeding. Examples: ambiguous scope, trust boundary change, out-of-contract action.

---

## G

### Governance artefact
See: Artefact.

---

## I

### Invariant
A governance rule that must be preserved unless explicitly amended through a user-approved governance change. Invariants are defined in `.github/carl/invariants.yml`.

### Instruction pack
A single-concern `.md` file in `.github/instructions/` that provides focused guidance for a language, platform, cloud provider, or core engineering concern.

---

## M

### Memory cache
The durable architectural truth cache at `.github/carl/memory.md`. Stores stable facts about the repository's purpose, invariants, trust boundaries, field findings, and open questions. Persists across sessions and model switches.

### Minimum sufficient reasoning depth
The lightest reasoning depth that can safely satisfy a task. cARLv2 governance requires starting at minimum depth and escalating only when uncertainty, risk, or novelty warrants the additional cost.

### Model fallback
Switching to an alternative AI model when the preferred model is unavailable or repeatedly misinterprets the PR contract. cARLv2 requires that the PR contract, invariants, and acceptance criteria are preserved across model switches.

---

## O

### Operating model
The root governance document at `.github/copilot-instructions.md`. Defines the agent's default behaviour, operating modes, core principles, security baseline, and cognition governance overview.

### Open question
A high-impact unresolved question persisted in the memory cache so future tasks do not rediscover the same ambiguity.

---

## P

### Pack
See: Instruction pack.

### Phase
One of the five cARLv2 work phases: shaping, planning, execution, validation, context reset. Phases should be kept distinct to reduce hidden branching.

### Plan-as-code
See: Prompt-as-code.

### Plan-only mode
The default operating mode. The agent outputs a plan without making code changes or running tests. Switches to Assisted implementation mode only on explicit user approval.

### PR contract
A scoped implementation contract for a pull request. Defines goal, approved scope, forbidden scope, architectural constraints, security constraints, stop conditions, and escalation triggers. Stored at `.github/carl/current-pr-contract.md`.

### Prompt-as-code
The practice of storing task instructions as committed plan files in `.github/carl/plans/` rather than writing them in the agent UI. Makes task contracts version-controlled, diffable, auditable, and immune to prompt truncation.

### Prompt ping-pong
A failure pattern in which repeated corrective prompts fail to fix a misunderstood contract. cARLv2 requires resetting the session or switching models after one corrective prompt rather than continuing to patch a failing mental frame.

---

## R

### Repo map
A structured file (`.github/carl/repo-map.example.json`) that provides a cognitive orientation map of a repository: directory purposes, key files, and pack registry. Reduces agent time-to-orientation.

### Runtime layer
The layer of committed files that an agent reads to understand its operating context. In cARL, this includes the root operating model, instruction packs, and governance artefacts.

---

## S

### Scoped write
A Tier 1 tool action that modifies a specific, bounded target inside approved PR scope. Requires intent declaration and scope confirmation before execution.

### Semantic rediscovery
The cost of an agent re-deriving the same architectural facts, constraints, and conventions in every session because they were not persisted. The memory cache eliminates semantic rediscovery.

### Session
See: Agent session.

### Shaping
The first cARLv2 phase. The agent clarifies scope, reduces ambiguity, and identifies constraints before planning or implementing.

### Stop condition
A condition defined in the PR contract that requires the agent to halt implementation immediately, regardless of progress.

---

## T

### Test drift
A failure pattern in which tests validate the implementation's current behaviour rather than the approved contract assertions. cARLv2 treats test drift as a contract-comprehension failure.

### Tier 0 / Tier 1 / Tier 2
The three tool permission tiers defined in `.github/carl/tool-policy.yml`:
- **Tier 0** — Read-only; no escalation required
- **Tier 1** — Scoped write; requires intent declaration and scope confirmation
- **Tier 2** — Destructive or broad; requires explicit user approval and contract coverage

### Trust boundary
A classification of an information source that defines how much validation is required before using it to make planning or implementation decisions. Defined in `.github/carl/trust-boundaries.md`.

---

## V

### Validation (phase)
The fourth cARLv2 phase. The agent compares the implementation and tests against the approved contract, not just against test passage. Rejects tests that encode drift rather than contract compliance.

### Version-controlled governance
The principle that engineering conventions, security rules, and planning contracts should be committed to source control — diffable, reviewable, and auditable — rather than stored in prompts, wikis, or tribal knowledge.

---

## Origin Terms

These terms refer to the predecessor project from which cARL evolved:

### AADLC
**Autonomous Agent Development Lifecycle.** The internal governance model that preceded cARL. The source of the core concepts: phase separation, durable memory, PR contracts, tool tiers, and prompt-as-code. Renamed to cARL for broader adoption.

### AADLCv2
The second major version of the AADLC model. Directly equivalent to cARLv2 in semantics and behaviour. The rename is cosmetic; runtime behaviour is unchanged.
