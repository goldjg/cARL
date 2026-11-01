# Managed Trust-Boundary Extension Plan

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
   interpreted as restrictive guidance only. Stop for unresolved conflicts.
4. Give doctor and repair precise preservation guidance while retaining
   exact-byte canonical comparison and explicit repair behavior.
5. Add disposable-fixture regression coverage and keep embedded copies in sync.
6. Document ownership, compatibility, migration, and the limits of prose
   validation; update durable memory.

## Contract assertions

1. Managed edits remain drifted; the fixed extension is not managed.
2. Explicit repair restores the managed baseline after warning and preserves
   extension bytes.
3. The extension loads only from the fixed non-symlink repository path and
   grants no execution/tool authority.
4. An absent extension remains compatible; runtime manifest and pack
   composition semantics are unchanged.
5. Canonical and embedded managed sources remain byte-identical.

## Test strategy

- Use `t.TempDir()` install fixtures; capture status/doctor output and assert
  managed drift, actionable extension guidance, canonical restoration, and
  extension preservation.
- Exercise missing, regular, symlink, and non-regular extension paths where
  platform support permits.
- Run focused Go tests, full tests, race tests where supported, vet, formatting,
  parity, and diff checks.

## Risks

The CLI can validate a fixed path and filesystem type but cannot determine
whether arbitrary Markdown prose conflicts semantically with a core invariant.
Agent and human review must resolve such conflicts; extension text never
grants permissions or execution authority.

## cARL/docs update expectation

Required: ownership and hydration behavior change durable project truth.
Update canonical trust-boundary guidance, memory, architecture/CLI/compatibility
documentation, and synchronized managed loader/pack assets.
