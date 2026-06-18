<!-- version: 1.0.0 -->
# cARL CLI Reference

The `carl` CLI installs and manages the cARL governance runtime inside a repository.
It is a self-contained binary with no runtime dependencies — all governance artefacts
are embedded at build time.

---

## Installation

### Download a pre-built binary

Pre-built binaries for Linux, macOS, and Windows are attached to every
[GitHub Release](https://github.com/goldjg/cARL/releases/latest).

**Linux (amd64)**
```sh
curl -L https://github.com/goldjg/cARL/releases/latest/download/carl-<tag>-linux-amd64 \
  -o carl && chmod +x carl && sudo mv carl /usr/local/bin/carl
```

**macOS (Apple Silicon)**
```sh
curl -L https://github.com/goldjg/cARL/releases/latest/download/carl-<tag>-darwin-arm64 \
  -o carl && chmod +x carl && sudo mv carl /usr/local/bin/carl
```

**macOS (Intel)**
```sh
curl -L https://github.com/goldjg/cARL/releases/latest/download/carl-<tag>-darwin-amd64 \
  -o carl && chmod +x carl && sudo mv carl /usr/local/bin/carl
```

**Windows (amd64)**

Download `carl-<tag>-windows-amd64.exe` from the
[releases page](https://github.com/goldjg/cARL/releases/latest) and add it to your `PATH`.

Replace `<tag>` with the release tag, for example `v1.0.0`.

### Build from source

Requires Go 1.24 or later.

```sh
git clone https://github.com/goldjg/cARL.git
cd cARL
go build -ldflags "-X main.cliVersion=$(git describe --tags --always)" -o carl ./cmd/carl
```

---

## Commands

### `carl init`

Installs the cARL runtime into the current repository.

**Usage**

```
carl init
```

**What it does**

1. Checks that the runtime is not already installed (no `.github/carl/runtime.json`).
2. Checks that none of the managed artefact paths already exist. If any do, it
   lists the conflicts and exits without writing any files.
3. Writes all governance artefacts into `.github/` (instruction packs, `carl/`
   governance files).
4. Creates `.github/carl/runtime.json` — the authoritative runtime manifest
   recording the installed version, source tag, commit, timestamp, and list of
   managed artefacts.

**Output (success)**

```
cARL runtime installed successfully.
  Runtime version:  1.0.0
  Source:           goldjg/cARL @ v1.0.0
  Artefacts:        32 files installed
```

**Errors**

| Error | Cause | Resolution |
|---|---|---|
| `cARL runtime already installed` | `runtime.json` already exists | Run `carl repair` to restore drift, or remove the runtime manually |
| `cARL artefacts already exist` | Individual managed files exist without a `runtime.json` | Remove the listed files or adopt them before running `init` |

---

### `carl repair`

Restores modified managed cARL artefacts to their canonical state.

**Usage**

```
carl repair
```

**What it does**

1. Reads `runtime.json` to discover the list of managed artefacts.
2. Compares each managed artefact against its embedded canonical version
   (byte-for-byte).
3. Reports any files that differ or are missing (drift).
4. Overwrites drifted files with their canonical versions.

**Protected files** — the following are never overwritten by `repair`:

- `.github/carl/memory.md` — per-repository state managed by humans and agents
- `.github/carl/runtime.json` — managed exclusively by `carl init`

**Output (no drift)**

```
No drift detected.
```

**Output (drift found)**

```
Drift detected:
  .github/instructions/core/security.instructions.md
  .github/copilot-instructions.md

Repairing...
Done.
```

**Errors**

| Error | Cause | Resolution |
|---|---|---|
| `no cARL runtime installed` | `runtime.json` not found | Run `carl init` first |

---

### `carl version`

Shows CLI and installed runtime version information, including installed packs
and a health status.

**Usage**

```
carl version
```

Aliases: `carl --version`, `carl -v`

**What it does**

1. Reads `runtime.json` for the installed runtime version, source, and artefact list.
2. Derives installed pack names from the managed artefact paths.
3. Performs a content-based health check: compares each managed artefact against
   its embedded canonical version and reports whether the runtime is healthy or
   has drifted.

**Output (runtime installed)**

```
CLI Version:      1.0.0
Runtime Version:  1.0.0
Source:           goldjg/cARL
Tag:              v1.0.0
Commit:           abc1234

Installed Packs:
  cloud/azure
  cloud/entra
  cloud/gcp
  cloud/microsoft-graph
  cloud/netlify
  core/baseline
  core/carl
  core/cognition-governance
  core/dependency
  core/identity
  core/memory-cache
  core/pr-contract
  core/security
  core/tool-permission-tiers
  languages/html
  languages/javascript
  languages/powershell
  languages/python
  languages/terraform
  languages/typescript
  platform/cicd
  platform/docker
  platform/kubernetes

Runtime Status:
  Healthy
```

**Output (no runtime installed)**

```
No cARL runtime installed.
```

**Runtime Status values**

| Status | Meaning |
|---|---|
| `Healthy` | All managed artefacts are present and match their canonical versions |
| `Drift detected (N artefact(s) modified or missing)` | One or more artefacts differ; run `carl repair` |

---

## Global options

| Flag | Alias | Effect |
|---|---|---|
| `--help` | `-h` | Print usage and available commands, then exit |
| `--version` | `-v` | Alias for `carl version` |

Run `carl` with no arguments to print usage.

---

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Command completed successfully |
| `1` | An error occurred; details are printed to stderr |

---

## Runtime manifest

`carl init` writes `.github/carl/runtime.json` — a JSON file that is the
authoritative source of truth for the installed runtime state.

```json
{
  "runtimeVersion": "1.0.0",
  "source": "goldjg/cARL",
  "sourceTag": "v1.0.0",
  "sourceCommit": "abc1234...",
  "installedAt": "2025-01-01T00:00:00Z",
  "managedArtifacts": [
    ".github/carl/invariants.yml",
    ".github/copilot-instructions.md",
    "..."
  ]
}
```

This file must not be edited manually and is never overwritten by `carl repair`.
