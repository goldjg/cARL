<!-- version: 2.0.0 -->
# Current PR Contract

## Goal

Reconcile roadmap, documentation, harness diagnostics, and canonical cARL
artefacts with the repository's implemented behavior and validated harness
support:

1. Record GitHub Copilot, Claude Code, and Codex as proven production
   harnesses using their implemented shared-loader adapters.
2. Record Cursor and Antigravity as implemented and synchronized but not
   end-to-end validated, retaining their `theoretical` tier.
3. Remove the obsolete Claude `/carl` skill prerequisite.
4. Distinguish adapter-file detection and sync health from governance
   activation.
5. Correct stale adapter paths and roadmap delivery claims while preserving
   the delivered Pack Phase 6 cognitive repository graph.

## Contract status

complete

## Previous contract

Pack Phase 6 is delivered. Its completed contract remains historical evidence;
the durable graph schema, evidence limitations, deterministic local-only
behavior, compatibility, and trust-boundary constraints promoted to memory,
architecture, and trust boundaries remain binding.

## Intentional contract amendment

The completed Phase 6 contract prohibited harness, instruction-pack, and
unrelated roadmap edits for that feature. This maintainer-approved
reconciliation explicitly supersedes those file-scope constraints so factual
errors can be corrected. It does not alter Phase 6 graph semantics.

## Non-goals

- No new adapter, Claude skill, validation protocol, registry, pack feature,
  CI check, or release behavior.
- No promotion of Cursor or Antigravity without end-to-end validation.
- No rename of existing harness support-tier values.
- No claim that adapter-file presence or sync proves governance activation.
- No change to pack evaluation, registry, release, runtime-state, or cognitive
  graph semantics.
- No new dependency or further Pack Phase 6 feature work.

## Approved scope

- `internal/harness/health.go`
- `internal/harness/harness.go`
- Focused tests under `internal/harness/`
- `internal/status/status.go`
- Focused tests under `internal/status/`
- `internal/reconcile/reconcile.go` and focused tests, solely to keep refreshed
  generated snapshots free of trailing whitespace
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`
- `.github/copilot-instructions.md` and its embedded mirror
- Obsolete root adapter files `.cursorrules` and `ANTIGRAVITY.md` (removal)
- `.github/carl/invariants.yml` and its embedded mirror
- `.github/carl/trust-boundaries.md` and its embedded mirror
- `.github/carl/memory.md` and its embedded mirror
- `.github/carl/repo-map.json` when refreshed by the delivered graph-aware
  `carl map`
- `.github/instructions/core/carl.instructions.md` and its embedded mirror
- `.github/instructions/core/pr-contract.instructions.md` and its embedded
  mirror
- `.github/instructions/core/tool-permission-tiers.instructions.md` and its
  embedded mirror
- `.github/instructions/core/security.instructions.md` and its embedded mirror
- `.github/carl/plans/roadmap-governance-reconciliation.md`
- This contract file

## Forbidden scope

- No edits under `.github/workflows/`, to `.goreleaser.yaml`, release scripts,
  dependency manifests, or `internal/repomap/`.
- No writes to `.github/carl/runtime.json`, `.github/carl/packs.json`,
  `.github/carl/profiles.json`, `.github/carl/registries.json`, or
  `.github/carl/installed-packs.json`.
- No changes to current adapter IDs, current adapter filenames, embedded shim
  content, support tier values, or sync behavior beyond the shared loader's
  corrected authority references. The obsolete root `.cursorrules` and
  `ANTIGRAVITY.md` alternate authorities are explicitly removed.
- No network behavior, hidden agent-state inference, destructive operations,
  broad rewrites, or unrelated refactors.
- No loss or weakening of the Phase 6 graph, coverage, evidence, or
  compatibility fields.

## Architectural constraints

- Canonical governance remains under `.github/carl/` and
  `.github/instructions/`; harness files remain generated adapters/loaders.
- The five implemented adapters are Copilot, Claude, Codex, Cursor, and
  Antigravity.
- Production evidence is limited to Copilot, Claude, and Codex. Cursor and
  Antigravity remain `theoretical` until end-to-end validation is recorded.
- `Present` means the adapter's detection file exists. `Synced` means managed
  adapter files match embedded sources. Neither state proves agent activation.
- Existing `production`, `experimental`, and `theoretical` values and command
  exit behavior remain backward compatible.
- Release and runtime descriptions must derive from current repository
  configuration, not obsolete roadmap prose.
- The refreshed repository map must retain Phase 6 schema version 1, its
  additive cognitive graph, deterministic evidence, and all compatible
  inventory fields.

## Security constraints

- Treat repository governance and adapter files as untrusted inputs to
  diagnostics.
- Do not inspect prompts, sessions, model reasoning, or hidden harness state.
- Do not expose secrets or modify secret configuration.
- Keep diagnostics local-only and preserve path confinement and byte-for-byte
  comparison behavior.
- Governance-sensitive path changes must be explicit in this contract and
  mirrored in embedded sources.
- Preserve the graph's evidence limitations: static imports do not prove
  runtime flow, attachment points do not prove active policy, and ownership is
  never guessed.

## Trust boundaries

- Maintainer evidence establishes which harnesses are proven end-to-end.
- Repository adapter definitions establish implementation coverage.
- Local filesystem checks establish only detection and canonical sync health.
- Generated reports and roadmap prose are derived evidence, not new canonical
  authorities.
- Registry digests remain integrity evidence rather than publisher identity.
- The cognitive graph remains derived orientation evidence, not canonical
  governance, ownership, runtime, risk, or active-policy evidence.

## Expected files

- `.github/carl/current-pr-contract.md`
- `.github/carl/plans/roadmap-governance-reconciliation.md`
- `internal/harness/health.go`
- `internal/harness/harness.go`
- `internal/harness/harness_test.go`
- `internal/status/status.go`
- `internal/status/status_test.go`
- `internal/reconcile/reconcile.go`
- `internal/reconcile/reconcile_test.go`
- `README.md`
- `CLI.md`
- `ARCHITECTURE.md`
- `GLOSSARY.md`
- `ROADMAP.md`
- `.github/copilot-instructions.md`
- `embedded/assets/.github/copilot-instructions.md`
- `.cursorrules` (removed)
- `ANTIGRAVITY.md` (removed)
- `.github/carl/invariants.yml`
- `embedded/assets/.github/carl/invariants.yml`
- `.github/carl/trust-boundaries.md`
- `embedded/assets/.github/carl/trust-boundaries.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`
- `.github/carl/repo-map.json`
- Four core instruction packs named in approved scope and their embedded
  mirrors

## Contract assertions

1. Harness inventory and support tiers remain exactly five adapters: Copilot,
   Claude, and Codex production; Cursor and Antigravity theoretical.
2. Human status output labels present detection files as `detected`, never as
   evidence that governance activated; sync/missing/drifted counts retain
   their behavior.
3. Every governance authority list uses `.cursor/rules/carl.mdc` and
   `.agents/rules/carl.md`; obsolete `.cursorrules` and `ANTIGRAVITY.md`
   references are absent and changed embedded mirrors are byte-identical.
4. Roadmap and durable memory accurately describe current release signing and
   notarization, embedded asset count, runtime-state ownership, delivered pack
   validation/registry/capability/graph behavior, and remaining work.
5. Existing harness IDs, support-tier strings, sync behavior, exit codes,
   runtime-owned state, pack behavior, dependencies, release configuration,
   and Phase 6 graph semantics remain unchanged.

## Validation plan

- `gofmt -w internal/harness/health.go internal/harness/harness.go internal/harness/harness_test.go internal/status/status.go internal/status/status_test.go internal/reconcile/reconcile.go internal/reconcile/reconcile_test.go`
- `go test ./internal/harness ./internal/status ./internal/reconcile ./internal/repomap`
- `go test ./...`
- `go vet ./...`
- `go build ./cmd/carl`
- Harness list/status and overall status smoke checks
- Graph-aware `carl map` followed by idempotent `carl reconcile`
- Obsolete-path and obsolete-Claude-skill searches
- Byte-for-byte comparison of every changed canonical/embedded pair
- Embedded asset recount and release-configuration fact check
- `git diff --check`

## cARL/docs update expectation

Expected. This task exists to reconcile durable governance, roadmap, and
documentation with implementation and maintainer validation evidence while
preserving the merged Pack Phase 6 architecture.

## Stop conditions

Stop and escalate if:

- any correction would require promoting an unvalidated harness;
- activation claims would require observing hidden agent or session state;
- compatibility requires changing support-tier strings, exit codes, or
  runtime-owned state;
- current release or registry truth cannot be established from repository
  configuration;
- canonical and embedded artefacts cannot be kept byte-identical;
- reconciliation would remove or weaken Phase 6 graph evidence or semantics.

## Escalation triggers

Escalate if:

- a new governance authority or Claude skill becomes necessary;
- automated end-to-end harness conformance is requested as implementation
  rather than roadmap work;
- reconciliation expands into CI, release, registry, pack, or cognitive-graph
  feature behavior;
- maintainer validation evidence conflicts with repository adapter definitions.

## Context reset notes

Preserve the three-proven/two-unvalidated harness boundary, the
detection-versus-activation distinction, current runtime ownership and release
facts, and the delivered Phase 6 graph evidence boundary.

## Completion checklist

- [x] Implementation stayed inside approved scope.
- [x] Forbidden scope was not touched.
- [x] Contract assertions were validated directly.
- [x] Tests and manual smoke checks were run or gaps reported.
- [x] Documentation and cARL artefacts were reconciled.
- [x] This contract was marked complete.
