<!-- version: 1.2.1 -->
# Current PR Contract

## Contract status

Completed

## Goal

Make opt-in enterprise-example adoption fail-safe throughout the complete
setup sequence while preserving the existing repository default.

## Intentional amendment

PR #44 originally committed active `packs.json` and `profiles.json` state,
which changed the repository's effective pack set from the pre-PR baseline.
This revision makes the fictional enterprise material present but inactive,
removes active selection/profile state from the PR, and proves exact default
effective-set parity with `main`.

Manual adoption testing then identified a sequencing hazard: selecting the 11
enterprise packs before copying the fixture temporarily makes all selected
packs active under profile-absent compatibility behaviour. This correction
requires copying the fixture first so its active `default` profile constrains
the effective set before enterprise selection is expanded.

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
- `.github/carl/profiles.enterprise.scenarios.example.json`
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

- “Installing or merging these examples does not change agent behaviour until
  a user explicitly activates or adopts one.”
- “During safe adoption, the effective set remains the existing 24-pack
  repository baseline at every step.”
- `main` is the default-behaviour authority for parity validation.
- With no PR-added `packs.json`, selection continues to use the existing
  `runtime.json` compatibility derivation.
- With no PR-added `profiles.json`, selected packs continue to be active under
  the existing compatibility behaviour.
- Enterprise pack definitions may be present/discoverable but are not selected
  merely because their files exist.
- The bootstrap fixture uses ordinary schema-version 1 data in
  `.github/carl/profiles.enterprise.example.json` and contains only the active
  `default` profile reproducing the 24-pack `main` baseline.
- The full catalogue uses ordinary schema-version 1 data in
  `.github/carl/profiles.enterprise.scenarios.example.json` and contains the
  same active `default` plus all six fictional profiles.
- Safe adoption must copy the fixture to `.github/carl/profiles.json` before
  selecting the 11 enterprise packs.
- Without `profiles.json`, selected packs are active compatibility seeds, so
  selecting the enterprise packs first would temporarily activate every
  enterprise example.
- Copying the fixture first establishes the active `default` profile; later
  selection makes all 11 enterprise packs selected but inactive.
- After selection, copying the full catalogue over `profiles.json` preserves
  `default` and the 24-pack effective baseline while making all seven profiles
  available for explicit activation.
- During safe adoption, the effective set remains the existing 24-pack
  repository baseline at every step.
- No fictional profile, role, or task becomes effective until a user
  explicitly activates one.
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
- `.github/carl/profiles.enterprise.scenarios.example.json`
- `.github/carl/plans/enterprise-profile-adherence-scenarios.md`
- Eleven `.github/instructions/enterprise/*.instructions.md` files
- No `.github/carl/packs.json`
- No `.github/carl/profiles.json`

## Contract assertions

1. The effective pack IDs and order after this revision are exactly equal to
   `main`: the same 24 packs, in the same order, with zero conflicts.
2. Merely merging the enterprise pack files and inactive fixture selects or
   activates none of them; no fictional profile, role, or task is active.
3. The ordinary schema-version 1 bootstrap contains only the safe `default`;
   the full catalogue contains that default plus all six fictional scenarios,
   five roles per scenario, three tasks per scenario, and only discoverable
   pack references.
4. From a clean state, copying the bootstrap fixture first preserves the 24-pack
   baseline; selecting all 11 enterprise packs next yields 35 selected packs
   while all enterprise packs remain inactive and ineffective.
5. Copying the full catalogue after selection preserves the active `default`
   and ordered 24-pack baseline.
6. Activating `brindleforge-process-pilot / implementer / sandbox-build`
   composes only that scenario, discipline, task packs, and their required
   dependencies; unrelated enterprise packs remain inactive.
7. Restoring `default` returns exactly to the ordered 24-pack baseline with
   zero conflicts while all 11 enterprise packs remain selected but inactive.
8. Documentation clearly distinguishes presence, selection, activation,
   effective composition, observable markers, actual model execution, and
   proof limitations.
9. Existing built-in pack definitions remain unchanged.

## Validation plan

- Evaluate and persist the `main` effective-set IDs/order as the parity
  baseline.
- Compare the post-change default IDs/order byte-for-byte or element-for-
  element against that baseline.
- Validate the fixture schema, every profile, every role/task reference, and
  all custom pack metadata.
- Exercise and record the isolated transition from clean state, to
  bootstrap-first adoption, to 35-pack selection, to full-catalogue adoption, to
  `brindleforge-process-pilot / implementer / sandbox-build`, and back to
  `default`.
- At each state, record selected, active, and effective totals and enterprise
  subsets; compare every baseline state against the original ordered 24 IDs.
- Validate `pack profile list` and `show` for all seven profiles, plus `pack
  list`, `pack effective`, and `trace` diagnostics.
- Restore and recheck the repository default, then remove temporary
  `packs.json` and `profiles.json`.
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
active policy. Future adoption must copy the bootstrap first, select the
enterprise packs second, copy the full catalogue third, activate a scenario
explicitly, and then restore the repository default.

## Completion checklist

- [x] Implementation stayed inside approved scope.
- [x] Every safe-adoption transition preserved the asserted state.
- [x] All example references and the representative context were validated.
- [x] Temporary user-owned policy state was removed.
- [x] Built-in packs remained unchanged.
- [x] Full tests, vet, build, and diff checks passed.
- [x] Documentation and PR metadata were reconciled.
- [x] This contract was marked complete.
