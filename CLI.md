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
Replace `v1.0.0` in the commands below with the desired release tag.

**Linux (amd64)**
```sh
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0/carl-v1.0.0-linux-amd64 \
  -o carl && chmod +x carl && sudo mv carl /usr/local/bin/carl
```

**macOS (Apple Silicon)**
```sh
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0/carl-v1.0.0-darwin-arm64 \
  -o carl && chmod +x carl && sudo mv carl /usr/local/bin/carl
```

**macOS (Intel)**
```sh
curl -L https://github.com/goldjg/cARL/releases/download/v1.0.0/carl-v1.0.0-darwin-amd64 \
  -o carl && chmod +x carl && sudo mv carl /usr/local/bin/carl
```

**Windows (amd64)**

Download `carl-v1.0.0-windows-amd64.exe` from the
[releases page](https://github.com/goldjg/cARL/releases/latest) and add it to your `PATH`.

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

### `carl doctor`

Diagnoses runtime issues and provides actionable remediation guidance.

**Usage**

```
carl doctor
```

**What it does**

1. Reads `runtime.json` to discover the installed runtime state. If the manifest
   is absent, reports it as an ERROR and suggests `carl init`.
2. Detects and categorises findings as ERROR, WARNING, or INFO:
   - **ERROR** — missing runtime manifest, unreadable manifest, artefact absent from disk
   - **WARNING** — artefact content differs from its canonical version (drifted)
   - **INFO** — no issues found; runtime is healthy
3. For each finding, provides a suggested remediation action.
4. Exits with code `0` regardless of whether issues are found — the command is
   diagnostic only and never modifies any files.

**Protected files** — the following are never inspected for drift:

- `.github/carl/memory.md` — per-repository state managed by humans and agents
- `.github/carl/runtime.json` — managed exclusively by `carl init`

**Output (no runtime installed)**

```
ERROR   missing runtime manifest (.github/carl/runtime.json)
        Action: run `carl init`

1 error(s), 0 warning(s), 0 info(s) found.
```

**Output (healthy runtime)**

```
INFO    runtime is healthy — all managed artefacts are present and canonical
```

**Output (missing and drifted artefacts)**

```
ERROR   .github/carl/invariants.yml — artefact is missing from disk
        Action: run `carl repair`
WARNING .github/copilot-instructions.md — artefact has drifted from its canonical version
        Action: run `carl repair`

1 error(s), 1 warning(s), 0 info(s) found.
```

**Finding levels**

| Level | Meaning |
|---|---|
| `ERROR` | Condition that prevents normal operation; immediate action required |
| `WARNING` | Condition that should be addressed; runtime still functional |
| `INFO` | Neutral observation; no action required |

---

Reports whether the installed cARL runtime is healthy, missing, or drifted.

**Usage**

```
carl status
```

**What it does**

1. Reads `runtime.json` to discover the installed runtime state.
2. Derives installed pack names from the managed artefact paths.
3. Compares each managed repairable artefact against its embedded canonical
   version (byte-for-byte), classifying files as missing (absent from disk) or
   drifted (present but content differs).
4. Reports overall status: `Healthy`, `Drifted`, or `Incomplete`.

**Protected files** — the following are never reported as missing or drifted:

- `.github/carl/memory.md` — per-repository state managed by humans and agents
- `.github/carl/runtime.json` — managed exclusively by `carl init`

**Output (runtime installed, healthy)**

```
CLI Version:      1.0.0
Runtime Version:  1.0.0
Source:           goldjg/cARL
Tag:              v1.0.0
Commit:           abc1234

Installed Packs:
  cloud/azure
  core/carl
  ...

Missing Artefacts:
  none

Drifted Artefacts:
  none

Status:           Healthy
```

**Output (missing and drifted artefacts)**

```
...
Missing Artefacts:
  .github/carl/invariants.yml

Drifted Artefacts:
  .github/copilot-instructions.md

Status:           Incomplete
```

**Output (no runtime installed)**

```
No cARL runtime installed.
```

**Status values**

| Status | Meaning |
|---|---|
| `Healthy` | All managed repairable artefacts are present and match their canonical versions |
| `Drifted` | All artefacts present, but one or more differ from their canonical versions; run `carl repair` |
| `Incomplete` | One or more managed artefacts are absent from disk; run `carl repair` |

---

### `carl map`

Generates and updates `.github/carl/repo-map.json` by deriving the repository
structure from the filesystem.

**Usage**

```
carl map
```

**What it does**

1. Walks the repository filesystem from the current working directory.
2. Detects programming languages from source file extensions.
3. Identifies project entry points (`go.mod`, `cmd/*/main.go`, `Makefile`, etc.).
4. Maps key directories (up to three levels deep) with human-readable purpose
   descriptions derived from Go package doc comments or known-path heuristics.
5. Lists GitHub Actions workflows from `.github/workflows/`.
6. Lists governance artefacts from `.github/carl/`.
7. Lists root-level documentation files.
8. Writes the result to `.github/carl/repo-map.json`.

The command is idempotent — running it again updates the file in place.
`.git/`, `node_modules/`, and `vendor/` are always excluded from the scan.

**Output**

```
Repo map updated: .github/carl/repo-map.json
  Languages:     Go
  Entry points:  2
  Directories:   22
  Workflows:     1
  Governance:    8
  Documentation: 7
```

**Generated file structure**

```json
{
  "_note": "Repository map derived by `carl map`. Re-run to update after structural changes.",
  "generated_by": "carl map",
  "last_updated": "2026-06-18",
  "languages": ["Go"],
  "entry_points": [
    { "path": "go.mod", "purpose": "Go module definition: github.com/org/repo" },
    { "path": "cmd/myapp/main.go", "purpose": "myapp CLI entry point" }
  ],
  "directories": {
    ".github/carl": "cARLv2 governance artefacts and templates",
    "internal/mylib": "Implements the mylib subsystem."
  },
  "workflows": [
    { "path": ".github/workflows/release.yml", "purpose": "release workflow" }
  ],
  "governance": [
    { "path": ".github/carl/invariants.yml", "purpose": "Runtime invariants enforced by all implementation PRs" }
  ],
  "documentation": [
    { "path": "README.md", "purpose": "Repository overview and pack catalogue" }
  ]
}
```

**Notes**

- Directory purpose descriptions are derived from Go `// Package ...` or
  `// Command ...` doc comments, well-known path heuristics, or left blank.
- The generated file itself appears in the `governance` section on subsequent runs.
- Run `carl map` after adding new packages, workflows, or documentation to keep
  the map current.

---

## carl version

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
