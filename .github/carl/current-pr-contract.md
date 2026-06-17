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

Add a GitHub Actions release workflow (`.github/workflows/release.yml`)
that builds and publishes cARL CLI binaries for five target platforms
when a semantic version tag (`v*`) is pushed. Update durable artefacts
(`ROADMAP.md`, `memory.md`) to reflect the new CI capability.

## Contract status

active

## Non-goals

- CI invariant enforcement or policy checks
- Structured memory schema changes
- Multi-repo governance tooling
- Changes to existing instruction packs
- Changes to the embedded asset set
- Changes to the cARL CLI binary itself

## Carry-forward rules

Promoted invariants from previous PRs remain in force:
- No secrets committed to any file.
- Security baseline (least privilege, no hard-coded credentials) applies.
- `current-pr-contract.md` must be read before implementation begins.
- The three-layer model must be preserved in structural changes.

## Approved scope

- `.github/workflows/release.yml` — new release workflow
- `.github/carl/current-pr-contract.md` — this file
- `.github/carl/memory.md` — durable facts update
- `ROADMAP.md` — mark release workflow as delivered

## Intentional amendments

No prior constraints are amended. This PR adds a new CI workflow outside
the governance artefact layer and CLI binary; both remain unchanged.

## Forbidden scope

- Changes to the cARL CLI source code (`cmd/`, `internal/`, `embedded/`)
- Changes to instruction packs under `.github/instructions/`
- Changes to `invariants.yml` (no new invariants are required)
- Changes to embedded asset copies unless `invariants.yml` changes

## Architectural constraints

- The workflow must use `GITHUB_TOKEN` with `contents: write` only.
- No secrets or credentials may be embedded in the workflow file.
- All GitHub-org actions may be pinned to major version tags;
  third-party actions must be pinned to a full commit SHA.
- `CGO_ENABLED=0` required to support cross-compilation on ubuntu-latest.

## Security constraints

- `GITHUB_TOKEN` is the only credential used; scoped to the repository.
- No user-controlled data is passed unsanitised to shell commands.
- CLI version is injected via ldflags only (no runtime secret exposure).

## Contract assertions

1. Workflow triggers on `v*` tags only (no branch or manual triggers).
2. Five build matrix entries: linux/amd64, linux/arm64, darwin/amd64,
   darwin/arm64, windows/amd64.
3. Each binary embeds the tag as `cliVersion` and commit SHA as
   `sourceCommit` via `-ldflags "-X main.cliVersion=... -X main.sourceCommit=..."`.
4. Artifacts are uploaded per-platform and attached to the GitHub Release.
5. Release is created if absent; artifacts are uploaded if it already exists.

## Files expected to change

- `.github/workflows/release.yml` (new)
- `.github/carl/current-pr-contract.md` (this file)
- `.github/carl/memory.md`
- `ROADMAP.md`

## Tests / validation

- `go build ./cmd/carl` — must succeed (CLI unchanged)
- `go test ./...` — must pass (no test files changed)
- YAML lint: workflow indentation and structure manually reviewed
- Parallel validation (code review + CodeQL) before PR is opened

## Stop conditions

- Any change that embeds a secret or credential.
- Any change to CLI source, embedded assets, or instruction packs.
- Any workflow change that grants permissions beyond `contents: write`.

## Escalation triggers

- Request to add a third-party action without a commit SHA.
- Request to grant additional permissions beyond the approved set.

## Context reset notes

This contract is active for the release workflow PR. Close it when the
PR is merged and reset for the next task.
