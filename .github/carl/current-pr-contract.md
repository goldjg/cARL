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

Implement harness health awareness across cARL so harness adapters are
treated as managed disposable artefacts. Extend `carl harness status`
with presence and sync-health reporting, surface missing/drifted harness
adapters in `carl doctor`, add a harness summary to `carl status`, and
update durable artefacts and CLI documentation.

## Contract status

active

## Non-goals

- Changes to `carl init`, `carl repair`, `carl map`, or `carl plan`
- Automatic harness repair or sync from `doctor` or `status`
- Network access or remote canonical sources
- Changes to instruction packs or embedded governance content
- Changing runtime status exit semantics

## Carry-forward rules

Promoted invariants from previous PRs remain in force:
- No secrets committed to any file.
- Security baseline (least privilege, no hard-coded credentials) applies.
- `current-pr-contract.md` must be read before implementation begins.

## Approved scope

- `internal/repair/repair.go` — expose shared canonical comparison helper
- `internal/harness/*.go` — add harness health inspection and richer status output
- `internal/doctor/*.go` — add harness health findings and remediation guidance
- `internal/status/*.go` — add harness summary section
- `CLI.md` — document new harness/doctor/status output
- `ROADMAP.md` — record harness health awareness as delivered
- `.github/carl/current-pr-contract.md` — this file
- `.github/carl/memory.md` — durable facts update for harness health

## Intentional amendments

This PR amends prior harness work by promoting adapter files from
presence-only detection artefacts to managed disposable outputs with
canonical drift awareness. Runtime `Status:` remains defined only by
managed runtime artefacts; harness health is surfaced separately.

## Forbidden scope

- Changes to instruction packs under `.github/instructions/`
- Changes to `invariants.yml` or either embedded/assets copy
- Automatic file repair outside explicit `carl harness sync`
- New external dependencies
- Network or GitHub API access

## Architectural constraints

- Harness sync health must compare adapter file bytes to the embedded
  canonical source, reusing the existing byte-comparison model.
- Shared comparison logic should not be duplicated across harness,
  doctor, status, and repair code paths.
- Harness adapters remain disposable generated outputs.
- `doctor` stays diagnostic-only and always exits 0 when findings are emitted.
- `status` keeps its existing runtime health semantics while adding a
  separate harness summary section.

## Security constraints

- No credentials, tokens, or secrets in any new file.
- No user-controlled data passed to shell commands.
- Filesystem reads remain bounded to the repository root.
- No automatic writes from read-only commands.

## Contract assertions

1. H1: a synced harness adapter is reported as healthy (`Present` + `Synced`) by `carl harness status`.
2. H2: a modified harness adapter is reported as drifted by `carl harness status`.
3. H3: a deleted harness adapter is reported as missing by `carl harness status`.
4. H4: `carl doctor` reports missing or drifted harness adapters as `WARNING` findings with `carl harness sync` remediation.
5. H5: `carl status` reports accurate active, missing, drifted, and healthy harness summary counts.

## Files expected to change

- `internal/repair/repair.go`
- `internal/harness/harness.go`
- `internal/harness/health.go`
- `internal/harness/harness_test.go`
- `internal/doctor/doctor.go`
- `internal/doctor/doctor_test.go`
- `internal/status/status.go`
- `internal/status/status_test.go`
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

- Any change that adds automatic repair behaviour to `doctor` or `status`
- Any request to alter `init`, `repair`, `map`, or `plan`
- Any need for network access
- Any change to embedded instruction content

## Escalation triggers

- Requirement to change overall runtime status semantics based on harness health
- Requirement to support harness-specific canonical sources beyond embedded cARL artefacts

## Context reset notes

This contract is active for the harness health awareness PR. Close it
when merged.
