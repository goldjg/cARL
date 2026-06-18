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

Add `carl plan` — a new CLI command that discovers, validates, and
summarises plan files in `.github/carl/plans/`. For each `.md` file it
shows title, status (lifecycle state), and purpose extracted from
standard template sections. Update durable artefacts (`CLI.md`,
`ROADMAP.md`, `memory.md`) to reflect the new command.

## Contract status

active

## Non-goals

- Writing or modifying plan files
- GitHub API calls or network access
- Changes to embedded assets or instruction packs
- Changes to `invariants.yml` (no new invariants required)
- Subcommands for creating or archiving plans

## Carry-forward rules

Promoted invariants from previous PRs remain in force:
- No secrets committed to any file.
- Security baseline (least privilege, no hard-coded credentials) applies.
- `current-pr-contract.md` must be read before implementation begins.

## Approved scope

- `internal/plan/plan.go` — new `carl plan` command implementation
- `internal/plan/plan_test.go` — tests
- `cmd/carl/main.go` — register `plan` command
- `CLI.md` — document `carl plan`
- `.github/carl/current-pr-contract.md` — this file
- `.github/carl/memory.md` — durable facts update
- `ROADMAP.md` — mark `carl plan` as delivered

## Intentional amendments

No prior constraints are amended. No existing command behaviour changes.

## Forbidden scope

- Changes to instruction packs under `.github/instructions/`
- Changes to `invariants.yml`
- Changes to embedded assets
- Modifying `repair`, `status`, `doctor`, `version`, `init`, or `map` commands
- Writing plan files or creating plans directory

## Architectural constraints

- `carl plan` is read-only; it never writes files.
- All output to stdout; errors to stderr via returned error.
- No network access; filesystem scan only.
- Always returns nil (exit 0) even when validation warnings are present.
- Plans directory is `.github/carl/plans/`; only `.md` files are scanned.
- Sorted lexicographically by filename.

## Security constraints

- No credentials, tokens, or secrets in any new file.
- No user-controlled data passed to shell commands.
- Filesystem scan is bounded to rootDir; no path traversal outside it.

## Contract assertions

1. No plans directory or empty directory → output "No plans found." and return nil.
2. Plan with all metadata fields is parsed with correct title, status, and purpose.
3. Plan missing `## Plan metadata` section → inline Warning line in output.
4. Plan with empty `Status:` field → inline Warning line in output.
5. Plans are listed sorted lexicographically by filename.
6. `carl plan` always returns nil (read-only).
7. Non-.md files in the plans directory are silently ignored.

## Files expected to change

- `internal/plan/plan.go` (new)
- `internal/plan/plan_test.go` (new)
- `cmd/carl/main.go`
- `CLI.md`
- `.github/carl/current-pr-contract.md` (this file)
- `.github/carl/memory.md`
- `ROADMAP.md`

## Tests / validation

- `go build ./cmd/carl` — must succeed
- `go test ./...` — must pass with new plan package tests
- Parallel validation (code review + CodeQL) before PR is opened

## Stop conditions

- Any file write other than displaying plan summaries.
- Any network or GitHub API access.
- Any change to embedded assets or instruction packs.

## Escalation triggers

- Request to create, modify, or delete plan files.
- Request to add remote download or GitHub API call.

## Context reset notes

This contract is active for the `carl plan` PR. Close it when merged.
