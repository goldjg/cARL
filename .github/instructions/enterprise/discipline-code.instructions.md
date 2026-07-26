<!-- version: 1.0.0 -->
<!-- requires: core/dependency, core/pr-contract -->
<!-- precedence-mode: additive -->
<!-- priority: 40 -->
# Enterprise Code Discipline Pack

Use this pack when the active profile role or task is code-focused.

## Observable marker fixture

When this pack is active, include `Discipline: code` in the first substantive
response. In the final response, include a `Code evidence` section containing
changed behaviour, tests, and any deliberately avoided dependency or refactor.

These markers are observable hooks for an actual agent-execution test. Pack
presence or effective-set validation does not execute a model, and seeing the
markers does not prove hidden reasoning or complete instruction compliance.

## Required behaviour

- Implement only behaviour supported by the active contract and scenario.
- Inspect existing naming, layout, error, validation, and test patterns first.
- Keep diffs focused and avoid unrelated formatting or refactoring.
- Validate external and persisted inputs at their trust boundaries.
- Add focused tests for the contract, including negative paths when the
  scenario is security-, privacy-, safety-, or availability-sensitive.
- Prefer existing dependencies or native code for bounded functionality.
- Do not perform deployment, migration, or external mutation unless the user
  also activates or requests implementation work.

## Handoff

State what changed, what did not change, how the contract is proved, and what
an implementer must do to release the change safely.
