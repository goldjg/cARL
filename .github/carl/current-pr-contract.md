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

> **Starting a new PR?** Copy `current-pr-contract.template.md` into
> this file, fill in each section, and set contract status to `active`.

---

## Goal

Add `carl status` — a new CLI command that reads `.github/carl/runtime.json`
and reports whether the installed cARL runtime is healthy, missing, or
drifted. Export `Inspect` from the `repair` package to provide a shared,
tested drift-classification function. Update durable artefacts (`CLI.md`,
`ROADMAP.md`, `memory.md`) to reflect the new command.

## Contract status

active

## Non-goals

- Repair changes (no files written)
- Upgrade or remote downloads
- GitHub API integration
- Pack install/remove
- Changes to existing instruction packs or embedded assets

## Carry-forward rules

Promoted invariants from previous PRs remain in force:
- No secrets committed to any file.
- Security baseline (least privilege, no hard-coded credentials) applies.
- `current-pr-contract.md` must be read before implementation begins.
- The three-layer model must be preserved in structural changes.

## Approved scope

- `internal/repair/repair.go` — export `Inspect` function
- `internal/status/status.go` — new `carl status` command
- `internal/status/status_test.go` — tests
- `cmd/carl/main.go` — register `status` command
- `CLI.md` — document `carl status`
- `.github/carl/current-pr-contract.md` — this file
- `.github/carl/memory.md` — durable facts update
- `ROADMAP.md` — mark `carl status` as delivered

## Intentional amendments

No prior constraints are amended. The `repair` package gains a new exported
function (`Inspect`) but its existing behaviour is preserved. No existing
command output changes.

## Forbidden scope

- Changes to instruction packs under `.github/instructions/`
- Changes to `invariants.yml` (no new invariants required)
- Changes to embedded assets
- Any command that modifies files on disk

## Architectural constraints

- `carl status` must be read-only: no file writes.
- `memory.md` and `runtime.json` must never appear in missing/drifted output.
- `Inspect` must behave consistently with `detectDrift` for repair parity.
- Output written to stdout only; errors to stderr via returned error.

## Security constraints

- No credentials, tokens, or secrets in any new file.
- No user-controlled data passed to shell commands.

## Contract assertions

1. When no runtime is installed, output is "No cARL runtime installed." and no error.
2. Healthy runtime: output includes CLI version, runtime version, source, tag,
   commit, installed packs, "none" for both artefact lists, and "Status: Healthy".
3. Missing artefact: listed under "Missing Artefacts:" and "Status: Incomplete".
4. Drifted (content-modified) artefact: listed under "Drifted Artefacts:" and "Status: Drifted".
5. `memory.md` and `runtime.json` never appear in artefact lists regardless of content.

## Files expected to change

- `internal/repair/repair.go`
- `internal/status/status.go` (new)
- `internal/status/status_test.go` (new)
- `cmd/carl/main.go`
- `CLI.md`
- `.github/carl/current-pr-contract.md` (this file)
- `.github/carl/memory.md`
- `ROADMAP.md`

## Tests / validation

- `go build ./cmd/carl` — must succeed
- `go test ./...` — must pass with new status package tests
- Parallel validation (code review + CodeQL) before PR is opened

## Stop conditions

- Any file write in the `status` command.
- Any change that exposes protected artefacts (memory.md, runtime.json) as drift.
- Any change to embedded assets or instruction packs.

## Escalation triggers

- Request to add a remote download or GitHub API call.
- Request to change the repair or version command output.

## Context reset notes

This contract is active for the `carl status` PR. Close it when the PR
is merged and reset for the next task.
