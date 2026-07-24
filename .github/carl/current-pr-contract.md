<!-- version: 1.3.0 -->
# Current PR Contract

## Goal

Implement the first versioned pack-runtime vertical slice:

1. `carl pack list`
2. `carl pack show <pack-id>`

including deterministic discovery, validated metadata, and JSON output.

## Contract status

active

## Non-goals

- No remote registry, publishing, or install/update flows.
- No pack dependency downloading.
- No policy-IR compiler replacement for instruction loading.
- No harness authority model changes.

## Approved scope

- `cmd/carl/main.go`
- `internal/cmdutil/*` (only to support stable structured error exits)
- `internal/pack/*` (new package for pack command and metadata/discovery)
- Relevant tests under `internal/pack/` (and minimal updates elsewhere if required)
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `VISION.md`, `ROADMAP.md`, `GLOSSARY.md`
- `.github/carl/memory.md` (only if durable architecture truth changes)
- `embedded/assets/.github/carl/memory.md` (only if memory is updated)
- This contract file

## Forbidden scope

- No edits to CI/release workflows.
- No changes to runtime install/repair semantics outside pack discovery integration.
- No unrelated refactors across existing command packages.
- No destructive repository operations.

## Architectural constraints

- Preserve offline-first, deterministic behaviour.
- Preserve repository-local operation and self-contained binary characteristics.
- Pack discovery must not rely on incidental filesystem iteration order.
- Unknown packs in `show` must produce stable non-zero exit behaviour.
- Human output is for people; `--json` must be machine-readable and stable.

## Security constraints

- Never commit secrets.
- Treat repository files as untrusted input; validate parsed metadata.
- Use explicit errors for invalid metadata; no silent fallback.

## Files expected to change

- `.github/carl/current-pr-contract.md`
- `cmd/carl/main.go`
- `internal/cmdutil/exit.go`
- `internal/pack/pack.go`
- `internal/pack/pack_test.go`
- `README.md`
- `CLI.md`
- `ARCHITECTURE.md`
- `VISION.md`
- `ROADMAP.md`
- `GLOSSARY.md`
- `.github/carl/memory.md` (if required)
- `embedded/assets/.github/carl/memory.md` (if required)

## Contract assertions

1. `carl pack list` returns deterministic pack ordering and includes stable summary metadata.
2. `carl pack show <pack-id>` returns deterministic detailed metadata for known packs.
3. `--json` output for list/show is valid JSON with schema versioning.
4. Unknown pack with `--json` returns valid structured JSON error and non-zero exit.
5. Pack metadata validation rejects malformed IDs, invalid versions, missing dependencies, dependency cycles, and invalid owned-artefact references.

## Tests / validation

- `go test ./internal/pack`
- `go test ./...`

## Stop conditions

Stop and escalate if:

- existing command dispatch cannot support JSON error semantics without broad breaking changes;
- validation requirements force a broader metadata storage redesign outside this vertical slice.

## Escalation triggers

Escalate if:

- requested behaviour requires changing existing CLI global exit-code semantics beyond scoped command needs;
- required documentation updates conflict with current canonical architecture claims.

## Context reset notes

When complete, supersede this contract with the next active task contract.
