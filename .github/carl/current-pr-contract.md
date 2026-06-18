<!-- version: 1.1.0 -->
# Current PR Contract

This contract constrains implementation scope for the active PR. Update
it when scope is explicitly amended. If a requested action falls outside
approved scope, stop and escalate before proceeding.

Use this contract to distinguish active PR constraints, completed PR
constraints, durable invariants, and intentional amendments. Completed
PR constraints are historical evidence unless explicitly promoted to
durable invariants.

---

## Goal

Add `carl map` — a new CLI command that derives a cognitive repository map
from the filesystem and writes it to `.github/carl/repo-map.json`.
The map includes: languages, entry points, key directories (with Go package
doc-derived purposes), GitHub Actions workflows, governance artefacts, and
root-level documentation. Update durable artefacts (`CLI.md`, `ROADMAP.md`,
`memory.md`) to reflect the new command.

## Contract status

active

## Non-goals

- Parsing existing `repo-map.json` to preserve user edits
- Remote downloads, GitHub API calls, or git log inspection
- Pack install/remove
- Changes to embedded assets or existing instruction packs
- Changes to `invariants.yml` (no new invariants required)

## Carry-forward rules

Promoted invariants from previous PRs remain in force:
- No secrets committed to any file.
- Security baseline (least privilege, no hard-coded credentials) applies.
- `current-pr-contract.md` must be read before implementation begins.

## Approved scope

- `internal/repomap/repomap.go` — new `carl map` command implementation
- `internal/repomap/repomap_test.go` — tests
- `cmd/carl/main.go` — register `map` command
- `CLI.md` — document `carl map`
- `.github/carl/current-pr-contract.md` — this file
- `.github/carl/memory.md` — durable facts update
- `ROADMAP.md` — mark repo map tooling as delivered

## Intentional amendments

No prior constraints are amended. No existing command behaviour changes.

## Forbidden scope

- Changes to instruction packs under `.github/instructions/`
- Changes to `invariants.yml`
- Changes to embedded assets
- Modifying `repair`, `status`, `doctor`, `version`, or `init` commands

## Architectural constraints

- `carl map` writes only `.github/carl/repo-map.json`; no other files.
- Output written to stdout only; errors to stderr via returned error.
- No network access; filesystem scan only.
- All path strings in JSON output use forward slashes.
- Repeated invocation must produce valid JSON (idempotent).
- `.git/`, `node_modules/`, `vendor/` are always excluded from scans.

## Security constraints

- No credentials, tokens, or secrets in any new file.
- No user-controlled data passed to shell commands.
- Filesystem walk is bounded to rootDir; no path traversal outside it.

## Contract assertions

1. Running `carl map` creates `.github/carl/repo-map.json` with valid JSON.
2. The JSON contains `generated_by: "carl map"`, a non-empty `last_updated`, and `_note`.
3. Go source files in the repo cause "Go" to appear in `languages`.
4. `.github/workflows/*.yml` files are listed in `workflows`.
5. Files directly under `.github/carl/` are listed in `governance`.
6. Root-level `*.md` files are listed in `documentation`.
7. Running twice (idempotent) still produces valid JSON.
8. `.git/` is never included in `directories` or `languages`.

## Files expected to change

- `internal/repomap/repomap.go` (new)
- `internal/repomap/repomap_test.go` (new)
- `cmd/carl/main.go`
- `CLI.md`
- `.github/carl/current-pr-contract.md` (this file)
- `.github/carl/memory.md`
- `ROADMAP.md`

## Tests / validation

- `go build ./cmd/carl` — must succeed
- `go test ./...` — must pass with new repomap package tests
- Parallel validation (code review + CodeQL) before PR is opened

## Stop conditions

- Any file writes other than `.github/carl/repo-map.json`.
- Any network or GitHub API access.
- Any change to embedded assets or instruction packs.

## Escalation triggers

- Request to add remote download or GitHub API call.
- Request to preserve user edits in repo-map.json across runs.

## Context reset notes

This contract is active for the `carl map` PR. Close it when merged.
