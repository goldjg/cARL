<!-- version: 1.0.0 -->
# Current PR Contract

## Contract status

Complete

## Goal

Create fictional enterprise policy profiles and custom instruction packs that
exercise agent adherence across security and non-security design,
architecture, code, implementation, and review scenarios.

## Non-goals

- No change to the cARL profile, selection, pack, or composition schemas.
- No CLI behaviour change.
- No registry, network, dependency, CI/CD, release, or generated-map change.
- No claim that instruction availability proves model adherence.

## Approved scope

- `.github/carl/packs.json`
- `.github/carl/profiles.json`
- `.github/carl/enterprise-profiles.md`
- `.github/instructions/enterprise/*.instructions.md`
- `.github/carl/plans/enterprise-profile-adherence-scenarios.md`
- This contract

## Forbidden scope

- Do not alter any existing built-in pack under `.github/instructions/cloud/`,
  `.github/instructions/core/`, `.github/instructions/languages/`, or
  `.github/instructions/platform/`.
- Do not modify embedded built-in assets, `runtime.json`, registries,
  dependencies, source code, tests, harness adapters, CI/CD, release files, or
  generated repository maps.
- Do not use custom overrides to weaken a built-in pack.

## Architectural constraints

- `packs.json` remains explicit user-owned selection authority.
- `profiles.json` uses only schema-version 1 defaults, profiles, roles, tasks,
  and active-context fields.
- Organisation/repository defaults, profile packs, roles, tasks, and required
  dependencies compose additively.
- Custom pack IDs and files use the canonical
  `.github/instructions/enterprise/<name>.instructions.md` convention.
- Lightweight scenarios are bounded to synthetic, isolated, non-production
  work and escalate to a stricter profile if that boundary is crossed.

## Security constraints

- No secrets, credentials, personal data, payment data, clinical data, or
  production endpoints may appear in profiles, packs, examples, or tests.
- Strict sector packs strengthen rather than override built-in security rules.
- Lightweight packs must not waive repository invariants or cARL lifecycle
  controls.

## Expected files

- `.github/carl/current-pr-contract.md`
- `.github/carl/plans/enterprise-profile-adherence-scenarios.md`
- `.github/carl/packs.json`
- `.github/carl/profiles.json`
- `.github/carl/enterprise-profiles.md`
- Eleven new `.github/instructions/enterprise/*.instructions.md` files

## Contract assertions

1. Six fictional-company profiles cover financial services, health, and
   manufacturing, with strict and non-security lightweight/balanced variants.
2. Every profile offers designer, architect, coder, implementer, and reviewer
   overlays plus task-specific activation, and every reference is selected.
3. Eleven custom packs are discoverable and compose without conflicts; strict
   packs add controls, while lightweight packs are explicitly restricted to
   isolated non-production work.
4. The committed active context resolves deterministically and exposes a
   visible adherence marker, while representative profile switches validate.
5. Existing built-in pack files remain byte-for-byte untouched by this task.

## Validation plan

- Validate pack discovery, profile listing/showing, selection references, and
  effective-set composition through the local cARL CLI.
- Exercise representative strict, balanced, and lightweight active contexts,
  then restore the committed active context.
- Run `go test ./...`, `go vet ./...`, `go build ./cmd/carl`, and
  `git diff --check`.
- Confirm `git diff` reports no modification to built-in instruction packs.

## cARL/docs update expectation

Expected. Add activation and adherence guidance in a dedicated document.
Do not duplicate mutable active profile state into memory.

## Stop conditions

Stop if the requested scenarios require a schema change, a built-in pack
modification, a weakening override, secret or regulated data, a network
authority, or changes outside approved scope.

## Escalation triggers

Escalate if valid profile composition cannot express the scenarios without
changing evaluator semantics or weakening repository invariants.

## Context reset notes

These profiles are test policy contexts, not claims about real companies or
legal compliance. Future work should activate the scenario, role, and task
needed for the adherence test and restore the documented default afterward.

## Completion checklist

- [x] Implementation stayed inside approved scope.
- [x] All contract assertions were validated.
- [x] Built-in packs remained unchanged.
- [x] Full tests, vet, build, and diff checks ran.
- [x] Durable cARL/docs were reconciled.
- [x] This contract was marked complete.
