<!-- version: 1.1.1 -->
# Opt-In Enterprise Pack and Profile Examples

## Status

Completed

## Goal

Add opt-in enterprise pack and profile examples while preserving the existing
repository default.

## Baseline

Before this revision, `main` has no `packs.json` or `profiles.json`. Its
compatibility evaluation selects and activates the 24 built-in packs derived
from `runtime.json`, ordered lexicographically because each has priority `0`.
The baseline evaluation has zero conflicts.

PR #44 initially added active selection/profile state and therefore changed
that default. This revision removes those active artefacts.

## Affected files

- `.github/carl/current-pr-contract.md`
- `.github/carl/enterprise-profiles.md`
- `.github/carl/profiles.enterprise.example.json`
- `.github/carl/plans/enterprise-profile-adherence-scenarios.md`
- `.github/instructions/enterprise/*.instructions.md`
- Remove PR-added `.github/carl/packs.json`
- Remove PR-added `.github/carl/profiles.json`
- PR #44 title and description

## Step-by-step changes

1. Capture the exact `main` effective pack IDs and ordering.
2. Remove PR-added active selection and profile state.
3. Move the six fictional profiles into a dedicated inactive
   `profiles.enterprise.example.json` fixture.
4. Include a `default` fixture profile that exactly lists the 24-pack baseline
   and remains active if the fixture is deliberately copied.
5. Retain the eleven repository-local enterprise pack definitions without
   selecting them.
6. Rewrite the guide around presence, selection, activation, composition,
   explicit adoption, restoration, and evidence limitations.
7. Validate all example references and representative contexts in an isolated
   adopted copy, then confirm the real repository still matches `main`.
8. Run full repository validation, commit, push, and update PR #44.

## Test strategy

- Compare `main` and post-change `carl pack effective --json` IDs/order.
- Assert no repository `packs.json` or `profiles.json` remains.
- Validate fixture shape: 7 profiles total, including `default`; six fictional
  profiles; five roles and three tasks per fictional profile.
- In an isolated copy:
  - create explicit selection containing all 24 built-ins and 11 enterprise
    packs;
  - copy the fixture to `profiles.json`;
  - validate the copied default against the baseline;
  - activate strict, balanced, and bounded-lightweight representative
    contexts and assert expected packs with zero conflicts.
- Run `go test ./...`.
- Run `go vet ./...`.
- Run `go build ./cmd/carl`.
- Run `git diff --check`.
- Confirm built-in pack diffs are empty.
- Stage only intended PR files.

## Evidence semantics

- Profile/effective-set validation proves deterministic policy composition.
- Adherence markers make selected behavior observable in model output.
- Actual agent-execution testing requires running a model in each activated
  context and assessing its output.
- Neither fixture presence nor marker text proves hidden reasoning or complete
  instruction compliance.

## Risks

- Users may copy the fixture without selecting enterprise packs. Documentation
  must require selection first because every profile reference must be
  selected.
- Users may mistake lightweight for weak governance. Documentation and pack
  wording must define it as a smaller safe operating boundary.
- Example files may be mistaken for active state. File names, the invariant,
  and default-parity proof must make inactivity explicit.

## cARL/docs update expectation

Update the contract, this plan, the operator guide, the inactive fixture, and
PR metadata. No memory, invariant, trust-boundary, runtime, or embedded update
is required because the default operating model remains unchanged.

## Completion evidence

- The branch default and `main` both resolve the same 24 pack IDs in the same
  order with zero conflicts; their full effective pack objects are equal.
- The fixture validates 7 profiles, including all 6 fictional scenarios, all
  30 role contexts, and all 18 task contexts.
- Representative strict, balanced, and bounded-lightweight contexts resolve
  to 11, 7, and 9 packs respectively, each with zero conflicts.
- The isolated adopted copy was restored to `default`, yielding the 24-pack
  baseline with zero conflicts.
- `go test ./...`, `go vet ./...`, `go build ./cmd/carl`, and
  `git diff --check` passed.
- The built-in instruction-pack diff is empty, and PR #44 metadata reflects
  the opt-in purpose and evidence limitations.
