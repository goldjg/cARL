# Roadmap and Governance Reconciliation

## Plan metadata

- PR / branch: `agent/roadmap-governance-reconciliation`
- Status: completed
- Author: Codex
- Created: 2026-07-25
- Last updated: 2026-07-25

## Task summary

Reconcile `ROADMAP.md`, user-facing documentation, harness health terminology,
and canonical cARL artefacts with the implementation now present in the
repository and the harness validation evidence supplied by the maintainer.

## Current repository context

The repository implements adapters for GitHub Copilot, Claude Code, Codex,
Cursor, and Antigravity. Copilot, Claude Code, and Codex are production
supported and have been proven through their shared-loader shims. Cursor and
Antigravity have implemented, synchronised shims but have not received
end-to-end harness validation, so their existing `theoretical` support-tier
value remains accurate. Several roadmap sections, governance packs, and CLI
labels still describe earlier assumptions or obsolete adapter paths.

## Previous contract status

The Pack Phase 5 contract is complete historical evidence. Its durable
offline-first, deterministic, read-only explanation constraints remain
recorded in canonical governance artefacts.

## Contract lifecycle note

Completed PR contracts are historical evidence, not binding scope. The active
authority for this implementation is the reconciliation contract in
`.github/carl/current-pr-contract.md`.

## Intentional contract amendments

This plan intentionally supersedes the completed Phase 5 prohibition on
harness, instruction-pack, and roadmap edits. The maintainer explicitly asked
for the audited governance corrections to be implemented. Changes remain
limited to factual reconciliation, adapter-path corrections, detected-versus-
activated terminology, tests, and canonical/embedded mirror updates.

## Goal

- Make the roadmap describe all five implemented adapters and the proven
  production status of Copilot, Claude Code, and Codex.
- Remove the false claim that a Claude `/carl` skill is required for
  production support.
- Describe Cursor and Antigravity as implemented but not end-to-end tested
  while retaining their `theoretical` tier.
- Stop presenting adapter-file presence as proof of governance activation.
- Correct stale adapter paths and keep every embedded canonical mirror
  byte-identical.
- Reconcile stale release, runtime-state, embedded-asset, pack-health,
  registry, capability-profile, and memory claims.

## Non-goals

- No support-tier promotion for Cursor or Antigravity.
- No new harness adapter, skill artefact, validation protocol, registry,
  release behavior, CI workflow, or pack runtime feature.
- No rename of the existing `production`, `experimental`, or `theoretical`
  support-tier values.
- No claim that local file detection proves end-to-end governance activation.
- No dependency or runtime-state changes.
- No inclusion of the separate Pack Phase 6 implementation.

## Approved scope

The approved files and surfaces are enumerated in the active PR contract.
Implementation may update harness summary terminology and its focused tests,
plus the documentation and canonical/embedded governance artefacts required
to restore factual consistency.

## Forbidden scope

- `.github/workflows/**`, `.goreleaser.yaml`, release scripts, and dependency
  manifests.
- Runtime-owned state files: `.github/carl/runtime.json`,
  `.github/carl/packs.json`, `.github/carl/profiles.json`,
  `.github/carl/registries.json`, and
  `.github/carl/installed-packs.json`.
- Adapter support-tier enum changes or production promotion without evidence.
- Destructive repository operations or unrelated refactors.

## Trust boundaries

- Maintainer-provided validation evidence establishes the proven harness set:
  Copilot, Claude Code, and Codex.
- Repository code and byte comparisons establish that all five adapter shims
  are implemented and synchronized.
- File presence and sync health are local diagnostic evidence only; neither
  proves that a harness loaded or obeyed governance.
- Canonical cARL artefacts remain authoritative; harness files remain
  disposable loaders.

## Invariants to preserve

- No hardcoded secrets.
- No broad rewrites or new dependencies.
- Deterministic, offline-capable, self-contained Go binary.
- Existing harness IDs, support-tier values, sync semantics, and exit behavior
  remain compatible.
- Canonical and embedded copies of managed artefacts remain byte-identical.
- Governance remains harness-independent.

## Expected files / directories

See `.github/carl/current-pr-contract.md`.

## Implementation phases

1. Activate this contract and record the factual baseline.
2. Correct adapter authority paths in the shared loader, invariants, trust
   boundaries, and core governance packs; remove obsolete root alternate
   authorities; mirror each embedded source.
3. Rename user-facing harness presence summaries from `active` to `detected`
   while retaining a documented compatibility alias in Go.
4. Reconcile roadmap support, lifecycle, release, foundation, pack-health,
   registry, capability-profile, and deferred-state claims.
5. Reconcile README, CLI, architecture, glossary, memory, embedded memory,
   repository map, and generated memory snapshot where needed.
6. Run focused and repository-wide validation, inspect mirror equality and
   stale-path searches, then mark the contract and plan complete.

## Acceptance criteria

- Documentation consistently reports three proven production harnesses and
  two implemented, unvalidated theoretical harnesses.
- Claude production support is based on the implemented `CLAUDE.md` shim and
  shared loader; no `/carl` skill is described as required or canonical.
- `carl harness status` and `carl status` describe adapter-file presence as
  `detected`, not `active`.
- Obsolete `.cursorrules` and `ANTIGRAVITY.md` alternate authorities and
  references are absent.
- Release and foundation descriptions match current configuration and assets.
- Canonical/embedded managed files compare byte-for-byte.

## Contract assertions

The five assertions in `.github/carl/current-pr-contract.md` are binding.

## Test strategy

- Assertion 1: inspect adapter registry and assert support-tier output remains
  three production and two theoretical.
- Assertion 2: update focused harness/status tests to assert `detected`
  summaries and the distinction from activation.
- Assertion 3: search all canonical and embedded governance surfaces for
  obsolete paths and compare every changed mirror byte-for-byte.
- Assertion 4: run documentation/state audits for asset counts, release
  configuration, registry delivery, pack-health delivery, and capability
  profile facts.
- Assertion 5: run `go test ./...`, `go vet ./...`, `go build ./cmd/carl`,
  CLI smoke checks, and `git diff --check`.

## Prompt ping-pong budget

One corrective prompt is acceptable. Two means reset the session. Three means
abandon the session/model and restart with the same contract.

## Model fallback strategy

Preserve the active contract, evidence boundary, support tiers, non-goals, and
acceptance criteria across any model or session reset.

## Stop conditions

Use the active contract's stop conditions. In particular, stop before any
support-tier promotion that lacks end-to-end validation evidence or any change
that would turn file presence into a claim of agent behavior.

## Context reset requirements

Record the reconciled harness truth, diagnostic evidence boundary, current
runtime ownership model, and current roadmap delivery state in durable memory.
Mark this plan delivered and the contract complete only after validation.
