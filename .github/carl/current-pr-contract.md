<!-- version: 1.2.0 -->
# Current PR Contract

This contract constrains implementation scope for the active PR. Update it when scope is explicitly amended. If a requested action falls outside approved scope, stop and escalate before proceeding.

Use this contract to distinguish active PR constraints, completed PR constraints, durable invariants, and intentional amendments. Completed PR constraints are historical evidence unless they are explicitly promoted to durable invariants.

---

## Previous contract (superseded)

The previous active contract (GoReleaser migration, PR #3) is now superseded by this contract.
Durable lesson carried forward: GoReleaser `skip_upload: auto` is not safe when a token secret is present but invalid — it still attempts to contact the tap repository. Use `skip_upload: true` to fully disable publishing until the tap is ready.

---

## Goal

Fix the GoReleaser Homebrew cask publishing failure discovered during the v0.4.1 release.

The v0.4.1 release successfully uploaded GitHub Release assets but then failed with
`401 Bad credentials` while attempting to contact `goldjg/homebrew-carl`. Root cause:
`skip_upload: auto` still contacts the tap repository when `HOMEBREW_TAP_GITHUB_TOKEN`
is present but invalid/unusable (even an empty or wrong-scope value triggers the attempt).

Replace `skip_upload: auto` with `skip_upload: true` to make the release pipeline
deterministic and green by default when Homebrew tap publishing has not been deliberately enabled.

## Contract status

active

## Non-goals

- No changes to core CLI command behaviour.
- No changes to runtime installation, repair, status, doctor, map, plan, convert, reconcile, or harness command behaviour.
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

- `.goreleaser.yaml` — change `skip_upload: auto` to `skip_upload: true`; update block comment.
- `.github/workflows/release.yml` — remove `HOMEBREW_TAP_GITHUB_TOKEN` env var; update header comment.
- `DISTRIBUTION.md` — update Homebrew section status and release pipeline table.
- `README.md` — update Homebrew install section note.
- `ROADMAP.md` — update release workflow description.
- `.github/carl/memory.md` — update release infrastructure note.
- `.github/carl/current-pr-contract.md` — this active contract.

## Forbidden scope

- No changes to Go source code or CLI command behaviour.
- No changes to cARL runtime governance artefacts (invariants.yml, trust-boundaries.md, tool-policy.yml).
- No embedded asset changes.
- No publishing to external package registries.
- No committing secrets, tokens, credentials, or organisation-internal URLs.
- No rewriting of instruction packs or harness adapters.
- No new Go dependencies.
- Do not delete GitHub Release publishing.
- Do not remove deb/rpm/apk generation.

## Architectural constraints

- `skip_upload: true` must be set in `homebrew_casks` in `.goreleaser.yaml`.
- `goreleaser check` must pass with the committed config.
- The `homebrew_casks` block is retained as documentation of the intended future configuration.
- No Homebrew tap access during normal releases.

## Security constraints

- No secrets, tokens, private keys, tenant data, or credentials in any new or modified file.
- Do not weaken authentication, authorization, validation, logging safety, dependency hygiene, or secret handling guidance.
- Treat CI/CD workflow files as governance-sensitive.

## Files expected to change

- `.goreleaser.yaml`
- `.github/workflows/release.yml`
- `DISTRIBUTION.md`
- `README.md`
- `ROADMAP.md`
- `.github/carl/memory.md`
- `.github/carl/current-pr-contract.md`

## Tests / validation

- `goreleaser check` passes on the committed `.goreleaser.yaml`.
- `go build ./cmd/carl` and `go test ./...` pass (no Go source changes, verifying no regression).
- No secrets in committed files.
- Release workflow no longer references `HOMEBREW_TAP_GITHUB_TOKEN`.

## Stop conditions

Stop and ask for confirmation if:

- Any change requires committing a token or credential.
- `goreleaser check` fails with the new config.

## Escalation triggers

Escalate if:

- `skip_upload: true` causes `goreleaser check` to reject the config.
- The problem statement requires changes beyond the approved scope above.

## Context reset notes

When this PR is complete, supersede or close this contract.

Durable lessons to carry forward:

- `skip_upload: auto` is not safe when a token secret is present but invalid; it still contacts the tap repository.
- Use `skip_upload: true` to fully disable Homebrew tap publishing until the tap repository is confirmed ready and the token is valid.
- Homebrew tap publishing is staged/documented, not active in CI.
