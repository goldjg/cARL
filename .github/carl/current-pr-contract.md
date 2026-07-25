<!-- version: 1.0.0 -->
# Current PR Contract

## Contract status

Complete

## Goal

Remove the bootstrap deadlock for repositories that contain cARL artefacts but
have no `.github/carl/runtime.json`, while preserving the existing
non-destructive default installation and protected-file repair behaviour.

## Non-goals

- No automatic ownership of pre-existing files during ordinary `carl init`.
- No overwrite of existing artefacts during adoption.
- No change to repair's protected artefacts.
- No manifest schema, pack, harness, registry, release, or dependency change.

## Approved scope

- `internal/install/` and focused tests
- `internal/manifest/` and focused tests when needed for safe manifest creation
- `internal/doctor/` and focused tests
- `README.md`, `CLI.md`, `ROADMAP.md`
- `.github/carl/memory.md`
- This contract

## Forbidden scope

- No changes to embedded runtime assets except the byte-identical durable
  memory mirror; no dependency, CI/CD, release configuration, pack, harness,
  or registry behaviour changes.
- No destructive writes to existing repository artefacts.
- No repair ownership of `memory.md` or `runtime.json`.

## Architectural constraints

- Ordinary `carl init` retains its existing collision failure.
- Adoption must be explicit, repository-local, offline, deterministic, and
  non-destructive.
- `runtime.json` remains init-owned and is created only after all missing
  embedded artefacts have been installed successfully.
- The manifest continues to record the complete bundled artefact list.

## Security constraints

- Treat existing repository files as untrusted and preserve them byte-for-byte.
- Keep every write confined to the existing embedded target list.
- Never replace an existing runtime manifest during adoption.

## Trust boundaries

- The embedded artefact list defines the only managed paths adoption may use.
- Filesystem state establishes presence only; adoption does not claim that
  existing bytes are canonical.
- The generated manifest establishes future repair ownership only after the
  user explicitly requests adoption.

## Expected files

- `.github/carl/current-pr-contract.md`
- `internal/install/install.go`
- `internal/install/install_test.go`
- `internal/manifest/manifest.go`
- `internal/manifest/manifest_test.go`
- `internal/doctor/doctor.go`
- `internal/doctor/doctor_test.go`
- `README.md`
- `CLI.md`
- `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`

## Contract assertions

1. Ordinary `carl init` still refuses every pre-existing managed artefact.
2. `carl init --adopt` preserves existing files, installs only missing bundled
   files, and then creates a valid manifest.
3. Adoption never replaces an existing manifest and does not create one after
   an earlier installation error.
4. Doctor recommends adoption when cARL artefacts exist without a manifest,
   while empty repositories retain the ordinary init recommendation.
5. After adoption, repair restores repairable drift while preserving
   `memory.md` and `runtime.json`.

## Validation plan

- `gofmt` on changed Go files
- `go test ./internal/install ./internal/manifest ./internal/doctor ./internal/repair`
- `go test ./...`
- `go vet ./...`
- `go build ./cmd/carl`
- Manual reproduction in a temporary repository
- `git diff --check`

## cARL/docs update expectation

Expected. Adoption is durable CLI recovery and runtime-ownership behaviour.

## Stop conditions

Stop if the fix requires overwriting existing artefacts without a separate
explicit repair action, changing the manifest schema, weakening protected-file
behaviour, or expanding outside approved scope.

## Escalation triggers

Escalate if dependency, CI/CD, release, registry, pack, harness, or embedded
asset changes become necessary.

## Context reset notes

Carry forward the explicit non-destructive adoption path and the distinction
between filesystem presence, canonical bytes, and manifest-established repair
ownership.

## Completion checklist

- [x] Implementation stayed inside approved scope.
- [x] Forbidden scope was not touched.
- [x] Contract assertions were validated.
- [x] Tests and manual checks were run or gaps reported.
- [x] Documentation and durable memory were updated.
- [x] This contract was marked complete.
