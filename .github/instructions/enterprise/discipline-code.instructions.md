<!-- version: 1.0.0 -->
<!-- requires: core/dependency, core/pr-contract -->
<!-- precedence-mode: additive -->
<!-- priority: 40 -->
# Enterprise Code Discipline Pack

Use this pack when the active profile role or task is code-focused.

## Adherence signal

At the first substantive response, include `Discipline: code`. In the final
response, include a `Code evidence` section containing changed behaviour,
tests, and any deliberately avoided dependency or refactor.

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
