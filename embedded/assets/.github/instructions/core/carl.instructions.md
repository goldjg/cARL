<!-- version: 1.2.0 -->
# cARLv2 Cognition Governance Pack

Defines the cARLv2 governance model that coordinates shaping, planning, execution, validation, and context reset.

-   **Delegated cognition is a governed resource.** Treat agent cognition as accountable project capacity, not ambient background activity.
-   **Separate work phases deliberately.** Keep shaping, planning, execution, validation, and context reset distinct to reduce hidden branching.
-   **Use the minimum sufficient reasoning depth.** Increase depth only when uncertainty, novelty, or risk warrants the additional cost.
-   **Preserve primary engineering goals.** Correctness, security, maintainability, and testability remain primary objectives across all phases.
-   **Reduce ambiguity before expensive or autonomous execution.** Clarify uncertain requirements before broad changes, high-impact tool use, or autonomous execution steps.
-   **Constrain execution with a PR contract.** Use `.github/carl/current-pr-contract.md` to define approved scope, constraints, and escalation triggers.
-   **Plan contract assertions before implementation.** For non-trivial work, identify contract-critical behaviors, choose 3-5 contract assertions, and map acceptance criteria to direct tests before execution begins.
-   **Reuse durable knowledge.** Use `.github/carl/memory.md` as a durable architectural truth cache to avoid repeated semantic rediscovery.
-   **Enforce tool-permission tiers.** Apply tiered tool governance via `.github/carl/tool-policy.yml` and `tool-permission-tiers.instructions.md`.
-   **Use prompt-as-code for substantial tasks.** Store long or boundary-sensitive task contracts in `.github/carl/plans/` so prompts are version-controlled, diffable, and line-addressable.
-   **Prefer committed plan files for substantial work.** Use committed plan files for long, nested, boundary-sensitive, or model-comparison tasks, preferably `.github/carl/plans/prN-short-description.md`.
-   **Read the plan before implementation.** For substantial work, the agent should read the plan file and respond in Plan-only mode before implementation.
-   **Archive temporary root plans before merge.** A temporary `PLAN.md` is acceptable on a feature branch, but it should be removed or archived before merge.
-   **Stop prompt ping-pong early.** If more than one corrective prompt is required to understand the PR contract, reset the session or switch models instead of continuing to patch a failing mental frame.
-   **Validate contract, implementation, and tests together.** During validation, compare the approved contract against the implementation and tests, reject tests that encode drift, and verify exact schema and failure semantics whenever the contract specifies them.
