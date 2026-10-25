<!-- version: 1.1.0 -->
# cARL Distribution Guide

This document describes how the cARL CLI is packaged and distributed, and how
to obtain it via various channels. It also covers enterprise mirroring and the
manual steps required to publish to package registries that are not yet
fully automated.

---

## Release artefacts

Tag pushes matching `v*` trigger the GoReleaser release workflow
(`.github/workflows/release.yml`) automatically. The same workflow also supports
manual `workflow_dispatch` runs for an existing tag (input `tag`) so a failed
tag release can be retried without deleting/recreating the Git tag.

The workflow attaches the following artefacts to the GitHub Release:

| Artefact | Platforms |
|---|---|
| `carl_<version>_linux_amd64.tar.gz` | Linux x86-64 |
| `carl_<version>_linux_arm64.tar.gz` | Linux ARM64 |
| `carl_<version>_darwin_amd64.tar.gz` | macOS Intel |
| `carl_<version>_darwin_arm64.tar.gz` | macOS Apple Silicon |
| `carl_<version>_windows_amd64.zip` | Windows x86-64 |
| `carl_<version>_linux_amd64.deb` | Debian/Ubuntu package |
| `carl_<version>_linux_arm64.deb` | Debian/Ubuntu ARM64 |
| `carl_<version>_linux_amd64.rpm` | Red Hat/Fedora package |
| `carl_<version>_linux_arm64.rpm` | Red Hat/Fedora ARM64 |
| `carl_<version>_linux_amd64.apk` | Alpine Linux package |
| `carl_<version>_linux_arm64.apk` | Alpine Linux ARM64 |
| `checksums.txt` | SHA-256 checksums for all artefacts |

Archives contain a single binary named `carl` (or `carl.exe` on Windows).

---

## Release workflow orchestration and retry behavior

### Trigger modes

- **Automatic release:** push a tag matching `v*`.
- **Manual recovery run:** run the **Release** workflow with `workflow_dispatch`
  and provide `tag` (required).

For manual runs:

1. `tag` must start with `v`.
2. The tag must already exist in `goldjg/cARL`.
3. The workflow checks out that exact tag commit before running GoReleaser.

### Concurrency and serialization

Release workflow executions are serialized with workflow-level concurrency:

```yaml
concurrency:
  group: release-${{ github.repository }}
  cancel-in-progress: false
```

This allows only one release workflow run at a time per repository, including
different tags and manual reruns, reducing contention on Apple notarization
credentials and App Store Connect rate limits.

### Apple notarization 429 handling

The release workflow runs `goreleaser release --clean` via
`scripts/release-with-retry.sh`.

- Attempt 1 runs normally.
- If attempt 1 fails and logs include either:
  - Apple's standalone `Exceeded hourly limit` signature, or
  - `429 Too Many Requests` / `RATE_LIMIT` on a line that also identifies
    Apple, App Store Connect, notarisation, or the notary service,
  then the workflow emits a warning and sleeps **1 hour**.
- After the cooldown, it performs exactly **one** retry.
- If the second attempt fails, the workflow fails.

Maximum automatic cooldown is **1 hour** (single outer retry).

### Failures deliberately not retried

No outer retry is performed for non-Apple-rate-limit failures, including:

- build/compile failures
- signing/certificate failures (except Apple 429 signatures)
- GoReleaser config errors
- auth/permission failures (GitHub, Homebrew, Apple credentials)
- GitHub Release publication failures unrelated to the Apple 429 signatures
- Homebrew publish failures
- a second Apple 429 after the cooldown retry

### Partial-publication and idempotency caveats

The pipeline intentionally remains a single all-in-one
`goreleaser release --clean` flow (build/sign/notarize/archive/publish).
GoReleaser preserves existing release notes and replaces same-name GitHub
Release assets on a same-tag rerun. That makes the GitHub Release portion
recoverable without deleting the tag or release.

After a successful wrapper run, the release job asserts the expected GitHub
Release assets exist for the resolved tag before the job is marked successful.
Downstream jobs such as WinGet only run when the release job succeeds.

The workflow does **not** delete/recreate tags or releases. A partial external
publication, such as a Homebrew tap push that cannot be repeated safely, can
still require operator inspection before rerun.

---

## Installation

### Direct download (Linux/macOS)

```sh
# Linux (amd64)
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0-rc.1/carl_1.0.0-rc.1_linux_amd64.tar.gz \
  | tar xz && sudo mv carl /usr/local/bin/carl

# macOS (Apple Silicon)
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0-rc.1/carl_1.0.0-rc.1_darwin_arm64.tar.gz \
  | tar xz && sudo mv carl /usr/local/bin/carl

# macOS (Intel)
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0-rc.1/carl_1.0.0-rc.1_darwin_amd64.tar.gz \
  | tar xz && sudo mv carl /usr/local/bin/carl
```

Verify the download against `checksums.txt` before installing:

```sh
curl -LO https://github.com/goldjg/cARL/releases/download/v1.0.0-rc.1/checksums.txt
sha256sum --check --ignore-missing checksums.txt
```

### Debian / Ubuntu (deb)

Download and install the `.deb` package directly:

```sh
curl -LO https://github.com/goldjg/cARL/releases/download/v1.0.0-rc.1/carl_1.0.0-rc.1_linux_amd64.deb
sudo dpkg -i carl_1.0.0-rc.1_linux_amd64.deb
```

> **Future:** A signed apt repository is a planned distribution channel.
> Until then, install from the `.deb` artefact directly.

### Red Hat / Fedora / SUSE (rpm)

```sh
curl -LO https://github.com/goldjg/cARL/releases/download/v1.0.0-rc.1/carl_1.0.0-rc.1_linux_amd64.rpm
sudo rpm -i carl_1.0.0-rc.1_linux_amd64.rpm
```

> **Future:** A signed yum/dnf repository is a planned distribution channel.
> Until then, install from the `.rpm` artefact directly.

### Alpine Linux (apk)

```sh
curl -LO https://github.com/goldjg/cARL/releases/download/v1.0.0-rc.1/carl_1.0.0-rc.1_linux_amd64.apk
sudo apk add --allow-untrusted carl_1.0.0-rc.1_linux_amd64.apk
```

> **Future:** A signed Alpine package repository is a planned distribution channel.
> Until then, install from the `.apk` artefact directly.

### Homebrew (macOS / Linux)

Homebrew publishing is **enabled**. GoReleaser publishes the cask definition to
the [`goldjg/homebrew-carl`](https://github.com/goldjg/homebrew-carl) tap
automatically on each tagged release. The `HOMEBREW_TAP_GITHUB_TOKEN` repository
secret must be set with `Contents: write` access to `goldjg/homebrew-carl` for
publishing to succeed.

macOS release artefacts are **codesigned and notarised** (Developer ID
Application, hardened runtime + App Store Connect API key notarisation). The
release workflow uses GoReleaser OSS `notarize.macos` so a single
`goreleaser release --clean` signs and notarises darwin binaries before
archiving.

```sh
brew tap goldjg/carl
brew trust goldjg/carl
brew install --cask carl

brew uninstall --cask carl
brew untrust goldjg/carl
brew untap goldjg/carl
```

#### Apple signing and notarisation secrets setup

The release workflow requires five repository secrets for macOS codesigning and
notarisation. Configure these under **Settings → Secrets and variables → Actions**
in the `goldjg/cARL` repository before tagging a release.

| Secret | Description |
|---|---|
| `MACOS_CERTIFICATE_P12_BASE64` | Base64-encoded Developer ID Application `.p12`. Export from Keychain Access (Certificate Assistant → Export), then encode: `base64 -i cert.p12 | tr -d '\n'` |
| `MACOS_CERTIFICATE_PASSWORD` | Password set when exporting the `.p12` |
| `NOTARIZE_ISSUER_ID` | Issuer UUID from App Store Connect API key |
| `NOTARIZE_KEY_ID` | Key ID from App Store Connect API key |
| `NOTARIZE_KEY` | Base64-encoded contents of the App Store Connect `.p8` API key file |

The release workflow validates that all five Apple secrets and the required
`HOMEBREW_TAP_GITHUB_TOKEN` are present before any publication begins, and
fails with a clear error message if any are missing.

### WinGet (Windows)

**Status: automated submission from the release workflow (when token is configured).**

The tag-triggered release workflow (`.github/workflows/release.yml`)
can submit/update the package in
[microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs) using
`wingetcreate update`.

Required secret:

- `WINGETCREATE_TOKEN` (repository secret) with permission to create PRs in
  `microsoft/winget-pkgs` via wingetcreate.
- Token guidance: use a classic PAT with `public_repo`, or a fine-grained token
  with read/write pull request access and `Contents: read` access to
  `microsoft/winget-pkgs`.
- In typical WinGet flows, this token belongs to the account that owns the fork
  used by wingetcreate for PR submission (not direct write access to upstream).

If the secret is not configured, the WinGet publish job is skipped and you can
still submit manually using the
[WinGet contributing guide](https://github.com/microsoft/winget-pkgs/blob/master/CONTRIBUTING.md).

Once accepted, users can install via:

```sh
winget install goldjg.cARL
```

### Build from source

Requires Go 1.26.10+:

```sh
go install github.com/goldjg/carl/cmd/carl@latest
```

### Compiler security baseline

CI, snapshot, and release workflows pin Go 1.26.10, verify the actual compiler,
and set `GOTOOLCHAIN=local` to prevent implicit toolchain switching. The release
pin is independent of the checked-out tag's `go.mod`, including manual runs of
older tags. Re-running a historical workflow revision still uses its historical
configuration; use the updated workflow for patched builds.

GoReleaser builds also set `GOTOOLCHAIN=local`; local packaging requires a
Go 1.26.10 compiler on `PATH`. Ordinary source builds may use Go's automatic
toolchain download. The patch-level `go` directive enforces the source minimum,
so a redundant `toolchain` suggestion is unnecessary. Update `go.mod` and both
workflow `GO_VERSION` pins together when raising the baseline.

Go 1.26.10 includes the `crypto/x509` fix for
[CVE-2026-111943 / GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809)
and earlier standard-library security fixes, including those in Go 1.26.9.
A newer major is not automatically patched: Go 1.27 needs at least 1.27.3.
The five supported OS/architecture targets, static `CGO_ENABLED=0` packaging,
and installation/bootstrap behaviour are unchanged. The embedded memory mirror
also carries the already-canonical Antigravity production-support statements
from PR #48; no other embedded assets change.

The Go 1.26 line is intentionally retained for macOS 12 Monterey compatibility;
Go 1.27 requires macOS 13 Ventura. Go 1.26 requires Linux kernel 3.2+,
Windows 10+/Server 2016+, and macOS 12+. macOS 11, supported by Go 1.24, is no
longer supported. A newer local Go major may raise the OS minimum; use the
release baseline for macOS 12 builds.

Existing v1.0.1 archives for all five platforms were inspected with
`go version -m`: they contain Go 1.24.13, not the patched compiler. Both Linux
deb packages contain the same binaries as their corresponding archives.
Changing `go.mod` does not patch these downloads or installed copies. A new
release must rebuild all archives and deb/rpm/apk packages and update
Homebrew, WinGet, caches, and internal mirrors. Inspect an extracted binary
with `go version -m /path/to/carl` before treating it as patched.

---

## Enterprise distribution and internal mirroring

cARL release artefacts can be mirrored into internal repositories such as
**JFrog Artifactory**, **Nexus Repository**, or similar binary managers.

### Mirroring with JFrog Artifactory

1. Create a Generic or Go repository in Artifactory.
2. After each GitHub Release, download the artefacts from
   `https://github.com/goldjg/cARL/releases/download/<tag>/` and upload them
   to your Artifactory instance:

   ```sh
   jf rt upload \
     "carl_*" \
     "cARL-local/<version>/" \
     --url https://your-artifactory.example.com/artifactory
   ```

3. Optionally configure a **Remote Repository** in Artifactory pointing at
   `https://github.com/goldjg/cARL/releases/download/` for on-demand proxying.

4. Configure internal client tooling to reference your Artifactory base URL
   instead of the GitHub Releases URL.

No credentials, organisation URLs, or internal configuration are committed to
this repository. All enterprise-specific setup lives in your own infrastructure.

### Mirroring native packages (deb/rpm/apk)

Native packages attached to the GitHub Release can be imported into:

- **Aptly** or **reprepro** for Debian/Ubuntu internal apt repositories.
- **Createrepo** + Nginx/Artifactory for internal RPM yum/dnf repositories.
- **Alpine abuild** / Artifactory for internal Alpine apk repositories.

Refer to your internal tooling documentation for the specific import steps.

---

## Checksums and verification

Every release includes `checksums.txt` with SHA-256 hashes for all artefacts.
Verify any download before installation:

```sh
sha256sum --check --ignore-missing checksums.txt
```

---

## Release pipeline summary

| Step | Tool | Status |
|---|---|---|
| Build (Linux, Windows, darwin) | GoReleaser (macos-latest) | ✅ Automated |
| macOS codesign (Developer ID, hardened runtime) | GoReleaser `notarize.macos.sign` (macos-latest) | ✅ Automated |
| macOS notarisation | GoReleaser `notarize.macos.notarize` (App Store Connect API key) + one 1-hour Apple-429 outer retry | ✅ Automated |
| Archives + checksums | GoReleaser (macos-latest) | ✅ Automated |
| deb / rpm / apk package artefacts | GoReleaser + nfpm | ✅ Automated |
| apt / yum / apk repository publishing | Internal/manual setup | 📋 Future — see mirroring section |
| GitHub Release | GoReleaser | ✅ Automated |
| Homebrew tap formula | GoReleaser | ✅ Automated |
| WinGet submission | GitHub Actions + wingetcreate | ✅ Automated when `WINGETCREATE_TOKEN` is set |
| Enterprise Artifactory mirroring | Manual / custom CI | 📋 Internal setup |
