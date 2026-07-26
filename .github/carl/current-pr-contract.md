<!-- version: 1.1.1 -->
# Current PR Contract

## Contract status

Completed

## Goal

Add opt-in enterprise pack and profile examples while preserving the existing
repository default.

## Intentional amendment

PR #44 originally committed active `packs.json` and `profiles.json` state,
which changed the repository's effective pack set from the pre-PR baseline.
This revision makes the fictional enterprise material present but inactive,
removes active selection/profile state from the PR, and proves exact default
effective-set parity with `main`.

## Non-goals

- No profile, selection, pack, composition, or trace schema change.
- No CLI or evaluator behaviour change.
- No hidden special case for enterprise examples.
- No registry, network-runtime, dependency, source-code, test-code, CI/CD,
  release, embedded-runtime, or generated-map change.
- No claim that fixture composition proves model execution, hidden reasoning,
  or complete instruction compliance.

## Approved scope

- `.github/carl/current-pr-contract.md`
- `.github/carl/enterprise-profiles.md`
- `.github/carl/profiles.enterprise.example.json`
- `.github/carl/plans/enterprise-profile-adherence-scenarios.md`
- `.github/instructions/enterprise/*.instructions.md`
- Removal of PR-added `.github/carl/packs.json`
- Removal of PR-added `.github/carl/profiles.json`
- PR #44 title and description

## Forbidden scope

- Do not modify `.github/carl/runtime.json` or the shipped
  `.github/carl/profiles.example.json`.
- Do not modify existing built-in packs under `.github/instructions/cloud/`,
  `.github/instructions/core/`, `.github/instructions/languages/`, or
  `.github/instructions/platform/`.
- Do not modify embedded assets, source code, tests, dependencies, harness
  adapters, CI/CD, release files, registries, or generated repository maps.
- Do not add active profile/selection state to the merged repository.
- Do not use overrides to weaken built-in controls.
- Do not include the unrelated local session record in PR #44.

## Architectural constraints

- `main` is the default-behaviour authority for parity validation.
- With no PR-added `packs.json`, selection continues to use the existing
  `runtime.json` compatibility derivation.
- With no PR-added `profiles.json`, selected packs continue to be active under
  the existing compatibility behaviour.
- Enterprise pack definitions may be present/discoverable but are not selected
  merely because their files exist.
- The enterprise profile fixture uses ordinary schema-version 1 data in
  `.github/carl/profiles.enterprise.example.json`.
- The fixture includes an active `default` profile that explicitly reproduces
  the 24-pack `main` baseline, so deliberate copying is safe before an
  enterprise profile is activated.
- Users must explicitly select the enterprise packs and copy/adopt the fixture
  before activating a fictional profile.
- Dependencies, priority, and overrides retain ordinary evaluator semantics.

## Security constraints

- No secrets, credentials, personal data, payment data, clinical data, live OT
  data, or production endpoints may appear in examples or validation.
- Strict examples strengthen built-in controls additively.
- Bounded-lightweight examples narrow the permitted operating boundary; they
  do not relax repository governance or override built-in controls.
- The documentation must state that fixtures are not certifications.

## Expected files

- `.github/carl/current-pr-contract.md`
- `.github/carl/enterprise-profiles.md`
- `.github/carl/profiles.enterprise.example.json`
- `.github/carl/plans/enterprise-profile-adherence-scenarios.md`
- Eleven `.github/instructions/enterprise/*.instructions.md` files
- No `.github/carl/packs.json`
- No `.github/carl/profiles.json`

## Contract assertions

1. The effective pack IDs and order after this revision are exactly equal to
   `main`: the same 24 packs, in the same order, with zero conflicts.
2. Merely merging the enterprise pack files and inactive fixture selects or
   activates none of them; no fictional profile, role, or task is active.
3. The ordinary schema-version 1 fixture contains the safe `default` profile,
   all six fictional scenarios, five roles per scenario, three tasks per
   scenario, and only discoverable pack references.
4. Representative strict, balanced, and bounded-lightweight contexts compose
   without conflicts after explicit temporary adoption and selection, and the
   repository default is restored afterward.
5. Documentation clearly distinguishes presence, selection, activation,
   effective composition, observable markers, actual model execution, and
   proof limitations.
6. Existing built-in pack definitions remain unchanged.

## Validation plan

- Evaluate and persist the `main` effective-set IDs/order as the parity
  baseline.
- Compare the post-change default IDs/order byte-for-byte or element-for-
  element against that baseline.
- Validate the fixture schema, every profile, every role/task reference, and
  all custom pack metadata.
- Exercise representative strict financial review, balanced health design,
  and bounded-lightweight manufacturing implementation contexts in an
  isolated copy with explicit selection/adoption.
- Restore and recheck the repository default.
- Run `go test ./...`, `go vet ./...`, `go build ./cmd/carl`, and
  `git diff --check`.
- Confirm no built-in instruction pack changed and only intended PR files are
  staged.

## cARL/docs update expectation

Expected. Update the contract, plan, operator documentation, inactive fixture,
and PR description. Do not update durable memory because evaluator behaviour
and the repository's durable default remain unchanged.

## Stop conditions

Stop if exact `main` parity requires a schema/evaluator change, modifying a
built-in or embedded asset, weakening governance, retaining active fictional
state, or including unrelated local files.

## Escalation triggers

Escalate if ordinary selection/profile semantics cannot validate all examples
while preserving exact default parity, or if remote branch/PR state diverges
from the local branch.

## Context reset notes

Enterprise definitions and fixtures are present examples, not selected or
active policy. Future tests must explicitly select, adopt, and activate a
scenario, then restore the repository default.

## Completion checklist

- [x] Implementation stayed inside approved scope.
- [x] Exact `main` default parity was proved.
- [x] All example references and representative contexts were validated.
- [x] Built-in packs remained unchanged.
- [x] Full tests, vet, build, and diff checks passed.
- [x] Documentation and PR metadata were reconciled.
- [x] This contract was marked complete.
