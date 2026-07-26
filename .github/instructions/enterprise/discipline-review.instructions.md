<!-- version: 1.0.0 -->
<!-- requires: core/cognition-governance, core/dependency, core/pr-contract -->
<!-- precedence-mode: additive -->
<!-- priority: 40 -->
# Enterprise Review Discipline Pack

Use this pack when the active profile role or task is review-focused.

## Adherence signal

At the first substantive response, include `Discipline: review`. Lead the
final response with findings ordered by severity and include a `Review
evidence` section naming the contract, tests, and boundaries checked.

## Required behaviour

- Review against the approved outcome and contract, not personal style.
- Inspect changed code, configuration, tests, and durable documentation
  together.
- Prioritise correctness, security, privacy, safety, data integrity,
  operability, and regression risks applicable to the active company scenario.
- Give each actionable finding a severity, exact location, consequence, and
  concrete remediation.
- Distinguish confirmed defects from questions, assumptions, and optional
  improvements.
- Verify that tests prove the intended contract and include relevant negative
  paths.
- Remain read-only unless the user explicitly asks to implement selected
  findings.

## No-findings response

If no actionable findings remain, state that directly and list residual test
or evidence gaps instead of inventing issues.
