# cARL Distribution Guide

This document describes how the cARL CLI is packaged and distributed, and how
to obtain it via various channels. It also covers enterprise mirroring and the
manual steps required to publish to package registries that are not yet
fully automated.

---

## Release artefacts

Every tagged release (`v*`) triggers the GoReleaser-based release workflow
(`.github/workflows/release.yml`). The workflow attaches the following
artefacts to the GitHub Release:

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

## Installation

### Direct download (Linux/macOS)

```sh
# Linux (amd64)
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0/carl_1.0.0_linux_amd64.tar.gz \
  | tar xz && sudo mv carl /usr/local/bin/carl

# macOS (Apple Silicon)
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0/carl_1.0.0_darwin_arm64.tar.gz \
  | tar xz && sudo mv carl /usr/local/bin/carl

# macOS (Intel)
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0/carl_1.0.0_darwin_amd64.tar.gz \
  | tar xz && sudo mv carl /usr/local/bin/carl
```

Verify the download against `checksums.txt` before installing:

```sh
curl -LO https://github.com/goldjg/cARL/releases/download/v1.0.0/checksums.txt
sha256sum --check --ignore-missing checksums.txt
```

### Debian / Ubuntu (deb)

Download and install the `.deb` package directly:

```sh
curl -LO https://github.com/goldjg/cARL/releases/download/v1.0.0/carl_1.0.0_linux_amd64.deb
sudo dpkg -i carl_1.0.0_linux_amd64.deb
```

> **Future:** A signed apt repository is a planned distribution channel.
> Until then, install from the `.deb` artefact directly.

### Red Hat / Fedora / SUSE (rpm)

```sh
curl -LO https://github.com/goldjg/cARL/releases/download/v1.0.0/carl_1.0.0_linux_amd64.rpm
sudo rpm -i carl_1.0.0_linux_amd64.rpm
```

> **Future:** A signed yum/dnf repository is a planned distribution channel.
> Until then, install from the `.rpm` artefact directly.

### Alpine Linux (apk)

```sh
curl -LO https://github.com/goldjg/cARL/releases/download/v1.0.0/carl_1.0.0_linux_amd64.apk
sudo apk add --allow-untrusted carl_1.0.0_linux_amd64.apk
```

> **Future:** A signed Alpine package repository is a planned distribution channel.
> Until then, install from the `.apk` artefact directly.

### Homebrew (macOS / Linux)

Homebrew publishing is **enabled**. GoReleaser publishes the cask definition to
the [`goldjg/homebrew-carl`](https://github.com/goldjg/homebrew-carl) tap
automatically on each tagged release. The `HOMEBREW_TAP_GITHUB_TOKEN` repository
secret must be set with `Contents: write` access to `goldjg/homebrew-carl` for
publishing to succeed.

> macOS binaries are not currently Apple signed or notarized. Homebrew installation works, but macOS Gatekeeper may require manual approval or removal of the quarantine attribute on first run. Signing and notarization are planned for a future release.

```sh
brew tap goldjg/carl
brew trust goldjg/carl
brew install --cask carl

# macOS unsigned binary workaround, only needed if Gatekeeper blocks first run
xattr -dr com.apple.quarantine "$(brew --prefix)/Caskroom/carl"

brew uninstall --cask carl
brew untrust goldjg/carl
brew untap goldjg/carl
```

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

Requires Go 1.24+:

```sh
go install github.com/goldjg/carl/cmd/carl@latest
```

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
| Build (all platforms) | GoReleaser | ✅ Automated |
| Archives + checksums | GoReleaser | ✅ Automated |
| deb / rpm / apk package artefacts | GoReleaser + nfpm | ✅ Automated |
| apt / yum / apk repository publishing | Internal/manual setup | 📋 Future — see mirroring section |
| GitHub Release | GoReleaser | ✅ Automated |
| Homebrew tap formula | GoReleaser | ✅ Automated |
| WinGet submission | GitHub Actions + wingetcreate | ✅ Automated when `WINGETCREATE_TOKEN` is set |
| Enterprise Artifactory mirroring | Manual / custom CI | 📋 Internal setup |
