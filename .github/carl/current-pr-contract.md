<!-- version: 1.2.0 -->
# Current PR Contract

This contract constrains implementation scope for the active PR. Update it when scope is explicitly amended. If a requested action falls outside approved scope, stop and escalate before proceeding.

Use this contract to distinguish active PR constraints, completed PR constraints, durable invariants, and intentional amendments. Completed PR constraints are historical evidence unless they are explicitly promoted to durable invariants.

---

## Goal

Refactor cARL harness loading semantics so harness-specific instruction files act as thin adapters/loaders for canonical cARL governance artefacts, rather than acting as independent repository constitutions.

The primary outcome is to make Copilot and future harness adapters explicitly perform the cARL lifecycle:

1. hydrate cARL before planning or implementation;
2. apply cARL governance during execution;
3. reconcile documentation and durable cARL artefacts before final response;
4. report whether cARL/docs updates were required.

## Contract status

active

## Non-goals

- No changes to core CLI command behaviour.
- No changes to runtime installation, repair, status, doctor, map, plan, convert, reconcile, or harness command behaviour unless required only to keep embedded assets in sync.
- No new harness implementations.
- No model benchmarking implementation.
- No network access.
- No external dependencies.
- No broad rewrite of instruction packs.
- No change to `memory.md` schema.
- No change to `runtime.json` semantics.

## Carry-forward rules

Promoted invariants from previous PRs remain in force:

- No secrets committed to any file.
- Security baseline applies.
- `current-pr-contract.md` must be read before implementation begins.
- Every implementation PR must update durable artefacts when behaviour, assumptions, commands, scope, roadmap, or operating model changes.
- cARL durable artefacts must not become a per-turn session diary.
- Completed PR contracts are historical evidence, not binding scope, unless explicitly promoted to durable invariants.

## Approved scope

- `.github/copilot-instructions.md` — refactor from full operating model into thin Copilot loader / adapter.
- `.github/carl/memory.md` — update durable architecture truth to reflect that cARL artefacts are canonical and harness files are adapters, not authorities.
- `.github/carl/current-pr-contract.md` — this active contract.
- `.github/carl/invariants.yml` — update only if a durable invariant needs to be added or clarified.
- `.github/carl/trust-boundaries.md` — update only if the trust boundary model needs clarification for harness adapters versus canonical cARL artefacts.
- `.github/instructions/core/*.instructions.md` — update only if needed to keep loader semantics consistent with canonical cARL governance.
- `embedded/assets/.github/**` — update only where required to keep embedded managed artefacts byte-identical with canonical runtime assets.
- `ROADMAP.md`, `ARCHITECTURE.md`, `README.md`, or `CLI.md` — update only if the loader/adaptor architecture or documented command behaviour changes.

## Intentional amendments

This PR intentionally amends the previous architectural wording that treated `.github/copilot-instructions.md` as the repository constitution.

New durable architecture:

- cARL artefacts are the canonical source of governance truth.
- Harness-specific files are adapters/loaders that inject or point agents toward canonical cARL artefacts.
- `.github/copilot-instructions.md` is the Copilot adapter, not the authority.
- If harness/session memory conflicts with `.github/carl/*`, the canonical cARL artefacts win unless current repository state proves them stale.
- If `.github/carl/memory.md` conflicts with current repository state, repository state wins and memory should be updated.

## Forbidden scope

- Introducing network-backed runtime updates.
- Adding new third-party dependencies.
- Rewriting all instruction packs as part of this PR.
- Removing existing safety, security, dependency, test, or PR contract controls.
- Turning `memory.md` into a per-turn execution log.
- Modifying runtime state semantics in `.github/carl/runtime.json`.
- Changing generated command behaviour unless explicitly amended.

## Architectural constraints

- Harness adapter files must remain disposable and regenerable.
- Canonical governance must live in `.github/carl/*` and reusable instruction packs, not be duplicated independently per harness.
- Loader files should be short, procedural, and explicit about the required lifecycle.
- Loader files should optimise for model salience, especially smaller models that may not infer governance lifecycle obligations.
- The cARL lifecycle must include a final reconciliation decision: update docs/cARL if durable truth changed, otherwise explicitly state why no update was required.
- Repository cARL artefacts outrank stale prompt/session memory when they conflict.
- Avoid contradictory authority language such as calling a harness adapter the “constitution” if cARL artefacts are canonical.

## Security constraints

- No secrets, tokens, private keys, tenant data, or credentials in any new or modified file.
- Do not weaken authentication, authorization, validation, logging safety, dependency hygiene, or secret handling guidance.
- Treat CI/CD and harness instruction files as governance-sensitive because they influence delegated agent behaviour.
- Do not introduce instructions that allow broad autonomous writes without PR contract coverage.

## Files expected to change

Expected:

- `.github/carl/current-pr-contract.md`
- `.github/copilot-instructions.md`
- `.github/carl/memory.md`

Possible, if required by the implementation:

- `.github/carl/invariants.yml`
- `.github/carl/trust-boundaries.md`
- `.github/instructions/core/carl.instructions.md`
- `.github/instructions/core/memory-cache.instructions.md`
- `.github/instructions/core/pr-contract.instructions.md`
- `.github/instructions/core/cognition-governance.instructions.md`
- `embedded/assets/.github/copilot-instructions.md`
- `embedded/assets/.github/instructions/core/*.instructions.md`
- `ROADMAP.md`
- `ARCHITECTURE.md`
- `README.md`
- `CLI.md`

## Tests / validation

- Review all changed governance files for contradictory authority language.
- Verify `.github/copilot-instructions.md` is a thin loader, not a duplicated full operating model.
- Verify the loader explicitly requires:
  - cARL hydration before planning/implementation;
  - PR contract check;
  - relevant instruction pack check;
  - docs/cARL reconciliation before final response;
  - explicit no-update reasoning when docs/cARL are unchanged.
- If embedded assets are changed, run `go test ./...`.
- If CLI code is changed, run:
  - `go build ./cmd/carl`
  - `go test ./...`
- If no CLI code is changed, state that Go build/test were not required and why.

## Stop conditions

Stop and ask for confirmation if the change requires:

- altering runtime.json semantics;
- changing CLI command behaviour;
- introducing a new harness;
- removing or weakening security/test/dependency governance;
- making broad instruction pack rewrites;
- resolving conflicting architecture language in a way not covered by this contract.

## Escalation triggers

Escalate if:

- GitHub Copilot instruction precedence requires a different file layout than expected;
- harness adapter semantics conflict with existing cARL runtime management behaviour;
- embedded asset sync requirements are unclear;
- a durable invariant needs amendment beyond loader/adaptor authority semantics;
- current repository state conflicts with memory in a way that cannot be safely resolved from the files alone.

## Context reset notes

When this PR is complete, close or supersede this contract.

The durable lesson to carry forward is:

Harness instruction files are adapter surfaces. They may load, summarise, or route to cARL, but they are not the canonical governance authority. cARL artefacts remain the source of durable project truth.