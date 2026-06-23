<!-- version: 1.2.0 -->
# Current PR Contract

This contract constrains implementation scope for the active PR. Update it when scope is explicitly amended. If a requested action falls outside approved scope, stop and escalate before proceeding.

Use this contract to distinguish active PR constraints, completed PR constraints, durable invariants, and intentional amendments. Completed PR constraints are historical evidence unless they are explicitly promoted to durable invariants.

---

## Previous contract (superseded)

The previous active contract (harness loader refactor) is now superseded by this contract.
Durable lesson carried forward: harness instruction files are adapter surfaces; cARL artefacts remain the canonical governance authority.

---

## Goal

Switch the cARL release/distribution pipeline from the current hand-rolled GitHub Actions build matrix to GoReleaser, and add first-class packaging artefacts for GitHub Releases, Homebrew, apt/deb, yum/rpm, and apk.

GoReleaser becomes the canonical release packaging layer: repeatable, checksummed, package-manager friendly, and easier to mirror internally.

## Contract status

active

## Non-goals

- No changes to core CLI command behaviour.
- No changes to runtime installation, repair, status, doctor, map, plan, convert, reconcile, or harness command behaviour unless required only to keep embedded assets in sync.
- No new harness implementations.
- No model benchmarking implementation.
- No network access.
- No external dependencies.
- No broad rewrite of instruction packs.
- No change to `memory.md` schema.
- No change to `runtime.json` semantics.

## Carry-forward rules

Promoted invariants from previous PRs remain in force:

- No secrets committed to any file.
- Security baseline applies.
- `current-pr-contract.md` must be read before implementation begins.
- Every implementation PR must update durable artefacts when behaviour, assumptions, commands, scope, roadmap, or operating model changes.
- cARL durable artefacts must not become a per-turn session diary.
- Completed PR contracts are historical evidence, not binding scope, unless explicitly promoted to durable invariants.

## Approved scope

- `.goreleaser.yaml` — new GoReleaser configuration (builds, archives, checksums, nfpm packages, homebrew tap).
- `.github/workflows/release.yml` — replace hand-rolled build matrix with GoReleaser workflow.
- `.github/workflows/goreleaser-check.yml` — new workflow for config validation and snapshot dry-run.
- `DISTRIBUTION.md` — new file documenting packaging, enterprise mirroring, and manual publishing steps.
- `README.md` — update install section to reflect new archive-based downloads and package manager options.
- `ROADMAP.md` — update release workflow entry to reflect GoReleaser adoption.
- `.github/carl/current-pr-contract.md` — this active contract.
- `.github/carl/memory.md` — update to record release infrastructure change.

## Forbidden scope

- No changes to Go source code or CLI command behaviour.
- No changes to cARL runtime governance artefacts (invariants.yml, trust-boundaries.md, tool-policy.yml) unless a durable invariant is materially affected.
- No embedded asset changes (no Go source changes → no embedded asset sync needed).
- No publishing to external package registries without explicit secret configuration.
- No committing secrets, tokens, credentials, or organisation-internal URLs.
- No rewriting of instruction packs or harness adapters.
- No new Go dependencies.

## Architectural constraints

- GoReleaser must use `version: 2` (GoReleaser v2 format).
- CGO_ENABLED=0 preserved for all build targets (static binaries).
- Build-time ldflags must inject `main.cliVersion` and `main.sourceCommit` as in the existing workflow.
- Homebrew publishing must be gated (`skip_upload: auto`) — no tap token committed, no silent failure.
- WinGet publishing is documented only; no auto-submission configured.
- nfpm packages (deb, rpm, apk) generated and attached to GitHub Release.
- No credentials in `.goreleaser.yaml` — tokens passed via GitHub Actions secrets only.
- `goreleaser check` must pass with the committed config.

## Security constraints

- No secrets, tokens, private keys, tenant data, or credentials in any new or modified file.
- Do not weaken authentication, authorization, validation, logging safety, dependency hygiene, or secret handling guidance.
- Treat CI/CD workflow files as governance-sensitive.
- Homebrew tap token must flow through GitHub Actions secrets, not be committed.

## Files expected to change

- `.goreleaser.yaml` — new
- `.github/workflows/release.yml` — replaced
- `.github/workflows/goreleaser-check.yml` — new
- `DISTRIBUTION.md` — new
- `README.md` — install section updated
- `ROADMAP.md` — release workflow item updated
- `.github/carl/current-pr-contract.md` — this file
- `.github/carl/memory.md` — release infrastructure note

## Tests / validation

- `goreleaser check` passes on the committed `.goreleaser.yaml`.
- `go build ./cmd/carl` and `go test ./...` pass (no Go source changes, but verifying no regression).
- No secrets in committed files (secret-scan changed files).

## Stop conditions

Stop and ask for confirmation if:

- GoReleaser configuration requires committing any token or credential.
- A GoReleaser feature requires a Go dependency or source change.
- WinGet auto-submission is requested without a clearly described token/process.

## Escalation triggers

Escalate if:

- GoReleaser v2 syntax differs materially from what is documented.
- Homebrew tap configuration cannot be safely gated without committing a token.
- nfpm package generation requires changes to Go build flags or source.

## Context reset notes

When this PR is complete, supersede or close this contract.

Durable lessons to carry forward:

- GoReleaser is the canonical release packaging layer for cARL CLI.
- Homebrew, WinGet, and Artifactory publishing are staged: artefacts generated now, live publishing gated on secrets/process.
- Package manager install instructions reference archive-based downloads from GitHub Releases.