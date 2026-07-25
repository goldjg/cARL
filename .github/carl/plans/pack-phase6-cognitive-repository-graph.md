<!-- version: 1.0.0 -->
# Pack Phase 6: Cognitive Repository Graph

## Plan metadata
- PR / branch: current working branch
- Status: Delivered
- Author: Codex
- Created: 2026-07-25
- Last updated: 2026-07-25

## Task summary

Extend `carl map` from a static repository inventory into a deterministic,
evidence-scoped cognitive graph without breaking existing repo-map consumers.

## Current repository context

`carl map` currently derives languages, entry points, key directories,
workflows, governance artefacts, and documentation into
`.github/carl/repo-map.json`. `carl reconcile` deserialises those fields into
`repomap.Map`; additive JSON fields are therefore backward compatible.

The repository is a dependency-free Go 1.24 module. Package imports can be
parsed safely with the Go standard library without building or executing the
repository.

## Previous contract status

The Pack Phase 5 contract is complete and historical. Its stable local-only,
deterministic, schema-versioned, and diagnostic-evidence boundaries remain
durable.

## Contract lifecycle note

Completed PR contracts are historical evidence, not binding scope. The active
authority for this implementation is the Phase 6 contract in
`.github/carl/current-pr-contract.md`.

## Intentional contract amendments

This plan supersedes the completed Phase 5 execution scope to implement the
roadmap's Phase 6 repository-graph slice. It does not amend pack evaluation,
policy provenance, registry, profile, harness, or reconcile semantics.

## Goal

Add a versioned `graph` object to repo-map output containing stable nodes,
evidence-backed edges, direct change impact, and explicit knowledge coverage.

## Non-goals

- Dynamic/runtime analysis, code execution, or runtime data-flow claims.
- Guessed owners or active policy assignments.
- A graph database, visualiser, query language, or remote graph service.
- Transitive change-impact guarantees.
- Removal or reinterpretation of existing repo-map fields.

## Approved scope

The approved files and surfaces are listed in the active PR contract.

## Forbidden scope

No harness, instruction-pack, runtime-state, pack-state, CI, release,
dependency, network, or destructive changes.

## Trust boundaries

- Repository paths and Go syntax are untrusted local inputs.
- Static imports prove a source dependency declaration, not runtime flow.
- Graph output is derived orientation evidence, not canonical governance.
- Policy attachment points do not represent active policy; use `carl trace`
  for active policy provenance.

## Invariants to preserve

- No hardcoded secrets or new dependencies.
- No broad rewrites.
- Deterministic, offline-capable, self-contained Go binary.
- Existing repo-map fields and reconcile behaviour remain compatible.
- Repository-relative paths only; symlinks are not traversed.

## Expected files / directories

See `.github/carl/current-pr-contract.md`.

## Implementation phases

1. Define graph schema types, evidence status vocabulary, and stable ID rules.
2. Build nodes from the repository root, mapped directories, Go packages,
   entry points, workflows, governance, documentation, and instruction packs.
3. Build containment and repository-local Go import edges with deterministic
   evidence ordering.
4. Derive direct change impact by reversing dependency edges.
5. Add graph coverage for ownership, dependencies, data flows, trust
   boundaries, criticality, policy attachment points, and change impact.
6. Add contract-focused tests and preserve reconcile compatibility.
7. Update docs, roadmap, memory, embedded memory, trust boundaries, example,
   and generated repo map.
8. Run the full validation matrix and inspect the final diff.

## Acceptance criteria

- Repo-map JSON has `schema_version: 1` and retains all legacy fields.
- Graph nodes and edges use stable unique IDs and deterministic sorting.
- Local Go imports create dependency edges with repository-relative evidence.
- Direct reverse dependants are recorded as change impact.
- Component/package nodes are policy attachment points, while coverage text
  explicitly denies active-policy inference.
- Ownership and runtime data-flow gaps are visible rather than guessed.
- Repeated runs on unchanged structure produce byte-identical output within
  the same UTC date.
- Existing reconcile tests and behaviour continue to pass.

## Contract assertions

The five assertions in `.github/carl/current-pr-contract.md` are binding.

## Test strategy

- Assertion 1: assert schema version, retained inventory fields, unique stable
  IDs, deterministic node/edge ordering, and repeated-run bytes.
- Assertion 2: create a temporary Go module with importing packages and assert
  dependency evidence plus reverse direct impact.
- Assertion 3: assert node criticality, trust-boundary classification, policy
  definition nodes, attachment-point flags, and agent context.
- Assertion 4: assert exact coverage statuses and limitations when no
  ownership or runtime-flow source exists.
- Assertion 5: run repomap, reconcile, repository-wide, vet, build, diff, and
  manual command smoke checks.

## Prompt ping-pong budget

One corrective prompt is acceptable. Two means reset the session. Three means
restart with the same Phase 6 contract and acceptance criteria.

## Model fallback strategy

Preserve the Phase 6 contract, graph schema, evidence boundaries, non-goals,
and acceptance criteria across any model or session reset.

## Stop conditions

Use the active contract's stop conditions.

## Context reset requirements

Record stable graph schema and evidence boundaries in memory, architecture,
roadmap, glossary, and trust boundaries. Mark the plan delivered and the active
contract complete only after all validation succeeds.
