# Pack Phase 5: Policy Provenance and Explanation

## Plan metadata

- PR / branch: `codex/pack-phase5-policy-provenance`
- Status: delivered
- Author: Codex
- Created: 2026-07-25
- Last updated: 2026-07-25

## Task summary

Deliver a local, read-only explanation surface for the existing pack policy
runtime. The commands must expose where effective policy packs came from and
why composition included, ordered, or overrode them without claiming access
to natural-language rule semantics or hidden model reasoning.

## Current repository context

Pack Phases 1 through 4 are delivered. The existing runtime already provides
strict pack discovery, explicit selection, profile-driven activation,
transitive dependencies, deterministic precedence, conservative overrides,
conflict detection, optional verified registries, and installed provenance.
Phase 5 explains those facts rather than creating a second evaluator.

## Previous contract status

The Phase 4 contract is delivered historical evidence. Its durable
offline-first, integrity, ownership, and trust-boundary constraints continue
through memory, architecture, and trust-boundary documentation.

## Contract lifecycle note

Completed PR contracts are historical evidence, not binding scope. The active
authority for this implementation is the Phase 5 contract in
`.github/carl/current-pr-contract.md`.

## Intentional contract amendments

This plan supersedes the Phase 4 execution scope to add the roadmap's Phase 5
top-level commands. It does not amend Phase 4 registry or installation
semantics.

## Goal

Add:

- `carl explain <pack-id> [--json]` for one discoverable pack;
- `carl trace [--json]` for the complete effective policy evaluation.

Both commands reuse the existing pack evaluator and produce deterministic
schema-versioned output.

## Non-goals

- Individual natural-language rule parsing or a compiled policy IR.
- Model prompt, response, session, reasoning, or chain-of-thought inspection.
- Repository state mutation, network access, or registry fetching.
- Automatic conflict resolution.
- Changes to selection, profile, precedence, or override semantics.

## Approved scope

The approved files and surfaces are listed in the active PR contract. New Go
implementation remains in `internal/pack`; only command registration changes
in `cmd/carl/main.go`.

## Forbidden scope

No harness, instruction-pack, runtime-manifest, registry-state, CI, release,
dependency, or destructive repository changes.

## Trust boundaries

- Repository pack/profile/selection/provenance files are untrusted validated
  inputs.
- Embedded packs are bundled inputs, not higher-order runtime policy.
- Explanation output is derived diagnostic evidence and never becomes a
  canonical governance authority.
- Registry digest provenance remains integrity evidence, not identity.

## Invariants to preserve

- No hardcoded secrets.
- No broad rewrites or unrelated refactors.
- No new dependency.
- Deterministic, offline-capable, self-contained Go binary.
- cARL artefacts remain the canonical governance source.
- Existing pack semantics and JSON contracts remain compatible.

## Expected files / directories

See `.github/carl/current-pr-contract.md`.

## Implementation phases

1. Define explanation models and a pure trace builder over discovered packs
   and the existing `ComputeEffectiveSet` result.
2. Represent structured activation steps for selection, defaults, profiles,
   role/task overlays, and dependency expansion with their canonical source
   artefact.
3. Represent resolved decisions for inclusion, precedence, additive
   strengthening, permitted overrides, and unresolved conflicts.
4. Add top-level command adapters with human and JSON rendering, structured
   errors, and non-zero conflict exits.
5. Add focused contract tests, including inactive packs, profile/dependency
   provenance, override/conflict decisions, output boundary notice, and
   no-write behaviour.
6. Update user documentation, architecture, glossary, roadmap, memory and
   embedded memory, trust boundaries, and contract status.
7. Run the full validation matrix and inspect the final diff.

## Acceptance criteria

- Known effective and inactive packs are both explainable.
- Canonical definition and source are explicit and repository-relative.
- Activation/dependency chains identify the selecting profile artefact or
  requiring pack.
- Trace order matches `carl pack effective`.
- Permitted overrides are explained; invalid overrides stay conflicts.
- JSON is schema version 1 and deterministic.
- Human and JSON output contain the no-chain-of-thought boundary.
- Commands do not write files or access registries.

## Contract assertions

The five assertions in `.github/carl/current-pr-contract.md` are binding.

## Test strategy

- Assertion 1: unit-test effective and inactive explanations, source paths,
  activation steps, precedence, and override state.
- Assertion 2: unit-test full trace ordering, structured decisions, permitted
  overrides, unresolved conflicts, and non-zero command exits.
- Assertion 3: snapshot repository files before/after command execution and
  use a fetcher that fails if called.
- Assertion 4: assert the boundary notice in human and JSON output and ensure
  no rule-level or reasoning fields exist.
- Assertion 5: run existing `internal/pack` and repository-wide suites,
  `go vet`, and `go build`.

## Prompt ping-pong budget

One corrective prompt is acceptable. Two means reset the session. Three means
abandon the session/model and restart with the same contract.

## Model fallback strategy

Preserve the Phase 5 contract, non-goals, invariants, output schema, and
acceptance criteria across any model or session reset.

## Stop conditions

Use the active contract's stop conditions.

## Context reset requirements

Record stable Phase 5 command behaviour and the diagnostic explanation
boundary in memory, architecture, roadmap, glossary, and trust boundaries.
Mark the plan delivered and the active contract complete after validation.
