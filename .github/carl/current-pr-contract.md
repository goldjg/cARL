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

Extend the harness adapter model introduced in PR #9 by promoting the
four planned adapters (Claude Code, Codex, Cursor, Antigravity) to
`supported` status. Add DetectionFile and AdapterFiles for each using
current known conventions. Implement only registry/status/list support
(no adapter file content generation or sync). Preserve all existing
Copilot adapter behaviour unchanged. Update tests, CLI docs, roadmap,
memory, and current PR contract.

## Contract status

active

## Non-goals

- Changes to embedded assets or instruction packs
- Changes to `invariants.yml` (no new invariants required)
- Modifying any existing command behaviour
- Generating or syncing adapter file contents
- Writing harness adapter files on disk (read-only command)
- Adapter file content generation or injection
- Network or GitHub API access

## Carry-forward rules

Promoted invariants from previous PRs remain in force:
- No secrets committed to any file.
- Security baseline (least privilege, no hard-coded credentials) applies.
- `current-pr-contract.md` must be read before implementation begins.

## Approved scope

- `internal/harness/harness.go` — promote claude/codex/cursor/antigravity to supported; add DetectionFile and AdapterFiles
- `internal/harness/harness_test.go` — add detection tests for new adapters; update SupportStatus test
- `CLI.md` — update list/status output examples and detection file table
- `ROADMAP.md` — mark item 15 (cARL for Non-Copilot Agents) as delivered
- `.github/carl/current-pr-contract.md` — this file
- `.github/carl/memory.md` — durable facts update (new supported adapters)

## Intentional amendments

No prior constraints are amended. No existing command behaviour changes.
The only change to `harness.go` registry is promoting four adapters from
`planned` to `supported` and adding their detection/adapter file fields.

## Forbidden scope

- Changes to instruction packs under `.github/instructions/`
- Changes to `invariants.yml` or either embedded/assets copy
- Modifying `repair`, `status`, `doctor`, `version`, `init`, `map`, or `plan` commands
- Creating or modifying harness adapter files on disk (harness command is read-only)
- Adapter file content generation or sync

## Architectural constraints

- `carl harness list` and `carl harness status` remain read-only; they never write files.
- All output to stdout; errors to stderr via returned error.
- No network access; filesystem check only (os.Stat for detection).
- Detection files: copilot → `.github/copilot-instructions.md`; claude → `CLAUDE.md`; codex → `AGENTS.md`; cursor → `.cursorrules`; antigravity → `ANTIGRAVITY.md`.
- Adapter registry is the canonical source of harness metadata; not read from disk.

## Security constraints

- No credentials, tokens, or secrets in any new file.
- No user-controlled data passed to shell commands.
- Filesystem check bounded to rootDir; no path traversal outside it.
- `os.Stat` is the only filesystem operation; no file reads.

## Contract assertions

1. `carl harness list` lists all 5 known adapters (copilot, claude, codex, cursor, antigravity).
2. `carl harness list` identifies all 5 adapters as "supported".
3. `carl harness status` shows "active" for copilot when `.github/copilot-instructions.md` exists.
4. `carl harness status` shows "active" for claude when `CLAUDE.md` exists.
5. `carl harness status` shows "active" for codex when `AGENTS.md` exists.
6. `carl harness status` shows "active" for cursor when `.cursorrules` exists.
7. `carl harness status` shows "active" for antigravity when `ANTIGRAVITY.md` exists.
8. `carl harness status` shows "active" for all 5 when all detection files are present.
9. All supported adapters have a non-empty DetectionFile in the registry.

## Files expected to change

- `internal/harness/harness.go`
- `internal/harness/harness_test.go`
- `CLI.md`
- `ROADMAP.md`
- `.github/carl/current-pr-contract.md` (this file)
- `.github/carl/memory.md`

## Tests / validation

- `go build ./cmd/carl` — must succeed
- `go test ./...` — must pass with updated and new harness package tests
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

This contract is active for the harness adapter implementation PR (PR #10). Close it when merged.
