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

Add `carl harness` — a new CLI command with two subcommands (`list`,
`status`) that introduce harness adapter support. Harness adapters bridge
cARL canonical artefacts to AI coding agent context injection mechanisms.
cARL artefacts are the canonical source of truth; harness files are
adapters, not authorities. Support GitHub Copilot as the first adapter
without changing existing behaviour. Register Claude Code, Codex, Cursor,
and Antigravity as planned adapters. Update durable artefacts (`CLI.md`,
`ROADMAP.md`, `memory.md`) to reflect the new command.

## Contract status

active

## Non-goals

- Changes to embedded assets or instruction packs
- Changes to `invariants.yml` (no new invariants required)
- Modifying any existing command behaviour
- Writing harness adapter files (read-only command)
- Adapter file content generation or injection
- Network or GitHub API access

## Carry-forward rules

Promoted invariants from previous PRs remain in force:
- No secrets committed to any file.
- Security baseline (least privilege, no hard-coded credentials) applies.
- `current-pr-contract.md` must be read before implementation begins.

## Approved scope

- `internal/harness/harness.go` — new package: Adapter registry, Command, list/status subcommands
- `internal/harness/harness_test.go` — tests
- `cmd/carl/main.go` — register `harness` command
- `CLI.md` — document `carl harness`, `carl harness list`, `carl harness status`
- `.github/carl/current-pr-contract.md` — this file
- `.github/carl/memory.md` — durable facts update
- `ROADMAP.md` — mark harness support as delivered

## Intentional amendments

No prior constraints are amended. No existing command behaviour changes.

## Forbidden scope

- Changes to instruction packs under `.github/instructions/`
- Changes to `invariants.yml` or either embedded/assets copy
- Modifying `repair`, `status`, `doctor`, `version`, `init`, `map`, or `plan` commands
- Creating or modifying harness adapter files (harness command is read-only)

## Architectural constraints

- `carl harness list` and `carl harness status` are read-only; they never write files.
- All output to stdout; errors to stderr via returned error.
- No network access; filesystem check only (os.Stat for detection).
- `carl harness` (no args) prints usage and returns nil.
- Unknown subcommand returns a non-nil error.
- Adapter registry is the canonical source of harness metadata; not read from disk.
- Detection is by presence of a single detection file per supported adapter.

## Security constraints

- No credentials, tokens, or secrets in any new file.
- No user-controlled data passed to shell commands.
- Filesystem check bounded to rootDir; no path traversal outside it.
- `os.Stat` is the only filesystem operation; no file reads.

## Contract assertions

1. `carl harness list` lists all 5 known adapters (copilot, claude, codex, cursor, antigravity).
2. `carl harness list` identifies copilot as "supported"; others as "planned".
3. `carl harness status` shows "active" for copilot when `.github/copilot-instructions.md` exists.
4. `carl harness status` shows "not active" for copilot when detection file is absent.
5. `carl harness` with no args (or --help) prints usage and returns nil.
6. Unknown subcommand returns a non-nil error containing "unknown subcommand".
7. Both `list` and `status` are read-only and always return nil.

## Files expected to change

- `internal/harness/harness.go` (new)
- `internal/harness/harness_test.go` (new)
- `cmd/carl/main.go`
- `CLI.md`
- `.github/carl/current-pr-contract.md` (this file)
- `.github/carl/memory.md`
- `ROADMAP.md`

## Tests / validation

- `go build ./cmd/carl` — must succeed
- `go test ./...` — must pass with new harness package tests
- Parallel validation (code review + CodeQL) before PR is opened

## Stop conditions

- Any file write in harness commands.
- Any network or GitHub API access.
- Any change to embedded assets or instruction packs.
- Any change to `invariants.yml`.

## Escalation triggers

- Request to modify existing command behaviour.
- Request to add remote download or GitHub API call.

## Context reset notes

This contract is active for the `carl harness` PR. Close it when merged.
