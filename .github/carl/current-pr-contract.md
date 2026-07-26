<!-- version: 1.0.0 -->
# Current PR Contract

## Contract status

Active

## Goal

Establish whether the current `main` branch is genuinely ready for the
`v1.0.0-rc.1` release candidate, close only demonstrated release-blocking
gaps, and leave auditable compatibility, lifecycle, distribution, and
release-pipeline evidence without creating a tag or publishing a release.

## Previous contract status

The enterprise-example adoption contract was completed by PR #44. It is
historical evidence, not active authority, and is superseded for this task.
Its durable opt-in and fail-safe adoption semantics remain binding through
memory, documentation, tests, and current repository behaviour.

## Non-goals

- No new product capability, command, TUI, policy IR, Pack Phase 7 publishing
  model, marketplace, or schema redesign.
- No support-tier promotion without native-harness execution evidence.
- No tag, GitHub Release, Homebrew publication, or WinGet submission.
- No change to release secrets or weakening of validation.
- No byte-for-byte stability promise for human-readable output or prose.
- No claim that cARL exposes hidden model reasoning or proves perfect
  instruction compliance.

## Approved scope

- Release-readiness governance and prompt-as-code artefacts under
  `.github/carl/`.
- A durable v1 compatibility policy.
- Release-readiness evidence and `v1.0.0-rc.1` prerelease notes.
- Release-facing documentation: `README.md`, `CLI.md`, `ARCHITECTURE.md`,
  `DISTRIBUTION.md`, `ROADMAP.md`, and directly related documentation.
- Tests, scripts, embedded assets, build metadata, GoReleaser configuration,
  and release workflows only when a demonstrated release blocker requires a
  narrowly scoped correction.
- Generated repository map and reconciled durable memory after final file
  changes.
- Branch, commit, push, and draft pull-request metadata for this readiness
  change.

## Forbidden scope

- Do not add major commands, a TUI, Pack Phase 7, a marketplace, or a new
  policy intermediate representation.
- Do not change a schema unless an observed release blocker cannot be fixed
  compatibly.
- Do not weaken validation, path safety, provenance, conflict handling,
  runtime ownership boundaries, or release failure behaviour.
- Do not alter secret values or expose secret material.
- Do not promote Cursor or Antigravity beyond theoretical support without
  native-harness evidence.
- Do not rewrite user-owned policy, provenance, memory, or profile state.
- Do not modify unrelated local files.
- Do not create or move a tag, publish a GitHub Release, update Homebrew, or
  submit WinGet manifests.

## Compatibility constraints

- Stable v1 contracts must cover documented commands and semantics, documented
  exits, schema-versioned JSON, runtime/pack/profile/registry/provenance state,
  repository-map schema, pack metadata, policy composition, ownership
  boundaries, and the documented lifecycle commands.
- Additive JSON fields are compatible unless an individual contract forbids
  them. Removing a field or changing its meaning requires a schema-version
  transition.
- User-owned policy files are never silently replaced. Repair remains limited
  to declared repairable runtime-owned assets.
- Intentional breaking changes to stable public contracts require a new major
  version.
- Human-readable formatting, undeclared wording/order, documentation prose,
  compatible bundled pack revisions, implementation details, and explicitly
  experimental/theoretical behaviour are not byte-for-byte stable.

## Architectural constraints

- Use release-equivalent host binaries built with the repository's actual
  GoReleaser ldflags and `v1.0.0-rc.1` provenance model.
- Run destructive lifecycle scenarios only in isolated temporary
  repositories, never in the source repository.
- Obtain or build the v0.4.3 state from the authoritative tag or release
  asset; do not simulate upgrade evidence.
- Preserve exact profile-absent/default-profile parity and the documented
  fail-safe enterprise adoption sequence.
- Harness adapters remain thin routes to the shared loader. Local
  detection/sync evidence is distinct from native-harness production evidence.
- Repository-map and reconcile output remain deterministic and
  evidence-scoped.
- Release-pipeline claims must be classified as statically validated,
  previously production-proven, or requiring `v1.0.0-rc.1` execution evidence.

## Security constraints

- Use no live customer, tenant, production, or secret data in fixtures.
- Do not print, inspect, modify, or infer release secret values.
- Registry checksum claims remain limited to integrity against the configured
  index, not publisher identity or a signing trust root.
- Adoption, repair, pack installation, map, reconcile, and harness tests must
  preserve repository path and symlink trust boundaries.
- Release automation must fail closed when required Apple credentials are
  absent and must not report partial publication as complete success.

## Expected files

- `.github/carl/current-pr-contract.md`
- `.github/carl/plans/v1.0.0-rc.1-release-readiness.md`
- `COMPATIBILITY.md`
- `RELEASE_READINESS.md`
- `RELEASE_NOTES_v1.0.0-rc.1.md`
- `README.md`
- `CLI.md`
- `ARCHITECTURE.md`
- `DISTRIBUTION.md`
- `ROADMAP.md`
- `.github/carl/memory.md`
- `.github/carl/repo-map.json`
- Additional implementation, test, embedded, or release files only when
  required by a demonstrated blocker and recorded in the readiness evidence.

## Contract assertions

1. A release-equivalent host binary identifies CLI and bundled-runtime
   provenance outside a repository and distinguishes CLI, bundled, and
   repository runtime layers inside one.
2. Isolated fresh install, adoption, v0.4.3 upgrade, profile, enterprise,
   harness, map, reconcile, and version scenarios produce the documented
   results without silently rewriting protected or user-owned state.
3. The v1 compatibility policy accurately separates stable public contracts,
   compatible evolution rules, and non-byte-stable implementation/presentation
   details.
4. Repository validation, release configuration, retry logic, workflow YAML,
   canonical/embedded parity, adapter routing, and generated-map consistency
   are either proven or recorded with an exact honest limitation.
5. Release notes and release-facing documentation agree with current
   behaviour, support tiers, distribution paths, upgrade steps, and evidence
   limitations.

## Validation requirements

- Build an RC binary with the exact host equivalent of GoReleaser metadata.
- Execute and preserve results for every lifecycle scenario in the linked plan.
- Run `gofmt` verification, `go test -count=1 ./...`, `go vet ./...`,
  `go build ./cmd/carl`, `git diff --check`, GoReleaser config validation,
  release retry-script syntax/tests, workflow YAML parsing, parity/routing/map
  checks, and stale-claim searches.
- Run `go test -race ./...` when supported; otherwise record the exact reason.
- Validate the full tag-to-release flow statically without publishing.
- Re-run relevant validation after every release-blocking correction.

## Stop conditions

Stop and report if:

- exact v0.4.3 state cannot be obtained or executed;
- a required fix needs a new feature, unsupported schema break, weakened
  validation, or user-owned-state rewrite;
- release evidence would require creating a tag or publishing externally;
- a requested support-tier claim lacks native-harness evidence;
- unrelated working-tree changes overlap required files;
- a required remote mutation other than the requested branch push/draft PR is
  needed.

## Escalation triggers

- Any demonstrated blocker requiring release-workflow, GoReleaser,
  trust-boundary, schema, embedded-runtime, or public CLI-contract changes.
- Any ambiguity about whether a file is runtime-owned or user-owned.
- Any need to clean up a partially published external release.
- Authentication or permission failure that prevents the requested branch
  push or draft PR.

## cARL/docs update expectation

Required. Establish the compatibility contract, active plan, evidence report,
release notes, reconciled release documentation, durable memory update, and
generated map. Mark this contract complete only after validation and draft PR
creation, or leave it active with the exact delivery blocker recorded.

## Context reset notes

This contract governs only `v1.0.0-rc.1` readiness. Findings must be classified
as BLOCKER, RC EXIT CRITERION, FOLLOW-UP, or NON-ISSUE. Do not carry transient
host-only evidence forward as a cross-platform guarantee.
