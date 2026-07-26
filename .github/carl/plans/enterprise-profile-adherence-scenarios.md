<!-- version: 1.2.1 -->
# Opt-In Enterprise Pack and Profile Examples

## Status

Completed

## Goal

Make opt-in enterprise-example adoption fail-safe throughout the complete
setup sequence while preserving the existing repository default.

## Baseline

Before this revision, `main` has no `packs.json` or `profiles.json`. Its
compatibility evaluation selects and activates the 24 built-in packs derived
from `runtime.json`, ordered lexicographically because each has priority `0`.
The baseline evaluation has zero conflicts.

PR #44 initially added active selection/profile state and therefore changed
that default. This revision removes those active artefacts.

Manual testing identified a second transition hazard. With no `profiles.json`,
all selected packs are active compatibility seeds. Selecting the 11 enterprise
packs before adopting the profile fixture therefore temporarily activates all
11 examples. Safe adoption must copy the fixture first, establishing its
active `default` profile before selection expands.

The correction preserves both invariants:

- “Installing or merging these examples does not change agent behaviour until
  a user explicitly activates or adopts one.”
- “During safe adoption, the effective set remains the existing 24-pack
  repository baseline at every step.”

## Affected files

- `.github/carl/current-pr-contract.md`
- `.github/carl/enterprise-profiles.md`
- `.github/carl/profiles.enterprise.example.json`
- `.github/carl/profiles.enterprise.scenarios.example.json`
- `.github/carl/plans/enterprise-profile-adherence-scenarios.md`
- `.github/instructions/enterprise/*.instructions.md`
- Remove PR-added `.github/carl/packs.json`
- Remove PR-added `.github/carl/profiles.json`
- PR #44 title and description

## Step-by-step changes

1. Capture the exact `main` effective pack IDs and ordering.
2. Remove PR-added active selection and profile state.
3. Keep a bootstrap `profiles.enterprise.example.json` containing only the
   safe `default` profile and a separate inactive
   `profiles.enterprise.scenarios.example.json` catalogue containing the
   default plus all six fictional profiles.
4. Ensure both ordinary schema-version 1 fixtures activate `default`, which
   exactly lists the 24-pack baseline.
5. Retain the eleven repository-local enterprise pack definitions without
   selecting them.
6. Rewrite the guide around presence, selection, activation, composition,
   explicit adoption, restoration, and evidence limitations.
7. Correct adoption guidance everywhere so the bootstrap fixture is copied
   before enterprise packs are selected, and the full catalogue is copied
   only after selection is valid.
8. Validate the complete transition from clean state through copy, selection,
   representative activation, and default restoration.
9. Confirm the real repository still matches `main`, run full validation,
   commit, push, and update PR #44.

## Test strategy

- Compare `main` and post-change `carl pack effective --json` IDs/order.
- Assert no repository `packs.json` or `profiles.json` remains.
- Validate bootstrap shape: one active `default` profile with the ordered
  24-pack baseline.
- Validate full-catalogue shape: 7 profiles total, including `default`; six
  fictional profiles; five roles and three tasks per fictional profile.
- In an isolated copy:
  - begin with no `packs.json` or `profiles.json`;
  - record 24 selected/effective built-ins and zero enterprise packs;
  - copy the bootstrap fixture to `profiles.json` first and prove the ordered
    effective set remains the same 24-pack baseline;
  - select all 11 enterprise packs and prove 35 packs are selected while the
    enterprise packs remain inactive and ineffective;
  - copy the full catalogue over `profiles.json` and prove `default` and the
    ordered 24-pack effective baseline remain unchanged;
  - activate `brindleforge-process-pilot / implementer / sandbox-build` and
    prove only the expected enterprise scenario and discipline become
    effective;
  - restore `default` and prove the ordered 24-pack baseline returns while all
    11 enterprise packs remain selected but inactive;
  - validate `pack profile list` and `show` for all seven profiles, `pack
    list`, `pack effective`, and `trace`;
  - remove the temporary user-owned state.
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

- Users may select enterprise packs before copying the fixture. Because
  profile-absent compatibility treats every selected pack as active, that
  order temporarily activates all 11 examples. Documentation must require the
  fixture-first order.
- A single full catalogue cannot be copied before selection because strict
  profile validation rejects its unselected enterprise references. The
  bootstrap/full-catalogue split preserves fail-closed validation without a
  CLI, evaluator, or schema special case.
- Users may mistake lightweight for weak governance. Documentation and pack
  wording must define it as a smaller safe operating boundary.
- Example files may be mistaken for active state. File names, the invariant,
  and default-parity proof must make inactivity explicit.

## cARL/docs update expectation

Update the contract, this plan, the operator guide, the inactive fixture, and
PR metadata. No memory, invariant, trust-boundary, runtime, or embedded update
is required because the default operating model remains unchanged.

## Completion evidence

- Initial clean state: 24 selected, 24 active, and 24 effective built-ins; no
  enterprise pack selected, active, or effective; zero conflicts.
- Bootstrap copied first: `default` active; the same ordered 24 packs selected
  and effective; no enterprise pack selected, active, or effective; zero
  conflicts.
- Enterprise selection: 35 selected, 24 active/effective; all 11 enterprise
  packs selected but none active or effective; zero conflicts.
- Full catalogue copied: all 7 profiles validate; `default` remains active;
  the same ordered 24-pack effective set remains; all 11 enterprise packs
  remain selected but inactive and ineffective; zero conflicts.
- Representative activation
  `brindleforge-process-pilot / implementer / sandbox-build`: 35 selected, 5
  active seeds, and 9 effective packs. Only
  `enterprise/brindleforge-process-pilot` and
  `enterprise/discipline-implementation` are enterprise-active/effective;
  zero conflicts.
- Default restoration: 35 selected and the original ordered 24 active/effective
  packs; all 11 enterprise packs selected but inactive and ineffective; zero
  conflicts.
- `pack profile list --json` and `pack profile show <id> --json` passed for
  all seven profiles.
- `go run ./cmd/carl pack list --json`,
  `go run ./cmd/carl pack effective --json`, and
  `go run ./cmd/carl trace --json` passed.
- `go test ./...`, `go vet ./...`, `go build ./cmd/carl`, and
  `git diff --check` passed.
- The isolated user-owned `packs.json` and `profiles.json` were confined to
  temporary validation state. The real branch contains neither file.
- The built-in instruction-pack diff is empty, and PR #44 metadata documents
  the safe order, transition evidence, and adherence limitations.
