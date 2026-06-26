<!-- version: 1.2.0 -->
# Current PR Contract

This contract constrains implementation scope for the active PR. Update it when scope is explicitly amended. If a requested action falls outside approved scope, stop and escalate before proceeding.

Use this contract to distinguish active PR constraints, completed PR constraints, durable invariants, and intentional amendments. Completed PR constraints are historical evidence unless they are explicitly promoted to durable invariants.

---

## Previous contract (superseded)

The previous active contract (WinGet publishing automation) is now superseded by this contract.

Durable lessons carried forward:

- Homebrew publishing is enabled and uses `HOMEBREW_TAP_GITHUB_TOKEN`.
- `skip_upload: auto` can still contact the tap repository when token state is invalid.
- WinGet publishing can be integrated into release automation with a post-release `wingetcreate update` job.
- `WINGETCREATE_TOKEN` should gate WinGet submission and allow no-token skip behaviour.

---

## Goal

Add macOS Developer ID codesigning for darwin release artefacts. darwin binaries
are signed inline via a GoReleaser `builds.hooks.post` hook before archiving, on
a macOS runner. A single `goreleaser release --clean` handles build, sign, archive,
checksum, and publish. Notarisation is not included: GoReleaser OSS `notarize.macos`
only supports App Store Connect API key auth and cannot use the apple-id +
app-specific-password model; this limitation is documented in DISTRIBUTION.md with
a clear path forward.

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
- No publishing of a release tag as part of this PR.
- No creation of a release tag.

## Carry-forward rules

Promoted invariants from previous PRs remain in force:

- No secrets committed to any file.
- Security baseline applies.
- `current-pr-contract.md` must be read before implementation begins.
- Every implementation PR must update durable artefacts when behaviour, assumptions, commands, scope, roadmap, or operating model changes.
- cARL durable artefacts must not become a per-turn session diary.
- Completed PR contracts are historical evidence, not binding scope, unless explicitly promoted to durable invariants.

## Approved scope

- `.goreleaser.yaml` — remove `id: darwin` prebuilt builder; add darwin targets to `id: carl` unified build; add `builds.hooks.post` calling `.github/scripts/codesign-darwin.sh`.
- `.github/scripts/codesign-darwin.sh` — new post-hook script: signs darwin binaries inline; skips gracefully on non-macOS runners.
- `.github/workflows/release.yml` — move release job to `macos-latest`; merge signing/notarisation inline; two-phase GoReleaser (`--skip=publish` then `publish`); remove separate `sign-darwin` job.
- `.github/workflows/goreleaser-check.yml` — remove placeholder darwin build step (no longer needed with cross-compilation).
- `DISTRIBUTION.md` — soften macOS signing claims to "configured from v0.4.2 onward"; update pipeline summary table; keep Apple signing secrets table.
- `README.md` — soften macOS signing claims to "configured from v0.4.2 onward".
- `ROADMAP.md` — update release workflow description.
- `.github/carl/memory.md` — update release infrastructure durable truth.
- `.github/carl/trust-boundaries.md` — add Apple signing secret trust-boundary rule.
- `.github/carl/current-pr-contract.md` — this active contract.

## Forbidden scope

- No changes to Go source code or CLI command behaviour.
- No changes to cARL runtime governance artefacts (invariants.yml, tool-policy.yml) unless explicitly required by durable governance change.
- No embedded asset changes.
- No publishing to external package registries during local validation.
- No committing secrets, tokens, credentials, certificate files, or organisation-internal URLs.
- No rewriting of instruction packs or harness adapters.
- No new Go dependencies.
- Do not delete GitHub Release publishing.
- Do not remove deb/rpm/apk generation.
- Do not remove checksum generation.
- Do not remove Homebrew tap publishing.
- Do not remove WinGet publishing.

## Architectural constraints

- The release job runs on `macos-latest` so that `codesign` is available. GoReleaser cross-compiles Linux and Windows binaries on the same runner — no separate ubuntu job.
- darwin binaries are signed inline by a GoReleaser `builds.hooks.post` script (`.github/scripts/codesign-darwin.sh`) immediately after each darwin binary is built, before archiving.
- The post-hook script uses `ARTIFACT_TARGET={{ .Target }}` (not `ARTIFACT_OS`) and skips gracefully when the target does not start with `darwin_` or when `codesign` is absent.
- Single-phase GoReleaser: `goreleaser release --clean` — builds, signs (via hook), archives, checksums, and publishes in one step. `goreleaser publish` (Pro-only) is not used.
- GoReleaser OSS `notarize.macos` does not support apple-id + app-specific-password auth; notarisation is not included in the current flow and is documented as a limitation in DISTRIBUTION.md.
- darwin artefacts are codesigned but not notarised; this is documented clearly.
- Checksums are generated by GoReleaser from the final (signed) archives; unsigned artefacts are never published.
- Homebrew cask checksums reference signed archives.
- WinGet job remains Windows-only and does not depend on Apple signing.
- Temporary keychain must be deleted in an `always()` cleanup step in the release job.
- Uses OSS GoReleaser only — no GoReleaser Pro features.

## Security constraints

- Certificate `.p12` must be removed from disk immediately after keychain import.
- Temporary keychain must be deleted in an `always()` cleanup step.
- Apple secrets must never be printed, logged, or exposed in workflow output.
- No secrets, tokens, private keys, tenant data, or credentials in any new or modified file.
- Do not weaken authentication, authorization, validation, logging safety, dependency hygiene, or secret handling guidance.
- Treat CI/CD workflow files as governance-sensitive.

## Files expected to change

- `.goreleaser.yaml`
- `.github/workflows/release.yml`
- `.github/workflows/goreleaser-check.yml`
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
- Release workflow includes macOS codesigning, certificate cleanup, and secret preflight check.

## Stop conditions

Stop and ask for confirmation if:

- Any change requires committing a certificate, credential, or private key.
- Apple signing requires additional architecture changes beyond this approved scope.

## Escalation triggers

Escalate if:

- GoReleaser prebuilt builder is confirmed Pro-only (alternative: move GoReleaser to macos-latest with documented justification).
- Apple notarisation timeout requires configuration changes that affect the ubuntu release job.

## Context reset notes

When this PR is complete, supersede or close this contract.

Durable lessons to carry forward:

- macOS codesign requires a macOS runner; running GoReleaser on `macos-latest` allows both cross-compilation (Linux/Windows) and inline darwin signing in a single job.
- GoReleaser OSS supports `builds.hooks.post` for inline signing; the `builder: prebuilt` approach requires GoReleaser Pro and must not be used with standard GoReleaser.
- `goreleaser publish` is a GoReleaser Pro-only subcommand; the two-phase `--skip=publish` + `goreleaser publish` design is not valid with OSS GoReleaser. Use a single `goreleaser release --clean` instead.
- GoReleaser OSS `notarize.macos` only supports App Store Connect API key auth (`issuer_id`/`key_id`/`key`); it does not support apple-id + app-specific-password. If notarisation is needed, switch to API key credentials.
- Temporary keychain cleanup must always run in CI to prevent certificate material persisting on runners.
- Two Apple repository secrets are required for macOS signing: `MACOS_CERTIFICATE_P12_BASE64` and `MACOS_CERTIFICATE_PASSWORD`.
