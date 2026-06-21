<!-- version: 1.1.0 -->
# Current PR Contract

This file defines the active scope, constraints, validation expectations, and escalation triggers for the current implementation task.

It is a working contract, not a permanent architectural document.

When the PR is complete, close, reset, or supersede this contract so future agents do not over-anchor on stale scope.

## Contract status

Status: Draft

Allowed values:

- Draft
- Active
- Superseded
- Complete

## Goal

<!-- Describe the single primary outcome this PR/task must achieve. -->

## Non-goals

<!-- List explicitly out-of-scope outcomes to prevent scope creep. -->

## Approved scope

<!-- List files, directories, behaviours, commands, docs, and governance artefacts that may be changed. -->

## Forbidden scope

<!-- List files, behaviours, commands, docs, or governance surfaces that must not be changed. -->

## Architectural constraints

<!-- List architecture rules, compatibility requirements, runtime assumptions, and source-of-truth decisions. -->

## Security constraints

<!-- List security properties that must be preserved, improved, or explicitly validated. -->

## Trust boundaries

<!-- List trust boundaries touched by this work and the validation expected for each. -->

## Expected files

<!-- List expected changed files. Update before implementation if scope changes. -->

## Contract assertions

For non-trivial work, define 3-5 assertions that must remain true after implementation.

Examples:

- Output schema remains backwards compatible.
- Managed artefact repair remains idempotent.
- Protected files are not overwritten.
- Harness adapters route agents to canonical cARL artefacts.
- Tests assert approved behaviour rather than implementation drift.

Assertions:

1. <!-- Assertion 1 -->
2. <!-- Assertion 2 -->
3. <!-- Assertion 3 -->

## Validation plan

<!-- List commands, tests, manual checks, and validation gaps. -->

Required validation:

- <!-- e.g. go test ./... -->
- <!-- e.g. go build ./cmd/carl -->
- <!-- e.g. carl doctor -->
- <!-- e.g. carl harness status -->

## cARL/docs update expectation

State whether this work is expected to update durable governance or documentation.

Allowed values:

- Expected
- Not expected
- Unknown until implementation

Decision:

<!-- Populate before final response. -->

Rationale:

<!-- Explain why docs/cARL updates are or are not expected. -->

## Stop conditions

Stop and ask for confirmation if:

- implementation requires files outside approved scope;
- non-goals appear necessary to satisfy the goal;
- security, identity, trust-boundary, or governance semantics change unexpectedly;
- validation cannot prove a contract-critical behaviour;
- tests begin encoding implementation drift;
- embedded managed asset requirements are unclear;
- model/session behaviour repeatedly misses the contract;
- assumptions conflict with current repository state.

## Escalation triggers

Escalate before proceeding if:

- the active authority order is unclear;
- current repository state conflicts with durable memory;
- prompt/session memory conflicts with cARL artefacts;
- an invariant must be amended;
- a trust boundary must be added or weakened;
- CI/CD, release, dependency, or security policy changes become necessary;
- harness adapter authority semantics change;
- runtime state semantics change;
- destructive or broad writes become necessary.

## Context reset notes

When the PR/task completes, record:

- what should carry forward into durable memory;
- what should be promoted to invariants or trust boundaries;
- what should remain historical only;
- what stale assumptions future agents should ignore.

## Completion checklist

Before final response:

- [ ] Implementation stayed inside approved scope.
- [ ] Forbidden scope was not touched.
- [ ] Contract assertions were validated.
- [ ] Tests or manual checks were run or gaps were reported.
- [ ] cARL/docs update decision was made.
- [ ] Durable changes were reflected in memory/docs/invariants/trust boundaries where required.
- [ ] Active contract was marked Complete, Superseded, or reset as appropriate.