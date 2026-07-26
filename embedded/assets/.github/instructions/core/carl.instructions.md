<!-- version: 1.5.0 -->
# cARLv2 Cognition Governance Pack

Defines the cARLv2 governance model that coordinates shaping, planning, execution, validation, reconciliation, and context reset.

## Authority model

cARL artefacts are the canonical source of durable governance truth.

Harness-specific files such as `.github/copilot-instructions.md`, `CLAUDE.md`, `AGENTS.md`, `.cursor/rules/carl.mdc`, and `.agents/rules/carl.md` are adapters/loaders. They may load, summarise, or route agents toward cARL, but they are not independent authorities.

If prompt/session memory conflicts with cARL artefacts, trust cARL and report the conflict.

If `.github/carl/memory.md` conflicts with current repository state, current repository state wins and memory should be updated.

## Pack/profile hydration

The shared loader at `.github/copilot-instructions.md` is the normative agent
procedure for consuming canonical pack/profile state. Follow its **Effective
instruction-pack hydration** section before planning or implementation.

In particular:

- directory presence, installation, or discovery does not select a pack;
- `.github/carl/packs.json` is the selection authority when present, otherwise
  selection uses the documented `.github/carl/runtime.json`
  `managedArtifacts` compatibility derivation;
- `.github/carl/profiles.json`, when present, makes organisation/repository
  defaults plus the configured active profile and role/task overlays the
  additive active seed set; without it, selected packs are active
  compatibility seeds;
- `.github/carl/profiles.example.json` is an inactive, cloneable reference
  baseline using the same profile schema; its presence selects or activates
  nothing, and only an adopted `.github/carl/profiles.json` participates in
  evaluation;
- transitive `requires:` dependencies, explicit precedence metadata, priority
  descending with pack-ID tie-breaking, and valid explicit overrides determine
  the composed set;
- only effective, non-overridden pack definitions are hydrated and applied;
- invalid, contradictory, cyclic, conflicting, or ambiguous policy state
  requires stopping and reporting rather than guessing or loading all packs.

The cARL CLI may validate or explain this repository-local evaluation when
available, but it is not a runtime dependency for hydration. Effective,
explain, and trace output is derived diagnostic evidence, not canonical policy
state or proof that a model adhered to loaded instructions.

## Lifecycle

- **Delegated cognition is a governed resource.** Treat agent cognition as accountable project capacity, not ambient background activity.
- **Separate work phases deliberately.** Keep shaping, planning, execution, validation, reconciliation, and context reset distinct to reduce hidden branching.
- **Hydrate before planning or implementation.** Read the active PR contract, durable memory, invariants, trust boundaries, tool policy, runtime/selection/profile state, relevant plans, and only the effective instruction packs before making changes.
- **Constrain execution with a PR contract.** Use `.github/carl/current-pr-contract.md` to define approved scope, constraints, non-goals, stop conditions, and escalation triggers.
- **Preserve invariants.** Preserve `.github/carl/invariants.yml` unless the user explicitly approves a governance amendment.
- **Use the minimum sufficient reasoning depth.** Increase depth only when uncertainty, novelty, or risk warrants the additional cost.
- **Preserve primary engineering goals.** Correctness, security, maintainability, testability, and explainability remain primary objectives across all phases.
- **Reduce ambiguity before expensive or autonomous execution.** Clarify uncertain requirements before broad changes, high-impact tool use, or autonomous execution steps.
- **Plan contract assertions before implementation.** For non-trivial work, identify contract-critical behaviours, choose 3-5 contract assertions, and map acceptance criteria to direct tests before execution begins.
- **Reuse durable knowledge.** Use `.github/carl/memory.md` as a durable architectural truth cache to avoid repeated semantic rediscovery.
- **Enforce tool-permission tiers.** Apply tiered tool governance via `.github/carl/tool-policy.yml` and `tool-permission-tiers.instructions.md`.
- **Use prompt-as-code for substantial tasks.** Store long, nested, boundary-sensitive, model-comparison, or trust-boundary-changing task contracts in `.github/carl/plans/` so prompts are version-controlled, diffable, and line-addressable.
- **Prefer committed plan files for substantial work.** Use committed plan files for long, nested, boundary-sensitive, or model-comparison tasks, preferably `.github/carl/plans/prN-short-description.md`.
- **Read the plan before implementation.** For substantial work, the agent should read the plan file and respond in Plan-only mode before implementation.
- **Archive temporary root plans before merge.** A temporary `PLAN.md` is acceptable on a feature branch, but it should be removed or archived before merge.
- **Validate contract, implementation, and tests together.** During validation, compare the approved contract against the implementation and tests. Reject tests that encode drift. Verify exact schema, output, error, and failure semantics whenever the contract specifies them.
- **Reconcile before final response.** Decide whether the change affects durable project truth. If it does, update relevant documentation and cARL artefacts. If it does not, explicitly state why no cARL/docs update was required.
- **Do not turn memory into a session diary.** Update durable artefacts only when stable facts, decisions, invariants, trust boundaries, roadmap items, operating model changes, or recurring workflow hazards should persist.
- **Stop prompt ping-pong early.** If more than one corrective prompt is required to understand the PR contract, reset the session or switch models instead of continuing to patch a failing mental frame.

## Model and harness reality

Instruction availability is not instruction adherence.

A harness may load an instruction file, but different models vary in their ability to operationalise the full governance lifecycle without explicit checkpoints.

Therefore harness loaders should make the cARL lifecycle explicit:

1. hydrate cARL;
2. apply governance;
3. validate against contract;
4. reconcile durable artefacts;
5. report the cARL/docs update decision.

Model availability and capability are not stable invariants. If the preferred model is unavailable or repeatedly misinterprets the contract, switch model/session without changing the PR contract, non-goals, invariants, or acceptance criteria unless the user explicitly amends them.
