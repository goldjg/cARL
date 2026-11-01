# Managed Trust-Boundary Extension Plan

## Plan metadata

- PR / branch: `goldjg/cARL#50` / `goldjg-governance-extension-drift`
- Status: Completed
- Author: Copilot
- Created: 2026-11-01
- Last updated: 2026-11-01

## Task summary

Add a fixed-path repository-owned trust-boundary extension while preserving
managed drift detection and repair behavior; address review findings about
unsafe status classification, advisory path checks, and repair overwrite
warnings.

## Goal

Provide a safe, obvious project-owned home for additional trust-boundary rules
without editing cARL's managed baseline or weakening its drift and repair
guarantees.

## Affected files

- `.github/carl/current-pr-contract.md`
- `.github/copilot-instructions.md` and its embedded mirror
- `.github/instructions/core/carl.instructions.md` and its embedded mirror
- `.github/carl/trust-boundaries.md` and its embedded mirror
- `internal/doctor/`, `internal/repair/`, and focused tests
- `embedded/embedded_test.go`
- `CLI.md`, `ARCHITECTURE.md`, `COMPATIBILITY.md`, `README.md`, and
  `.github/carl/memory.md`

## Step-by-step changes

1. Reproduce managed-file drift in temporary install fixtures.
2. Define `.github/carl/trust-boundaries.repository.md` as a fixed, optional,
   repository-owned Markdown extension, not runtime state or a general file
   discovery mechanism.
3. Make the shared loader and core cARL pack require the extension to be a
   regular in-repository file, subordinate to system/core invariants, and
   interpreted as restrictive guidance only. Require a path re-check
   immediately before reading. CLI checks are point-in-time and advisory;
   they do not mediate agent reads or prevent path replacement. Stop for
   unresolved conflicts.
4. Give doctor, status, and repair precise safety/preservation guidance.
   Invalid extension paths report overall `Unsafe` status, ahead of
   `Incomplete` and `Drifted`, while managed missing/drifted lists and exit
   behavior remain unchanged. Explicit repair still overwrites managed local
   edits and does not create a backup.
5. Add disposable-fixture regression coverage and keep embedded copies in sync.
6. Document ownership, compatibility, migration, and the limits of prose
   validation; update durable memory.

## Contract assertions

1. Managed edits remain drifted; the fixed extension is not managed.
2. Explicit repair restores the managed baseline after warning and preserves
   extension bytes.
3. Agents re-check immediately before reading the fixed path; CLI checks are
   advisory point-in-time diagnostics and do not enforce later reads. The
   extension grants no execution/tool authority.
4. An absent extension remains compatible; runtime manifest and pack
   composition semantics are unchanged.
5. Invalid extension paths yield `Unsafe`, taking precedence over managed
   missing/drifted status without changing those lists or exit behavior.
6. Explicit repair overwrites managed local edits without a backup and
   preserves the repository-owned extension.
7. Canonical and embedded managed sources remain byte-identical.

## Test strategy

- Use `t.TempDir()` install fixtures; capture status/doctor output and assert
  managed drift, actionable extension guidance, canonical restoration, and
  extension preservation.
- Exercise missing, regular, symlink, and non-regular extension paths where
  platform support permits; assert final status for absent, valid, symlinked
  file, non-regular directory, symlinked parent, and invalid-plus-drift cases.
- Assert repair warns about overwrite/no-backup/recovery before restoring
  canonical bytes and preserving the extension.
- Run focused Go tests, full tests, race tests where supported, vet, formatting,
  parity, and diff checks.

## Risks

The CLI can validate a fixed path and filesystem type but cannot determine
whether arbitrary Markdown prose conflicts semantically with a core invariant.
This is a point-in-time advisory diagnostic: the CLI does not mediate agent
reads and cannot prevent path replacement between inspection and read. Agents
must re-check immediately before reading. A secure-read helper is a future
option only if consumers actually use that API. Agent and human review must
resolve semantic conflicts; extension text never grants permissions or
execution authority.

## Follow-up owner decisions

- Whether to pursue a secure-read helper. It would only enforce safe path
  handling for consumers that actually use it; this PR keeps CLI inspection
  advisory by design.
- Whether to add non-destructive backups of drifted managed files before
  repair. That adds side-effect files and requires decisions about naming,
  idempotence, symlink safety, and ensuring backups cannot be mistaken for
  active policy extensions. This PR preserves existing explicit-overwrite
  behavior.

## cARL/docs update expectation

Required: ownership and hydration behavior change durable project truth.
Update canonical trust-boundary guidance, memory, architecture/CLI/compatibility
documentation, and synchronized managed loader/pack assets.
