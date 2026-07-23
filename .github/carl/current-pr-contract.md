<!-- version: 1.3.0 -->
# Current PR Contract

This contract constrains implementation scope for the active PR.

---

## Goal

Implement macOS notarisation in the release pipeline using App Store Connect API
key authentication, as described in `DISTRIBUTION.md`.

## Contract status

active

## Non-goals

- No changes to CLI command behaviour.
- No changes to harness adapter authority semantics.
- No new dependencies.
- No changes to release targets/platform matrix.

## Approved scope

- `.goreleaser.yaml` — add `notarize.macos` configuration for darwin artefacts.
- `.github/workflows/release.yml` — wire required notarisation secrets and remove
  obsolete manual signing-keychain steps no longer needed by GoReleaser notarize.
- `.github/workflows/goreleaser-check.yml` — keep snapshot/check notes aligned
  with notarisation gating behaviour.
- `DISTRIBUTION.md` — update notarisation status, setup, and release summary.
- `.github/carl/memory.md` and `.github/carl/trust-boundaries.md` — reconcile
  durable release/trust assumptions changed by notarisation.
- `embedded/assets/.github/carl/memory.md` and
  `embedded/assets/.github/carl/trust-boundaries.md` — keep embedded canonical
  copies aligned with source artefacts.
- `.github/carl/current-pr-contract.md` — this contract update.

## Forbidden scope

- No edits to CLI implementation packages under `cmd/` or `internal/`.
- No changes to unrelated workflows.
- No destructive repository operations.

## Architectural constraints

- Release continues through a single `goreleaser release --clean` step.
- Notarisation config must use App Store Connect API key fields:
  `issuer_id`, `key_id`, and `key`.
- Snapshot/check workflows must remain functional without notarisation secrets.

## Security constraints

- Never commit secrets or secret material.
- Validate required notarisation secrets in release workflow before publish.
- Keep CI token/secret values out of logs.

## Files expected to change

- `.github/carl/current-pr-contract.md`
- `.goreleaser.yaml`
- `.github/workflows/release.yml`
- `.github/workflows/goreleaser-check.yml`
- `DISTRIBUTION.md`
- `.github/carl/memory.md`
- `.github/carl/trust-boundaries.md`
- `embedded/assets/.github/carl/memory.md`
- `embedded/assets/.github/carl/trust-boundaries.md`

## Contract assertions

1. Tagged releases must produce signed and notarised darwin artefacts when all
   required secrets are present.
2. Release workflow must fail early with explicit error messages when notarise
   secrets are missing.
3. `goreleaser check` and snapshot dry-run must stay usable without notarisation
   secrets.
4. Documentation and durable cARL artefacts must no longer describe darwin
   artefacts as "codesigned but not notarised."

## Tests / validation

- `goreleaser check`
- `go test ./...`

## Stop conditions

Stop and escalate if:

- GoReleaser OSS cannot satisfy required notarisation behaviour with current
  release structure.
- implementing notarisation requires broad pipeline redesign outside approved
  scope.

## Escalation triggers

Escalate if:

- a required secret name or encoding model conflicts with deployment policy;
- notarisation settings break snapshot/check workflows in a way that cannot be
  resolved with scoped conditional config.

## Context reset notes

When complete, supersede this contract with the next active task contract.
