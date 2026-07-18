<!-- version: 1.2.0 -->
# Current PR Contract

This contract constrains implementation scope for the active PR.

---

## Goal

Update `carl version` to model and report three distinct version layers:

1. CLI executable version.
2. Bundled canonical runtime version + provenance.
3. Repository-installed runtime version (from `.github/carl/runtime.json`).

Also add component-level pack and harness shim version reporting, including
bundled vs installed comparisons via `carl version --components`.

## Contract status

active

## Non-goals

- No changes to command semantics outside `version` and the metadata wiring needed
  for `init` runtime manifest provenance.
- No changes to security/auth flows.
- No new dependencies.
- No CI workflow behaviour changes.
- No changes to runtime repair semantics.

## Approved scope

- `cmd/carl/main.go` — introduce explicit build-time metadata vars for CLI and bundled runtime.
- `internal/install/install.go` (+ tests if required) — write runtime manifest from bundled runtime metadata.
- `internal/version/version.go`
- `internal/version/version_test.go`
- `.goreleaser.yaml` — inject bundled runtime metadata at build time.
- `CLI.md`, `README.md`, and `ROADMAP.md` — document the three-layer version model and updated build/version semantics.
- `.github/carl/memory.md` — update durable command-behavior truth if command semantics change durably.
- `embedded/assets/.github/carl/memory.md` — keep bundled canonical memory copy aligned when durable memory truth changes.
- `.github/carl/current-pr-contract.md` — this contract update.

## Forbidden scope

- No edits to unrelated commands (`doctor`, `status`, `repair`, `plan`, etc.) except compile-safe constructor signature propagation.
- No harness adapter authority model changes.
- No instruction-pack policy rewrites unrelated to version reporting.
- No external API/network integration.
- No destructive repository operations.

## Architectural constraints

- CLI version and bundled runtime version must be represented by distinct fields.
- Bundled runtime metadata must be available without reading repository state.
- Repository runtime metadata is optional and only read from runtime manifest when present.
- Pack and shim installed versions are derived from installed file metadata headers.
- Harness shim listing must use the canonical harness adapter registry and avoid shared-loader duplication in shim rows.
- Output must be deterministic.

## Security constraints

- No secrets in code or tests.
- No unsafe path handling; read files only under repository root derived from known paths.
- Unknown/malformed metadata must degrade safely (`unknown`) rather than fail the command.

## Files expected to change

- `.github/carl/current-pr-contract.md`
- `cmd/carl/main.go`
- `internal/install/install.go`
- `internal/install/install_test.go` (if constructor/metadata assertions require updates)
- `internal/version/version.go`
- `internal/version/version_test.go`
- `.goreleaser.yaml`
- `CLI.md`
- `README.md`
- `ROADMAP.md`
- `.github/carl/memory.md`
- `embedded/assets/.github/carl/memory.md`

## Tests / validation

- `go test ./internal/version ./internal/install`
- `go test ./...`
- `go build ./cmd/carl`

## Stop conditions

Stop and escalate if:

- requested output requires changing runtime.json schema;
- requested behaviour requires broad rewrites outside approved scope.

## Escalation triggers

Escalate if:

- compatibility requirements for existing `carl version` parsers conflict with requested output shape;
- bundled metadata source-of-truth cannot be made immutable at build time without broader release-pipeline changes.

## Context reset notes

When complete, supersede this contract with the next active task contract.
