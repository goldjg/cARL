<!-- version: 1.2.0 -->
# PR Contract Pack

Defines the PR contract controls that constrain implementation scope and govern escalation.

-   **Create a contract before implementation.** Use `.github/carl/current-pr-contract.md` to capture goal, non-goals, scope, constraints, and stop conditions.
-   **Treat approved scope as execution boundary.** Planned implementation must remain inside approved scope unless explicitly amended.
-   **Enforce forbidden scope explicitly.** Do not execute changes that fall into forbidden scope without user-approved contract amendment.
-   **Honor architectural and security constraints.** Contract constraints are mandatory guardrails during implementation and validation.
-   **Validate against expected file and test surfaces.** Keep changes aligned to expected files and listed validation commands.
-   **Record contract assertions when the work is test-critical.** For output-schema-heavy, validation-heavy, trust-boundary-changing, CLI/API-facing, persistence-format-sensitive, security-relevant, or failure-mode-sensitive work, list 3-5 explicit contract assertions in the active PR contract or linked plan before implementation.
-   **Require direct test linkage to contract assertions.** Tests that do not trace directly to listed contract assertions are insufficient evidence of compliance, even when they pass.
-   **Use escalation triggers and reset notes.** Escalate when trigger conditions are met and document context reset expectations on completion.
-   **Avoid stale contract anchoring.** Do not treat a completed PR constraint as durable law unless it has been promoted to an invariant.
-   **Carry forward only promoted invariants.** Completed PR contracts are historical evidence, not binding scope, unless the current contract explicitly promotes a constraint.
-   **Contract amendments must be explicit.** If a PR intentionally changes a previous constraint or introduces a new trust boundary, record the amendment in the current PR contract or plan file before implementation.
-   **Do not over-anchor on prior contracts.** New PRs may intentionally amend prior constraints when the amendment is explicit and scoped in the current contract.
-   **Approve new trust boundaries explicitly.** Treat a newly introduced trust boundary as approved only when the current PR contract or plan file records it as an intentional amendment.
