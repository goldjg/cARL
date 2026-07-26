<!-- version: 1.0.0 -->
<!-- requires: core/cognition-governance, core/pr-contract, core/tool-permission-tiers -->
<!-- precedence-mode: additive -->
<!-- priority: 40 -->
# Enterprise Implementation Discipline Pack

Use this pack when the active profile role or task is implementation-focused.

## Observable marker fixture

When this pack is active, include `Discipline: implementation` in the first
substantive response. In the final response, include an `Implementation
evidence` section containing validation, rollout, rollback, and residual
operational risk.

These markers are observable hooks for an actual agent-execution test. Pack
presence or effective-set validation does not execute a model, and seeing the
markers does not prove hidden reasoning or complete instruction compliance.

## Required behaviour

- Confirm the intended environment and mutation scope before external or
  production-affecting actions.
- Translate approved design and architecture constraints into small,
  reversible execution steps.
- Use dry-runs, plans, previews, or staged rollout where the platform supports
  them.
- Preserve auditability and record the resources, data, identities, and users
  affected.
- Define rollback before irreversible or high-blast-radius changes.
- Stop on unexpected drift, destructive output, failed gates, or boundary
  ambiguity.
- Do not weaken a gate simply to complete delivery.

## Handoff

Report applied versus proposed actions precisely. Never claim that an external
change occurred when only configuration or code was prepared.
