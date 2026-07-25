<!-- version: 1.4.0 -->
# Current PR Contract

## Goal

Implement Pack Phase 2: Pack Composition —

1. `carl pack select <pack-id>...` / `carl pack unselect <pack-id>...`
2. `carl pack effective`
3. dependency expansion, effective pack set computation, conflict detection,
   precedence rules, and explicit override handling.

Composition must remain conservative: packs add constraints; no pack silently
disables another. Override authority must be explicit metadata, never inferred
from load order.

## Contract status

active

## Non-goals

- No profiles or agent roles (`state.active` remains derived from `selected`).
- No remote registry, publishing, or install/update flows.
- No pack dependency downloading.
- No policy-IR compiler.
- No harness authority model changes.
- No repo-map redesign.

## Approved scope

- `cmd/carl/main.go` (subcommand help text only, if required)
- `internal/pack/*` (selection persistence, metadata header parsing, composition)
- Relevant tests under `internal/pack/`
- `.github/carl/packs.json` as a new user-owned selection artefact (schema only; not written in this repository unless selection is exercised)
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`
- `.github/carl/memory.md` and `embedded/assets/.github/carl/memory.md`
- This contract file

## Forbidden scope

- No edits to CI/release workflows.
- No writes to `.github/carl/runtime.json` (init-only invariant preserved).
- No changes to install/repair semantics.
- No unrelated refactors across existing command packages.
- No destructive repository operations.
- No new third-party dependencies.

## Architectural constraints

- Offline-first, deterministic behaviour; no network access.
- Selection state lives in an explicit committed artefact (`.github/carl/packs.json`), never inferred from filesystem order.
- Absent composition metadata must be handled gracefully (defaults: no dependencies, additive mode, no overrides).
- Precedence ordering: explicit priority, ties broken by pack ID — never load order.
- Overridden packs remain in the effective set, flagged, never removed.
- `--json` output is schema-versioned and stable; JSON errors are structured with non-zero exit.
- Legacy behaviour preserved: when `packs.json` is absent, selection falls back to runtime.json-derived managed-artefact selection.

## Security constraints

- Never commit secrets.
- Treat `packs.json` and pack file headers as untrusted input; validate before use.
- Explicit errors for invalid selection state or metadata; no silent fallback.
- Writes stay inside the repository root.

## Files expected to change

- `.github/carl/current-pr-contract.md`
- `internal/pack/pack.go`
- `internal/pack/selection.go` (new)
- `internal/pack/compose.go` (new)
- `internal/pack/pack_test.go`
- `internal/pack/selection_test.go` (new)
- `internal/pack/compose_test.go` (new)
- `README.md`, `CLI.md`, `ARCHITECTURE.md`, `GLOSSARY.md`, `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`

## Contract assertions

1. `carl pack select` / `unselect` persist selection deterministically (sorted, deduplicated) in `.github/carl/packs.json`, validating that selected packs exist.
2. `carl pack effective` computes the effective set as explicit selection plus transitive required dependencies, each entry carrying an explicit reason (`selected` or `dependency of <id>`).
3. Effective output is ordered by explicit precedence (priority desc, then pack ID) — never filesystem or load order — and `--json` output is valid, schema-versioned JSON.
4. An override is honoured only when declared in explicit pack metadata and the target pack declares mode `overridable`; overridden packs remain listed with `overriddenBy`; overriding a non-overridable pack or mutual overrides are reported as conflicts with non-zero exit.
5. Absent composition metadata yields safe defaults; malformed metadata, unknown selected packs, or malformed `packs.json` produce explicit errors (structured JSON with `--json`).

## Tests / validation

- `go test ./internal/pack`
- `go test ./...`
- `go build ./cmd/carl`

## Stop conditions

Stop and escalate if:

- selection persistence cannot avoid touching `runtime.json`;
- composition requires redesigning the phase-1 metadata schema rather than extending it.

## Escalation triggers

Escalate if:

- override semantics require removing packs from the effective set;
- documentation updates conflict with canonical architecture claims.

## Context reset notes

When complete, supersede this contract with the next active task contract.
