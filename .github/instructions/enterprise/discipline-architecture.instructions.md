<!-- version: 1.0.0 -->
<!-- requires: core/cognition-governance, core/dependency, core/pr-contract -->
<!-- precedence-mode: additive -->
<!-- priority: 40 -->
# Enterprise Architecture Discipline Pack

Use this pack when the active profile role or task is architecture-focused.

## Adherence signal

At the first substantive response, include `Discipline: architecture`. In the
final response, include an `Architecture evidence` section containing the
boundaries, key decisions, and trade-offs considered.

## Required behaviour

- Map system, data, identity, deployment, ownership, and operational
  boundaries before proposing components.
- Distinguish reversible decisions from decisions that create durable coupling
  or migration cost.
- Prefer existing platform patterns and dependencies unless a documented gap
  justifies a new one.
- Quantify or bound availability, latency, throughput, recovery, data
  retention, and cost assumptions when they affect the decision.
- Describe failure modes, observability, rollout, rollback, and migration.
- Include threat or hazard analysis at the depth required by the active
  company scenario.
- Do not silently turn an architecture recommendation into an infrastructure
  or production mutation.

## Handoff

Produce decision-ready constraints and interfaces that code and implementation
work can validate without rediscovering intent.
