<!-- version: 1.0.5 -->
# Current PR Contract

## Contract status

Complete implementation and available local validation

## Goal

Migrate source, CI, snapshots, and release builds to Go 1.26.10 for
CVE-2026-111943 and earlier standard-library fixes. This explicitly supersedes
the completed PR #48 contract without changing cARL governance semantics.

## Non-goals

- No unrelated dependencies, runtime behaviour, CLI/schema changes, policy
  composition changes, or harness-adapter changes.
- No embedded changes except the explicitly approved memory-mirror sync below.
- No tag/release publication, PR merge, secrets/settings changes, cloud
  mutations, or system-wide toolchain installation.

## Approved scope

- `go.mod`, `.github/workflows/goreleaser-check.yml`,
  `.github/workflows/release.yml`, and `.goreleaser.yaml`: patched compiler
  selection, verification, fail-closed packaging, and compatibility gates.
- `README.md`, `CLI.md`, `DISTRIBUTION.md`, and `RELEASE_READINESS.md`:
  current requirements and a clear distinction from historical evidence.
- This contract: scope and build-policy reconciliation.
- `embedded/assets/.github/carl/memory.md`: user-approved minimal sync to the
  existing canonical memory from PR #48. The pre-existing mismatch otherwise
  fails `TestCanonicalAssetsMatchEmbeddedMirrors`; do not change canonical
  memory or any other embedded assets.
- Create a focused branch from fetched main, commit, push, and open a PR.
- Preserve existing work and avoid unrelated module churn.

## Plan

1. Audit compiler declarations, build/release paths, packaging, and existing
   binary provenance; derive effective packs from canonical runtime state.
2. Raise the patch-level source minimum, pin/verify workflow compilers, and
   disable implicit toolchain switching, including manual old-tag releases.
3. Sync only the approved pre-existing memory drift, then run tests, vet,
   formatting, release-wrapper tests, vulnerability checks,
   config validation, and a non-publishing five-target packaging snapshot.
4. Smoke-test packaged Linux bootstrap/install lifecycle, reconcile docs,
   and submit without publishing.

## Architectural and security constraints

- Preserve all invariants and canonical cARL authority.
- Preserve existing five OS/architecture targets, `CGO_ENABLED=0`, metadata
  flags, embedded assets except the approved memory sync, native package
  formats, signing/notarisation,
  Homebrew/WinGet publishing, and the Apple retry policy.
- Do not print or alter credentials. Tests and snapshots must not publish.
- The user's request covers Tier 2 workflow edits and branch/PR submission;
  scoped module/docs/contract writes are Tier 1. No destructive or
  infrastructure actions are authorised.
- Use patched Go 1.26 rather than 1.27 by explicit user choice to retain
  macOS 12 support. Document that macOS 11 is no longer supported.

## Contract assertions

1. Source requires Go 1.26.10; an older local compiler fails. CI/release verify
   their explicit patched compiler and cannot silently fall back.
2. Manual release checkout of an old tag cannot lower the workflow compiler
   pin; historical workflow reruns remain historical and are documented.
3. Full test/vet/build and wrapper checks pass without module dependencies.
4. GoReleaser snapshots retain all five targets and deb/rpm/apk packaging;
   embedded build information identifies Go 1.26.10.
5. cARL bootstrap and doctor/status retain their existing behaviour; the
   embedded memory matches existing canonical Antigravity support statements.
   Published Go 1.24.13 binaries are not mislabelled as patched by source edits.

## Validation requirements

- `go mod tidy` with reviewed diff; `gofmt -l .`; `go test ./...`;
  `go vet ./...`; native CLI build; release retry wrapper tests.
- `go test -race ./...` on Ubuntu CI; report Gaia's ARM64 VMA limitation if
  the local race runtime cannot start.
- Existing GoReleaser version's `check` and non-publishing snapshot;
  five-target compiler metadata and archive/package inspection.
- `govulncheck`, workflow syntax validation, `git diff --check`.
- Bootstrap smoke tests in disposable directories, never this checkout.
- Report native-platform, signing/notarisation, publication, or blocked
  security-check gaps rather than treating cross-builds as native tests.

## Risks and stop conditions

Stop on incompatible public behaviour, required target removal, embedded
asset drift beyond the approved memory sync, unrelated module churn,
conflicting local work, or a required
publication/secret change. Older binaries, cached installers, and mirrors
remain unpatched until rebuilt and replaced. Linux cross-compilation does not
prove native Windows/macOS execution, signing, or package-manager acceptance.

## cARL/docs update expectation

Update this contract and source/distribution guidance with the durable
compiler policy. Keep historical readiness evidence identifiable as historical.
Keep canonical memory unchanged and sync only its stale embedded mirror.
The existing memory declares no compiler version, and build-policy authority
remains `go.mod`, workflows, and distribution documentation. No invariants,
runtime manifest, other embedded assets, or trust-boundary semantics change.

## Validation outcome and reset

Tidy added no dependencies. Formatting, full tests, vet, retry-wrapper tests,
workflow syntax, GoReleaser v2.17.0 check/snapshot, source and binary
vulnerability scans, and older-compiler rejection checks passed. All five
snapshot binaries embed Go 1.26.10; both macOS targets declare minimum 12.0.
All archives and six native Linux packages passed checksum validation; deb
payloads match the corresponding binaries. Packaged Linux ARM64 init,
status, doctor, repair, repeated-init refusal, adoption, and effective-pack
smoke checks passed with 38 managed assets.

Local race tests cannot start on Gaia's 39-bit ARM64 VMA layout
(`ThreadSanitizer: unsupported VMA range; Found 39 - Supported 48`).
Ubuntu CI now runs the race suite. Native macOS/Windows execution,
signing/notarisation, and publication were not performed.

Published v1.0.1 archives still embed Go 1.24.13 and require replacement.
PR/hosted-check state belongs in the PR; future work must establish a new
contract rather than treating this completed scope as active.
