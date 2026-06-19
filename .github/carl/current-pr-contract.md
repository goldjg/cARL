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

Implement `carl reconcile` — a command that updates repository-specific
durable artefacts so cARL memory reflects the current repository structure,
not just the upstream default runtime.

## Contract status

active

## Non-goals

- Changes to `carl init`, `carl repair`, `carl doctor`, `carl status`,
  `carl map`, `carl plan`, or `carl harness` behaviour except command registration
- Automatic harness repair or sync
- Network access or remote canonical sources
- Changes to instruction packs or embedded governance content
- Memory schema changes (memory.md remains freeform markdown)
- Modifying `runtime.json` or any harness adapter file

## Carry-forward rules

Promoted invariants from previous PRs remain in force:
- No secrets committed to any file.
- Security baseline (least privilege, no hard-coded credentials) applies.
- `current-pr-contract.md` must be read before implementation begins.
- Every implementation PR must update durable artefacts when behaviour,
  assumptions, commands, scope, or operating model changes.

## Approved scope

- `internal/reconcile/reconcile.go` — new reconcile package and command
- `internal/reconcile/reconcile_test.go` — R1–R6 contract tests
- `cmd/carl/main.go` — register reconcile command
- `CLI.md` — document new reconcile command
- `ROADMAP.md` — record reconcile as delivered
- `.github/carl/current-pr-contract.md` — this file
- `.github/carl/memory.md` — durable facts update for reconcile command

## Intentional amendments

This PR extends cARL with a repo-specific knowledge update command.
reconcile is distinct from repair: repair restores canonical runtime artefacts;
reconcile updates human/agent-readable memory with current repo structure.

## Forbidden scope

- Changes to instruction packs under `.github/instructions/`
- Changes to `invariants.yml` or either embedded/assets copy
- Automatic writes from any existing read-only command
- New external dependencies
- Network or GitHub API access
- Modifying `runtime.json`, harness adapter files, or embedded assets

## Architectural constraints

- reconcile reads only `repo-map.json` and `memory.md`; writes only `memory.md`
- Generated section uses `<!-- BEGIN GENERATED: reconcile -->` /
  `<!-- END GENERATED: reconcile -->` markers; content outside markers is preserved
- Command is idempotent: identical repo-map + same day = no file write
- No embedded asset changes required (reconcile is not a managed artefact)

## Security constraints

- No credentials, tokens, or secrets in any new file.
- No user-controlled data passed to shell commands.
- Filesystem reads and writes remain bounded to the repository root.

## Contract assertions

1. R1: missing repo-map returns actionable error suggesting `carl map`
2. R2: missing memory.md returns actionable error suggesting `carl init`
3. R3: reconcile updates the repo-specific snapshot section from repo-map data
4. R4: reconcile preserves human-authored notes outside the generated section
5. R5: reconcile is idempotent — second run on same repo-map reports no changes
6. R6: reconcile does not modify runtime.json or harness adapter files

## Files expected to change

- `internal/reconcile/reconcile.go` (new)
- `internal/reconcile/reconcile_test.go` (new)
- `cmd/carl/main.go`
- `CLI.md`
- `ROADMAP.md`
- `.github/carl/current-pr-contract.md`
- `.github/carl/memory.md`

## Tests / validation

- `go build ./cmd/carl`
- `go test ./...`
- Secret scan on modified files before final commit
- Parallel validation (code review + CodeQL) before completion

## Stop conditions

- Any change that modifies init, repair, doctor, status, map, plan, or harness
  behaviour beyond command registration
- Any need for network access
- Any change to embedded instruction content or runtime.json

## Escalation triggers

- Requirement to change memory.md schema from freeform markdown to structured data
- Requirement to reconcile files other than memory.md

## Context reset notes

This contract is active for the carl reconcile PR. Close it when merged.
