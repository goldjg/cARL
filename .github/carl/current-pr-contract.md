<!-- version: 1.2.0 -->
# Current PR Contract

This contract constrains implementation scope for the active PR. Update it when scope is explicitly amended. If a requested action falls outside approved scope, stop and escalate before proceeding.

Use this contract to distinguish active PR constraints, completed PR constraints, durable invariants, and intentional amendments. Completed PR constraints are historical evidence unless they are explicitly promoted to durable invariants.

---

## Previous contract (superseded)

The previous active contract (Homebrew publishing status update) is now superseded by this contract.

Durable lessons carried forward:

- Homebrew publishing is enabled and uses `HOMEBREW_TAP_GITHUB_TOKEN`.
- `skip_upload: auto` can still contact the tap repository when token state is invalid.

---

## Goal

Add WinGet publishing automation to the release workflow by following the
`oh-my-posh` release workflow pattern for `wingetcreate` submission.

## Contract status

active

## Non-goals

- No changes to core CLI command behaviour.
- No changes to runtime installation, repair, status, doctor, map, plan, convert, reconcile, or harness command behaviour.
- No new harness implementations.
- No network access from cARL CLI commands.
- No external dependencies for the Go codebase.
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

- `.github/workflows/release.yml` — add WinGet publish job using `wingetcreate update` and gated secret usage.
- `DISTRIBUTION.md` — update WinGet distribution status and release pipeline summary.
- `README.md` — add WinGet install path for Windows users.
- `ROADMAP.md` — update release workflow description for WinGet automation.
- `.github/carl/memory.md` — update release infrastructure durable truth.
- `.github/carl/trust-boundaries.md` — add secret-gated CI publishing trust-boundary guidance.
- `.github/carl/current-pr-contract.md` — this active contract.

## Forbidden scope

- No changes to Go source code or CLI command behaviour.
- No changes to cARL runtime governance artefacts (invariants.yml, trust-boundaries.md, tool-policy.yml) unless explicitly required by durable governance change.
- No embedded asset changes.
- No publishing to external package registries during local validation.
- No committing secrets, tokens, credentials, or organisation-internal URLs.
- No rewriting of instruction packs or harness adapters.
- No new Go dependencies.
- Do not delete GitHub Release publishing.
- Do not remove deb/rpm/apk generation.

## Architectural constraints

- Release tags remain the trigger (`v*`).
- WinGet submission runs as a separate job after release publication.
- WinGet submission uses `wingetcreate update goldjg.cARL`.
- WinGet job is skipped when `WINGETCREATE_TOKEN` is not configured.
- Existing GoReleaser release behaviour for archives and packages remains unchanged.

## Security constraints

- No secrets, tokens, private keys, tenant data, or credentials in any new or modified file.
- Do not weaken authentication, authorization, validation, logging safety, dependency hygiene, or secret handling guidance.
- Treat CI/CD workflow files as governance-sensitive.
- Keep token usage scoped to the WinGet job and avoid logging token values.

## Files expected to change

- `.github/workflows/release.yml`
- `DISTRIBUTION.md`
- `README.md`
- `ROADMAP.md`
- `.github/carl/memory.md`
- `.github/carl/trust-boundaries.md`
- `.github/carl/current-pr-contract.md`

## Tests / validation

- `go build ./cmd/carl` and `go test ./...` pass.
- `goreleaser check` passes (or report environment limitation if unavailable).
- No secrets in committed files.
- Release workflow includes WinGet job and secret gate.

## Stop conditions

Stop and ask for confirmation if:

- Any change requires committing a token or credential.
- WinGet automation requires additional credentials beyond repository secret configuration.

## Escalation triggers

Escalate if:

- The requested WinGet flow requires permissions or workflow changes beyond this approved scope.
- `wingetcreate` command semantics require package identifier or artefact changes not covered by this contract.

## Context reset notes

When this PR is complete, supersede or close this contract.

Durable lessons to carry forward:

- WinGet publishing can be integrated into release automation with a post-release `wingetcreate update` job.
- `WINGETCREATE_TOKEN` should gate WinGet submission and allow no-token skip behaviour.
