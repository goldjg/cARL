<!-- version: 1.0.0 -->
# Current PR Contract

## Contract status

Complete

## Goal

Wire repository-local pack selection, profile activation, dependency,
precedence, override, and validation state into the shared agent hydration
contract so every supported harness applies only effective, non-overridden
instruction packs without requiring the cARL CLI, and ship a cloneable default
profile example that explicitly reproduces the repository's legacy effective
pack baseline without activating or owning mutable user profile state.

## Intentional amendment

The original hydration contract is extended with a canonical
`.github/carl/profiles.example.json` asset. It is an ordinary schema-version 1
profile document installed as a reference example, not active policy.
Repositories retain selected-as-active compatibility until a user deliberately
copies or otherwise adopts the example as `.github/carl/profiles.json`.

## Non-goals

- No mandatory CLI invocation during agent hydration.
- No harness-specific pack evaluator or expanded harness shim.
- No new pack, dependency, registry, network, or schema.
- No claim that instruction availability, adapter sync, or agent self-report
  proves model adherence.
- No unrelated CLI, installation, repair, release, or repository-map change.
- No automatic creation or ownership of mutable `.github/carl/profiles.json`.
- No profile cloning command, TUI, menu editor, schema change, implicit
  "default means all packs" behaviour, or other hidden evaluator special case.

## Approved scope

- `.github/copilot-instructions.md` and its embedded mirror
- `.github/instructions/core/carl.instructions.md` and its embedded mirror
- `.github/instructions/core/cognition-governance.instructions.md` and its
  embedded mirror for the matching hydration checkpoint correction
- `.github/carl/profiles.example.json` and its embedded mirror
- Focused embedded/canonical parity and hydration-contract tests
- `internal/pack/` for focused default-profile validation/effective-set tests
  and only where evaluator or trace behaviour conflicts with the required
  fail-closed, non-overridden hydration semantics
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`, and
  `VISION.md`
  only where durable public wording requires reconciliation
- `.github/carl/memory.md`, `.github/carl/trust-boundaries.md`, and their
  embedded mirrors where stable operating or trust-boundary truth changes
- `.github/carl/plans/pack-profile-agent-hydration.md`
- This contract

## Forbidden scope

- No changes to `runtime.json`, `packs.json`, active `profiles.json`, registry
  state, harness shims, dependencies, CI/CD, release configuration, or
  generated repository maps.
- No network access or inspection of prompts, sessions, hidden harness state,
  model reasoning, or chain-of-thought.
- No load-all fallback when selection, profile, dependency, precedence, or
  override state is invalid or ambiguous.
- No inference of policy activation or order from directory presence or
  filesystem enumeration.

## Architectural constraints

- Canonical policy state remains in repository files; the shared loader
  explains how agents consume it.
- The CLI remains authoritative for management, validation, explanation, and
  trace diagnostics but optional for governance hydration.
- `packs.json` is selection authority when present; otherwise selection uses
  the tested `runtime.json` managed-artefact compatibility derivation.
- With `profiles.json`, active seeds are additive organisation/repository
  defaults plus the active profile and optional role/task overlays; without
  it, selected packs are active compatibility seeds.
- Effective evaluation expands required dependencies transitively, validates
  state, orders priority descending with pack-ID tie-breaking, and applies
  explicit overrides only to overridable targets.
- Overridden packs may remain visible in diagnostic evaluation output but
  their instruction definitions are not applied by agents.
- All harness shims remain thin pointers to the shared loader.
- The shipped default is `.github/carl/profiles.example.json`, an embedded
  canonical example installed by `carl init`; it is not read by the evaluator
  until deliberately adopted as `.github/carl/profiles.json`.
- The example uses only schema-version 1 fields and explicit sorted pack IDs.
  It has profile ID `default`, a clear baseline description, empty
  organisation/repository defaults, all current legacy-active packs in the
  profile pack list, and active profile `default`.
- Existing compatibility context has no named role or task. The example keeps
  role/task context unset; no new role identity or overlay semantics may be
  invented.
- Future interactive profile creation/cloning/editing remains non-binding
  roadmap direction and must operate on the same `profiles.json` schema.

## Security constraints

- Hydration fails closed: invalid, conflicting, missing, cyclic, or ambiguous
  policy state requires stopping and reporting rather than guessing.
- Repository-local files are the only hydration inputs; no network or hidden
  harness evidence is required.
- Generated explanation and trace output remain derived evidence, never a new
  canonical authority or proof of adherence.

## Expected files

- `.github/carl/current-pr-contract.md`
- `.github/carl/plans/pack-profile-agent-hydration.md`
- `.github/carl/profiles.example.json`
- `embedded/assets/.github/carl/profiles.example.json`
- `.github/copilot-instructions.md`
- `embedded/assets/.github/copilot-instructions.md`
- `.github/instructions/core/carl.instructions.md`
- `embedded/assets/.github/instructions/core/carl.instructions.md`
- `.github/instructions/core/cognition-governance.instructions.md`
- `embedded/assets/.github/instructions/core/cognition-governance.instructions.md`
- `embedded/embedded_test.go`
- Focused `internal/pack/` implementation/tests if reconciliation is required
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`
- `.github/carl/trust-boundaries.md`
- `embedded/assets/.github/carl/trust-boundaries.md`
- Public documentation only where contradictory wording is found, including
  `VISION.md`

## Contract assertions

1. The default profile example is valid ordinary schema-version 1 profile
   data with deterministic formatting, explicit references to all 24 current
   legacy-active pack IDs, active profile `default`, empty global defaults,
   and no invented role/task context.
2. When deliberately adopted with those 24 packs selected, the example
   resolves to exactly the same complete effective pack IDs and order as the
   no-profile selected-as-active compatibility path.
3. Every example reference exists and profile reference validation requires
   it to be selected; removing a selected reference fails closed. Dependency,
   precedence, conflict, and override evaluation remains the ordinary shared
   `ComputeEffectiveSet` path with no default-profile special case.
4. Merely shipping/installing `profiles.example.json` does not create active
   profile state. Repositories without `.github/carl/profiles.json` retain
   current legacy selection and role-neutral compatibility behaviour.
5. The shared loader identifies the example as the cloneable reference
   baseline without making the CLI mandatory or blurring present, selected,
   active, effective, and overridden states.
6. Canonical loader/profile/core-pack assets and embedded mirrors are
   byte-identical, every harness shim remains a thin shared-loader route, and
   trace/evidence limitations remain explicit.

## Validation plan

- Focused tests for default-example schema validity, exact pack/reference
  completeness, role-neutral active context, selected-reference validation,
  legacy compatibility parity, unchanged dependency/precedence/override
  semantics, loader wording, mirror parity, shim routing, malformed legacy
  selection, and overridden explanation state
- `gofmt -w` on changed Go files
- `/usr/local/go/bin/go test ./...`
- `/usr/local/go/bin/go vet ./...`
- `/usr/local/go/bin/go build ./cmd/carl`
- `git diff --check`
- Repository searches for obsolete load-all, mandatory-CLI, filesystem-order,
  and overridden-as-applied wording
- Direct byte comparisons for canonical/embedded mirrors

## cARL/docs update expectation

Expected. This change alters the stable shared-loader operating contract and
the trust boundary between repository-local policy state, shipped inactive
profile examples, derived CLI diagnostics, and agent instruction hydration.

## Stop conditions

Stop if the work requires a new schema, mandatory CLI runtime dependency,
harness-specific pack logic, network authority, hidden model-state inspection,
weakened validation, load-all fallback, automatic active `profiles.json`
ownership, a hidden default-profile evaluator case, an invented role identity,
or changes outside approved scope.

## Escalation triggers

Escalate if tested pack semantics cannot be reconciled with the required
non-overridden hydration contract without a breaking public schema change, or
if unrelated pre-existing workspace changes overlap a required target.

## Context reset notes

Carry forward the repository-local effective-pack hydration contract, the
inactive canonical default-profile example, the distinction between example
availability and active profile state, the distinction between evaluation
visibility and instruction application, and the limitation that loader
availability cannot prove model adherence.

## Completion checklist

- [x] Implementation stayed inside approved scope.
- [x] Contract assertions were validated.
- [x] Canonical and embedded assets are byte-identical.
- [x] Harness shims remain thin shared-loader pointers.
- [x] Full tests, vet, build, diff checks, and contradiction searches ran.
- [x] Durable documentation and cARL artefacts were reconciled.
- [x] This contract was marked complete.
