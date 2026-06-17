<!-- version: 1.2.0 -->
# Cognition Governance Pack

Defines how reasoning depth is selected and when ambiguity must be reduced before execution.

-   **Use the minimum sufficient reasoning depth.** Start with the lightest depth that can safely satisfy the task, then escalate only if required.
-   **Reasoning depth is orthogonal to operating mode.** Plan-only, Assisted implementation, and Automatic modes do not force a fixed reasoning depth.
-   **Classify uncertainty before escalation.** Identify whether uncertainty is routine, ambiguous, novel, or conflicting before increasing reasoning depth.
-   **Reduce ambiguity before costly execution.** If scope, constraints, or trust-boundary effects are unclear, pause and clarify before executing.
-   **Avoid assumption-driven execution.** Do not convert unresolved ambiguity into implicit assumptions that drive code or configuration changes.
-   **Escalate when risk dominates.** If wrong assumptions could violate invariants or create high rework cost, escalate to explicit user clarification.
-   **Treat model availability as unstable.** Do not assume a named model will remain available or equally capable across sessions.
-   **Preserve the contract across model fallback.** If switching models, keep the same goal, non-goals, scope, invariants, and acceptance criteria unless the user explicitly amends them.
-   **Use a correction budget.** One corrective prompt is acceptable. Two means reset the session. Three means abandon the session/model and restart from a clearer plan.
-   **Treat test drift as comprehension failure.** If tests validate implementation drift instead of the approved contract, treat that as a contract-comprehension failure rather than successful validation.
-   **Escalate repeated assertion corrections.** If the same contract assertion must be corrected more than once in implementation or tests, consume the correction budget and reset the session or switch models instead of continuing prompt repair.
